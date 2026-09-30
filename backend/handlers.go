package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultUserID = "demo"
	maxAskLength  = 500
	maxTitleLen   = 200
	maxBodySize   = 1 << 20
)

type Server struct {
	store    *Store
	geocoder Geocoder
	ai       AI
	now      func() time.Time
}

func NewServer(store *Store, geocoder Geocoder, ai AI) *Server {
	return &Server{store: store, geocoder: geocoder, ai: ai, now: time.Now}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/tasks/today", s.todayTasks)
	mux.HandleFunc("POST /api/tasks", s.createTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.updateTask)
	mux.HandleFunc("GET /api/features", s.listFeatures)
	mux.HandleFunc("POST /api/ask", s.ask)
	mux.HandleFunc("GET /api/profile", s.getProfile)
	mux.HandleFunc("PUT /api/profile", s.saveProfile)
	mux.HandleFunc("PUT /api/profile/address", s.saveAddress)
	mux.HandleFunc("GET /api/for-you", s.forYou)
	mux.HandleFunc("GET /api/geo/search", s.geoSearch)
	mux.HandleFunc("GET /api/geo/reverse", s.geoReverse)
	mux.HandleFunc("GET /api/medicines/suggest", s.medicineSuggest)
	mux.HandleFunc("GET /api/medicines/{id}/offers", s.medicineOffers)
	mux.HandleFunc("GET /api/pharmacies/nearby", s.nearbyPharmacies)
	mux.HandleFunc("GET /api/products/suggest", s.productSuggest)
	mux.HandleFunc("GET /api/products/{id}/offers", s.productOffers)
	mux.HandleFunc("GET /api/shops/nearby", s.nearbyShops)
	mux.HandleFunc("GET /api/shops/{id}", s.shopDetails)
	mux.HandleFunc("PUT /api/profile/registration", s.saveRegistration)
	mux.HandleFunc("PUT /api/profile/clinic", s.saveClinic)
	mux.HandleFunc("GET /api/clinics", s.listClinics)
	mux.HandleFunc("GET /api/clinics/{id}", s.clinicDetails)
	mux.HandleFunc("GET /api/doctors/{id}", s.doctorDetails)
	mux.HandleFunc("GET /api/social/nearby", s.nearbySocial)
	mux.HandleFunc("GET /api/benefits", s.listBenefits)
	mux.HandleFunc("GET /api/benefits/{id}", s.benefitDetails)
	mux.HandleFunc("PUT /api/benefits/{id}", s.saveBenefit)
	mux.HandleFunc("GET /api/guides", s.listGuides)
	mux.HandleFunc("GET /api/guides/{id}", s.guideDetails)
	mux.HandleFunc("PUT /api/guides/{id}/read", s.markGuide)
	return mux
}

func userID(r *http.Request) string {
	id := strings.TrimSpace(r.Header.Get("X-User-Id"))
	if id == "" {
		return defaultUserID
	}
	return id
}

func (s *Server) today() string {
	return s.now().Format("2006-01-02")
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	writeError(w, http.StatusInternalServerError, "Что-то пошло не так. Попробуйте ещё раз")
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) todayTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.store.TasksByDate(r.Context(), userID(r), s.today())
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": s.today(), "tasks": tasks})
}

var taskKinds = map[string]bool{
	"medicine": true,
	"doctor":   true,
	"call":     true,
	"other":    true,
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
		Time  string `json:"time"`
		Kind  string `json:"kind"`
		Date  string `json:"date"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "Напишите, что нужно сделать")
		return
	}
	if utf8.RuneCountInString(title) > maxTitleLen {
		writeError(w, http.StatusBadRequest, "Слишком длинное название")
		return
	}

	t, err := time.Parse("15:04", strings.TrimSpace(body.Time))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Время должно быть в формате ЧЧ:ММ")
		return
	}

	kind := body.Kind
	if kind == "" {
		kind = "other"
	}
	if !taskKinds[kind] {
		writeError(w, http.StatusBadRequest, "Неизвестный вид задачи")
		return
	}

	// дату можно не передавать — тогда задача на сегодня. Запись к врачу ставится на день приёма
	date := s.today()
	if body.Date != "" {
		d, err := time.Parse("2006-01-02", body.Date)
		today, _ := time.Parse("2006-01-02", s.today())
		if err != nil || d.Before(today) || d.After(today.AddDate(1, 0, 0)) {
			writeError(w, http.StatusBadRequest, "Дата должна быть не раньше сегодня и не дальше чем через год")
			return
		}
		date = body.Date
	}

	task, err := s.store.AddTask(r.Context(), userID(r), date, Task{Time: t.Format("15:04"), Title: title, Kind: kind})
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Неверный id задачи")
		return
	}

	var body struct {
		Done *bool `json:"done"`
	}
	if err := decodeBody(w, r, &body); err != nil || body.Done == nil {
		writeError(w, http.StatusBadRequest, "Ожидается {\"done\": true или false}")
		return
	}

	task, err := s.store.SetDone(r.Context(), userID(r), id, *body.Done)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Задача не найдена")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) listFeatures(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, features)
}

func (s *Server) ask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	text := strings.TrimSpace(body.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, "Пустой запрос")
		return
	}
	if utf8.RuneCountInString(text) > maxAskLength {
		writeError(w, http.StatusBadRequest, "Запрос слишком длинный")
		return
	}

	answer := answerQuestion(text)
	if answer.Type == "unknown" {
		if aiAnswer, ok := s.aiAnswer(r.Context(), text); ok {
			answer = aiAnswer
		}
	}
	writeJSON(w, http.StatusOK, answer)
}
