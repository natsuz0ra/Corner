package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Adapter struct {
	token string
	http  *http.Client
}

func NewAdapter(token string) *Adapter {
	return &Adapter{
		token: strings.TrimSpace(token),
		http:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (a *Adapter) getAPIURL(method string) string {
	return "https://api.telegram.org/bot" + a.token + "/" + method
}

func (a *Adapter) getFileURL(filePath string) string {
	return "https://api.telegram.org/file/bot" + a.token + "/" + strings.TrimLeft(strings.TrimSpace(filePath), "/")
}

// postJSON sends a JSON POST to the Telegram API and checks the status code.
func (a *Adapter) postJSON(method string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, a.getAPIURL(method), bytes.NewReader(raw))
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
		return fmt.Errorf("telegram %s failed: status=%d", method, resp.StatusCode)
	}
	return nil
}

// GetUpdates calls Telegram getUpdates long-polling for incremental updates.
func (a *Adapter) GetUpdates(ctx context.Context, offset int64, timeoutSeconds int) ([]update, error) {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return nil, fmt.Errorf("telegram token is empty")
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
		return nil, fmt.Errorf("telegram getUpdates failed: status=%d", resp.StatusCode)
	}
	var payload getUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates returned ok=false")
	}
	return payload.Result, nil
}

func (a *Adapter) ResolveFilePath(ctx context.Context, fileID string) (string, error) {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return "", fmt.Errorf("telegram token is empty")
	}
	trimmedFileID := strings.TrimSpace(fileID)
	if trimmedFileID == "" {
		return "", fmt.Errorf("telegram file id is empty")
	}
	query := url.Values{}
	query.Set("file_id", trimmedFileID)
	apiURL := a.getAPIURL("getFile") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("telegram getFile failed: status=%d", resp.StatusCode)
	}
	var payload getFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if !payload.OK || payload.Result == nil || strings.TrimSpace(payload.Result.FilePath) == "" {
		return "", fmt.Errorf("telegram getFile returned empty file path")
	}
	return payload.Result.FilePath, nil
}

func (a *Adapter) DownloadFile(ctx context.Context, fileID string, maxBytes int64) ([]byte, string, error) {
	filePath, err := a.ResolveFilePath(ctx, fileID)
	if err != nil {
		return nil, "", err
	}
	data, err := a.DownloadFileByPath(ctx, filePath, maxBytes)
	if err != nil {
		return nil, "", err
	}
	return data, filepath.Base(strings.TrimSpace(filePath)), nil
}

func (a *Adapter) DownloadFileByPath(ctx context.Context, filePath string, maxBytes int64) ([]byte, error) {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return nil, fmt.Errorf("telegram token is empty")
	}
	trimmedPath := strings.TrimSpace(filePath)
	if trimmedPath == "" {
		return nil, fmt.Errorf("telegram file path is empty")
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("max bytes must be greater than 0")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.getFileURL(trimmedPath), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram file download failed: status=%d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("telegram file exceeds max size limit")
	}
	return body, nil
}

func collectMediaCandidates(msg *message) []mediaCandidate {
	if msg == nil {
		return nil
	}
	items := make([]mediaCandidate, 0, 4)
	if len(msg.Photo) > 0 {
		best := msg.Photo[0]
		for _, p := range msg.Photo[1:] {
			if p.FileSize > best.FileSize {
				best = p
			}
		}
		if strings.TrimSpace(best.FileID) != "" {
			items = append(items, mediaCandidate{
				Source:         "photo",
				ProviderFileID: strings.TrimSpace(best.FileID),
				Name:           "photo.jpg",
				MimeType:       "image/jpeg",
				SizeBytes:      best.FileSize,
			})
		}
	}
	if msg.Voice != nil && strings.TrimSpace(msg.Voice.FileID) != "" {
		items = append(items, mediaCandidate{
			Source:         "voice",
			ProviderFileID: strings.TrimSpace(msg.Voice.FileID),
			Name:           "voice.ogg",
			MimeType:       strings.TrimSpace(msg.Voice.MimeType),
			SizeBytes:      msg.Voice.FileSize,
		})
	}
	if msg.Audio != nil && strings.TrimSpace(msg.Audio.FileID) != "" {
		items = append(items, mediaCandidate{
			Source:         "audio",
			ProviderFileID: strings.TrimSpace(msg.Audio.FileID),
			Name:           strings.TrimSpace(msg.Audio.FileName),
			MimeType:       strings.TrimSpace(msg.Audio.MimeType),
			SizeBytes:      msg.Audio.FileSize,
		})
	}
	if msg.Document != nil && strings.TrimSpace(msg.Document.FileID) != "" {
		items = append(items, mediaCandidate{
			Source:         "document",
			ProviderFileID: strings.TrimSpace(msg.Document.FileID),
			Name:           strings.TrimSpace(msg.Document.FileName),
			MimeType:       strings.TrimSpace(msg.Document.MimeType),
			SizeBytes:      msg.Document.FileSize,
		})
	}
	return items
}

// SendText sends a plain text message.
func (a *Adapter) SendText(chatID string, text string) error {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return fmt.Errorf("telegram token is empty")
	}
	return a.postJSON("sendMessage", map[string]any{
		"chat_id": strings.TrimSpace(chatID),
		"text":    text,
	})
}

// SendApprovalKeyboard sends an approval prompt with an inline keyboard.
func (a *Adapter) SendApprovalKeyboard(chatID string, text string, approveData string, rejectData string) error {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return fmt.Errorf("telegram token is empty")
	}
	return a.postJSON("sendMessage", map[string]any{
		"chat_id": strings.TrimSpace(chatID),
		"text":    text,
		"reply_markup": map[string]any{
			"inline_keyboard": []any{
				[]map[string]string{
					{"text": "Approve", "callback_data": approveData},
					{"text": "Reject", "callback_data": rejectData},
				},
			},
		},
	})
}

// AnswerCallbackQuery acknowledges a button press so Telegram stops the loading spinner.
func (a *Adapter) AnswerCallbackQuery(callbackQueryID string, text string) error {
	if a == nil || strings.TrimSpace(a.token) == "" {
		return fmt.Errorf("telegram token is empty")
	}
	payload := map[string]any{
		"callback_query_id": strings.TrimSpace(callbackQueryID),
	}
	if strings.TrimSpace(text) != "" {
		payload["text"] = text
	}
	return a.postJSON("answerCallbackQuery", payload)
}
