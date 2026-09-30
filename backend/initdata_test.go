package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

const testBotToken = "test_bot_token_123"

// подписываем так же, как в тестах официального клиента MAX
func signInitData(token string, params map[string]string) string {
	pairs := []string{}
	for k, v := range params {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(pairs, "\n")))

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func initParams(userID int64, authDate time.Time) map[string]string {
	return map[string]string{
		"user":        fmt.Sprintf(`{"id":%d,"first_name":"Анна"}`, userID),
		"auth_date":   fmt.Sprintf("%d", authDate.Unix()),
		"start_param": "med_paracetamol",
		"chat_type":   "private",
	}
}

func TestValidateInitData(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	valid := signInitData(testBotToken, initParams(123456, now.Add(-time.Hour)))

	data, err := validateInitData(valid, testBotToken, now)
	if err != nil || data.UserID != "123456" || data.StartParam != "med_paracetamol" {
		t.Fatalf("%+v %v", data, err)
	}

	if _, err := validateInitData(url.QueryEscape(valid), testBotToken, now); err != nil {
		t.Errorf("строка целиком в URL-кодировке тоже должна проходить: %v", err)
	}

	bad := map[string]string{
		"чужой токен":      signInitData("other", initParams(123456, now)),
		"подменили id":     strings.Replace(valid, "123456", "999999", 1),
		"нет подписи":      "user=%7B%22id%22%3A1%7D&auth_date=1",
		"пусто":            "",
		"старше суток":     signInitData(testBotToken, initParams(123456, now.Add(-25*time.Hour))),
		"нет пользователя": signInitData(testBotToken, map[string]string{"auth_date": fmt.Sprintf("%d", now.Unix())}),
		"id ноль":          signInitData(testBotToken, initParams(0, now)),
	}
	for name, raw := range bad {
		if _, err := validateInitData(raw, testBotToken, now); err == nil {
			t.Errorf("%s: должно быть отклонено", name)
		}
	}
	if _, err := validateInitData(valid, "", now); err == nil {
		t.Error("без токена бота проверять нечем")
	}
}

func TestIdentifyMiddleware(t *testing.T) {
	srv, store := newTestServer(t)
	srv.botToken = testBotToken
	srv.now = func() time.Time { return time.Now() }
	store.SaveProfile(context.Background(), "123456", Profile{Name: "Анна из MAX"})
	store.SaveProfile(context.Background(), "demo", Profile{Name: "Демо"})
	store.SaveProfile(context.Background(), "victim", Profile{Name: "Чужой"})

	valid := signInitData(testBotToken, initParams(123456, time.Now()))
	rec := doRequest(srv, "GET", "/api/profile", "", map[string]string{"X-Max-Init-Data": valid})
	if !strings.Contains(rec.Body.String(), "Анна из MAX") {
		t.Errorf("пользователь MAX определяется по подписи: %s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/profile", "", map[string]string{"X-User-Id": "victim"})
	if strings.Contains(rec.Body.String(), "Чужой") || !strings.Contains(rec.Body.String(), "Демо") {
		t.Errorf("при включённом боте X-User-Id игнорируется: %s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/profile", "", map[string]string{"X-Max-Init-Data": valid + "x", "X-User-Id": "victim"})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("испорченная подпись — 401, получили %d", rec.Code)
	}

	srv.botToken = ""
	rec = doRequest(srv, "GET", "/api/profile", "", map[string]string{"X-User-Id": "victim"})
	if !strings.Contains(rec.Body.String(), "Чужой") {
		t.Error("без бота (разработка) X-User-Id работает как раньше")
	}
}
