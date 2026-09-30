package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) (*Server, *Store) {
	t.Helper()
	store := testStore(t)
	ctx := context.Background()
	store.AddTask(ctx, "demo", "2026-09-30", Task{Time: "13:00", Title: "Витамин", Kind: "medicine"})
	store.AddTask(ctx, "demo", "2026-09-30", Task{Time: "08:00", Title: "Таблетка", Kind: "medicine"})
	store.AddTask(ctx, "demo", "2026-09-29", Task{Time: "09:00", Title: "Вчерашняя", Kind: "other"})
	store.AddTask(ctx, "other", "2026-09-30", Task{Time: "10:00", Title: "Чужая", Kind: "other"})

	srv := NewServer(store, &fakeGeocoder{}, nil)
	srv.now = func() time.Time {
		return time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	}
	return srv, store
}

func doRequest(srv *Server, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	srv, _ := newTestServer(t)
	if rec := doRequest(srv, "GET", "/api/health", "", nil); rec.Code != http.StatusOK {
		t.Fatalf("код %d", rec.Code)
	}
}

func TestTodayTasks(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doRequest(srv, "GET", "/api/tasks/today", "", nil)

	var resp struct {
		Date  string `json:"date"`
		Tasks []Task `json:"tasks"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if rec.Code != http.StatusOK || resp.Date != "2026-09-30" || len(resp.Tasks) != 2 || resp.Tasks[0].Title != "Таблетка" {
		t.Errorf("%d %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/tasks/today", "", map[string]string{"X-User-Id": "new"})
	if !strings.Contains(rec.Body.String(), `"tasks":[]`) {
		t.Errorf("пустой список должен быть []: %s", rec.Body.String())
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		body     string
		wantCode int
	}{
		{"отметить", "/api/tasks/1", `{"done": true}`, http.StatusOK},
		{"нет такой задачи", "/api/tasks/999", `{"done": true}`, http.StatusNotFound},
		{"чужая задача", "/api/tasks/4", `{"done": true}`, http.StatusNotFound},
		{"id не число", "/api/tasks/abc", `{"done": true}`, http.StatusBadRequest},
		{"id ноль", "/api/tasks/0", `{"done": true}`, http.StatusBadRequest},
		{"нет поля done", "/api/tasks/1", `{}`, http.StatusBadRequest},
		{"лишнее поле", "/api/tasks/1", `{"done": true, "title": "x"}`, http.StatusBadRequest},
		{"битый json", "/api/tasks/1", `{"done": tr`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			rec := doRequest(srv, "PATCH", tt.path, tt.body, nil)
			if rec.Code != tt.wantCode {
				t.Errorf("код %d, ожидали %d: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantTime string
	}{
		{"обычная", `{"title": "Купить хлеб", "time": "10:30", "kind": "other"}`, http.StatusCreated, "10:30"},
		{"час одной цифрой", `{"title": "Зарядка", "time": "9:05"}`, http.StatusCreated, "09:05"},
		{"пустое название", `{"title": " ", "time": "10:00"}`, http.StatusBadRequest, ""},
		{"201 символ", `{"title": "` + strings.Repeat("я", 201) + `", "time": "10:00"}`, http.StatusBadRequest, ""},
		{"25 часов", `{"title": "Дело", "time": "25:00"}`, http.StatusBadRequest, ""},
		{"неизвестный вид", `{"title": "Дело", "time": "10:00", "kind": "party"}`, http.StatusBadRequest, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			rec := doRequest(srv, "POST", "/api/tasks", tt.body, nil)
			if rec.Code != tt.wantCode {
				t.Fatalf("код %d: %s", rec.Code, rec.Body.String())
			}
			if tt.wantCode == http.StatusCreated {
				var task Task
				json.Unmarshal(rec.Body.Bytes(), &task)
				if task.ID == 0 || task.Time != tt.wantTime {
					t.Errorf("%+v", task)
				}
				list := doRequest(srv, "GET", "/api/tasks/today", "", nil)
				if !strings.Contains(list.Body.String(), `"time":"`+tt.wantTime+`"`) {
					t.Error("новая задача не попала в список")
				}
			}
		})
	}
}

func TestAsk(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantType string
	}{
		{"аптека", `{"text": "где аптека"}`, http.StatusOK, "feature"},
		{"лекарство", `{"text": "где купить нурафен"}`, http.StatusOK, "medicine"},
		{"товар", `{"text": "нужно молоко"}`, http.StatusOK, "product"},
		{"непонятно без ИИ", `{"text": "привет"}`, http.StatusOK, "unknown"},
		{"пустой", `{"text": "  "}`, http.StatusBadRequest, ""},
		{"501 символ", `{"text": "` + strings.Repeat("я", 501) + `"}`, http.StatusBadRequest, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			rec := doRequest(srv, "POST", "/api/ask", tt.body, nil)
			if rec.Code != tt.wantCode {
				t.Fatalf("код %d: %s", rec.Code, rec.Body.String())
			}
			var answer AskAnswer
			json.Unmarshal(rec.Body.Bytes(), &answer)
			if tt.wantType != "" && answer.Type != tt.wantType {
				t.Errorf("type = %q", answer.Type)
			}
		})
	}
}

func TestAskFallsBackToAI(t *testing.T) {
	srv, _ := newTestServer(t)
	ai := &fakeAI{answer: `{"type": "doctor", "target": "neurologist"}`}
	srv.ai = ai

	rec := doRequest(srv, "POST", "/api/ask", `{"text": "голова кружится по утрам"}`, nil)
	var answer AskAnswer
	json.Unmarshal(rec.Body.Bytes(), &answer)
	if answer.Type != "doctor" || answer.Specialty == nil || answer.Specialty.ID != "neurologist" || answer.Button == "" {
		t.Errorf("ИИ должен был понять запрос: %s", rec.Body.String())
	}

	doRequest(srv, "POST", "/api/ask", `{"text": "где аптека"}`, nil)
	if ai.calls != 1 {
		t.Errorf("понятные запросы не должны идти в ИИ, вызовов %d", ai.calls)
	}

	srv.ai = &fakeAI{fail: true}
	rec = doRequest(srv, "POST", "/api/ask", `{"text": "расскажи анекдот"}`, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"unknown"`) {
		t.Errorf("если ИИ упал, отвечаем «не поняли»: %s", rec.Body.String())
	}
}

func TestProfileHandlers(t *testing.T) {
	srv, _ := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}

	body := `{"name": " Анна ", "birthDate": "1956-03-12", "address": "Тверская, 7", "lat": 55.757, "lon": 37.613,
		"hasLocation": true, "contactName": "Дочь", "contactPhone": "+7 (999) 123-45-67", "health": "аллергия"}`
	if rec := doRequest(srv, "PUT", "/api/profile", body, h); rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}

	rec := doRequest(srv, "GET", "/api/profile", "", h)
	var p Profile
	json.Unmarshal(rec.Body.Bytes(), &p)
	if p.Name != "Анна" || p.BirthDate != "1956-03-12" || !p.HasLocation {
		t.Errorf("%+v", p)
	}
	if strings.Contains(doRequest(srv, "GET", "/api/profile", "", nil).Body.String(), "Анна") {
		t.Error("профиль виден другому пользователю")
	}
}

func TestProfileValidation(t *testing.T) {
	for _, body := range []string{
		`{"birthDate": "2027-01-01"}`,
		`{"birthDate": "1956-02-30"}`,
		`{"birthDate": "12.03.1956"}`,
		`{"age": 70}`,
		`{"address": "Тверская 7"}`,
		`{"contactPhone": "12345"}`,
		`{"name":`,
	} {
		srv, _ := newTestServer(t)
		if rec := doRequest(srv, "PUT", "/api/profile", body, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: код %d", body, rec.Code)
		}
	}
}

func TestSaveAddressHandler(t *testing.T) {
	srv, _ := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}
	doRequest(srv, "PUT", "/api/profile", `{"name": "Анна"}`, h)

	rec := doRequest(srv, "PUT", "/api/profile/address", `{"address": "Мясницкая, 20", "lat": 55.76, "lon": 37.63}`, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/pharmacies/nearby", "", h)
	if !strings.Contains(rec.Body.String(), "Мясницкая, 20") {
		t.Errorf("аптеки должны искаться от нового адреса сразу: %s", rec.Body.String())
	}
	if !strings.Contains(doRequest(srv, "GET", "/api/profile", "", h).Body.String(), "Анна") {
		t.Error("сохранение адреса стёрло имя")
	}

	for _, bad := range []string{
		`{"address": "Мясницкая"}`,
		`{"address": "x", "lat": 91, "lon": 37}`,
		`{"address": "x", "lat": 55, "lon": 37, "extra": 1}`,
	} {
		if rec := doRequest(srv, "PUT", "/api/profile/address", bad, h); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: код %d", bad, rec.Code)
		}
	}

	rec = doRequest(srv, "PUT", "/api/profile/address", `{"address": ""}`, h)
	var loc Location
	json.Unmarshal(rec.Body.Bytes(), &loc)
	if rec.Code != http.StatusOK || !loc.IsDefault {
		t.Errorf("пустой адрес стирает его: %s", rec.Body.String())
	}
}

func TestGeoHandlers(t *testing.T) {
	tests := []struct {
		path     string
		fail     bool
		wantCode int
	}{
		{"/api/geo/search?q=" + url.QueryEscape("Тверская 7"), false, http.StatusOK},
		{"/api/geo/search?q=" + url.QueryEscape("Тв"), false, http.StatusBadRequest},
		{"/api/geo/search?q=" + url.QueryEscape("Тверская 7"), true, http.StatusBadGateway},
		{"/api/geo/reverse?lat=55.75&lon=37.61", false, http.StatusOK},
		{"/api/geo/reverse?lat=0&lon=0", false, http.StatusNotFound},
		{"/api/geo/reverse?lat=95&lon=37.61", false, http.StatusBadRequest},
	}
	for _, tt := range tests {
		srv, _ := newTestServer(t)
		srv.geocoder = &fakeGeocoder{fail: tt.fail}
		if rec := doRequest(srv, "GET", tt.path, "", nil); rec.Code != tt.wantCode {
			t.Errorf("%s: код %d, ожидали %d", tt.path, rec.Code, tt.wantCode)
		}
	}
}

func TestSuggestHandlers(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(srv, "GET", "/api/medicines/suggest?q="+url.QueryEscape("парацетомол"), "", nil)
	var list []Suggestion
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) == 0 || list[0].ID != "paracetamol" || !list[0].Corrected || list[0].Form == "" {
		t.Errorf("%s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/products/suggest?q="+url.QueryEscape("малако"), "", nil)
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) == 0 || list[0].ID != "milk" || list[0].Unit != "1 л" {
		t.Errorf("%s", rec.Body.String())
	}

	if rec := doRequest(srv, "GET", "/api/products/suggest", "", nil); rec.Body.String() != "[]\n" {
		t.Errorf("пустой запрос: %s", rec.Body.String())
	}
	if rec := doRequest(srv, "GET", "/api/products/suggest?q="+strings.Repeat("a", 201), "", nil); rec.Code != http.StatusBadRequest {
		t.Errorf("длинный запрос: %d", rec.Code)
	}
}

func TestSuggestUsesAIOnlyWhenNothingFound(t *testing.T) {
	srv, _ := newTestServer(t)
	ai := &fakeAI{answer: "Молоко 2,5%"}
	srv.ai = ai

	doRequest(srv, "GET", "/api/products/suggest?q="+url.QueryEscape("гречка"), "", nil)
	doRequest(srv, "GET", "/api/products/suggest?q="+url.QueryEscape("мол"), "", nil)
	if ai.calls != 0 {
		t.Fatal("если нашли сами, ИИ не вызываем")
	}

	rec := doRequest(srv, "GET", "/api/products/suggest?q="+url.QueryEscape("белое коровье питьё"), "", nil)
	var list []Suggestion
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != "milk" || !list[0].Corrected {
		t.Errorf("ИИ должен был подсказать молоко: %s", rec.Body.String())
	}
}

func TestPlacesHandlers(t *testing.T) {
	srv, store := newTestServer(t)
	pharmacy := addPlace(t, store, Place{Kind: "pharmacy", Name: "Ригла", Lat: 55.755, Lon: 37.622, Rating: 4.8, Reviews: 100})
	shop := addPlace(t, store, Place{Kind: "shop", Name: "Пятёрочка", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 100, BirthdayDiscount: 10})
	addMedicinePrice(t, store, pharmacy.ID, "ibuprofen", 70)
	addProductPrice(t, store, shop.ID, "milk", 100, 80, "2026-10-03")

	rec := doRequest(srv, "GET", "/api/pharmacies/nearby", "", nil)
	if !strings.Contains(rec.Body.String(), "Ригла") || !strings.Contains(rec.Body.String(), `"isDefault":true`) {
		t.Errorf("аптеки: %s", rec.Body.String())
	}
	rec = doRequest(srv, "GET", "/api/shops/nearby", "", nil)
	if !strings.Contains(rec.Body.String(), "Пятёрочка") || strings.Contains(rec.Body.String(), "Ригла") {
		t.Errorf("магазины: %s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/medicines/ibuprofen/offers", "", nil)
	if !strings.Contains(rec.Body.String(), `"price":70`) || !strings.Contains(rec.Body.String(), `"cheapest":true`) {
		t.Errorf("цены лекарства: %s", rec.Body.String())
	}
	rec = doRequest(srv, "GET", "/api/products/milk/offers", "", nil)
	if !strings.Contains(rec.Body.String(), `"price":80`) || !strings.Contains(rec.Body.String(), `"oldPrice":100`) {
		t.Errorf("цены товара: %s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/shops/"+itoa(shop.ID), "", nil)
	if !strings.Contains(rec.Body.String(), `"discount":20`) || !strings.Contains(rec.Body.String(), `"distanceKm"`) {
		t.Errorf("страница магазина: %s", rec.Body.String())
	}

	for path, code := range map[string]int{
		"/api/medicines/nope/offers":      http.StatusNotFound,
		"/api/products/nope/offers":       http.StatusNotFound,
		"/api/shops/abc":                  http.StatusBadRequest,
		"/api/shops/0":                    http.StatusBadRequest,
		"/api/shops/99999":                http.StatusNotFound,
		"/api/shops/" + itoa(pharmacy.ID): http.StatusNotFound,
	} {
		if rec := doRequest(srv, "GET", path, "", nil); rec.Code != code {
			t.Errorf("%s: код %d, ожидали %d", path, rec.Code, code)
		}
	}
}

func TestForYouHandler(t *testing.T) {
	srv, store := newTestServer(t)
	addPlace(t, store, Place{Kind: "shop", Name: "Скидка ДР", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 100, BirthdayDiscount: 10})
	addPlace(t, store, Place{Kind: "shop", Name: "Без скидки", Lat: 55.755, Lon: 37.622, Rating: 4.9, Reviews: 100})
	h := map[string]string{"X-User-Id": "u1"}

	rec := doRequest(srv, "GET", "/api/for-you", "", h)
	if rec.Body.String() != "[]\n" {
		t.Errorf("без даты рождения ничего нет: %s", rec.Body.String())
	}

	doRequest(srv, "PUT", "/api/profile", `{"birthDate": "1981-10-05"}`, h)
	rec = doRequest(srv, "GET", "/api/for-you", "", h)
	var items []ForYouItem
	json.Unmarshal(rec.Body.Bytes(), &items)
	if len(items) != 2 || items[0].Type != "birthday" || items[1].Type != "passport" {
		t.Fatalf("ждём день рождения и паспорт: %s", rec.Body.String())
	}
	if len(items[0].Places) != 1 || items[0].Places[0].Name != "Скидка ДР" {
		t.Errorf("показываем только места со скидкой именинникам: %+v", items[0].Places)
	}
}

func TestWrongMethodAndPath(t *testing.T) {
	srv, _ := newTestServer(t)
	if rec := doRequest(srv, "DELETE", "/api/tasks/1", "", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE: %d", rec.Code)
	}
	if rec := doRequest(srv, "GET", "/api/unknown", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("неизвестный путь: %d", rec.Code)
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
