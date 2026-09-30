package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type fakeReader struct {
	text  string
	err   error
	calls int
}

func (f *fakeReader) ReadText(ctx context.Context, data []byte, mime string) (string, error) {
	f.calls++
	return f.text, f.err
}

const ocrRecipe = `РЕЦЕПТ
ГИ № 2, врач Петрова А.С.

Вр: Лизиноприл 10 мг — по 1 таб. утром
Вр: Аторвастатин 20 мг — 1 таб. вечером
Вр: Кардиомагнил 75 мг — 1 таб. в обед
Вр: Омепразол 20 мг — 1 капс. до еды
Подпись врача`

func recipeIDs(items []RecipeItem) string {
	ids := []string{}
	for _, it := range items {
		ids = append(ids, it.MedicineID)
	}
	return strings.Join(ids, ",")
}

func TestRecipeByRules(t *testing.T) {
	items := recipeByRules(ocrRecipe)
	if recipeIDs(items) != "lizinopril,atorvastatin,cardiomagnil,omeprazol" {
		t.Fatalf("%+v", items)
	}
	if items[0].Title != "Лизиноприл" || items[0].Dose != "10 мг, по 1 таб. утром" {
		t.Errorf("%+v", items[0])
	}

	// ошибки распознавания и повтор одного лекарства
	items = recipeByRules("Rp: Лизинаприл 10мг\nRp: Лизиноприл 10 мг\nМетформен 850 мг 2 раза")
	if recipeIDs(items) != "lizinopril,metformin" || items[1].Dose != "850 мг" {
		t.Errorf("%+v", items)
	}

	// как принимать — следующей строкой, «мг» распознан как «Mr»
	items = recipeByRules("Вр: Конкор 5 Mr\n1 таб. утром\nВр: Омепразол 20 мг\nВрач Петрова")
	if recipeIDs(items) != "bisoprolol,omeprazol" || items[0].Title != "Конкор (Бисопролол)" || items[0].Dose != "5 Mr, 1 таб. утром" ||
		items[1].Title != "Омепразол" || items[1].Dose != "20 мг" {
		t.Errorf("%+v", items)
	}

	for _, text := range []string{"", "Врач Петрова А.С.\nПоликлиника № 2\nПодпись, печать", "Анализ крови общий"} {
		if items := recipeByRules(text); len(items) != 0 {
			t.Errorf("%q: лишнее %+v", text, items)
		}
	}
}

func TestDoseFromLine(t *testing.T) {
	cases := map[string]string{
		"Лизиноприл 10 мг — по 1 таб. утром": "10 мг, по 1 таб. утром",
		"Витамин D3 2000 МЕ":                 "2000 МЕ",
		"Нитроглицерин 0,5 мг":               "0,5 мг",
		"Омепразол — перед едой":             "перед едой",
		"Смекта": "",
	}
	for line, want := range cases {
		if got := doseFromLine(line); got != want {
			t.Errorf("%q: %q, ожидали %q", line, got, want)
		}
	}
}

func TestRecognizeRecipeOrder(t *testing.T) {
	srv, _ := newTestServer(t)
	skip := &fakeReader{err: errUnsupported}
	broken := &fakeReader{err: errors.New("упал")}
	empty := &fakeReader{text: "Врач Петрова"}
	good := &fakeReader{text: ocrRecipe}
	last := &fakeReader{text: ocrRecipe}
	srv.readers = []TextReader{skip, broken, empty, good, last}

	items, err := srv.recognizeRecipe(context.Background(), []byte("фото"), "image/png")
	if err != nil || len(items) != 4 {
		t.Fatalf("%+v %v", items, err)
	}
	if last.calls != 0 {
		t.Error("платный способ не нужен, если бесплатный уже нашёл лекарства")
	}

	items, _ = srv.recognizeRecipe(context.Background(), []byte("фото"), "image/png")
	if good.calls != 1 || len(items) != 4 {
		t.Error("то же фото второй раз берём из кеша")
	}
}

func TestRecognizeRecipeNothing(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.readers = []TextReader{&fakeReader{text: "Справка для бассейна"}}
	items, err := srv.recognizeRecipe(context.Background(), []byte("x"), "image/png")
	if err != nil || len(items) != 0 {
		t.Errorf("%+v %v", items, err)
	}

	srv.readers = []TextReader{&fakeReader{err: errors.New("нет")}, &fakeReader{err: errUnsupported}}
	if _, err := srv.recognizeRecipe(context.Background(), []byte("y"), "image/png"); err == nil {
		t.Error("если ни один способ не прочитал файл — ошибка")
	}
}

// ИИ дополняет правила лекарствами не из справочника. Выдуманное отбрасываем
func TestRecipeByAI(t *testing.T) {
	srv, _ := newTestServer(t)
	text := "Rp: Тромбопол 75 мг — 1 таб. в обед\nRp: Физиотенз 0,4 мг"
	srv.ai = &fakeAI{answer: "```json\n" + `{"items": [
		{"name": "Тромбопол", "dose": "75 мг, 1 таб. в обед"},
		{"name": "Моксонидин", "dose": "0,4 мг"},
		{"name": "Супертаблетка", "dose": "1 шт"}
	]}` + "\n```"}
	srv.readers = []TextReader{&fakeReader{text: text}}

	items, err := srv.recognizeRecipe(context.Background(), []byte("z"), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].MedicineID != "moxonidin" || items[1].Title != "Тромбопол" || items[1].MedicineID != "" ||
		items[1].Dose != "75 мг, 1 таб. в обед" {
		t.Errorf("правила находят справочное, ИИ — остальное, выдуманное отбрасываем: %+v", items)
	}

	srv.ai = &fakeAI{answer: "не json"}
	if items := srv.recipeByAI(context.Background(), text, "k2"); len(items) != 0 {
		t.Errorf("%+v", items)
	}
}

func TestTesseractAndPdfText(t *testing.T) {
	png, _ := os.ReadFile("testdata/recipe.png")
	pdf, _ := os.ReadFile("testdata/recipe.pdf")

	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Log("tesseract не установлен, пропускаем")
	} else {
		text, err := Tesseract{}.ReadText(context.Background(), png, "image/png")
		if err != nil || recipeIDs(recipeByRules(text)) != "lizinopril,atorvastatin,omeprazol" {
			t.Errorf("%q %v", text, err)
		}
		if _, err := (Tesseract{}).ReadText(context.Background(), pdf, "application/pdf"); !errors.Is(err, errUnsupported) {
			t.Error("PDF читает pdftotext")
		}
	}

	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Log("pdftotext не установлен, пропускаем")
	} else {
		text, err := PdfText{}.ReadText(context.Background(), pdf, "application/pdf")
		if err != nil || recipeIDs(recipeByRules(text)) != "metformin,cardiomagnil" {
			t.Errorf("%q %v", text, err)
		}
	}
}

func TestYandexOCR(t *testing.T) {
	var got map[string]any
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Api-Key k" || r.Header.Get("x-folder-id") != "f" {
			t.Errorf("заголовки: %v", r.Header)
		}
		json.NewDecoder(r.Body).Decode(&got)
		if got["mimeType"] == "PDF" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"result": {"textAnnotation": {"fullText": "Rp: Лизиноприл 10 мг"}}}`))
	}))
	defer fake.Close()

	y := NewYandexOCR("k", "f", "handwritten")
	y.url = fake.URL
	text, err := y.ReadText(context.Background(), []byte("jpg"), "image/jpeg")
	if err != nil || text != "Rp: Лизиноприл 10 мг" || got["model"] != "handwritten" || got["mimeType"] != "JPEG" || got["content"] != "anBn" {
		t.Errorf("%q %v %v", text, err, got)
	}
	if _, err := y.ReadText(context.Background(), []byte("pdf"), "application/pdf"); !errors.Is(err, errUnsupported) {
		t.Error("рукописная модель PDF не принимает")
	}

	y = NewYandexOCR("k", "f", "page")
	y.url = fake.URL
	if _, err := y.ReadText(context.Background(), []byte("pdf"), "application/pdf"); err == nil {
		t.Error("ошибка API должна вернуться")
	}
	if _, err := y.ReadText(context.Background(), []byte("gif"), "image/gif"); !errors.Is(err, errUnsupported) {
		t.Error("gif не поддерживается")
	}
}

func uploadRecipe(t *testing.T, srv *Server, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if data != nil {
		part, _ := form.CreateFormFile("photo", "recipe.png")
		part.Write(data)
	}
	form.Close()
	req := httptest.NewRequest("POST", "/api/recipe", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

func TestRecipeHandler(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.readers = []TextReader{&fakeReader{text: ocrRecipe}}
	png, _ := os.ReadFile("testdata/recipe.png")

	rec := uploadRecipe(t, srv, png)
	var resp struct {
		Items []RecipeItem `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if rec.Code != http.StatusOK || len(resp.Items) != 4 || resp.Items[0].MedicineID != "lizinopril" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}

	bad := map[string][]byte{
		"без файла":   nil,
		"не картинка": []byte("просто текст, а не фото"),
		"больше 5 МБ": append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, maxRecipeSize)...),
	}
	for name, data := range bad {
		if rec := uploadRecipe(t, srv, data); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}

	srv.readers = []TextReader{&fakeReader{err: errors.New("нет")}}
	if rec := uploadRecipe(t, srv, append(png, 1)); rec.Code != http.StatusBadGateway || !strings.Contains(rec.Body.String(), "Не получилось прочитать") {
		t.Errorf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestRecipeReaders(t *testing.T) {
	readers := recipeReaders("", "")
	for _, r := range readers {
		if _, ok := r.(*YandexOCR); ok {
			t.Error("без ключа Yandex не подключаем")
		}
	}
	withKey := recipeReaders("k", "f")
	if len(withKey) != len(readers)+2 {
		t.Errorf("с ключом добавляются две модели Yandex: %d", len(withKey))
	}
}
