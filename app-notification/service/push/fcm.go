package push

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FCMConfig 是 Firebase HTTP v1 的服务账号配置。
type FCMConfig struct {
	ProjectID   string `json:"projectId"`
	ClientEmail string `json:"clientEmail"`
	PrivateKey  string `json:"privateKey"`
}

// FCMClient 调用 FCM HTTP v1。
type FCMClient struct {
	http  *http.Client
	mu    sync.Mutex
	token string
	exp   time.Time
}

func NewFCMClient() *FCMClient {
	return &FCMClient{http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *FCMClient) Send(ctx context.Context, cfg FCMConfig, token, title, body string, data map[string]string) (SendResult, error) {
	access, err := c.accessToken(ctx, cfg)
	if err != nil {
		return SendResult{}, err
	}
	payload := map[string]any{
		"message": map[string]any{
			"token": token,
			"notification": map[string]string{
				"title": title,
				"body":  trimPush(body),
			},
			"data": data,
		},
	}
	raw, _ := json.Marshal(payload)
	endpoint := "https://fcm.googleapis.com/v1/projects/" + cfg.ProjectID + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(raw)))
	if err != nil {
		return SendResult{}, err
	}
	req.Header.Set("authorization", "Bearer "+access)
	req.Header.Set("content-type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return SendResult{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	result := SendResult{Status: resp.StatusCode}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var okBody struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(respBody, &okBody)
		result.MessageID = okBody.Name
		return result, nil
	}
	bodyText := string(respBody)
	result.Reason = bodyText
	result.InvalidToken = strings.Contains(bodyText, "UNREGISTERED") || strings.Contains(bodyText, "NOT_FOUND")
	return result, fmt.Errorf("fcm %d", resp.StatusCode)
}

func (c *FCMClient) accessToken(ctx context.Context, cfg FCMConfig) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.exp) {
		return c.token, nil
	}
	key, err := parseRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return "", err
	}
	now := time.Now()
	assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   cfg.ClientEmail,
		"sub":   cfg.ClientEmail,
		"aud":   "https://oauth2.googleapis.com/token",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
	}).SignedString(key)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("fcm oauth %d", resp.StatusCode)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return "", err
	}
	c.token = tok.AccessToken
	ttl := time.Duration(tok.ExpiresIn) * time.Second
	if ttl > time.Minute {
		ttl -= time.Minute
	}
	c.exp = time.Now().Add(ttl)
	return c.token, nil
}

func parseRSAPrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.ReplaceAll(raw, `\n`, "\n")))
	if block == nil {
		return nil, fmt.Errorf("FCM 私钥不是 PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		if pk, err2 := x509.ParsePKCS1PrivateKey(block.Bytes); err2 == nil {
			return pk, nil
		}
		return nil, err
	}
	rk, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("FCM 私钥不是 RSA 密钥")
	}
	return rk, nil
}
