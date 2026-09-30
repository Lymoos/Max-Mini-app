package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeAI struct {
	answer string
	fail   bool
	calls  int
}

func (f *fakeAI) Complete(ctx context.Context, system, user string) (string, error) {
	f.calls++
	if f.fail {
		return "", errors.New("нет сети")
	}
	return f.answer, nil
}

func TestYandexGPTRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Api-Key secret" {
			t.Errorf("неверная авторизация: %q", r.Header.Get("Authorization"))
		}
		var body struct {
			ModelURI string       `json:"modelUri"`
			Messages []gptMessage `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.ModelURI != "gpt://folder1/yandexgpt-lite/latest" {
			t.Errorf("modelUri = %q", body.ModelURI)
		}
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[1].Text != "нурафен" {
			t.Errorf("сообщения: %+v", body.Messages)
		}
		w.Write([]byte(`{"result":{"alternatives":[{"message":{"role":"assistant","text":" Ибупрофен \n"}}]}}`))
	}))
	defer srv.Close()

	y := NewYandexGPT("secret", "folder1")
	y.url = srv.URL
	got, err := y.Complete(context.Background(), "система", "нурафен")
	if err != nil || got != "Ибупрофен" {
		t.Errorf("получили %q %v", got, err)
	}
}

func TestYandexGPTErrors(t *testing.T) {
	for _, tt := range []struct {
		status int
		body   string
	}{
		{http.StatusUnauthorized, `{"error":"bad key"}`},
		{http.StatusOK, `не json`},
		{http.StatusOK, `{"result":{"alternatives":[]}}`},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
			w.Write([]byte(tt.body))
		}))
		y := NewYandexGPT("k", "f")
		y.url = srv.URL
		if _, err := y.Complete(context.Background(), "s", "u"); err == nil {
			t.Errorf("ожидали ошибку для %d %s", tt.status, tt.body)
		}
		srv.Close()
	}
}

func TestParseIntent(t *testing.T) {
	tests := []struct {
		answer string
		want   aiIntent
		ok     bool
	}{
		{`{"type": "medicine", "target": "Ибупрофен"}`, aiIntent{"medicine", "Ибупрофен"}, true},
		{"```json\n{\"type\": \"tab\", \"target\": \"help\"}\n```", aiIntent{"tab", "help"}, true},
		{"```\n{\"type\": \"unknown\"}\n```", aiIntent{Type: "unknown"}, true},
		{`не знаю`, aiIntent{}, false},
	}
	for _, tt := range tests {
		got, ok := parseIntent(tt.answer)
		if ok != tt.ok || got != tt.want {
			t.Errorf("parseIntent(%q) = %+v %v", tt.answer, got, ok)
		}
	}
}

func TestAICorrect(t *testing.T) {
	tests := []struct {
		answer string
		wantID string
	}{
		{"Парацетамол", "paracetamol"},
		{"«парацетамол».", "paracetamol"},
		{"НЕТ", ""},
		{"Супертаблетка", ""},
	}
	for _, tt := range tests {
		t.Run(tt.answer, func(t *testing.T) {
			s := testStore(t)
			srv := NewServer(s, &fakeGeocoder{}, &fakeAI{answer: tt.answer})
			item, ok := srv.aiCorrect(context.Background(), "medicine", medicineItems(), "что-то от температуры")
			if tt.wantID == "" && ok {
				t.Errorf("ИИ не должен выдумывать: %+v", item)
			}
			if tt.wantID != "" && (!ok || item.ID != tt.wantID) {
				t.Errorf("получили %+v %v", item, ok)
			}
		})
	}
}

func TestAIUsesCache(t *testing.T) {
	s := testStore(t)
	ai := &fakeAI{answer: "Молоко 2,5%"}
	srv := NewServer(s, &fakeGeocoder{}, ai)

	for i := 0; i < 3; i++ {
		item, ok := srv.aiCorrect(context.Background(), "product", productItems(), "молочко коровье")
		if !ok || item.ID != "milk" {
			t.Fatalf("попытка %d: %+v %v", i, item, ok)
		}
	}
	if ai.calls != 1 {
		t.Errorf("ИИ вызвали %d раз, должен один — дальше кеш", ai.calls)
	}
}

func TestAIDownOrMissing(t *testing.T) {
	s := testStore(t)
	srv := NewServer(s, &fakeGeocoder{}, &fakeAI{fail: true})
	if _, ok := srv.aiCorrect(context.Background(), "medicine", medicineItems(), "от головы"); ok {
		t.Error("при ошибке ИИ результата быть не должно")
	}
	if _, ok, _ := s.CacheGet(context.Background(), "correct:medicine:"+string(normalize("от головы"))); ok {
		t.Error("ошибки не кешируются")
	}

	srv = NewServer(s, &fakeGeocoder{}, nil)
	if _, ok := srv.aiAnswer(context.Background(), "что-нибудь"); ok {
		t.Error("без ИИ ответа нет")
	}
}

func TestAIAnswer(t *testing.T) {
	tests := []struct {
		answer     string
		wantType   string
		wantTarget string
	}{
		{`{"type": "medicine", "target": "Ибупрофен"}`, "medicine", "ibuprofen"},
		{`{"type": "product", "target": "Гречка"}`, "product", "buckwheat"},
		{`{"type": "feature", "target": "social"}`, "feature", "social"},
		{`{"type": "feature", "target": "taxi"}`, "", ""},
		{`{"type": "tab", "target": "documents"}`, "tab", "documents"},
		{`{"type": "tab", "target": "settings"}`, "", ""},
		{`{"type": "medicine", "target": "Выдуманное"}`, "", ""},
		{`{"type": "unknown"}`, "", ""},
		{`мусор`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.answer, func(t *testing.T) {
			s := testStore(t)
			srv := NewServer(s, &fakeGeocoder{}, &fakeAI{answer: tt.answer})
			got, ok := srv.aiAnswer(context.Background(), "у меня болит голова")
			if tt.wantType == "" {
				if ok {
					t.Errorf("не ожидали ответа: %+v", got)
				}
				return
			}
			if !ok || got.Type != tt.wantType || got.Target != tt.wantTarget || got.Message == "" {
				t.Errorf("получили %+v %v", got, ok)
			}
			if tt.wantType == "medicine" && got.Medicine == nil || tt.wantType == "product" && got.Product == nil {
				t.Error("нет объекта лекарства или товара")
			}
		})
	}
}

func TestAIPromptHasCatalog(t *testing.T) {
	names := catalogNames(medicineItems())
	if !strings.Contains(names, "Парацетамол") || !strings.Contains(names, "; ") {
		t.Errorf("в подсказке для ИИ нет каталога: %s", names)
	}
}
