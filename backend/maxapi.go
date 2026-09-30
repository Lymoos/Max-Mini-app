package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// клиент Bot API MAX: https://dev.max.ru/docs-api
type MaxClient struct {
	token   string
	baseURL string
	client  *http.Client
}

type maxUser struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
}

type maxAttachment struct {
	Type      string  `json:"type"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type maxUpdate struct {
	UpdateType string `json:"update_type"`
	Timestamp  int64  `json:"timestamp"`
	Message    *struct {
		Sender maxUser `json:"sender"`
		Body   struct {
			Text        string          `json:"text"`
			Attachments []maxAttachment `json:"attachments"`
		} `json:"body"`
	} `json:"message"`
	Callback *struct {
		CallbackID string  `json:"callback_id"`
		Payload    string  `json:"payload"`
		User       maxUser `json:"user"`
	} `json:"callback"`
	User    *maxUser `json:"user"`
	Payload string   `json:"payload"`
}

type maxButton struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	URL     string `json:"url,omitempty"`
	Payload string `json:"payload,omitempty"`
	WebApp  string `json:"web_app,omitempty"`
}

type maxMessage struct {
	Text        string           `json:"text"`
	Format      string           `json:"format,omitempty"`
	Attachments []map[string]any `json:"attachments,omitempty"`
}

func keyboard(rows ...[]maxButton) map[string]any {
	return map[string]any{"type": "inline_keyboard", "payload": map[string]any{"buttons": rows}}
}

// сертификат API MAX выдан Минцифры, его нет в стандартном списке доверенных.
// Путь к корневому сертификату задаётся в MAX_CA_FILE и добавляется только в этот клиент
func NewMaxClient(token, caFile string) (*MaxClient, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if caFile != "" {
		pem, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать MAX_CA_FILE: %w", err)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("в MAX_CA_FILE нет сертификата в формате PEM")
		}
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}
	return &MaxClient{
		token:   token,
		baseURL: "https://platform-api2.max.ru",
		client:  &http.Client{Timeout: 60 * time.Second, Transport: transport},
	}, nil
}

func (c *MaxClient) do(ctx context.Context, method, path string, query url.Values, body, dst any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path+"?"+query.Encode(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		text, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("max api %s %s: %d %s", method, path, resp.StatusCode, text)
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *MaxClient) Updates(ctx context.Context, marker int64) ([]maxUpdate, int64, error) {
	query := url.Values{}
	query.Set("timeout", "30")
	query.Set("types", "message_created,message_callback,bot_started")
	if marker != 0 {
		query.Set("marker", strconv.FormatInt(marker, 10))
	}

	var resp struct {
		Updates []maxUpdate `json:"updates"`
		Marker  int64       `json:"marker"`
	}
	if err := c.do(ctx, http.MethodGet, "/updates", query, nil, &resp); err != nil {
		return nil, marker, err
	}
	return resp.Updates, resp.Marker, nil
}

func (c *MaxClient) Send(ctx context.Context, userID string, msg maxMessage) error {
	query := url.Values{}
	query.Set("user_id", userID)
	return c.do(ctx, http.MethodPost, "/messages", query, msg, nil)
}

func (c *MaxClient) Answer(ctx context.Context, callbackID, notification string) error {
	query := url.Values{}
	query.Set("callback_id", callbackID)
	return c.do(ctx, http.MethodPost, "/answers", query, map[string]string{"notification": notification}, nil)
}
