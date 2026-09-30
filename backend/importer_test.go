package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOsmToPlace(t *testing.T) {
	node := osmElement{Type: "node", ID: 1, Lat: 55.75, Lon: 37.62,
		Tags: map[string]string{"name": "Ригла", "addr:street": "Тверская улица", "addr:housenumber": "7"}}
	p, osmID, ok := osmToPlace(node, "pharmacy", "Москва")
	if !ok || osmID != "node/1" || p.Name != "Ригла" || p.Address != "Тверская улица, 7" || p.Lat != 55.75 {
		t.Errorf("%+v %s %v", p, osmID, ok)
	}

	way := osmElement{Type: "way", ID: 2, Tags: map[string]string{"brand": "Пятёрочка", "addr:street": "Арбат"}}
	way.Center = &struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}{55.74, 37.59}
	p, osmID, ok = osmToPlace(way, "shop", "Москва")
	if !ok || osmID != "way/2" || p.Name != "Пятёрочка" || p.Address != "Арбат" || p.Lat != 55.74 {
		t.Errorf("здание берём по центру: %+v %s %v", p, osmID, ok)
	}

	if p, _, ok := osmToPlace(osmElement{Type: "node", ID: 3, Lat: 55, Lon: 37}, "pharmacy", "Москва"); !ok || p.Name != "Аптека" {
		t.Errorf("аптека без названия называется «Аптека»: %+v", p)
	}
	if _, _, ok := osmToPlace(osmElement{Type: "node", ID: 4, Lat: 55, Lon: 37}, "shop", "Москва"); ok {
		t.Error("магазин без названия пропускаем")
	}
	if _, _, ok := osmToPlace(osmElement{Type: "way", ID: 5, Tags: map[string]string{"name": "x"}}, "shop", "Москва"); ok {
		t.Error("без координат пропускаем")
	}
}

func TestOverpassQuery(t *testing.T) {
	q := overpassQuery("pharmacy", importCities[0])
	if !strings.Contains(q, `"amenity"="pharmacy"`) || !strings.Contains(q, "out center") || !strings.Contains(q, "55.49") {
		t.Errorf("запрос аптек: %s", q)
	}
	if q := overpassQuery("shop", importCities[1]); !strings.Contains(q, "supermarket|convenience") {
		t.Errorf("запрос магазинов: %s", q)
	}
}

func TestDemoDataIsStable(t *testing.T) {
	r1, r2 := placeRand("node/42"), placeRand("node/42")
	a1, b1 := demoRating(r1)
	a2, b2 := demoRating(r2)
	if a1 != a2 || b1 != b2 {
		t.Error("для одного места рейтинг при повторном импорте должен совпадать")
	}
	for i := 0; i < 500; i++ {
		rating, reviews := demoRating(placeRand("node/x" + string(rune(i))))
		if rating < 3.3 || rating > 5 || reviews < 0 {
			t.Fatalf("рейтинг %.1f, отзывов %d", rating, reviews)
		}
	}
}

func TestImportPlaces(t *testing.T) {
	s := testStore(t)
	var requests atomic.Int32
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		r.ParseForm()
		if !strings.Contains(r.Form.Get("data"), "pharmacy") {
			t.Errorf("запрос: %s", r.Form.Get("data"))
		}
		w.Write([]byte(`{"elements": [
			{"type": "node", "id": 1, "lat": 55.755, "lon": 37.622, "tags": {"name": "Ригла", "addr:street": "Тверская", "addr:housenumber": "1"}},
			{"type": "node", "id": 2, "lat": 55.756, "lon": 37.623, "tags": {}},
			{"type": "way", "id": 3, "tags": {"name": "без координат"}}
		]}`))
	}))
	defer fake.Close()
	old, oldPause := overpassURL, overpassPause
	overpassURL, overpassPause = fake.URL, 0
	defer func() { overpassURL, overpassPause = old, oldPause }()

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		n, err := importPlaces(ctx, s, "pharmacy", importCities[0])
		if err != nil || n != 2 {
			t.Fatalf("импорт %d: %d мест, %v", i, n, err)
		}
	}

	places, _ := s.PlacesNear(ctx, "pharmacy", 55.755, 37.622, 1)
	if len(places) != 2 {
		t.Errorf("повторный импорт не должен дублировать места: %d", len(places))
	}

	var prices int
	s.db.QueryRow(ctx, `SELECT count(*) FROM medicine_prices`).Scan(&prices)
	if prices == 0 || prices > 2*len(medicines) {
		t.Errorf("цен %d", prices)
	}
}

func TestCityTiles(t *testing.T) {
	c := city{Name: "Москва", South: 55, West: 37, North: 56, East: 38}
	tiles := cityTiles(c)
	if len(tiles) != 4 {
		t.Fatal(len(tiles))
	}
	area := 0.0
	for _, tile := range tiles {
		if tile.Name != "Москва" || tile.South < c.South || tile.North > c.North || tile.West < c.West || tile.East > c.East {
			t.Errorf("часть выходит за город: %+v", tile)
		}
		area += (tile.North - tile.South) * (tile.East - tile.West)
	}
	if area != 1 {
		t.Errorf("части должны покрывать весь город, площадь %v", area)
	}
}

func TestIsSocialService(t *testing.T) {
	good := []string{
		"Отдел социальной защиты населения Таганского района",
		"Центр московского долголетия «Орехово-Борисово Южное»",
		"ГБУ ТЦСО «Мещанский»",
		"КЦСОН Вахитовского района",
		"Совет ветеранов",
		"Приют человека",
		"Центр социальной помощи пожилым",
	}
	bad := []string{
		"Школа интимной гармонии",
		"Московское ПО «Металлпластизделие» (Галантерея)",
		"Магазин социальных цен",
		"Центр социальной помощи семье и детям",
		"Молочно-раздаточный пункт",
		"Детская городская поликлиника №98 Филиал №1 молочно раздаточный пункт",
		"Кафе",
	}
	for _, name := range good {
		if !isSocialService(name) {
			t.Errorf("%q должна пройти", name)
		}
	}
	for _, name := range bad {
		if isSocialService(name) {
			t.Errorf("%q не должна пройти", name)
		}
	}
	if _, _, ok := osmToPlace(osmElement{Type: "node", ID: 1, Lat: 55, Lon: 37, Tags: map[string]string{"name": "Кафе"}}, "social", "Москва"); ok {
		t.Error("osmToPlace должен отсеивать лишнее в соцпомощи")
	}
}

func TestOsmPhone(t *testing.T) {
	cases := map[string]map[string]string{
		"+7 495 123-45-67": {"phone": "+7 495 123-45-67"},
		"+7 843 111-22-33": {"contact:phone": "+7 843 111-22-33"},
		"+7 495 000-00-01": {"phone": "+7 495 000-00-01; +7 495 000-00-02"},
		"+7 495 000-00-03": {"phone": " +7 495 000-00-03, +7 495 000-00-04"},
		"":                 {},
	}
	for want, tags := range cases {
		if got := osmPhone(tags); got != want {
			t.Errorf("osmPhone(%v) = %q", tags, got)
		}
	}
}

func TestOverpassQueryClinicSocial(t *testing.T) {
	if q := overpassQuery("clinic", importCities[0]); !strings.Contains(q, "clinic|hospital|doctors") {
		t.Errorf("клиники: %s", q)
	}
	q := overpassQuery("social", importCities[1])
	if !strings.Contains(q, "social_facility") || !strings.Contains(q, "КЦСОН") || strings.Count(q, "55.700000") != 3 {
		t.Errorf("соцпомощь: %s", q)
	}
}

func TestImportClinicsCreatesDoctors(t *testing.T) {
	s := testStore(t)
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"elements": [
			{"type": "node", "id": 10, "lat": 55.755, "lon": 37.622, "tags": {"name": "Городская поликлиника № 5", "phone": "+7 495 111-11-11"}},
			{"type": "node", "id": 11, "lat": 55.756, "lon": 37.623, "tags": {"name": "Медси"}}
		]}`))
	}))
	defer fake.Close()
	old, oldPause := overpassURL, overpassPause
	overpassURL, overpassPause = fake.URL, 0
	defer func() { overpassURL, overpassPause = old, oldPause }()

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if n, err := importPlaces(ctx, s, "clinic", importCities[0]); err != nil || n != 2 {
			t.Fatalf("%d %v", n, err)
		}
	}

	places, _ := s.PlacesNear(ctx, "clinic", 55.755, 37.622, 1)
	var state, private Place
	for _, p := range places {
		if p.Name == "Медси" {
			private = p
		} else {
			state = p
		}
	}
	if !state.IsState || private.IsState || state.Phone != "+7 495 111-11-11" || state.BirthdayDiscount != 0 {
		t.Errorf("гос: %+v, частная: %+v", state, private)
	}

	var doctors int
	s.db.QueryRow(ctx, `SELECT count(*) FROM doctors`).Scan(&doctors)
	if doctors < 12 || doctors > 28 {
		t.Errorf("повторный импорт не должен дублировать врачей, их %d", doctors)
	}
}

func TestRunImportUnknownKind(t *testing.T) {
	s := testStore(t)
	if err := runImport(context.Background(), s, []string{"zoo"}); err == nil {
		t.Error("неизвестный вид мест должен давать ошибку")
	}
}
