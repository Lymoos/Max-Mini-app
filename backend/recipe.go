package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const maxRecipeSize = 5 << 20

var errUnsupported = errors.New("этот тип файла не поддерживается")

// читает текст с фото или PDF. Способов несколько: сначала бесплатные на своём сервере, потом Yandex Vision
type TextReader interface {
	ReadText(ctx context.Context, data []byte, mime string) (string, error)
}

// Tesseract — бесплатная программа распознавания текста, ставится на сервер (tesseract-ocr, tesseract-ocr-rus)
type Tesseract struct{}

func (Tesseract) ReadText(ctx context.Context, data []byte, mime string) (string, error) {
	if mime != "image/jpeg" && mime != "image/png" {
		return "", errUnsupported
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tesseract", "stdin", "stdout", "-l", "rus+eng")
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	return string(out), err
}

// у электронного рецепта в PDF обычно уже есть текст, его достаёт pdftotext (poppler-utils)
type PdfText struct{}

func (PdfText) ReadText(ctx context.Context, data []byte, mime string) (string, error) {
	if mime != "application/pdf" {
		return "", errUnsupported
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", "-", "-")
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	return string(out), err
}

// платный, но читает лучше, в том числе рукописный текст (model = "handwritten")
type YandexOCR struct {
	apiKey   string
	folderID string
	model    string
	url      string
	client   *http.Client
}

func NewYandexOCR(apiKey, folderID, model string) *YandexOCR {
	return &YandexOCR{
		apiKey:   apiKey,
		folderID: folderID,
		model:    model,
		url:      "https://ocr.api.cloud.yandex.net/ocr/v1/recognizeText",
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

var yandexMimes = map[string]string{"image/jpeg": "JPEG", "image/png": "PNG", "application/pdf": "PDF"}

func (y *YandexOCR) ReadText(ctx context.Context, data []byte, mime string) (string, error) {
	yandexMime, ok := yandexMimes[mime]
	if !ok || (y.model == "handwritten" && mime == "application/pdf") {
		return "", errUnsupported
	}
	body, err := json.Marshal(map[string]any{
		"mimeType":      yandexMime,
		"languageCodes": []string{"ru", "en"},
		"model":         y.model,
		"content":       base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, y.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Api-Key "+y.apiKey)
	req.Header.Set("x-folder-id", y.folderID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := y.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("yandex ocr: статус %d", resp.StatusCode)
	}

	var result struct {
		Result struct {
			TextAnnotation struct {
				FullText string `json:"fullText"`
			} `json:"textAnnotation"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Result.TextAnnotation.FullText, nil
}

// бесплатные способы — если программы установлены, Yandex — если есть ключ
func recipeReaders(apiKey, folderID string) []TextReader {
	readers := []TextReader{}
	if _, err := exec.LookPath("tesseract"); err == nil {
		readers = append(readers, Tesseract{})
	} else {
		log.Print("tesseract не установлен — фото рецептов читаем только через Yandex Vision")
	}
	if _, err := exec.LookPath("pdftotext"); err == nil {
		readers = append(readers, PdfText{})
	}
	if apiKey != "" && folderID != "" {
		readers = append(readers, NewYandexOCR(apiKey, folderID, "page"), NewYandexOCR(apiKey, folderID, "handwritten"))
	}
	return readers
}

type RecipeItem struct {
	Title      string `json:"title"`
	Dose       string `json:"dose"`
	MedicineID string `json:"medicineId"`
}

// «мг» распознавание иногда путает с латинскими «mg» или «Mr»
var doseRe = regexp.MustCompile(`(?i)\d+(?:[.,]\d+)?\s*(?:мг|mg|mr|мкг|мл|ме|ед|г|%)`)

// строка с тем, как принимать: «по 1 таб. в обед», «2 раза в день»
var howRe = regexp.MustCompile(`(?i)^(?:по\s|\d+\s*(?:таб|капс|кап|раз))`)

// дозу и как принимать берём из строки рецепта: «10 мг — по 1 таб. утром»
func doseFromLine(line string) string {
	dose := doseRe.FindString(line)
	if _, how, ok := strings.Cut(line, "—"); ok {
		how = strings.TrimSpace(how)
		if dose != "" && how != "" {
			return dose + ", " + how
		}
		if how != "" {
			return how
		}
	}
	return dose
}

// лекарства из справочника, найденные по строкам рецепта. Одно лекарство — один пункт
func recipeByRules(text string) []RecipeItem {
	items := []RecipeItem{}
	seen := map[string]bool{}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		item, ok := findInText(medicineItems(), line)
		if !ok || seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		m, _ := findMedicine(item.ID)
		dose := doseFromLine(line)
		// как принимать часто пишут следующей строкой
		if i+1 < len(lines) && !strings.Contains(line, "—") {
			next := strings.TrimSpace(lines[i+1])
			if howRe.MatchString(next) {
				dose = strings.TrimPrefix(dose+", "+next, ", ")
			}
		}
		items = append(items, RecipeItem{Title: writtenName(m, line), Dose: dose, MedicineID: m.ID})
	}
	return items
}

// в рецепте часто торговое название — показываем его, чтобы человек узнал своё лекарство: «Конкор (Бисопролол)»
func writtenName(m Medicine, line string) string {
	text := string(normalize(line))
	for _, a := range m.Aliases {
		if len(normalize(a)) >= 4 && strings.Contains(text, string(normalize(a))) {
			return a + " (" + m.Name + ")"
		}
	}
	return m.Name
}

// в рецепте бывают лекарства не из справочника, их выписывает ИИ. Каждое название должно быть в тексте рецепта
func (s *Server) recipeByAI(ctx context.Context, text, cacheKey string) []RecipeItem {
	system := "Тебе дают текст рецепта врача, распознанный с фото, в нём бывают ошибки распознавания. " +
		"Выпиши лекарства, которые назначены, ровно как они написаны в тексте, и дозировку. Ничего не добавляй от себя. " +
		`Ответь только JSON: {"items": [{"name": "...", "dose": "..."}]}. Если лекарств нет — {"items": []}.`
	answer, ok := s.askAI(ctx, cacheKey, system, text)
	if !ok {
		return nil
	}
	answer = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(answer), "```json"), "```"), "```"))
	var parsed struct {
		Items []struct {
			Name string `json:"name"`
			Dose string `json:"dose"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(answer), &parsed); err != nil {
		return nil
	}

	items := []RecipeItem{}
	for _, p := range parsed.Items {
		name := strings.TrimSpace(p.Name)
		if name == "" || len([]rune(name)) > maxTitleLen || !writtenIn(name, text) {
			continue
		}
		item := RecipeItem{Title: name, Dose: strings.TrimSpace(p.Dose)}
		if found, ok := findInText(medicineItems(), name); ok {
			m, _ := findMedicine(found.ID)
			item.Title, item.MedicineID = m.Name, m.ID
		}
		items = append(items, item)
	}
	return items
}

// к найденному правилами добавляем то, что нашёл ИИ, без повторов
func mergeRecipe(rules, ai []RecipeItem) []RecipeItem {
	result := append([]RecipeItem{}, rules...)
	for _, item := range ai {
		repeat := false
		for _, have := range result {
			if (item.MedicineID != "" && item.MedicineID == have.MedicineID) || string(normalize(item.Title)) == string(normalize(have.Title)) {
				repeat = true
			}
		}
		if !repeat {
			result = append(result, item)
		}
	}
	return result
}

// название от ИИ должно встречаться в распознанном тексте, хотя бы с опечаткой
func writtenIn(name, text string) bool {
	if strings.Contains(string(normalize(text)), string(normalize(name))) {
		return true
	}
	_, ok := findInText([]CatalogItem{{ID: "x", Name: name}}, text)
	return ok
}

func (s *Server) recognizeRecipe(ctx context.Context, data []byte, mime string) ([]RecipeItem, error) {
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	if cached, ok, err := s.store.CacheGet(ctx, "recipe:"+hash); err == nil && ok {
		var items []RecipeItem
		if json.Unmarshal([]byte(cached), &items) == nil {
			return items, nil
		}
	}

	readOnce := false
	for i, r := range s.readers {
		text, err := r.ReadText(ctx, data, mime)
		if errors.Is(err, errUnsupported) {
			continue
		}
		if err != nil {
			log.Printf("рецепт, способ %d: %v", i, err)
			continue
		}
		readOnce = true
		items := recipeByRules(text)
		if strings.TrimSpace(text) != "" {
			items = mergeRecipe(items, s.recipeByAI(ctx, text, fmt.Sprintf("recipe-ai:%s:%d", hash, i)))
		}
		if len(items) > 0 {
			if saved, err := json.Marshal(items); err == nil {
				s.store.CachePut(ctx, "recipe:"+hash, string(saved))
			}
			return items, nil
		}
	}
	if !readOnce {
		return nil, errors.New("не удалось прочитать файл")
	}
	return []RecipeItem{}, nil
}

func (s *Server) readRecipe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRecipeSize+1<<20)
	file, _, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Пришлите фото рецепта (до 5 МБ)")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxRecipeSize+1))
	if err != nil || len(data) > maxRecipeSize {
		writeError(w, http.StatusBadRequest, "Файл слишком большой, нужен до 5 МБ")
		return
	}

	mime := http.DetectContentType(data)
	if _, ok := yandexMimes[mime]; !ok {
		writeError(w, http.StatusBadRequest, "Пришлите фото (JPG или PNG) или PDF электронного рецепта")
		return
	}

	items, err := s.recognizeRecipe(r.Context(), data, mime)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Не получилось прочитать рецепт. Попробуйте ещё раз")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
