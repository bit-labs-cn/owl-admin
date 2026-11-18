package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// APNsConfig 是苹果推送的 token 认证配置。
type APNsConfig struct {
	TeamID     string `json:"teamId"`
	KeyID      string `json:"keyId"`
	BundleID   string `json:"bundleId"`
	PrivateKey string `json:"privateKey"`
	Production bool   `json:"production"`
}

// APNsClient 用 HTTP/2 发送苹果推送。
type APNsClient struct {
	http *http.Client
	mu   sync.Mutex
	jwt  string
	exp  time.Time
}

func NewAPNsClient() *APNsClient {
	return &APNsClient{http: &http.Client{Timeout: 15 * time.Second}}
}

// SendResult 是一次设备推送的结果。
type SendResult struct {
	MessageID    string
	Status       int
	Reason       string
	InvalidToken bool
}

func (c *APNsClient) Send(ctx context.Context, cfg APNsConfig, token, title, body string, data map[string]string) (SendResult, error) {
	key, err := parseECPrivateKey(cfg.PrivateKey)
	if err != nil {
		return SendResult{}, err
	}
	bearer, err := c.token(cfg, key)
	if err != nil {
		return SendResult{}, err
	}
	payload := map[string]any{
		"aps": map[string]any{
			"alert": map[string]string{"title": title, "body": trimPush(body)},
			"sound": "default",
		},
		"data": data,
	}
	raw, _ := json.Marshal(payload)
	host := "https://api.sandbox.push.apple.com"
	if cfg.Production {
		host = "https://api.push.apple.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/3/device/"+token, strings.NewReader(string(raw)))
	if err != nil {
		return SendResult{}, err
	}
	req.Header.Set("authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", cfg.BundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("content-type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return SendResult{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	result := SendResult{Status: resp.StatusCode, MessageID: resp.Header.Get("apns-id")}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return result, nil
	}
	var reason struct {
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal(respBody, &reason)
	result.Reason = reason.Reason
	result.InvalidToken = resp.StatusCode == http.StatusGone || reason.Reason == "BadDeviceToken" || reason.Reason == "Unregistered"
	return result, fmt.Errorf("apns %d %s", resp.StatusCode, result.Reason)
}

func (c *APNsClient) token(cfg APNsConfig, key *ecdsa.PrivateKey) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.jwt != "" && time.Now().Before(c.exp) {
		return c.jwt, nil
	}
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": cfg.TeamID,
		"iat": now.Unix(),
	})
	tok.Header["kid"] = cfg.KeyID
	signed, err := tok.SignedString(key)
	if err != nil {
		return "", err
	}
	c.jwt = signed
	c.exp = now.Add(50 * time.Minute)
	return signed, nil
}

func parseECPrivateKey(raw string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("苹果推送私钥不是 PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("苹果推送私钥不是 EC 密钥")
	}
	return ec, nil
}

func trimPush(s string) string {
	r := []rune(s)
	if len(r) > 120 {
		return string(r[:120])
	}
	return s
}
