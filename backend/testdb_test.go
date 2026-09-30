package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// тесты с базой идут на отдельной базе maxapp_test из docker-compose
func testStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://maxapp:maxapp@localhost:5440/maxapp_test?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := connectDB(ctx, url)
	if err != nil {
		t.Fatalf("нет тестовой базы, запустите `docker compose up -d`: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `TRUNCATE tasks, profiles, places, medicine_prices, product_prices, doctors, ai_cache, user_benefits, user_guides RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(pool)
	if err := store.SeedCatalogs(ctx); err != nil {
		t.Fatal(err)
	}
	return store
}

var placeCounter int

func addPlace(t *testing.T, s *Store, p Place) Place {
	t.Helper()
	placeCounter++
	if p.City == "" {
		p.City = "Москва"
	}
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO places (osm_id, kind, name, address, city, lat, lon, rating, reviews, birthday_discount, phone, is_state)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`,
		fmt.Sprintf("node/test%d", placeCounter), p.Kind, p.Name, p.Address, p.City, p.Lat, p.Lon, p.Rating, p.Reviews, p.BirthdayDiscount,
		p.Phone, p.IsState,
	).Scan(&p.ID)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func addMedicinePrice(t *testing.T, s *Store, placeID int64, medicineID string, price int) {
	t.Helper()
	_, err := s.db.Exec(context.Background(),
		`INSERT INTO medicine_prices (place_id, medicine_id, price) VALUES ($1, $2, $3)`, placeID, medicineID, price)
	if err != nil {
		t.Fatal(err)
	}
}

func addProductPrice(t *testing.T, s *Store, placeID int64, productID string, price int, promoPrice int, promoUntil string) {
	t.Helper()
	var promo, until any
	if promoPrice > 0 {
		promo, until = promoPrice, promoUntil
	}
	_, err := s.db.Exec(context.Background(),
		`INSERT INTO product_prices (place_id, product_id, price, promo_price, promo_until) VALUES ($1, $2, $3, $4, $5)`,
		placeID, productID, price, promo, until)
	if err != nil {
		t.Fatal(err)
	}
}

func addDoctor(t *testing.T, s *Store, d Doctor) Doctor {
	t.Helper()
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO doctors (place_id, name, specialty, experience, category, rating, reviews)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		d.ClinicID, d.Name, d.Specialty, d.Experience, d.Category, d.Rating, d.Reviews,
	).Scan(&d.ID)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
