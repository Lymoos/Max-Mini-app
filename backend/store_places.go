package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Promo struct {
	ProductID  string `json:"productId"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	Price      int    `json:"price"`
	OldPrice   int    `json:"oldPrice"`
	Discount   int    `json:"discount"`
	PromoUntil string `json:"promoUntil"`
}

const placeColumns = `p.id, p.kind, p.name, p.address, p.city, p.lat, p.lon, p.rating, p.reviews, p.birthday_discount, p.phone, p.is_state`

func scanPlace(row pgx.Row, extra ...any) (Place, error) {
	var p Place
	var rating float32
	dest := []any{&p.ID, &p.Kind, &p.Name, &p.Address, &p.City, &p.Lat, &p.Lon, &rating, &p.Reviews, &p.BirthdayDiscount, &p.Phone, &p.IsState}
	err := row.Scan(append(dest, extra...)...)
	p.Rating = float64(rating)
	return p, err
}

func (s *Store) SeedCatalogs(ctx context.Context) error {
	batch := &pgx.Batch{}
	for _, m := range medicines {
		batch.Queue(
			`INSERT INTO medicines (id, name, form, base_price) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, form = EXCLUDED.form, base_price = EXCLUDED.base_price`,
			m.ID, m.Name, m.Form, m.BasePrice,
		)
	}
	for _, p := range products {
		batch.Queue(
			`INSERT INTO products (id, name, unit, base_price) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, unit = EXCLUDED.unit, base_price = EXCLUDED.base_price`,
			p.ID, p.Name, p.Unit, p.BasePrice,
		)
	}
	return s.db.SendBatch(ctx, batch).Close()
}

func (s *Store) PlacesNear(ctx context.Context, kind string, lat, lon, km float64) ([]Place, error) {
	minLat, maxLat, minLon, maxLon := boundingBox(lat, lon, km)
	rows, err := s.db.Query(ctx,
		`SELECT `+placeColumns+` FROM places p
		 WHERE p.kind = $1 AND p.lat BETWEEN $2 AND $3 AND p.lon BETWEEN $4 AND $5`,
		kind, minLat, maxLat, minLon, maxLon,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := []Place{}
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, p)
	}
	return places, rows.Err()
}

func (s *Store) GetPlace(ctx context.Context, kind string, id int64) (Place, error) {
	p, err := scanPlace(s.db.QueryRow(ctx, `SELECT `+placeColumns+` FROM places p WHERE p.id = $1 AND p.kind = $2`, id, kind))
	if errors.Is(err, pgx.ErrNoRows) {
		return Place{}, ErrNotFound
	}
	return p, err
}

func (s *Store) MedicineOffers(ctx context.Context, medicineID string, lat, lon, km float64) ([]Offer, error) {
	minLat, maxLat, minLon, maxLon := boundingBox(lat, lon, km)
	rows, err := s.db.Query(ctx,
		`SELECT `+placeColumns+`, mp.price FROM medicine_prices mp
		 JOIN places p ON p.id = mp.place_id
		 WHERE mp.medicine_id = $1 AND p.lat BETWEEN $2 AND $3 AND p.lon BETWEEN $4 AND $5`,
		medicineID, minLat, maxLat, minLon, maxLon,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := []Offer{}
	for rows.Next() {
		var price int
		p, err := scanPlace(rows, &price)
		if err != nil {
			return nil, err
		}
		r := rankPlace(p, lat, lon)
		if r.DistanceKm <= km {
			offers = append(offers, Offer{RankedPlace: r, Price: price})
		}
	}
	return offers, rows.Err()
}

func (s *Store) ProductOffers(ctx context.Context, productID, today string, lat, lon, km float64) ([]Offer, error) {
	minLat, maxLat, minLon, maxLon := boundingBox(lat, lon, km)
	rows, err := s.db.Query(ctx,
		`SELECT `+placeColumns+`, pp.price, pp.promo_price, to_char(pp.promo_until, 'YYYY-MM-DD')
		 FROM product_prices pp
		 JOIN places p ON p.id = pp.place_id
		 WHERE pp.product_id = $1 AND p.lat BETWEEN $2 AND $3 AND p.lon BETWEEN $4 AND $5`,
		productID, minLat, maxLat, minLon, maxLon,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := []Offer{}
	for rows.Next() {
		var price int
		var promoPrice *int
		var promoUntil *string
		p, err := scanPlace(rows, &price, &promoPrice, &promoUntil)
		if err != nil {
			return nil, err
		}
		r := rankPlace(p, lat, lon)
		if r.DistanceKm > km {
			continue
		}

		o := Offer{RankedPlace: r, Price: price}
		if promoPrice != nil && promoUntil != nil && *promoUntil >= today {
			o.OldPrice = price
			o.Price = *promoPrice
			o.PromoUntil = *promoUntil
		}
		offers = append(offers, o)
	}
	return offers, rows.Err()
}

func (s *Store) ShopPromos(ctx context.Context, placeID int64, today string) ([]Promo, error) {
	rows, err := s.db.Query(ctx,
		`SELECT pr.id, pr.name, pr.unit, pp.promo_price, pp.price, to_char(pp.promo_until, 'YYYY-MM-DD')
		 FROM product_prices pp
		 JOIN products pr ON pr.id = pp.product_id
		 WHERE pp.place_id = $1 AND pp.promo_price IS NOT NULL AND pp.promo_until >= $2
		 ORDER BY (pp.price - pp.promo_price)::float / pp.price DESC, pr.name`,
		placeID, today,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	promos := []Promo{}
	for rows.Next() {
		var p Promo
		if err := rows.Scan(&p.ProductID, &p.Name, &p.Unit, &p.Price, &p.OldPrice, &p.PromoUntil); err != nil {
			return nil, err
		}
		p.Discount = int(float64(p.OldPrice-p.Price) / float64(p.OldPrice) * 100)
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (s *Store) CountPlaces(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.Query(ctx, `SELECT kind, count(*) FROM places GROUP BY kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			return nil, err
		}
		counts[kind] = n
	}
	return counts, rows.Err()
}
