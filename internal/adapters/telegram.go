package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SendAlert intentionally never returns the URL/token or response body in errors.
func SendAlert(ctx context.Context, token, chat, message string) error {
	if token == "" || chat == "" {
		return nil
	}
	if strings.ContainsAny(token, "/\r\n ?#") {
		return fmt.Errorf("invalid Telegram token format")
	}
	payload, _ := json.Marshal(map[string]string{"chat_id": chat, "text": "Scooby PAPER: " + message})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.telegram.org/bot"+token+"/sendMessage", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("alert delivery failed")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != 200 {
		return fmt.Errorf("alert HTTP %d", resp.StatusCode)
	}
	return nil
}
