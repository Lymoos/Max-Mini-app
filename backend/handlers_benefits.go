package main

import (
	"context"
	"net/http"
	"time"
)

const staleDays = 30

var benefitStatuses = []string{"not_started", "collecting", "submitted", "review", "approved", "rejected"}

type BenefitState struct {
	Status      string `json:"status"`
	SubmittedAt string `json:"submittedAt,omitempty"`
	Docs        []int  `json:"docs"`
}

func (s *Store) BenefitStates(ctx context.Context, userID string) (map[string]BenefitState, error) {
	rows, err := s.db.Query(ctx,
		`SELECT benefit_id, status, to_char(submitted_at, 'YYYY-MM-DD'), docs FROM user_benefits WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := map[string]BenefitState{}
	for rows.Next() {
		var id string
		var st BenefitState
		var submitted *string
		var docs []int32
		if err := rows.Scan(&id, &st.Status, &submitted, &docs); err != nil {
			return nil, err
		}
		if submitted != nil {
			st.SubmittedAt = *submitted
		}
		st.Docs = []int{}
		for _, d := range docs {
			st.Docs = append(st.Docs, int(d))
		}
		states[id] = st
	}
	return states, rows.Err()
}

func (s *Store) SaveBenefitState(ctx context.Context, userID, benefitID string, st BenefitState) error {
	var submitted any
	if st.SubmittedAt != "" {
		submitted = st.SubmittedAt
	}
	docs := []int32{}
	for _, d := range st.Docs {
		docs = append(docs, int32(d))
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO user_benefits (user_id, benefit_id, status, submitted_at, docs, updated_at)
		 VALUES ($1, $2, $3, $4, $5, now())
		 ON CONFLICT (user_id, benefit_id) DO UPDATE SET status = EXCLUDED.status,
		   submitted_at = EXCLUDED.submitted_at, docs = EXCLUDED.docs, updated_at = now()`,
		userID, benefitID, st.Status, submitted, docs,
	)
	return err
}

func (s *Store) BenefitState(ctx context.Context, userID, benefitID string) (BenefitState, error) {
	states, err := s.BenefitStates(ctx, userID)
	if err != nil {
		return BenefitState{}, err
	}
	st, ok := states[benefitID]
	if !ok {
		return BenefitState{Status: "not_started", Docs: []int{}}, nil
	}
	return st, nil
}

// подача без ответа дольше месяца — стоит проверить статус на Госуслугах
func isStale(st BenefitState, today time.Time) bool {
	if st.Status != "submitted" && st.Status != "review" {
		return false
	}
	d, err := time.Parse("2006-01-02", st.SubmittedAt)
	return err == nil && daysBetween(d, today) > staleDays
}

type benefitContext struct {
	birth  *time.Time
	region string
	today  time.Time
}

func (s *Server) benefitContext(ctx context.Context, r *http.Request) (benefitContext, error) {
	p, err := s.store.GetProfile(ctx, userID(r))
	if err != nil {
		return benefitContext{}, err
	}
	bc := benefitContext{region: regionOf(p.Location()), today: s.now()}
	if d, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
		bc.birth = &d
	}
	return bc, nil
}

func (s *Server) listBenefits(w http.ResponseWriter, r *http.Request) {
	bc, err := s.benefitContext(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	states, err := s.store.BenefitStates(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}

	type item struct {
		ID          string `json:"id"`
		Category    string `json:"category"`
		Title       string `json:"title"`
		Short       string `json:"short"`
		Auto        bool   `json:"auto"`
		Deadline    string `json:"deadline,omitempty"`
		Fit         bool   `json:"fit"`
		FitReason   string `json:"fitReason,omitempty"`
		Status      string `json:"status"`
		SubmittedAt string `json:"submittedAt,omitempty"`
		Stale       bool   `json:"stale"`
		DocsDone    int    `json:"docsDone"`
		DocsTotal   int    `json:"docsTotal"`
	}

	items := []item{}
	summary := map[string]int{"inProgress": 0, "approved": 0, "fit": 0}
	for _, b := range benefits {
		if b.Region != "" && b.Region != bc.region {
			continue
		}
		st, ok := states[b.ID]
		if !ok {
			st = BenefitState{Status: "not_started"}
		}
		fit, reason := benefitFit(b, bc.birth, bc.region, bc.today)

		items = append(items, item{
			ID: b.ID, Category: b.Category, Title: b.Title, Short: b.Short, Auto: b.Auto, Deadline: b.Deadline,
			Fit: fit, FitReason: reason, Status: st.Status, SubmittedAt: st.SubmittedAt,
			Stale: isStale(st, bc.today), DocsDone: len(st.Docs), DocsTotal: len(b.Documents),
		})
		switch st.Status {
		case "collecting", "submitted", "review":
			summary["inProgress"]++
		case "approved":
			summary["approved"]++
		}
		if fit {
			summary["fit"]++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": benefitCategories, "summary": summary, "items": items})
}

func (s *Server) benefitDetails(w http.ResponseWriter, r *http.Request) {
	b, ok := findBenefit(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Такой льготы нет")
		return
	}
	bc, err := s.benefitContext(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	st, err := s.store.BenefitState(r.Context(), userID(r), b.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	fit, reason := benefitFit(b, bc.birth, bc.region, bc.today)
	writeJSON(w, http.StatusOK, map[string]any{
		"benefit": b, "fit": fit, "fitReason": reason, "state": st, "stale": isStale(st, bc.today),
	})
}

func (s *Server) saveBenefit(w http.ResponseWriter, r *http.Request) {
	b, ok := findBenefit(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Такой льготы нет")
		return
	}

	var body struct {
		Status      *string `json:"status"`
		SubmittedAt *string `json:"submittedAt"`
		Docs        *[]int  `json:"docs"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	st, err := s.store.BenefitState(r.Context(), userID(r), b.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	if body.Status != nil {
		if b.Auto {
			writeError(w, http.StatusBadRequest, "Эта выплата назначается автоматически, подавать ничего не нужно")
			return
		}
		if !contains(benefitStatuses, *body.Status) {
			writeError(w, http.StatusBadRequest, "Неизвестный статус")
			return
		}
		st.Status = *body.Status
		if st.Status == "not_started" || st.Status == "collecting" {
			st.SubmittedAt = ""
		} else if st.SubmittedAt == "" {
			st.SubmittedAt = s.today()
		}
	}

	if body.SubmittedAt != nil {
		d, err := time.Parse("2006-01-02", *body.SubmittedAt)
		today, _ := time.Parse("2006-01-02", s.today())
		if err != nil || d.After(today) || d.Before(today.AddDate(-2, 0, 0)) {
			writeError(w, http.StatusBadRequest, "Проверьте дату подачи")
			return
		}
		if st.Status == "not_started" || st.Status == "collecting" {
			writeError(w, http.StatusBadRequest, "Дату подачи можно указать только после подачи")
			return
		}
		st.SubmittedAt = *body.SubmittedAt
	}

	if body.Docs != nil {
		seen := map[int]bool{}
		docs := []int{}
		for _, d := range *body.Docs {
			if d < 0 || d >= len(b.Documents) {
				writeError(w, http.StatusBadRequest, "Неверный номер документа")
				return
			}
			if !seen[d] {
				seen[d] = true
				docs = append(docs, d)
			}
		}
		st.Docs = docs
	}

	if err := s.store.SaveBenefitState(r.Context(), userID(r), b.ID, st); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}
