package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Task struct {
	ID    int    `json:"id"`
	Time  string `json:"time"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	Done  bool   `json:"done"`
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) AddTask(ctx context.Context, userID, date string, t Task) (Task, error) {
	err := s.db.QueryRow(ctx,
		`INSERT INTO tasks (user_id, date, time, title, kind) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, date, t.Time, t.Title, t.Kind,
	).Scan(&t.ID)
	return t, err
}

func (s *Store) TasksByDate(ctx context.Context, userID, date string) ([]Task, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, time, title, kind, done FROM tasks WHERE user_id = $1 AND date = $2 ORDER BY time, id`,
		userID, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Time, &t.Title, &t.Kind, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) CountTasks(ctx context.Context, userID, date string) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM tasks WHERE user_id = $1 AND date = $2`, userID, date).Scan(&n)
	return n, err
}

func (s *Store) SetDone(ctx context.Context, userID string, id int, done bool) (Task, error) {
	var t Task
	err := s.db.QueryRow(ctx,
		`UPDATE tasks SET done = $3 WHERE id = $1 AND user_id = $2 RETURNING id, time, title, kind, done`,
		id, userID, done,
	).Scan(&t.ID, &t.Time, &t.Title, &t.Kind, &t.Done)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (s *Store) GetProfile(ctx context.Context, userID string) (Profile, error) {
	var p Profile
	var birth *string
	var lat, lon, regLat, regLon *float64
	var clinicID *int64
	err := s.db.QueryRow(ctx,
		`SELECT name, to_char(birth_date, 'YYYY-MM-DD'), address, lat, lon, contact_name, contact_phone, health,
		        reg_address, reg_lat, reg_lon, clinic_id
		 FROM profiles WHERE user_id = $1`,
		userID,
	).Scan(&p.Name, &birth, &p.Address, &lat, &lon, &p.ContactName, &p.ContactPhone, &p.Health,
		&p.RegAddress, &regLat, &regLon, &clinicID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, nil
	}
	if err != nil {
		return Profile{}, err
	}

	if birth != nil {
		p.BirthDate = *birth
	}
	if lat != nil && lon != nil {
		p.Lat = *lat
		p.Lon = *lon
		p.HasLocation = true
	}
	if regLat != nil && regLon != nil {
		p.RegLat = *regLat
		p.RegLon = *regLon
		p.HasRegistration = true
	}
	if clinicID != nil {
		p.ClinicID = *clinicID
	}
	return p, nil
}

func (s *Store) SaveProfile(ctx context.Context, userID string, p Profile) error {
	var birth, lat, lon any
	if p.BirthDate != "" {
		birth = p.BirthDate
	}
	if p.HasLocation {
		lat = p.Lat
		lon = p.Lon
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO profiles (user_id, name, birth_date, address, lat, lon, contact_name, contact_phone, health, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
		 ON CONFLICT (user_id) DO UPDATE SET
		   name = EXCLUDED.name, birth_date = EXCLUDED.birth_date, address = EXCLUDED.address,
		   lat = EXCLUDED.lat, lon = EXCLUDED.lon, contact_name = EXCLUDED.contact_name,
		   contact_phone = EXCLUDED.contact_phone, health = EXCLUDED.health, updated_at = now()`,
		userID, p.Name, birth, p.Address, lat, lon, p.ContactName, p.ContactPhone, p.Health,
	)
	return err
}

// адрес сохраняется отдельно и сразу, чтобы он не терялся, если окно профиля закрыли без «Сохранить»
func (s *Store) SaveAddress(ctx context.Context, userID string, address string, lat, lon *float64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO profiles (user_id, address, lat, lon, updated_at) VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (user_id) DO UPDATE SET address = EXCLUDED.address, lat = EXCLUDED.lat, lon = EXCLUDED.lon, updated_at = now()`,
		userID, address, lat, lon,
	)
	return err
}

func (s *Store) SaveRegistration(ctx context.Context, userID string, address string, lat, lon *float64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO profiles (user_id, reg_address, reg_lat, reg_lon, updated_at) VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (user_id) DO UPDATE SET reg_address = EXCLUDED.reg_address, reg_lat = EXCLUDED.reg_lat,
		   reg_lon = EXCLUDED.reg_lon, updated_at = now()`,
		userID, address, lat, lon,
	)
	return err
}

// clinicID = nil — поликлиника снова выбирается автоматически по прописке
func (s *Store) SaveClinic(ctx context.Context, userID string, clinicID *int64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO profiles (user_id, clinic_id, updated_at) VALUES ($1, $2, now())
		 ON CONFLICT (user_id) DO UPDATE SET clinic_id = EXCLUDED.clinic_id, updated_at = now()`,
		userID, clinicID,
	)
	return err
}

func (s *Store) CacheGet(ctx context.Context, key string) (string, bool, error) {
	var answer string
	err := s.db.QueryRow(ctx, `SELECT answer FROM ai_cache WHERE key = $1`, key).Scan(&answer)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return answer, err == nil, err
}

func (s *Store) CachePut(ctx context.Context, key, answer string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO ai_cache (key, answer) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET answer = EXCLUDED.answer, created_at = now()`,
		key, answer,
	)
	return err
}

func seedDemoTasks(ctx context.Context, s *Store, userID, date string) error {
	n, err := s.CountTasks(ctx, userID, date)
	if err != nil || n > 0 {
		return err
	}

	demo := []Task{
		{Time: "08:00", Title: "Выпить таблетку от давления", Kind: "medicine"},
		{Time: "09:30", Title: "Измерить давление", Kind: "other"},
		{Time: "13:00", Title: "Выпить витамин D", Kind: "medicine"},
		{Time: "15:30", Title: "Приём у терапевта, кабинет 214", Kind: "doctor"},
		{Time: "19:00", Title: "Позвонить дочери", Kind: "call"},
	}
	for _, t := range demo {
		if _, err := s.AddTask(ctx, userID, date, t); err != nil {
			return err
		}
	}
	return nil
}
