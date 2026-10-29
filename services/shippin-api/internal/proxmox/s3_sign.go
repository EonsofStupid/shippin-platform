package proxmox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"
)

const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func signAWS4(req *http.Request, cfg S3Config, payloadHash string, now time.Time) {
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	signedHeaderNames, canonicalHeaders := canonicalAWSHeaders(req)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req.URL.Path),
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaderNames,
		payloadHash,
	}, "\n")
	scope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hashSHA256([]byte(canonicalRequest)),
	}, "\n")
	signingKey := aws4SigningKey(cfg.SecretKey, dateStamp, region, "s3")
	sig := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization",
		"AWS4-HMAC-SHA256 Credential="+cfg.AccessKey+"/"+scope+
			", SignedHeaders="+signedHeaderNames+
			", Signature="+sig)
}

func canonicalURI(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func canonicalAWSHeaders(req *http.Request) (names, canon string) {
	type kv struct{ k, v string }
	var pairs []kv
	for k, vs := range req.Header {
		lk := strings.ToLower(k)
		if lk != "host" && lk != "content-type" && !strings.HasPrefix(lk, "x-amz-") {
			continue
		}
		pairs = append(pairs, kv{lk, strings.Join(vs, ",")})
	}
	if req.Header.Get("Host") == "" && req.URL.Host != "" {
		pairs = append(pairs, kv{"host", req.URL.Host})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	var nameList []string
	var b strings.Builder
	for _, p := range pairs {
		nameList = append(nameList, p.k)
		b.WriteString(p.k)
		b.WriteByte(':')
		b.WriteString(strings.TrimSpace(p.v))
		b.WriteByte('\n')
	}
	return strings.Join(nameList, ";"), b.String()
}

func aws4SigningKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write([]byte(data))
	return m.Sum(nil)
}

func hashSHA256(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
