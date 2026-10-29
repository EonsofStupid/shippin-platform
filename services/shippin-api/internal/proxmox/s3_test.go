package proxmox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseAndRedactS3Env(t *testing.T) {
	var cfg S3Config
	parseS3Env(`
# comment
S3_PROVIDER=r2
S3_ENDPOINT=https://abc.r2.cloudflarestorage.com
S3_REGION=auto
S3_BUCKET=shippin-backups
S3_ACCESS_KEY=AKIAEXAMPLEKEY
S3_SECRET_KEY=super-secret
S3_PATH_STYLE=true
S3_PREFIX=devstation
RESTIC_PASSWORD=repo-pass
`, &cfg)
	if cfg.Provider != "r2" || cfg.Bucket != "shippin-backups" || cfg.Prefix != "devstation" {
		t.Fatalf("parse: %+v", cfg)
	}
	red := cfg.Redacted()
	if red["access_key_suffix"] != "EKEY" {
		t.Fatalf("suffix=%v", red["access_key_suffix"])
	}
	if red["restic_password_set"] != true {
		t.Fatal("restic_password_set")
	}
	blob, _ := jsonish(red)
	if strings.Contains(blob, "super-secret") || strings.Contains(blob, "repo-pass") {
		t.Fatal("redacted payload leaked a secret")
	}
}

func jsonish(v map[string]any) (string, error) {
	var b strings.Builder
	for k, val := range v {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(strings.TrimSpace(toStr(val)))
		b.WriteByte(' ')
	}
	return b.String(), nil
}

func toStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func TestResticRepoURL(t *testing.T) {
	cfg := S3Config{
		Endpoint: "https://abc.r2.cloudflarestorage.com",
		Bucket:   "bucket",
		Prefix:   "shippin-backups",
	}
	got := cfg.resticRepo()
	want := "s3:https://abc.r2.cloudflarestorage.com/bucket/shippin-backups"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestMergeKeepsSecret(t *testing.T) {
	cur := S3Config{Bucket: "b", AccessKey: "ak", SecretKey: "sk", ResticPassword: "rp", Provider: "r2"}
	next := mergeS3(cur, S3SaveInput{Bucket: "b2", AccessKey: "ak2"})
	if next.SecretKey != "sk" || next.ResticPassword != "rp" {
		t.Fatalf("merge dropped secrets: %+v", next)
	}
	if next.Bucket != "b2" || next.AccessKey != "ak2" {
		t.Fatalf("merge did not apply patch: %+v", next)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SHIPPIN_SECRETS_DIR", filepath.Join(root, "run"))
	in := S3Config{
		Provider: "minio", Endpoint: "http://127.0.0.1:9000", Region: "us-east-1",
		Bucket: "lab", AccessKey: "minioadmin", SecretKey: "minioadmin",
		PathStyle: true, Prefix: "shippin-backups", ResticPassword: "pw",
	}
	if err := saveS3Config(root, in); err != nil {
		t.Fatal(err)
	}
	got := loadS3Config(root)
	if !got.Configured() || got.Bucket != "lab" || got.SecretKey != "minioadmin" {
		t.Fatalf("load: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(root, "secrets", "s3.env")); err != nil {
		t.Fatal("expected gitignored secrets/s3.env")
	}
}

func TestSignV4HeadAgainstFakeS3(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodHead {
			t.Errorf("method %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := S3Config{
		Provider: "minio", Endpoint: srv.URL, Region: "us-east-1",
		Bucket: "lab", AccessKey: "AKIA", SecretKey: "secret", PathStyle: true,
	}
	if err := testS3(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(sawAuth, "AWS4-HMAC-SHA256 Credential=AKIA/") {
		t.Fatalf("authorization=%q", sawAuth)
	}
}

func TestSignIsDeterministic(t *testing.T) {
	cfg := S3Config{AccessKey: "AKID", SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", Region: "us-east-1"}
	req, _ := http.NewRequest(http.MethodHead, "https://s3.amazonaws.com/examplebucket", nil)
	now := time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC)
	signAWS4(req, cfg, emptyPayloadHash, now)
	if req.Header.Get("X-Amz-Date") != "20130524T000000Z" {
		t.Fatalf("date %s", req.Header.Get("X-Amz-Date"))
	}
	if !strings.Contains(req.Header.Get("Authorization"), "Signature=") {
		t.Fatalf("missing signature: %s", req.Header.Get("Authorization"))
	}
}
