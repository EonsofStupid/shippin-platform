package platform

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ConsoleProxyURL is the local operator UI (Vite) this API reverse-proxies
// for non-/api routes so LAN/Tailscale clients hit a single origin on :8081.
func ConsoleProxyURL() string {
	if v := strings.TrimSpace(os.Getenv("SHIPPIN_CONSOLE_PROXY")); v != "" {
		return v
	}
	return "http://127.0.0.1:4176"
}

// ConsoleProxy forwards page/asset/HMR traffic to the operator console.
// API routes must be registered on the router first so they win.
func ConsoleProxy() http.Handler {
	raw := ConsoleProxyURL()
	target, err := url.Parse(raw)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Error(w, http.StatusBadGateway, "console proxy misconfigured")
		})
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		orig(req)
		req.Host = target.Host
		req.URL.Host = target.Host
		req.URL.Scheme = target.Scheme
	}
	proxy.ModifyResponse = stripHTMLNULs
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		Error(w, http.StatusBadGateway, "operator console is not running")
	}
	return proxy
}

// Vite/TanStack SSR streams occasionally embed NUL bytes. Some browsers then
// treat the page as binary and show a blank tab ("does not even try to load").
func stripHTMLNULs(res *http.Response) error {
	ct := res.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		return nil
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	clean := bytes.ReplaceAll(body, []byte{0}, nil)
	res.Body = io.NopCloser(bytes.NewReader(clean))
	res.ContentLength = int64(len(clean))
	res.Header.Set("Content-Length", strconv.Itoa(len(clean)))
	res.Header.Del("Transfer-Encoding")
	return nil
}
