package proxmox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// WorkspaceExcludes are rebuildable trees that must not be shipped to S3.
var WorkspaceExcludes = []string{
	"cache",
	"scratch/tmp",
	".pnpm-store",
	".Trash-1000",
	"**/node_modules",
	"**/target",
}

const defaultWorkspaceRoot = "/workspace/warden-storage"

// BackupRunInput selects a backup job.
type BackupRunInput struct {
	Kind      string `json:"kind"` // workspace | guest-image
	Node      string `json:"node,omitempty"`
	GuestKind string `json:"guest_kind,omitempty"`
	VMID      int    `json:"vmid,omitempty"`
	Storage   string `json:"storage,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

// BackupStatus is operator-facing (no secrets).
type BackupStatus struct {
	S3              map[string]any   `json:"s3"`
	ResticAvailable bool             `json:"restic_available"`
	ResticVersion   string           `json:"restic_version,omitempty"`
	WorkspaceRoot   string           `json:"workspace_root"`
	Excludes        []string         `json:"excludes"`
	PVEStorages     []Storage        `json:"pve_backup_storages,omitempty"`
	RecentJobs      []map[string]any `json:"recent_jobs"`
	Message         string           `json:"message"`
}

func resticVersion() (string, bool) {
	out, err := exec.Command("restic", "version").Output()
	if err != nil {
		return "", false
	}
	line := strings.TrimSpace(string(out))
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return line, true
}

func (s *Service) BackupStatus(ctx context.Context) BackupStatus {
	s3 := loadS3Config(s.repoRoot)
	ver, ok := resticVersion()
	st := BackupStatus{
		S3:              s3.Redacted(),
		ResticAvailable: ok,
		ResticVersion:   ver,
		WorkspaceRoot:   workspaceRoot(),
		Excludes:        append([]string(nil), WorkspaceExcludes...),
		RecentJobs:      s.recentBackupJobs(ctx),
	}
	if s.cli != nil && s.cfg.Configured() {
		if stores, err := s.cli.ListStorage(ctx); err == nil {
			for _, it := range stores {
				if strings.Contains(it.Content, "backup") {
					st.PVEStorages = append(st.PVEStorages, it)
				}
			}
		}
	}
	switch {
	case !s3.Configured():
		st.Message = "Save an S3-compatible connector (R2, MinIO, AWS, B2) on this tab, then Test, then Run."
	case !s3.ResticPasswordSet():
		st.Message = "S3 connector saved. Set a restic password (encrypts the repo) before workspace backup."
	case !ok:
		st.Message = "S3 connector saved. restic is not on PATH; guest-image vzdump still works."
	default:
		st.Message = "S3 connector ready. Workspace backup uses restic → object storage. Guest images use Proxmox vzdump."
	}
	return st
}

func (c S3Config) ResticPasswordSet() bool { return c.ResticPassword != "" }

func workspaceRoot() string {
	if v := os.Getenv("SHIPPIN_WORKSPACE_ROOT"); v != "" {
		return v
	}
	return defaultWorkspaceRoot
}

func (s *Service) SaveS3(in S3SaveInput) (map[string]any, error) {
	cur := loadS3Config(s.repoRoot)
	next := mergeS3(cur, in)
	if strings.TrimSpace(next.Bucket) == "" || strings.TrimSpace(next.AccessKey) == "" {
		return nil, fmt.Errorf("bucket and access_key are required")
	}
	if next.SecretKey == "" {
		return nil, fmt.Errorf("secret_key is required on first save")
	}
	if err := saveS3Config(s.repoRoot, next); err != nil {
		return nil, err
	}
	return next.Redacted(), nil
}

func (s *Service) TestS3(ctx context.Context) error {
	return testS3(ctx, loadS3Config(s.repoRoot))
}

func (s *Service) RunBackup(ctx context.Context, in BackupRunInput) (*ActionResult, error) {
	switch in.Kind {
	case "workspace":
		return s.runWorkspaceBackup(ctx)
	case "guest-image":
		return s.runGuestImageBackup(ctx, in)
	default:
		return nil, fmt.Errorf("kind must be workspace or guest-image")
	}
}

func (s *Service) insertAction(ctx context.Context, kind string, payload any) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database unavailable")
	}
	b, _ := json.Marshal(payload)
	var id string
	err := s.db.QueryRow(ctx,
		`INSERT INTO shippin_core.action_requests (kind, status, payload)
		 VALUES ($1, 'planned', $2) RETURNING id::text`, kind, b).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create action_request: %w", err)
	}
	_, _ = s.db.Exec(ctx, `UPDATE shippin_core.action_requests SET status='approved', updated_at=now() WHERE id=$1`, id)
	_, _ = s.db.Exec(ctx, `UPDATE shippin_core.action_requests SET status='executing', updated_at=now() WHERE id=$1`, id)
	return id, nil
}

func (s *Service) finishAction(ctx context.Context, id, status, errMsg string, result any) {
	var resultJSON []byte
	if result != nil {
		resultJSON, _ = json.Marshal(result)
	}
	_, _ = s.db.Exec(ctx,
		`UPDATE shippin_core.action_requests
		 SET status=$2, result=$3, error=NULLIF($4,''), updated_at=now() WHERE id=$1`,
		id, status, resultJSON, errMsg)
}

func (s *Service) runWorkspaceBackup(ctx context.Context) (*ActionResult, error) {
	cfg := loadS3Config(s.repoRoot)
	if !cfg.Configured() {
		return nil, fmt.Errorf("s3 connector not configured")
	}
	if cfg.ResticPassword == "" {
		return nil, fmt.Errorf("restic password is required")
	}
	if _, ok := resticVersion(); !ok {
		return nil, fmt.Errorf("restic not found on PATH")
	}
	root := workspaceRoot()
	id, err := s.insertAction(ctx, "backup_workspace", map[string]any{
		"root": root, "repo": cfg.resticRepo(), "excludes": WorkspaceExcludes,
	})
	if err != nil {
		return nil, err
	}
	go s.execRestic(id, cfg, root)
	return &ActionResult{
		ActionRequestID: id, Kind: "backup_workspace", Status: "executing",
	}, nil
}

func (s *Service) execRestic(actionID string, cfg S3Config, root string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	env := resticEnv(cfg)
	if err := resticEnsureRepo(ctx, env); err != nil {
		s.finishAction(ctx, actionID, "failed", err.Error(), nil)
		return
	}
	args := []string{"backup", "--one-file-system", root, "--tag", "shippin", "--tag", "workspace"}
	for _, ex := range WorkspaceExcludes {
		args = append(args, "--exclude", ex)
	}
	out, err := runRestic(ctx, env, args...)
	if err != nil {
		s.finishAction(ctx, actionID, "failed", err.Error(), map[string]any{"output": clip(out, 2000)})
		return
	}
	s.finishAction(ctx, actionID, "succeeded", "", map[string]any{"output": clip(out, 2000)})
}

func resticEnv(cfg S3Config) []string {
	env := os.Environ()
	env = append(env,
		"AWS_ACCESS_KEY_ID="+cfg.AccessKey,
		"AWS_SECRET_ACCESS_KEY="+cfg.SecretKey,
		"RESTIC_PASSWORD="+cfg.ResticPassword,
		"RESTIC_REPOSITORY="+cfg.resticRepo(),
	)
	if cfg.Region != "" {
		env = append(env, "AWS_DEFAULT_REGION="+cfg.Region)
	}
	if dir := os.Getenv("RESTIC_CACHE_DIR"); dir == "" {
		env = append(env, "RESTIC_CACHE_DIR=/workspace/warden-storage/cache/wardenop/restic")
	}
	return env
}

func resticEnsureRepo(ctx context.Context, env []string) error {
	if _, err := runRestic(ctx, env, "snapshots", "--json"); err == nil {
		return nil
	}
	out, err := runRestic(ctx, env, "init")
	if err != nil && !strings.Contains(strings.ToLower(out), "already") {
		return fmt.Errorf("restic init: %w (%s)", err, clip(out, 400))
	}
	return nil
}

func runRestic(ctx context.Context, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "restic", args...)
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func (s *Service) runGuestImageBackup(ctx context.Context, in BackupRunInput) (*ActionResult, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	if in.GuestKind != "qemu" && in.GuestKind != "lxc" {
		return nil, fmt.Errorf("guest_kind must be qemu or lxc")
	}
	if in.VMID <= 0 {
		return nil, fmt.Errorf("vmid is required")
	}
	node := in.Node
	if node == "" {
		node = s.cfg.Node
	}
	mode := in.Mode
	if mode == "" {
		mode = "snapshot"
	}
	form := url.Values{}
	form.Set("vmid", strconv.Itoa(in.VMID))
	form.Set("mode", mode)
	form.Set("compress", "zstd")
	form.Set("remove", "0")
	form.Set("notes", "shippin-backup")
	if in.Storage != "" {
		form.Set("storage", in.Storage)
	}

	id, err := s.insertAction(ctx, "backup_guest_image", map[string]any{
		"node": node, "guest_kind": in.GuestKind, "vmid": in.VMID, "mode": mode, "storage": in.Storage,
	})
	if err != nil {
		return nil, err
	}

	upid, err := s.cli.Vzdump(ctx, node, form)
	if err != nil {
		s.finishAction(ctx, id, "failed", err.Error(), nil)
		return &ActionResult{
			ActionRequestID: id, Kind: "backup_guest_image", Node: node,
			GuestKind: in.GuestKind, VMID: in.VMID, Status: "failed", Error: err.Error(),
		}, nil
	}
	s.finishAction(ctx, id, "executing", "", map[string]any{"upid": upid})
	go s.watchVzdump(id, node, upid)
	return &ActionResult{
		ActionRequestID: id, Kind: "backup_guest_image", Node: node,
		GuestKind: in.GuestKind, VMID: in.VMID, UPID: upid, Status: "executing",
	}, nil
}

func (s *Service) watchVzdump(actionID, node, upid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()
	st, err := s.cli.WaitTask(ctx, node, upid, 6*time.Hour)
	exit := ""
	if st != nil {
		exit = st.ExitStatus
	}
	final := "succeeded"
	errMsg := ""
	if err != nil {
		final = "failed"
		errMsg = err.Error()
	} else if exit != "" && exit != "OK" {
		final = "failed"
		errMsg = "proxmox exitstatus=" + exit
	}
	s.finishAction(ctx, actionID, final, errMsg, map[string]any{"upid": upid, "exitstatus": exit})
}

func (s *Service) recentBackupJobs(ctx context.Context) []map[string]any {
	if s.db == nil {
		return nil
	}
	rows, err := s.db.Query(ctx, `
		SELECT id::text, kind, status::text, COALESCE(error,''), created_at, updated_at
		FROM shippin_core.action_requests
		WHERE kind IN ('backup_workspace', 'backup_guest_image')
		ORDER BY created_at DESC LIMIT 8`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, kind, status, errText string
		var created, updated time.Time
		if err := rows.Scan(&id, &kind, &status, &errText, &created, &updated); err != nil {
			continue
		}
		item := map[string]any{
			"id": id, "kind": kind, "status": status,
			"created_at": created, "updated_at": updated,
		}
		if errText != "" {
			item["error"] = errText
		}
		out = append(out, item)
	}
	if out == nil {
		return []map[string]any{}
	}
	return out
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
