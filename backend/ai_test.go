package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
		{`{"type": "medicine", "target": "Ибупрофен"}`, aiIntent{Type: "medicine", Target: "Ибупрофен"}, true},
		{"```json\n{\"type\": \"tab\", \"target\": \"help\"}\n```", aiIntent{Type: "tab", Target: "help"}, true},
		{"```\n{\"type\": \"unknown\"}\n```", aiIntent{Type: "unknown"}, true},
		{`{"type": "task", "title": "Выпить таблетку", "time": "09:00", "date": "2026-10-01"}`,
			aiIntent{Type: "task", Title: "Выпить таблетку", Time: "09:00", Date: "2026-10-01"}, true},
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
		{`{"type": "product", "target": "Гречка"}`, "product", "buckwheat"},
		{`{"type": "product", "target": "buckwheat"}`, "product", "buckwheat"},
		{`{"type": "product", "target": "Космолёт"}`, "", ""},
		{`{"type": "pharmacy"}`, "feature", "pharmacy"},
		{`{"type": "shops"}`, "feature", "goods"},
		{`{"type": "social"}`, "feature", "social"},
		{`{"type": "taxi"}`, "", ""},
		{`{"type": "doctor", "target": "neurologist"}`, "doctor", "neurologist"},
		{`{"type": "doctor", "target": ""}`, "doctor", ""},
		{`{"type": "doctor", "target": "экстрасенс"}`, "", ""},
		{`{"type": "benefit", "target": "overhaul"}`, "benefit", "overhaul"},
		{`{"type": "benefit", "target": "free-car"}`, "", ""},
		{`{"type": "guide", "target": "font"}`, "guide", "font"},
		{`{"type": "guide", "target": "hack-bank"}`, "", ""},
		{`{"type": "documents"}`, "tab", "documents"},
		{`{"type": "settings"}`, "", ""},
		{`{"type": "profile"}`, "profile", ""},
		{`{"type": "treatment"}`, "tab", "medicines"},
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
			if tt.wantType == "product" && got.Product == nil {
				t.Error("нет объекта товара")
			}
		})
	}
}

// ИИ назвал препарат, которого человек не называл, — это назначение лечения, не пропускаем
func TestAIDoesNotPrescribe(t *testing.T) {
	tests := []struct {
		text     string
		answer   string
		wantType string
	}{
		{"что выпить чтобы сердце не болело", `{"type": "medicine", "target": "Валидол"}`, "tab"},
		{"давление 160 что делать", `{"type": "medicine", "target": "Каптоприл"}`, "tab"},
		{"голова раскалывается", `{"type": "medicine", "target": "Цитрамон"}`, "tab"},
		{"нужен этот цитромон", `{"type": "medicine", "target": "Цитрамон"}`, "medicine"},
		{"есть ли каптоприл", `{"type": "medicine", "target": "captopril"}`, "tab"},
		{"хочу капотен", `{"type": "medicine", "target": "kaptopril"}`, "medicine"},
		{"что-нибудь посильнее", `{"type": "medicine", "target": "Выдуманное"}`, "tab"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			srv := NewServer(testStore(t), &fakeGeocoder{}, &fakeAI{answer: tt.answer})
			got, ok := srv.aiAnswer(context.Background(), tt.text)
			if !ok || got.Type != tt.wantType {
				t.Fatalf("получили %+v %v", got, ok)
			}
			if tt.wantType == "tab" && (got.Medicine != nil || got.Target != "medicines" || !strings.Contains(got.Message, "врач")) {
				t.Errorf("нужен раздел лекарств и совет обратиться к врачу: %+v", got)
			}
		})
	}
}

func TestAIPromptForbidsTreatment(t *testing.T) {
	prompt := intentPrompt("2026-09-30")
	for _, want := range []string{"не назначаешь лечение", "Никогда не называй лекарство", "type=treatment", "только JSON",
		"cardiologist", "overhaul", "font", "Молоко 2,5%", "Сегодня 2026-09-30"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("в системном промпте нет %q", want)
		}
	}
}

func TestAITask(t *testing.T) {
	srv := NewServer(testStore(t), &fakeGeocoder{}, nil)
	srv.now = func() time.Time { return askNow }
	tests := []struct {
		answer string
		ok     bool
		want   TaskDraft
	}{
		{`{"type": "task", "title": "выпить таблетку", "time": "09:00"}`, true, TaskDraft{"Выпить таблетку", "09:00", "2026-09-30", "medicine"}},
		{`{"type": "task", "title": "позвонить сыну", "time": "18:30", "date": "2026-10-03"}`, true, TaskDraft{"Позвонить сыну", "18:30", "2026-10-03", "call"}},
		{`{"type": "task", "title": "полить цветы", "time": ""}`, true, TaskDraft{"Полить цветы", "", "2026-09-30", "other"}},
		{`{"type": "task", "title": "x", "time": "25:00"}`, false, TaskDraft{}},
		{`{"type": "task", "title": "x", "time": "9 утра"}`, false, TaskDraft{}},
		{`{"type": "task", "title": "x", "time": "09:00", "date": "2020-01-01"}`, false, TaskDraft{}},
		{`{"type": "task", "title": "x", "time": "09:00", "date": "2030-01-01"}`, false, TaskDraft{}},
		{`{"type": "task", "title": "", "time": "09:00"}`, false, TaskDraft{}},
	}
	for _, tt := range tests {
		srv.ai = &fakeAI{answer: tt.answer}
		got, ok := srv.aiAnswer(context.Background(), "дела на потом "+tt.answer)
		if ok != tt.ok {
			t.Errorf("%s: ok=%v", tt.answer, ok)
			continue
		}
		if ok && (got.Task == nil || *got.Task != tt.want) {
			t.Errorf("%s: %+v", tt.answer, got.Task)
		}
	}
}

func TestAIPromptHasCatalog(t *testing.T) {
	names := catalogNames(medicineItems())
	if !strings.Contains(names, "Парацетамол") || !strings.Contains(names, "; ") {
		t.Errorf("в подсказке для ИИ нет каталога: %s", names)
	}
}
