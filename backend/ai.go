package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type AI interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

type YandexGPT struct {
	apiKey   string
	folderID string
	url      string
	client   *http.Client
}

func NewYandexGPT(apiKey, folderID string) *YandexGPT {
	return &YandexGPT{
		apiKey:   apiKey,
		folderID: folderID,
		url:      "https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
		client:   &http.Client{Timeout: 8 * time.Second},
	}
}

type gptMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

func (y *YandexGPT) Complete(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"modelUri": "gpt://" + y.folderID + "/yandexgpt-lite/latest",
		"completionOptions": map[string]any{
			"stream":      false,
			"temperature": 0.1,
			"maxTokens":   100,
		},
		"messages": []gptMessage{
			{Role: "system", Text: system},
			{Role: "user", Text: user},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, y.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Api-Key "+y.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := y.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("yandexgpt: статус %d", resp.StatusCode)
	}

	var data struct {
		Result struct {
			Alternatives []struct {
				Message gptMessage `json:"message"`
			} `json:"alternatives"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Result.Alternatives) == 0 {
		return "", errors.New("yandexgpt: пустой ответ")
	}
	return strings.TrimSpace(data.Result.Alternatives[0].Message.Text), nil
}

// ответы ИИ сохраняем в базе: одинаковые запросы не тратят деньги и время
func (s *Server) askAI(ctx context.Context, cacheKey, system, user string) (string, bool) {
	if s.ai == nil {
		return "", false
	}
	if answer, ok, err := s.store.CacheGet(ctx, cacheKey); err == nil && ok {
		return answer, true
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	answer, err := s.ai.Complete(ctx, system, user)
	if err != nil {
		log.Printf("ai: %v", err)
		return "", false
	}
	if err := s.store.CachePut(ctx, cacheKey, answer); err != nil {
		log.Printf("ai cache: %v", err)
	}
	return answer, true
}

func catalogNames(items []CatalogItem) string {
	names := []string{}
	for _, item := range items {
		names = append(names, item.Name)
	}
	return strings.Join(names, "; ")
}

func findByName(items []CatalogItem, name string) (CatalogItem, bool) {
	want := string(normalize(name))
	for _, item := range items {
		if string(normalize(item.Name)) == want {
			return item, true
		}
	}
	return CatalogItem{}, false
}

// ИИ выбирает название только из нашего списка, поэтому выдуманное лекарство не пройдёт
func (s *Server) aiCorrect(ctx context.Context, kind string, items []CatalogItem, query string) (CatalogItem, bool) {
	what := "лекарство"
	if kind == "product" {
		what = "товар в магазине"
	}
	system := "Пользователь ищет " + what + ". Он мог ошибиться в написании или описать его своими словами. " +
		"Выбери одно подходящее название из списка и ответь только им, без пояснений. " +
		"Если ничего не подходит, ответь НЕТ.\nСписок: " + catalogNames(items)

	answer, ok := s.askAI(ctx, "correct:"+kind+":"+string(normalize(query)), system, query)
	if !ok {
		return CatalogItem{}, false
	}
	return findByName(items, strings.Trim(answer, " .«»\"'"))
}

type aiIntent struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

func parseIntent(answer string) (aiIntent, bool) {
	answer = strings.TrimSpace(answer)
	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimSuffix(answer, "```")

	var intent aiIntent
	if err := json.Unmarshal([]byte(strings.TrimSpace(answer)), &intent); err != nil {
		return aiIntent{}, false
	}
	return intent, true
}

// сюда попадают запросы, которые не поняли простые правила
func (s *Server) aiAnswer(ctx context.Context, text string) (AskAnswer, bool) {
	featureIDs := []string{}
	for _, f := range features {
		featureIDs = append(featureIDs, f.ID+" ("+f.Title+")")
	}
	system := "Ты помощник в приложении для пожилых людей. Определи, что нужно пользователю, и ответь только JSON вида " +
		`{"type": "...", "target": "..."}` + ". Варианты:\n" +
		"- type=medicine, target — название лекарства из списка: " + catalogNames(medicineItems()) + "\n" +
		"- type=product, target — название товара из списка: " + catalogNames(productItems()) + "\n" +
		"- type=feature, target — id из списка: " + strings.Join(featureIDs, ", ") + "\n" +
		"- type=tab, target — medicines (лекарства), documents (документы) или help (помощь)\n" +
		"- type=unknown, если ничего не подходит."

	answer, ok := s.askAI(ctx, "intent:"+string(normalize(text)), system, text)
	if !ok {
		return AskAnswer{}, false
	}
	intent, ok := parseIntent(answer)
	if !ok {
		return AskAnswer{}, false
	}

	switch intent.Type {
	case "medicine":
		if item, ok := findByName(medicineItems(), intent.Target); ok {
			return medicineAnswer(item.ID), true
		}
	case "product":
		if item, ok := findByName(productItems(), intent.Target); ok {
			return productAnswer(item.ID), true
		}
	case "feature":
		if f, ok := findFeature(intent.Target); ok {
			return AskAnswer{Type: "feature", Target: f.ID, Message: "Открываю «" + f.Title + "»", Feature: &f}, true
		}
	case "tab":
		if title, ok := tabTitles[intent.Target]; ok {
			return AskAnswer{Type: "tab", Target: intent.Target, Message: "Открываю раздел «" + title + "»"}, true
		}
	}
	return AskAnswer{}, false
}
