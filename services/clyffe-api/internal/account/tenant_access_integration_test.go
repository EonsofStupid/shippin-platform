package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shippin/clyffe-api/internal/platform"
)

// This test creates and deletes fixtures and a login role. It must be run only
// against the disposable database named by this explicitly scoped variable.
func TestTenantAccessIntegration(t *testing.T) {
	url := os.Getenv("SHIPPIN_TEST_DISPOSABLE_DATABASE_URL")
	if url == "" {
		t.Skip("disposable database not configured")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if err := platform.ValidateCustomerRole(ctx, admin); err == nil {
		t.Fatal("privileged database login must be rejected")
	}
	var existing int
	if err := admin.QueryRow(ctx, `SELECT (SELECT count(*) FROM shippin_core.subjects)
		+ (SELECT count(*) FROM shippin_core.tenants) + (SELECT count(*) FROM clyffe_core.orders)
		+ (SELECT count(*) FROM identity.sessions)`).Scan(&existing); err != nil {
		t.Fatal(err)
	}
	if existing != 0 {
		t.Fatal("integration database must be empty and disposable")
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	suffix := hex.EncodeToString(random)
	login := "shippin_test_" + suffix
	password := suffix + suffix
	quotedLogin := pgx.Identifier{login}.Sanitize()
	if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD '%s'", quotedLogin, password)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP ROLE "+quotedLogin)
	})
	if _, err := admin.Exec(ctx, "GRANT shippin_customer_api TO "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.User = login
	cfg.ConnConfig.Password = password
	cfg.MaxConns = 1 // exercise reuse of the same pooled connection
	appDB, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(appDB.Close)
	if err := platform.ValidateCustomerRole(ctx, appDB); err != nil {
		t.Fatalf("restricted login rejected: %v", err)
	}
	if _, err := admin.Exec(ctx, "GRANT pg_read_all_data TO "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if err := platform.ValidateCustomerRole(ctx, appDB); err == nil {
		t.Fatal("login with an extra read-all role must be rejected")
	}
	if _, err := admin.Exec(ctx, "REVOKE pg_read_all_data FROM "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "GRANT UPDATE ON clyffe_core.orders TO "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if err := platform.ValidateCustomerRole(ctx, appDB); err == nil {
		t.Fatal("login with order write access must be rejected")
	}
	if _, err := admin.Exec(ctx, "REVOKE UPDATE ON clyffe_core.orders FROM "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "GRANT EXECUTE ON FUNCTION identity.verify_password(citext,text) TO "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if err := platform.ValidateCustomerRole(ctx, appDB); err == nil {
		t.Fatal("login with password-verification permission must be rejected")
	}
	if _, err := admin.Exec(ctx, "REVOKE EXECUTE ON FUNCTION identity.verify_password(citext,text) FROM "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "GRANT SELECT ON identity.credentials TO "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if err := platform.ValidateCustomerRole(ctx, appDB); err == nil {
		t.Fatal("login with password-hash access must be rejected")
	}
	if _, err := admin.Exec(ctx, "REVOKE SELECT ON identity.credentials FROM "+quotedLogin); err != nil {
		t.Fatal(err)
	}
	if _, err := appDB.Exec(ctx, `SELECT password_hash FROM identity.credentials`); err == nil {
		t.Fatal("customer role must not read password hashes")
	}
	var canCreateCustomer, canVerifyPassword bool
	if err := appDB.QueryRow(ctx, `SELECT has_function_privilege(current_user, 'identity.create_customer(citext,text)', 'EXECUTE'),
		has_function_privilege(current_user, 'identity.verify_password(citext,text)', 'EXECUTE')`).Scan(&canCreateCustomer, &canVerifyPassword); err != nil {
		t.Fatal(err)
	}
	if canCreateCustomer || canVerifyPassword {
		t.Fatal("customer role must not execute identity administration or password verification")
	}

	var tenantA, tenantB, subjectA, subjectB, unverified, suspended, disabled string
	for _, item := range []struct {
		prefix string
		out    *string
	}{{"a", &tenantA}, {"b", &tenantB}} {
		if err := admin.QueryRow(ctx, `INSERT INTO shippin_core.tenants(slug,name) VALUES ($1,$2) RETURNING id::text`, item.prefix+suffix, item.prefix).Scan(item.out); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, `DELETE FROM identity.sessions WHERE subject_id IN ($1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid)`, subjectA, subjectB, unverified, suspended, disabled)
		_, _ = admin.Exec(ctx, `DELETE FROM identity.credentials WHERE subject_id IN ($1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid)`, subjectA, subjectB, unverified, suspended, disabled)
		_, _ = admin.Exec(ctx, `DELETE FROM clyffe_core.orders WHERE tenant_id IN ($1::uuid,$2::uuid)`, tenantA, tenantB)
		_, _ = admin.Exec(ctx, `DELETE FROM shippin_core.subject_tenant WHERE tenant_id IN ($1::uuid,$2::uuid)`, tenantA, tenantB)
		_, _ = admin.Exec(ctx, `DELETE FROM shippin_core.subjects WHERE id IN ($1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid)`, subjectA, subjectB, unverified, suspended, disabled)
		_, _ = admin.Exec(ctx, `DELETE FROM shippin_core.tenants WHERE id IN ($1::uuid,$2::uuid)`, tenantA, tenantB)
	})
	for _, item := range []struct {
		name, status string
		verified     bool
		out          *string
	}{
		{"a", "active", true, &subjectA}, {"b", "active", true, &subjectB},
		{"unverified", "active", false, &unverified}, {"suspended", "suspended", true, &suspended}, {"disabled", "active", true, &disabled},
	} {
		if err := admin.QueryRow(ctx, `INSERT INTO shippin_core.subjects(kind,display_name,email,status,email_verified_at)
			VALUES ('customer',$1,$2,$3,CASE WHEN $4 THEN now() ELSE NULL END) RETURNING id::text`,
			item.name, item.name+suffix+"@example.test", item.status, item.verified).Scan(item.out); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := admin.Exec(ctx, `UPDATE shippin_core.subjects SET is_active=false WHERE id=$1`, disabled); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ subject, tenant, role string }{
		{subjectA, tenantA, "customer"}, {subjectA, tenantA, "viewer"}, {subjectB, tenantB, "customer"},
	} {
		if _, err := admin.Exec(ctx, `INSERT INTO shippin_core.subject_tenant(subject_id,tenant_id,role) VALUES($1,$2,$3)`, item.subject, item.tenant, item.role); err != nil {
			t.Fatal(err)
		}
	}
	for _, tenant := range []string{tenantA, tenantB} {
		if _, err := admin.Exec(ctx, `INSERT INTO clyffe_core.orders(tenant_id,product,tier) VALUES($1,'fixture','test')`, tenant); err != nil {
			t.Fatal(err)
		}
	}
	var unscoped int
	if err := appDB.QueryRow(ctx, `SELECT count(*) FROM clyffe_core.orders`).Scan(&unscoped); err != nil || unscoped != 0 {
		t.Fatalf("unscoped RLS query with fixture orders: count=%d error=%v", unscoped, err)
	}
	newSession := func(subject, duration string) string {
		t.Helper()
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			t.Fatal(err)
		}
		token := hex.EncodeToString(b)
		hash := sha256.Sum256([]byte(token))
		if _, err := admin.Exec(ctx, `INSERT INTO identity.sessions(token_hash,subject_id,expires_at) VALUES($1,$2,now()+$3::interval)`, hex.EncodeToString(hash[:]), subject, duration); err != nil {
			t.Fatal(err)
		}
		return token
	}
	// Exercise the real shippin-auth login against the same session table, rather
	// than reproducing its token-generation behavior in this test.
	const testPassword = "disposable-integration-password"
	if _, err := admin.Exec(ctx, `INSERT INTO identity.credentials(subject_id,password_hash) VALUES($1,crypt($2,gen_salt('bf',4)))`, subjectB, testPassword); err != nil {
		t.Fatal(err)
	}
	authURL := startTestAuthService(t, url)
	authClient := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	loginResponse, err := authClient.PostForm(authURL+"/auth/login", urlpkgValues("b"+suffix+"@example.test", testPassword))
	if err != nil {
		t.Fatal(err)
	}
	loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusFound {
		t.Fatalf("auth login returned %d", loginResponse.StatusCode)
	}
	var tokenB string
	for _, cookie := range loginResponse.Cookies() {
		if cookie.Name == "test_session" {
			tokenB = cookie.Value
		}
	}
	if tokenB == "" {
		t.Fatal("auth login did not set the session cookie")
	}
	tokenA := newSession(subjectA, "1 hour")
	expired, pending, blocked, inactive := newSession(subjectA, "-1 hour"), newSession(unverified, "1 hour"), newSession(suspended, "1 hour"), newSession(disabled, "1 hour")
	router := chi.NewRouter()
	NewHandler(NewStore(appDB), "test_session").Routes(router)
	server := httptest.NewServer(router)
	defer server.Close()
	request := func(path, token string, want int) []byte {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "test_session", Value: token})
		}
		// A caller-controlled identity header must never override the session.
		req.Header.Set("X-Auth-Subject", subjectB)
		response, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var body json.RawMessage
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s: got %d, want %d: %s", path, response.StatusCode, want, body)
		}
		if response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("customer response must not be cached")
		}
		return body
	}
	for _, token := range []string{"", "invalid", expired, pending, blocked, inactive} {
		request("/api/clyffe/accounts", token, http.StatusUnauthorized)
	}
	var accounts struct {
		Items []Account `json:"items"`
	}
	if err := json.Unmarshal(request("/api/clyffe/accounts", tokenA, http.StatusOK), &accounts); err != nil {
		t.Fatal(err)
	}
	if len(accounts.Items) != 1 || accounts.Items[0].ID != tenantA {
		t.Fatalf("subject A memberships: %+v", accounts.Items)
	}
	request("/api/clyffe/orders", tokenA, http.StatusBadRequest)
	request("/api/clyffe/orders?tenant_id=not-a-uuid", tokenA, http.StatusNotFound)
	request("/api/clyffe/orders?tenant_id="+tenantB, tokenA, http.StatusNotFound)
	var orders struct {
		Items []Order `json:"items"`
	}
	if err := json.Unmarshal(request("/api/clyffe/orders?tenant_id="+tenantA, tokenA, http.StatusOK), &orders); err != nil {
		t.Fatal(err)
	}
	if len(orders.Items) != 1 || orders.Items[0].TenantID != tenantA {
		t.Fatalf("subject A orders: %+v", orders.Items)
	}
	if err := json.Unmarshal(request("/api/clyffe/orders?tenant_id="+tenantB, tokenB, http.StatusOK), &orders); err != nil {
		t.Fatal(err)
	}
	if len(orders.Items) != 1 || orders.Items[0].TenantID != tenantB {
		t.Fatalf("subject B orders: %+v", orders.Items)
	}
	home := request("/api/clyffe/home", tokenA, http.StatusOK)
	if strings.Contains(string(home), tenantB) {
		t.Fatal("home leaked another organization's identifier")
	}
	var overview struct {
		Accounts []Account `json:"accounts"`
		Summary  struct {
			Orders int `json:"orders"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(home, &overview); err != nil {
		t.Fatal(err)
	}
	if len(overview.Accounts) != 1 || overview.Accounts[0].ID != tenantA || overview.Summary.Orders != 1 {
		t.Fatalf("unexpected subject A overview: %+v", overview)
	}
	var retained *string
	if err := appDB.QueryRow(ctx, `SELECT current_setting('app.current_tenant', true)`).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != nil && *retained != "" {
		t.Fatalf("tenant context leaked through pool: %q", *retained)
	}
	if _, err := admin.Exec(ctx, `DELETE FROM shippin_core.subject_tenant WHERE subject_id=$1`, subjectA); err != nil {
		t.Fatal(err)
	}
	request("/api/clyffe/orders?tenant_id="+tenantA, tokenA, http.StatusNotFound)
	logout, err := http.NewRequest(http.MethodGet, authURL+"/auth/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	logout.AddCookie(&http.Cookie{Name: "test_session", Value: tokenB})
	logoutResponse, err := authClient.Do(logout)
	if err != nil {
		t.Fatal(err)
	}
	logoutResponse.Body.Close()
	if logoutResponse.StatusCode != http.StatusFound {
		t.Fatalf("auth logout returned %d", logoutResponse.StatusCode)
	}
	request("/api/clyffe/accounts", tokenB, http.StatusUnauthorized)
}

func urlpkgValues(email, password string) url.Values {
	return url.Values{"email": {email}, "password": {password}, "next": {"/"}}
}

func startTestAuthService(t *testing.T, databaseURL string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate source tree")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	binary := filepath.Join(t.TempDir(), "shippin-auth")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./services/shippin-auth/cmd/shippin-auth")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build auth service: %v: %s", err, out)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "SHIPPIN_AUTH_ADDR="+address, "SHIPPIN_AUTH_DB_URL="+databaseURL,
		"SHIPPIN_AUTH_COOKIE=test_session", "SHIPPIN_AUTH_SECURE_COOKIE=0")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = command.Process.Kill(); _ = command.Wait() })
	endpoint := "http://" + address
	client := &http.Client{Timeout: time.Second}
	for i := 0; i < 100; i++ {
		response, err := client.Get(endpoint + "/auth/healthz")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusNoContent {
				return endpoint
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("disposable auth service did not become ready")
	return ""
}
