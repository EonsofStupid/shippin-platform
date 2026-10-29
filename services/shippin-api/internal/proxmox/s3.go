package proxmox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shippin/shippin-api/internal/platform"
)

// S3Config is an S3-compatible object-store connector (AWS, R2, B2, MinIO).
// Never log AccessKey, SecretKey, or ResticPassword.
type S3Config struct {
	Provider       string
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	PathStyle      bool
	Prefix         string
	ResticPassword string
}

// S3SaveInput is the operator POST body. Empty secret fields mean "keep existing".
type S3SaveInput struct {
	Provider       string `json:"provider"`
	Endpoint       string `json:"endpoint"`
	Region         string `json:"region"`
	Bucket         string `json:"bucket"`
	AccessKey      string `json:"access_key"`
	SecretKey      string `json:"secret_key"`
	PathStyle      *bool  `json:"path_style,omitempty"`
	Prefix         string `json:"prefix"`
	ResticPassword string `json:"restic_password"`
}

func (c S3Config) Configured() bool {
	return strings.TrimSpace(c.Bucket) != "" &&
		strings.TrimSpace(c.AccessKey) != "" &&
		strings.TrimSpace(c.SecretKey) != ""
}

func (c S3Config) Redacted() map[string]any {
	suffix := ""
	if n := len(c.AccessKey); n > 0 {
		if n <= 4 {
			suffix = "set"
		} else {
			suffix = c.AccessKey[n-4:]
		}
	}
	return map[string]any{
		"configured":          c.Configured(),
		"provider":            c.Provider,
		"endpoint":            c.Endpoint,
		"region":              c.Region,
		"bucket":              c.Bucket,
		"prefix":              c.Prefix,
		"path_style":          c.PathStyle,
		"access_key_suffix":   suffix,
		"restic_password_set": c.ResticPassword != "",
	}
}

func (c S3Config) resticRepo() string {
	prefix := strings.Trim(c.Prefix, "/")
	if prefix == "" {
		prefix = "shippin-backups"
	}
	ep := strings.TrimRight(c.Endpoint, "/")
	if ep == "" {
		region := c.Region
		if region == "" || region == "auto" {
			region = "us-east-1"
		}
		return fmt.Sprintf("s3:s3.%s.amazonaws.com/%s/%s", region, c.Bucket, prefix)
	}
	return fmt.Sprintf("s3:%s/%s/%s", ep, c.Bucket, prefix)
}

func defaultPathStyle(provider, endpoint string) bool {
	p := strings.ToLower(provider)
	if p == "r2" || p == "minio" || p == "b2" {
		return true
	}
	return endpoint != "" && !strings.Contains(endpoint, "amazonaws.com")
}

func mergeS3(existing S3Config, in S3SaveInput) S3Config {
	out := existing
	if in.Provider != "" {
		out.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	}
	if in.Endpoint != "" {
		out.Endpoint = strings.TrimSpace(in.Endpoint)
	}
	if in.Region != "" {
		out.Region = strings.TrimSpace(in.Region)
	}
	if in.Bucket != "" {
		out.Bucket = strings.TrimSpace(in.Bucket)
	}
	if in.AccessKey != "" {
		out.AccessKey = strings.TrimSpace(in.AccessKey)
	}
	if in.SecretKey != "" {
		out.SecretKey = in.SecretKey
	}
	if in.Prefix != "" {
		out.Prefix = strings.Trim(in.Prefix, "/")
	}
	if in.ResticPassword != "" {
		out.ResticPassword = in.ResticPassword
	}
	if in.PathStyle != nil {
		out.PathStyle = *in.PathStyle
	} else if in.Provider != "" || in.Endpoint != "" {
		out.PathStyle = defaultPathStyle(out.Provider, out.Endpoint)
	}
	if out.Provider == "" {
		out.Provider = "s3"
	}
	if out.Prefix == "" {
		out.Prefix = "shippin-backups"
	}
	if out.Region == "" {
		if out.Provider == "r2" {
			out.Region = "auto"
		} else {
			out.Region = "us-east-1"
		}
	}
	return out
}

func loadS3Config(repoRoot string) S3Config {
	cfg := S3Config{}
	for _, path := range s3FileCandidates(repoRoot) {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		parseS3Env(string(b), &cfg)
	}
	// Process env wins (Infisical / systemd).
	overlayS3Env(&cfg)
	if cfg.Prefix == "" {
		cfg.Prefix = "shippin-backups"
	}
	if cfg.Provider == "" {
		cfg.Provider = "s3"
	}
	return cfg
}

func s3FileCandidates(repoRoot string) []string {
	var out []string
	out = append(out, filepath.Join(platform.SecretsDir(), "s3.env"))
	if platform.SecretsDir() != platform.CanonicalSecretsDir {
		out = append(out, filepath.Join(platform.CanonicalSecretsDir, "s3.env"))
	}
	out = append(out, filepath.Join(platform.CompatibilitySecretsDir, "s3.env"))
	if repoRoot != "" {
		out = append(out, filepath.Join(repoRoot, "secrets", "s3.env"))
	}
	return out
}

func parseS3Env(body string, cfg *S3Config) {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		switch k {
		case "S3_PROVIDER":
			cfg.Provider = strings.ToLower(v)
		case "S3_ENDPOINT":
			cfg.Endpoint = v
		case "S3_REGION":
			cfg.Region = v
		case "S3_BUCKET":
			cfg.Bucket = v
		case "S3_ACCESS_KEY":
			cfg.AccessKey = v
		case "S3_SECRET_KEY":
			cfg.SecretKey = v
		case "S3_PATH_STYLE":
			cfg.PathStyle = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
		case "S3_PREFIX":
			cfg.Prefix = strings.Trim(v, "/")
		case "RESTIC_PASSWORD":
			cfg.ResticPassword = v
		}
	}
}

func overlayS3Env(cfg *S3Config) {
	if v := os.Getenv("S3_PROVIDER"); v != "" {
		cfg.Provider = strings.ToLower(v)
	}
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("S3_REGION"); v != "" {
		cfg.Region = v
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		cfg.Bucket = v
	}
	if v := os.Getenv("S3_ACCESS_KEY"); v != "" {
		cfg.AccessKey = v
	}
	if v := os.Getenv("S3_SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	}
	if v := os.Getenv("S3_PREFIX"); v != "" {
		cfg.Prefix = strings.Trim(v, "/")
	}
	if v := os.Getenv("RESTIC_PASSWORD"); v != "" {
		cfg.ResticPassword = v
	}
	if v := os.Getenv("S3_PATH_STYLE"); v != "" {
		cfg.PathStyle = v == "1" || strings.EqualFold(v, "true")
	}
}

func formatS3Env(c S3Config) string {
	var b strings.Builder
	b.WriteString("# shippin s3 connector — never commit\n")
	fmt.Fprintf(&b, "S3_PROVIDER=%s\n", c.Provider)
	fmt.Fprintf(&b, "S3_ENDPOINT=%s\n", c.Endpoint)
	fmt.Fprintf(&b, "S3_REGION=%s\n", c.Region)
	fmt.Fprintf(&b, "S3_BUCKET=%s\n", c.Bucket)
	fmt.Fprintf(&b, "S3_ACCESS_KEY=%s\n", c.AccessKey)
	fmt.Fprintf(&b, "S3_SECRET_KEY=%s\n", c.SecretKey)
	fmt.Fprintf(&b, "S3_PATH_STYLE=%s\n", strconv.FormatBool(c.PathStyle))
	fmt.Fprintf(&b, "S3_PREFIX=%s\n", c.Prefix)
	fmt.Fprintf(&b, "RESTIC_PASSWORD=%s\n", c.ResticPassword)
	return b.String()
}

func saveS3Config(repoRoot string, c S3Config) error {
	body := []byte(formatS3Env(c))
	var last error
	wrote := false
	if repoRoot != "" {
		p := filepath.Join(repoRoot, "secrets", "s3.env")
		if err := platform.WriteFile0600(p, body); err != nil {
			last = err
		} else {
			wrote = true
		}
	}
	runtime := filepath.Join(platform.SecretsDir(), "s3.env")
	if err := platform.WriteFile0600(runtime, body); err != nil {
		if last == nil {
			last = err
		}
	} else {
		wrote = true
	}
	if !wrote {
		if last != nil {
			return last
		}
		return fmt.Errorf("s3 connector: no writable secret path")
	}
	return nil
}

func testS3(ctx context.Context, cfg S3Config) error {
	if !cfg.Configured() {
		return fmt.Errorf("s3 connector not configured")
	}
	req, err := newS3HeadBucketRequest(ctx, cfg)
	if err != nil {
		return err
	}
	signAWS4(req, cfg, emptyPayloadHash, time.Now().UTC())
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("head bucket: HTTP %s", res.Status)
}

func newS3HeadBucketRequest(ctx context.Context, cfg S3Config) (*http.Request, error) {
	ep := strings.TrimRight(cfg.Endpoint, "/")
	if ep == "" {
		region := cfg.Region
		if region == "" || region == "auto" {
			region = "us-east-1"
		}
		ep = "https://s3." + region + ".amazonaws.com"
	}
	u, err := url.Parse(ep)
	if err != nil {
		return nil, fmt.Errorf("endpoint: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
		u.Host = ep
		u.Path = ""
	}
	var raw string
	if cfg.PathStyle {
		u.Path = "/" + strings.Trim(cfg.Bucket, "/")
		raw = u.String()
	} else {
		u.Host = cfg.Bucket + "." + u.Host
		u.Path = "/"
		raw = u.String()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, raw, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
