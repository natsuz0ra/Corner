package platforms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TelegramAdapter struct {
	token string
	http  *http.Client
}

type telegramGetUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []telegramUpdate `json:"result"`
}

type telegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *telegramMessage `json:"message"`
}

type telegramMessage struct {
	Chat telegramChat `json:"chat"`
	Text string       `json:"text"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

func NewTelegramAdapter(token string) *TelegramAdapter {
	return &TelegramAdapter{
		token: strings.TrimSpace(token),
		http:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (a *TelegramAdapter) getAPIURL(method string) string {
	return "https://api.telegram.org/bot" + a.token + "/" + method
}

func (a *TelegramAdapter) GetUpdates(ctx context.Context, offset int64, timeoutSeconds int) ([]telegramUpdate, error) {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return nil, fmt.Errorf("telegram token 为空")
	}

	query := url.Values{}
	query.Set("timeout", strconv.Itoa(timeoutSeconds))
	if offset > 0 {
		query.Set("offset", strconv.FormatInt(offset, 10))
	}
	apiURL := a.getAPIURL("getUpdates") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram getUpdates 失败: status=%d", resp.StatusCode)
	}
	var payload telegramGetUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates 返回 ok=false")
	}
	return payload.Result, nil
}

// SendText 发送纯文本消息，作为平台统一回包能力。
func (a *TelegramAdapter) SendText(chatID string, text string) error {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return fmt.Errorf("telegram token 为空")
	}
	payload := map[string]any{
		"chat_id": strings.TrimSpace(chatID),
		"text":    text,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, a.getAPIURL("sendMessage"), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage 失败: status=%d", resp.StatusCode)
	}
	return nil
}
