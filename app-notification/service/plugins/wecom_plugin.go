package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
)

// WeComPlugin 企业微信机器人/群通知插件
type WeComPlugin struct {
	client     *http.Client
	webhookURL string
}

func NewWeComPlugin(webhookURL string) *WeComPlugin {
	return &WeComPlugin{
		client:     &http.Client{Timeout: 5 * time.Second},
		webhookURL: webhookURL,
	}
}

func (p *WeComPlugin) ChannelCode() string {
	return contract.ChannelWeCom
}

func (p *WeComPlugin) Send(ctx context.Context, req *contract.ChannelRequest) error {
	url := p.webhookURL
	if strings.TrimSpace(req.Webhook) != "" {
		url = req.Webhook
	} else if req.Payload != nil {
		if customURL, ok := req.Payload["webhook"].(string); ok && customURL != "" {
			url = customURL
		}
	}
	if strings.TrimSpace(url) == "" {
		return nil
	}

	content := fmt.Sprintf("### %s\n> %s", req.Title, req.Content)
	if req.Link != "" {
		content += fmt.Sprintf("\n\n[点击前往处理](%s)", req.Link)
	}

	bodyData := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}
	raw, err := json.Marshal(bodyData)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("企业微信返回 HTTP %d", resp.StatusCode)
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("企业微信响应无法解析: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信返回错误 %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
