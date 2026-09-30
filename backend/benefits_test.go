package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBenefitsCatalog(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range benefits {
		if b.ID == "" || b.Title == "" || b.Short == "" || b.Who == "" || len(b.Steps) == 0 || len(b.Links) == 0 || seen[b.ID] {
			t.Errorf("плохая льгота %+v", b)
		}
		seen[b.ID] = true
		if !contains(benefitCategories, b.Category) {
			t.Errorf("%s: неизвестная категория %q", b.ID, b.Category)
		}
		if b.Region != "" && b.Region != "moscow" && b.Region != "tatarstan" {
			t.Errorf("%s: неизвестный регион", b.ID)
		}
		if !b.Auto && len(b.Documents) == 0 {
			t.Errorf("%s: для подачи нужен список документов", b.ID)
		}
		for _, l := range b.Links {
			if !strings.HasPrefix(l.URL, "https://") || !(strings.Contains(l.URL, "gosuslugi.ru") || strings.Contains(l.URL, "sfr.gov.ru") ||
				strings.Contains(l.URL, "mos.ru") || strings.Contains(l.URL, "nalog.gov.ru")) {
				t.Errorf("%s: ссылка не на официальный сайт: %s", b.ID, l.URL)
			}
		}
	}
}

func TestBenefitFit(t *testing.T) {
	today := date("2026-09-30")
	birth := func(s string) *time.Time { d := date(s); return &d }

	overhaul, _ := findBenefit("overhaul")
	if fit, _ := benefitFit(overhaul, birth("1950-01-01"), "moscow", today); !fit {
		t.Error("76 лет — компенсация за капремонт подходит")
	}
	if fit, reason := benefitFit(overhaul, birth("1957-06-01"), "moscow", today); !fit || !strings.Contains(reason, "Скоро") {
		t.Errorf("69 лет — скоро подойдёт: %v %q", fit, reason)
	}
	if fit, _ := benefitFit(overhaul, birth("1960-01-01"), "moscow", today); fit {
		t.Error("66 лет — рано")
	}
	if fit, _ := benefitFit(overhaul, nil, "moscow", today); fit {
		t.Error("без даты рождения не угадываем")
	}

	passport, _ := findBenefit("passport")
	if fit, _ := benefitFit(passport, birth("1981-10-05"), "other", today); !fit {
		t.Error("44 года, скоро 45 — паспорт")
	}
	if fit, _ := benefitFit(passport, birth("1950-10-05"), "other", today); fit {
		t.Error("76 лет — паспорт менять не нужно")
	}

	card, _ := findBenefit("moscow-card")
	if fit, _ := benefitFit(card, birth("1950-01-01"), "tatarstan", today); fit {
		t.Error("карта москвича не для Казани")
	}
}

func TestBenefitsList(t *testing.T) {
	srv, _ := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}

	rec := doRequest(srv, "GET", "/api/benefits", "", h)
	var resp struct {
		Categories []string         `json:"categories"`
		Summary    map[string]int   `json:"summary"`
		Items      []map[string]any `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if rec.Code != http.StatusOK || len(resp.Items) != len(benefits) || len(resp.Categories) != 4 {
		t.Fatalf("без адреса — Москва по умолчанию, видны все: %d %d", rec.Code, len(resp.Items))
	}
	if resp.Items[0]["status"] != "not_started" || resp.Summary["fit"] != 0 {
		t.Errorf("новый пользователь: %+v", resp)
	}

	doRequest(srv, "PUT", "/api/profile", `{"birthDate": "1950-01-01"}`, h)
	doRequest(srv, "PUT", "/api/profile/address", `{"address": "Казань", "lat": 55.79, "lon": 49.12}`, h)
	rec = doRequest(srv, "GET", "/api/benefits", "", h)
	json.Unmarshal(rec.Body.Bytes(), &resp)
	for _, it := range resp.Items {
		if it["id"] == "moscow-card" || it["id"] == "moscow-longevity" {
			t.Error("в Казани не показываем московские льготы")
		}
	}
	if resp.Summary["fit"] == 0 {
		t.Error("76-летнему что-то должно подойти")
	}
}

func TestBenefitStatusFlow(t *testing.T) {
	srv, _ := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}

	put := func(body string) (int, BenefitState) {
		rec := doRequest(srv, "PUT", "/api/benefits/overhaul", body, h)
		var st BenefitState
		json.Unmarshal(rec.Body.Bytes(), &st)
		return rec.Code, st
	}

	if code, st := put(`{"status": "collecting", "docs": [0, 2, 2]}`); code != 200 || st.Status != "collecting" || len(st.Docs) != 2 || st.SubmittedAt != "" {
		t.Fatalf("%d %+v", code, st)
	}
	if code, st := put(`{"status": "submitted"}`); code != 200 || st.SubmittedAt != "2026-09-30" || len(st.Docs) != 2 {
		t.Errorf("при подаче ставим сегодняшнюю дату и не теряем документы: %+v", st)
	}
	if code, st := put(`{"submittedAt": "2026-08-01"}`); code != 200 || st.SubmittedAt != "2026-08-01" || st.Status != "submitted" {
		t.Errorf("%d %+v", code, st)
	}

	rec := doRequest(srv, "GET", "/api/benefits", "", h)
	if !strings.Contains(rec.Body.String(), `"id":"overhaul","category":"Льготы"`) || !strings.Contains(rec.Body.String(), `"stale":true`) {
		t.Errorf("подача 2 месяца назад — пора проверить статус: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"inProgress":1`) {
		t.Errorf("в сводке одна заявка в процессе: %s", rec.Body.String())
	}

	put(`{"status": "approved"}`)
	rec = doRequest(srv, "GET", "/api/benefits/overhaul", "", h)
	var details struct {
		Benefit Benefit      `json:"benefit"`
		State   BenefitState `json:"state"`
		Stale   bool         `json:"stale"`
	}
	json.Unmarshal(rec.Body.Bytes(), &details)
	if details.State.Status != "approved" || details.Stale || details.Benefit.Title == "" || len(details.Benefit.Steps) == 0 {
		t.Errorf("%s", rec.Body.String())
	}

	if code, st := put(`{"status": "not_started"}`); code != 200 || st.SubmittedAt != "" {
		t.Errorf("сброс статуса стирает дату подачи: %+v", st)
	}

	if strings.Contains(doRequest(srv, "GET", "/api/benefits", "", nil).Body.String(), `"inProgress":1`) {
		t.Error("статусы одного пользователя не видны другому")
	}
}

func TestBenefitValidation(t *testing.T) {
	srv, _ := newTestServer(t)
	cases := []struct {
		path, body string
		code       int
	}{
		{"/api/benefits/nope", `{"status": "collecting"}`, http.StatusNotFound},
		{"/api/benefits/overhaul", `{"status": "done"}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"docs": [99]}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"docs": [-1]}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"submittedAt": "2026-09-01"}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"status": "submitted", "submittedAt": "2026-10-10"}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"status": "submitted", "submittedAt": "2020-01-01"}`, http.StatusBadRequest},
		{"/api/benefits/pension-80", `{"status": "collecting"}`, http.StatusBadRequest},
		{"/api/benefits/overhaul", `{"extra": 1}`, http.StatusBadRequest},
	}
	for _, c := range cases {
		if rec := doRequest(srv, "PUT", c.path, c.body, nil); rec.Code != c.code {
			t.Errorf("%s %s: %d, ожидали %d (%s)", c.path, c.body, rec.Code, c.code, rec.Body.String())
		}
	}
	if rec := doRequest(srv, "GET", "/api/benefits/nope", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("GET несуществующей: %d", rec.Code)
	}
}

func TestIsStale(t *testing.T) {
	today := date("2026-09-30")
	cases := []struct {
		st   BenefitState
		want bool
	}{
		{BenefitState{Status: "submitted", SubmittedAt: "2026-08-01"}, true},
		{BenefitState{Status: "review", SubmittedAt: "2026-08-31"}, false},
		{BenefitState{Status: "review", SubmittedAt: "2026-08-30"}, true},
		{BenefitState{Status: "approved", SubmittedAt: "2026-01-01"}, false},
		{BenefitState{Status: "submitted"}, false},
	}
	for _, c := range cases {
		if got := isStale(c.st, today); got != c.want {
			t.Errorf("%+v: %v", c.st, got)
		}
	}
}

func TestGuides(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range guides {
		if g.ID == "" || g.Title == "" || g.Icon == "" || seen[g.ID] {
			t.Errorf("плохая инструкция %+v", g)
		}
		seen[g.ID] = true
		if len(g.Steps) == 0 && (len(g.Android) == 0 || len(g.IPhone) == 0) {
			t.Errorf("%s: нужны общие шаги или шаги для обеих систем", g.ID)
		}
	}
	if !guides[0].Important || guides[0].ID != "scam" {
		t.Error("первой должна идти инструкция про мошенников")
	}

	srv, _ := newTestServer(t)
	h := map[string]string{"X-User-Id": "u1"}
	if rec := doRequest(srv, "PUT", "/api/guides/call/read", `{"read": true}`, h); rec.Code != http.StatusOK {
		t.Fatalf("%d", rec.Code)
	}
	doRequest(srv, "PUT", "/api/guides/call/read", `{"read": true}`, h)

	rec := doRequest(srv, "GET", "/api/guides", "", h)
	if strings.Count(rec.Body.String(), `"read":true`) != 1 {
		t.Errorf("прочитана одна инструкция: %s", rec.Body.String())
	}
	rec = doRequest(srv, "GET", "/api/guides/call", "", h)
	if !strings.Contains(rec.Body.String(), `"read":true`) || !strings.Contains(rec.Body.String(), `"iphone"`) {
		t.Errorf("%s", rec.Body.String())
	}

	doRequest(srv, "PUT", "/api/guides/call/read", `{"read": false}`, h)
	if strings.Contains(doRequest(srv, "GET", "/api/guides", "", h).Body.String(), `"read":true`) {
		t.Error("отметку можно снять")
	}

	for path, code := range map[string]int{"/api/guides/nope": 404} {
		if rec := doRequest(srv, "GET", path, "", nil); rec.Code != code {
			t.Errorf("%s: %d", path, rec.Code)
		}
	}
	if rec := doRequest(srv, "PUT", "/api/guides/nope/read", `{"read": true}`, nil); rec.Code != 404 {
		t.Errorf("%d", rec.Code)
	}
	if rec := doRequest(srv, "PUT", "/api/guides/call/read", `{}`, nil); rec.Code != 400 {
		t.Errorf("%d", rec.Code)
	}
}
