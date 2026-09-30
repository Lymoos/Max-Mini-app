package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
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
			"maxTokens":   400,
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
	Title  string `json:"title"`
	Time   string `json:"time"`
	Date   string `json:"date"`
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

func intentPrompt(today string) string {
	specialtyIDs := []string{}
	for _, sp := range doctorSpecialties {
		specialtyIDs = append(specialtyIDs, sp.ID+" ("+sp.Name+")")
	}
	benefitIDs := []string{}
	for _, b := range benefits {
		benefitIDs = append(benefitIDs, b.ID+" ("+b.Title+")")
	}
	guideIDs := []string{}
	for _, g := range guides {
		guideIDs = append(guideIDs, g.ID+" ("+g.Title+")")
	}

	return "Ты помощник в приложении для пожилых людей. Определи, что нужно пользователю, и ответь только JSON без пояснений: " +
		`{"type": "...", "target": "...", "title": "", "time": "", "date": ""}` + ".\n" +
		"ВАЖНО: ты не врач и не назначаешь лечение. Никогда не называй лекарство, если пользователь сам не написал его название. " +
		"На вопросы вроде «что выпить от давления», «чем лечить сердце», «что принять от боли» отвечай type=treatment.\n" +
		"Варианты type:\n" +
		"- medicine: пользователь сам назвал лекарство, target — это название как он его написал\n" +
		"- treatment: спрашивает, чем лечиться, но лекарство не назвал\n" +
		"- product: товар в магазине, target — название из списка: " + catalogNames(productItems()) + "\n" +
		"- pharmacy: аптеки рядом; shops: магазины рядом; social: соцпомощь, соцзащита, волонтёры\n" +
		"- doctor: запись к врачу, target — id специальности или пусто: " + strings.Join(specialtyIDs, ", ") + "\n" +
		"- benefit: льгота или документ, target — id: " + strings.Join(benefitIDs, ", ") + "\n" +
		"- guide: инструкция по телефону, target — id: " + strings.Join(guideIDs, ", ") + "\n" +
		"- task: напоминание или дело, title — что сделать, time — ЧЧ:ММ, date — ГГГГ-ММ-ДД. Сегодня " + today + "\n" +
		"- medicines: раздел лекарств; documents: документы и льготы; help: инструкции; profile: профиль, адрес, имя\n" +
		"- unknown: ничего не подходит.\n" +
		"Используй только id и названия из списков."
}

var clockOnlyRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// лекарство принимаем от ИИ, только если человек сам его назвал, пусть и с ошибкой
func namedInText(item CatalogItem, text string) bool {
	if _, ok := findInText([]CatalogItem{item}, text); ok {
		return true
	}
	for _, m := range fuzzySuggest([]CatalogItem{item}, text, 1) {
		if m.ID == item.ID {
			return true
		}
	}
	return false
}

func catalogByNameOrID(items []CatalogItem, target string) (CatalogItem, bool) {
	if item, ok := findByName(items, target); ok {
		return item, true
	}
	want := string(normalize(target))
	for _, item := range items {
		if item.ID == target {
			return item, true
		}
		for _, a := range item.Aliases {
			if string(normalize(a)) == want {
				return item, true
			}
		}
	}
	return CatalogItem{}, false
}

// каждое значение от ИИ сверяем со справочниками: выдуманное не пропускаем
func checkIntent(intent aiIntent, text string, now time.Time) (AskAnswer, bool) {
	switch intent.Type {
	case "medicine":
		item, ok := catalogByNameOrID(medicineItems(), intent.Target)
		if !ok || !namedInText(item, text) {
			return treatmentAnswer(), true
		}
		return medicineAnswer(item.ID), true
	case "treatment":
		return treatmentAnswer(), true
	case "product":
		if item, ok := catalogByNameOrID(productItems(), intent.Target); ok {
			return productAnswer(item.ID), true
		}
	case "pharmacy", "social":
		return featureAnswer(intent.Type), true
	case "shops":
		return featureAnswer("goods"), true
	case "doctor":
		if intent.Target == "" {
			return doctorAnswer(""), true
		}
		if _, ok := findSpecialty(intent.Target); ok {
			return doctorAnswer(intent.Target), true
		}
	case "benefit":
		if _, ok := findBenefit(intent.Target); ok {
			return benefitAnswer(intent.Target), true
		}
	case "guide":
		if _, ok := findGuide(intent.Target); ok {
			return guideAnswer(intent.Target), true
		}
	case "task":
		return checkTask(intent, now)
	case "medicines", "documents", "help":
		return tabAnswers[intent.Type], true
	case "profile":
		return profileAnswer(), true
	}
	return AskAnswer{}, false
}

func checkTask(intent aiIntent, now time.Time) (AskAnswer, bool) {
	title := strings.TrimSpace(intent.Title)
	if title == "" || utf8.RuneCountInString(title) > maxTitleLen {
		return AskAnswer{}, false
	}
	if intent.Time != "" && !clockOnlyRe.MatchString(intent.Time) {
		return AskAnswer{}, false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	date := today
	if intent.Date != "" {
		d, err := time.ParseInLocation("2006-01-02", intent.Date, now.Location())
		if err != nil || d.Before(today) || d.After(today.AddDate(1, 0, 0)) {
			return AskAnswer{}, false
		}
		date = d
	}
	r := []rune(title)
	title = strings.ToUpper(string(r[0])) + string(r[1:])
	t := TaskDraft{Title: title, Time: intent.Time, Date: date.Format("2006-01-02"), Kind: taskKind(strings.ToLower(title))}
	return taskAnswer(t, now), true
}

// сюда попадают запросы, которые не поняли простые правила
func (s *Server) aiAnswer(ctx context.Context, text string) (AskAnswer, bool) {
	now := s.now()
	today := now.Format("2006-01-02")
	// в ответе могут быть «сегодня» и «завтра», поэтому кешируем на один день
	answer, ok := s.askAI(ctx, "intent2:"+today+":"+string(normalize(text)), intentPrompt(today), text)
	if !ok {
		return AskAnswer{}, false
	}
	intent, ok := parseIntent(answer)
	if !ok {
		return AskAnswer{}, false
	}
	return checkIntent(intent, text, now)
}

// сначала правила и нечёткий поиск, ИИ — только если они не справились
func (s *Server) understand(ctx context.Context, text string) AskAnswer {
	answer := answerQuestion(text, s.now())
	if answer.Type == "unknown" {
		if aiAnswer, ok := s.aiAnswer(ctx, text); ok {
			return aiAnswer
		}
	}
	return answer
}
