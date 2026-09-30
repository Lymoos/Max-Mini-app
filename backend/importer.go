package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var overpassURL = "https://overpass-api.de/api/interpreter"

var overpassPause = 5 * time.Second

type city struct {
	Name                     string
	South, West, North, East float64
}

var importCities = []city{
	{Name: "Москва", South: 55.49, West: 37.30, North: 55.96, East: 37.97},
	{Name: "Казань", South: 55.70, West: 48.95, North: 55.90, East: 49.35},
}

type osmElement struct {
	Type   string  `json:"type"`
	ID     int64   `json:"id"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Center *struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"center"`
	Tags map[string]string `json:"tags"`
}

func overpassQuery(kind string, c city) string {
	bbox := fmt.Sprintf("(%f,%f,%f,%f)", c.South, c.West, c.North, c.East)
	var filter string
	switch kind {
	case "shop":
		filter = `nwr["shop"~"^(supermarket|convenience)$"]["name"]` + bbox + ";"
	case "clinic":
		filter = `nwr["amenity"~"^(clinic|hospital|doctors)$"]["name"]` + bbox + ";"
	case "social":
		filter = `(nwr["amenity"~"^(social_facility|social_centre)$"]["name"]` + bbox + ";" +
			`nwr["office"="government"]["government"~"social"]["name"]` + bbox + ";" +
			`nwr["office"]["name"~"социальн|соцзащит|КЦСОН|ЦСО",i]` + bbox + ";);"
	default:
		filter = `nwr["amenity"="pharmacy"]` + bbox + ";"
	}
	return "[out:json][timeout:240];" + filter + "out center tags;"
}

// в OSM бывает несколько номеров через «;» — берём первый
func osmPhone(tags map[string]string) string {
	phone := tags["phone"]
	if phone == "" {
		phone = tags["contact:phone"]
	}
	phone = strings.TrimSpace(strings.Split(strings.Split(phone, ";")[0], ",")[0])
	return phone
}

// у Overpass лимит на число запросов, поэтому при отказе ждём и пробуем снова
func fetchOverpass(ctx context.Context, query string) ([]osmElement, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, overpassURL,
			strings.NewReader(url.Values{"data": {query}}.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "max-miniapp-hackathon/0.1")

		resp, err := client.Do(req)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil && resp.StatusCode == http.StatusOK {
				var data struct {
					Elements []osmElement `json:"elements"`
				}
				if err := json.Unmarshal(body, &data); err == nil {
					return data.Elements, nil
				}
			}
			log.Printf("overpass: статус %d, попытка %d", resp.StatusCode, attempt)
		} else {
			log.Printf("overpass: %v, попытка %d", err, attempt)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt*30) * time.Second):
		}
	}
	return nil, fmt.Errorf("overpass не ответил после 5 попыток")
}

// большие запросы Overpass часто обрывает по таймауту, поэтому город грузим четырьмя частями
func cityTiles(c city) []city {
	midLat := (c.South + c.North) / 2
	midLon := (c.West + c.East) / 2
	return []city{
		{Name: c.Name, South: c.South, West: c.West, North: midLat, East: midLon},
		{Name: c.Name, South: c.South, West: midLon, North: midLat, East: c.East},
		{Name: c.Name, South: midLat, West: c.West, North: c.North, East: midLon},
		{Name: c.Name, South: midLat, West: midLon, North: c.North, East: c.East},
	}
}

func fetchCity(ctx context.Context, kind string, c city) ([]osmElement, error) {
	seen := map[string]bool{}
	result := []osmElement{}
	for i, tile := range cityTiles(c) {
		if i > 0 {
			time.Sleep(overpassPause)
		}
		elements, err := fetchOverpass(ctx, overpassQuery(kind, tile))
		if err != nil {
			return nil, err
		}
		for _, e := range elements {
			key := fmt.Sprintf("%s/%d", e.Type, e.ID)
			if !seen[key] {
				seen[key] = true
				result = append(result, e)
			}
		}
	}
	return result, nil
}

var socialNameRe = regexp.MustCompile(`(?i)социальн|соцзащит|кцсон|цсо|долголети|ветеран|пенсионер|инвалид|милосерд|помощ|опек|попечител|пожил|приют|собес|волонт`)
var socialStopRe = regexp.MustCompile(`(?i)интим|галантере|магазин|салон|клуб знакомств|детск|детей|детям|детьми|дети|ребен|ребён|молочн`)

// в OSM под соцпомощь попадают случайные организации, поэтому оставляем только понятные по названию
func isSocialService(name string) bool {
	return socialNameRe.MatchString(name) && !socialStopRe.MatchString(name)
}

func osmToPlace(e osmElement, kind, cityName string) (Place, string, bool) {
	lat, lon := e.Lat, e.Lon
	if e.Center != nil {
		lat, lon = e.Center.Lat, e.Center.Lon
	}
	if lat == 0 && lon == 0 {
		return Place{}, "", false
	}

	name := e.Tags["name"]
	if name == "" {
		name = e.Tags["brand"]
	}
	if name == "" {
		if kind != "pharmacy" {
			return Place{}, "", false
		}
		name = "Аптека"
	}

	if kind == "social" && !isSocialService(name) {
		return Place{}, "", false
	}

	address := e.Tags["addr:street"]
	if address != "" && e.Tags["addr:housenumber"] != "" {
		address += ", " + e.Tags["addr:housenumber"]
	}

	osmID := fmt.Sprintf("%s/%d", e.Type, e.ID)
	p := Place{Kind: kind, Name: name, Address: address, City: cityName, Lat: lat, Lon: lon, Phone: osmPhone(e.Tags)}
	if kind == "clinic" {
		p.IsState = isStateClinic(name, e.Tags["operator"])
	}
	return p, osmID, true
}

// случайность привязана к id места, поэтому при повторном импорте рейтинги и цены те же
func placeRand(osmID string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(osmID))
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

func demoRating(r *rand.Rand) (float64, int) {
	rating := math.Round((3.3+r.Float64()*1.7)*10) / 10
	reviews := r.Intn(600)
	if r.Float64() < 0.2 {
		reviews = r.Intn(10)
	}
	return rating, reviews
}

func demoBirthdayDiscount(r *rand.Rand) int {
	if r.Float64() < 0.2 {
		return []int{5, 7, 10, 15}[r.Intn(4)]
	}
	return 0
}

func importPlaces(ctx context.Context, s *Store, kind string, c city) (int, error) {
	elements, err := fetchCity(ctx, kind, c)
	if err != nil {
		return 0, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// соцпомощь целиком перезаписываем, чтобы ушли места, которые больше не проходят фильтр
	if kind == "social" {
		if _, err := tx.Exec(ctx, `DELETE FROM places WHERE city = $1 AND kind = 'social'`, c.Name); err != nil {
			return 0, err
		}
	}

	batch := &pgx.Batch{}
	type saved struct {
		osmID string
		rng   *rand.Rand
		tags  map[string]string
	}
	list := []saved{}
	for _, e := range elements {
		p, osmID, ok := osmToPlace(e, kind, c.Name)
		if !ok {
			continue
		}
		r := placeRand(osmID)
		rating, reviews := demoRating(r)
		discount := 0
		if kind == "pharmacy" || kind == "shop" {
			discount = demoBirthdayDiscount(r)
		}
		batch.Queue(
			`INSERT INTO places (osm_id, kind, name, address, city, lat, lon, rating, reviews, birthday_discount, phone, is_state, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'osm')
			 ON CONFLICT (osm_id) DO UPDATE SET name = EXCLUDED.name, address = EXCLUDED.address,
			   lat = EXCLUDED.lat, lon = EXCLUDED.lon, city = EXCLUDED.city, phone = EXCLUDED.phone, is_state = EXCLUDED.is_state
			 RETURNING id`,
			osmID, kind, p.Name, p.Address, c.Name, p.Lat, p.Lon, rating, reviews, discount, p.Phone, p.IsState,
		)
		list = append(list, saved{osmID: osmID, rng: r, tags: e.Tags})
	}

	results := tx.SendBatch(ctx, batch)
	ids := make([]int64, len(list))
	for i := range list {
		if err := results.QueryRow().Scan(&ids[i]); err != nil {
			results.Close()
			return 0, err
		}
	}
	if err := results.Close(); err != nil {
		return 0, err
	}

	if kind == "clinic" {
		_, err = tx.Exec(ctx, `DELETE FROM doctors WHERE place_id IN (SELECT id FROM places WHERE city = $1 AND kind = 'clinic')`, c.Name)
		if err != nil {
			return 0, err
		}
		doctorRows := [][]any{}
		for i, item := range list {
			for _, d := range demoDoctors(item.rng, ids[i]) {
				doctorRows = append(doctorRows, []any{d.ClinicID, d.Name, d.Specialty, d.Experience, d.Category, float32(d.Rating), d.Reviews})
			}
		}
		_, err = tx.CopyFrom(ctx, pgx.Identifier{"doctors"},
			[]string{"place_id", "name", "specialty", "experience", "category", "rating", "reviews"},
			pgx.CopyFromRows(doctorRows))
		if err != nil {
			return 0, err
		}
		return len(list), tx.Commit(ctx)
	}
	if kind == "social" {
		return len(list), tx.Commit(ctx)
	}

	priceRows := [][]any{}
	for i, item := range list {
		factor := 0.85 + item.rng.Float64()*0.3
		if kind == "pharmacy" {
			for _, m := range medicines {
				if item.rng.Float64() < 0.15 {
					continue
				}
				priceRows = append(priceRows, []any{ids[i], m.ID, demoPrice(m.BasePrice, factor, item.rng)})
			}
		} else {
			have := 0.9
			if item.tags["shop"] == "convenience" {
				have = 0.6
			}
			for _, p := range products {
				if item.rng.Float64() > have {
					continue
				}
				priceRows = append(priceRows, []any{ids[i], p.ID, demoPrice(p.BasePrice, factor, item.rng)})
			}
		}
	}

	table, column := "medicine_prices", "medicine_id"
	if kind == "shop" {
		table, column = "product_prices", "product_id"
	}
	_, err = tx.Exec(ctx,
		`DELETE FROM `+table+` WHERE place_id IN (SELECT id FROM places WHERE city = $1 AND kind = $2)`,
		c.Name, kind,
	)
	if err != nil {
		return 0, err
	}
	_, err = tx.CopyFrom(ctx, pgx.Identifier{table}, []string{"place_id", column, "price"}, pgx.CopyFromRows(priceRows))
	if err != nil {
		return 0, err
	}
	return len(list), tx.Commit(ctx)
}

func demoPrice(base int, factor float64, r *rand.Rand) int {
	return int(math.Round(float64(base) * factor * (0.95 + r.Float64()*0.1)))
}

// акции генерируются от текущей даты: перед показом можно запустить `go run . promos`
func refreshPromos(ctx context.Context, s *Store) error {
	_, err := s.db.Exec(ctx, `UPDATE product_prices SET promo_price = NULL, promo_until = NULL`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx,
		`UPDATE product_prices
		 SET promo_price = round(price * (0.6 + random() * 0.3))::int,
		     promo_until = current_date + (2 + floor(random() * 13))::int
		 WHERE random() < 0.12`,
	)
	return err
}

var importKinds = []string{"pharmacy", "shop", "clinic", "social"}

func runImport(ctx context.Context, s *Store, kinds []string) error {
	if len(kinds) == 0 {
		kinds = importKinds
	}
	for _, kind := range kinds {
		if !contains(importKinds, kind) {
			return fmt.Errorf("неизвестный вид мест %q, есть: %s", kind, strings.Join(importKinds, ", "))
		}
	}

	for _, c := range importCities {
		for _, kind := range kinds {
			log.Printf("загружаем %s: %s…", c.Name, kind)
			n, err := importPlaces(ctx, s, kind, c)
			if err != nil {
				return fmt.Errorf("%s %s: %w", c.Name, kind, err)
			}
			log.Printf("%s: %s — %d мест", c.Name, kind, n)
			time.Sleep(10 * time.Second)
		}
	}
	if contains(kinds, "shop") {
		return refreshPromos(ctx, s)
	}
	return nil
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
