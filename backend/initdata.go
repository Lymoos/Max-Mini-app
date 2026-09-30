package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const initDataMaxAge = 24 * time.Hour

type InitData struct {
	UserID     string
	StartParam string
}

// проверка подписи данных миниаппа, как в документации MAX:
// secret = HMAC_SHA256("WebAppData", токен бота), hash = HMAC_SHA256(secret, отсортированные пары key=value)
func validateInitData(raw, botToken string, now time.Time) (InitData, error) {
	if raw == "" || botToken == "" {
		return InitData{}, errors.New("нет данных для проверки")
	}
	if decoded, err := url.QueryUnescape(raw); err == nil {
		raw = decoded
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return InitData{}, err
	}

	hash := values.Get("hash")
	if hash == "" {
		return InitData{}, errors.New("нет подписи")
	}
	values.Del("hash")

	pairs := []string{}
	for key := range values {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(pairs, "\n")))
	if subtle.ConstantTimeCompare([]byte(hash), []byte(hex.EncodeToString(mac.Sum(nil)))) != 1 {
		return InitData{}, errors.New("подпись не совпадает")
	}

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return InitData{}, errors.New("нет даты входа")
	}
	if now.Sub(time.Unix(authDate, 0)) > initDataMaxAge {
		return InitData{}, errors.New("данные устарели")
	}

	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil || user.ID == 0 {
		return InitData{}, errors.New("нет пользователя")
	}
	return InitData{UserID: strconv.FormatInt(user.ID, 10), StartParam: values.Get("start_param")}, nil
}

// в MAX пользователь определяется только по подписанным данным миниаппа.
// Заголовок X-User-Id от клиента при включённом боте не принимаем, иначе можно выдать себя за другого.
// Без данных MAX (открыли в обычном браузере) — общий демо-пользователь
func (s *Server) identify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.botToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		r.Header.Del("X-User-Id")
		raw := r.Header.Get("X-Max-Init-Data")
		if raw != "" {
			data, err := validateInitData(raw, s.botToken, s.now())
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Не удалось проверить вход. Откройте приложение заново")
				return
			}
			r.Header.Set("X-User-Id", data.UserID)
		}
		next.ServeHTTP(w, r)
	})
}
