package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type clinicsResponse struct {
	HasRegistration bool          `json:"hasRegistration"`
	RegAddress      string        `json:"regAddress"`
	MyClinic        *RankedPlace  `json:"myClinic"`
	MyClinicChosen  bool          `json:"myClinicChosen"`
	Nearby          []RankedPlace `json:"nearby"`
}

func getClinics(t *testing.T, srv *Server, h map[string]string) clinicsResponse {
	t.Helper()
	rec := doRequest(srv, "GET", "/api/clinics", "", h)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var resp clinicsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	return resp
}

func TestClinicsFlow(t *testing.T) {
	srv, store := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}
	home := addPlace(t, store, Place{Kind: "clinic", Name: "Городская поликлиника № 2", IsState: true, Lat: 55.765, Lon: 37.640, Rating: 3.9, Reviews: 50})
	near := addPlace(t, store, Place{Kind: "clinic", Name: "Медси", Lat: 55.754, Lon: 37.621, Rating: 4.8, Reviews: 300})
	kids := addPlace(t, store, Place{Kind: "clinic", Name: "Детская поликлиника № 1", IsState: true, Lat: 55.7655, Lon: 37.6405, Rating: 4.5, Reviews: 10})
	addPlace(t, store, Place{Kind: "pharmacy", Name: "Аптека", Lat: 55.754, Lon: 37.621, Rating: 5, Reviews: 10})

	resp := getClinics(t, srv, h)
	if resp.HasRegistration || resp.MyClinic != nil {
		t.Fatalf("без прописки своей поликлиники нет: %+v", resp)
	}
	if len(resp.Nearby) != 2 || resp.Nearby[0].ID != near.ID {
		t.Errorf("рядом две взрослые клиники, лучшая — Медси, детская не показывается: %+v", resp.Nearby)
	}

	rec := doRequest(srv, "PUT", "/api/profile/registration", `{"address": "Мясницкая, 20", "lat": 55.765, "lon": 37.6401}`, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}

	resp = getClinics(t, srv, h)
	if !resp.HasRegistration || resp.RegAddress != "Мясницкая, 20" {
		t.Errorf("прописка не сохранилась: %+v", resp)
	}
	if resp.MyClinic == nil || resp.MyClinic.ID != home.ID || resp.MyClinicChosen {
		t.Fatalf("своя — ближайшая взрослая поликлиника к прописке, а не детская: %+v", resp.MyClinic)
	}
	for _, c := range resp.Nearby {
		if c.ID == home.ID {
			t.Error("своя поликлиника не должна повторяться в списке рядом")
		}
	}

	rec = doRequest(srv, "PUT", "/api/profile/clinic", `{"clinicId": `+itoa(kids.ID)+`}`, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	resp = getClinics(t, srv, h)
	if resp.MyClinic == nil || resp.MyClinic.ID != kids.ID || !resp.MyClinicChosen {
		t.Errorf("выбранная вручную поликлиника важнее автоматической: %+v", resp.MyClinic)
	}

	doRequest(srv, "PUT", "/api/profile/clinic", `{"clinicId": 0}`, h)
	resp = getClinics(t, srv, h)
	if resp.MyClinic == nil || resp.MyClinic.ID != home.ID || resp.MyClinicChosen {
		t.Errorf("после сброса снова по прописке: %+v", resp.MyClinic)
	}
}

func TestRegistrationAndClinicValidation(t *testing.T) {
	srv, store := newTestServer(t)
	pharmacy := addPlace(t, store, Place{Kind: "pharmacy", Name: "Аптека", Lat: 55.75, Lon: 37.62, Rating: 5, Reviews: 1})

	for path, body := range map[string]string{
		"/api/profile/registration": `{"address": "Мясницкая"}`,
		"/api/profile/clinic":       `{"clinicId": -1}`,
	} {
		if rec := doRequest(srv, "PUT", path, body, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d", path, body, rec.Code)
		}
	}
	if rec := doRequest(srv, "PUT", "/api/profile/registration", `{"address": "x", "lat": 91, "lon": 0}`, nil); rec.Code != http.StatusBadRequest {
		t.Errorf("широта 91: %d", rec.Code)
	}
	if rec := doRequest(srv, "PUT", "/api/profile/clinic", `{"clinicId": `+itoa(pharmacy.ID)+`}`, nil); rec.Code != http.StatusNotFound {
		t.Errorf("аптека — не поликлиника: %d", rec.Code)
	}
	if rec := doRequest(srv, "PUT", "/api/profile/clinic", `{"clinicId": 99999}`, nil); rec.Code != http.StatusNotFound {
		t.Errorf("нет такой: %d", rec.Code)
	}
	if rec := doRequest(srv, "PUT", "/api/profile/registration", `{"address": ""}`, nil); rec.Code != http.StatusOK {
		t.Errorf("пустой адрес стирает прописку: %d", rec.Code)
	}
}

func TestClinicAndDoctorDetails(t *testing.T) {
	srv, store := newTestServer(t)
	clinic := addPlace(t, store, Place{Kind: "clinic", Name: "ГП №5", Lat: 55.755, Lon: 37.622, Rating: 4, Reviews: 1})
	addDoctor(t, store, Doctor{ClinicID: clinic.ID, Name: "Иванова Анна Сергеевна", Specialty: "Терапевт", Experience: 12, Rating: 4.7, Reviews: 10})
	addDoctor(t, store, Doctor{ClinicID: clinic.ID, Name: "Петров Олег Иванович", Specialty: "Терапевт", Experience: 30, Rating: 4.9, Reviews: 10})
	surgeon := addDoctor(t, store, Doctor{ClinicID: clinic.ID, Name: "Сафин Рустам Ренатович", Specialty: "Хирург", Experience: 20, Category: "Высшая категория", Rating: 4.8, Reviews: 10})

	rec := doRequest(srv, "GET", "/api/clinics/"+itoa(clinic.ID), "", nil)
	var resp struct {
		Clinic      RankedPlace `json:"clinic"`
		Doctors     []Doctor    `json:"doctors"`
		Specialties []string    `json:"specialties"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Doctors) == 3 && (resp.Doctors[0].SpecialtyID != "therapist" || resp.Doctors[2].SpecialtyID != "surgeon") {
		t.Errorf("у врача должен быть id специальности для фильтра из поиска: %+v", resp.Doctors)
	}
	if len(resp.Doctors) != 3 || strings.Join(resp.Specialties, ",") != "Терапевт,Хирург" || resp.Clinic.DistanceKm == 0 {
		t.Errorf("%s", rec.Body.String())
	}

	rec = doRequest(srv, "GET", "/api/doctors/"+itoa(surgeon.ID), "", nil)
	var doc struct {
		Doctor  Doctor      `json:"doctor"`
		Clinic  RankedPlace `json:"clinic"`
		Booking Booking     `json:"booking"`
	}
	json.Unmarshal(rec.Body.Bytes(), &doc)
	if doc.Doctor.Name != "Сафин Рустам Ренатович" || doc.Clinic.Name != "ГП №5" || doc.Booking.Region != "moscow" {
		t.Errorf("%s", rec.Body.String())
	}

	h := map[string]string{"X-User-Id": "kazan"}
	doRequest(srv, "PUT", "/api/profile/registration", `{"address": "Казань, Баумана, 1", "lat": 55.79, "lon": 49.12}`, h)
	rec = doRequest(srv, "GET", "/api/doctors/"+itoa(surgeon.ID), "", h)
	json.Unmarshal(rec.Body.Bytes(), &doc)
	if doc.Booking.Region != "tatarstan" {
		t.Errorf("запись идёт по региону прописки: %+v", doc.Booking)
	}

	for path, code := range map[string]int{
		"/api/clinics/abc":   http.StatusBadRequest,
		"/api/clinics/99999": http.StatusNotFound,
		"/api/doctors/0":     http.StatusBadRequest,
		"/api/doctors/99999": http.StatusNotFound,
	} {
		if rec := doRequest(srv, "GET", path, "", nil); rec.Code != code {
			t.Errorf("%s: %d, ожидали %d", path, rec.Code, code)
		}
	}
}

func TestSocialNearby(t *testing.T) {
	srv, store := newTestServer(t)
	addPlace(t, store, Place{Kind: "social", Name: "ЦСО Мещанский", Phone: "+7 495 123-45-67", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 20})
	addPlace(t, store, Place{Kind: "clinic", Name: "ГП", Lat: 55.755, Lon: 37.622, Rating: 4, Reviews: 1})

	rec := doRequest(srv, "GET", "/api/social/nearby", "", nil)
	if !strings.Contains(rec.Body.String(), `"phone":"+7 495 123-45-67"`) || strings.Contains(rec.Body.String(), `"ГП"`) {
		t.Errorf("%s", rec.Body.String())
	}
}

func TestCreateTaskWithDate(t *testing.T) {
	tests := []struct {
		date     string
		wantCode int
	}{
		{"2026-10-02", http.StatusCreated},
		{"2026-09-30", http.StatusCreated},
		{"2027-09-30", http.StatusCreated},
		{"2026-09-29", http.StatusBadRequest},
		{"2027-10-01", http.StatusBadRequest},
		{"02.10.2026", http.StatusBadRequest},
	}
	for _, tt := range tests {
		srv, _ := newTestServer(t)
		rec := doRequest(srv, "POST", "/api/tasks", `{"title": "Приём у терапевта", "time": "10:00", "kind": "doctor", "date": "`+tt.date+`"}`, nil)
		if rec.Code != tt.wantCode {
			t.Errorf("%s: %d %s", tt.date, rec.Code, rec.Body.String())
		}
	}

	srv, store := newTestServer(t)
	doRequest(srv, "POST", "/api/tasks", `{"title": "Приём", "time": "10:00", "kind": "doctor", "date": "2026-10-02"}`, nil)
	today := doRequest(srv, "GET", "/api/tasks/today", "", nil)
	if strings.Contains(today.Body.String(), `"Приём"`) {
		t.Error("задача на другой день не должна попасть в сегодня")
	}
	later, _ := store.TasksByDate(context.Background(), "demo", "2026-10-02")
	if len(later) != 1 {
		t.Error("задача должна сохраниться на день приёма")
	}
}
