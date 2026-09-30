package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeGeocoder struct {
	fail bool
}

func (f *fakeGeocoder) Search(ctx context.Context, query string) ([]GeoPlace, error) {
	if f.fail {
		return nil, errors.New("нет сети")
	}
	return []GeoPlace{{Address: "Тверская улица, 7, Москва", Lat: 55.757, Lon: 37.613}}, nil
}

func (f *fakeGeocoder) Reverse(ctx context.Context, lat, lon float64) (GeoPlace, error) {
	if f.fail {
		return GeoPlace{}, errors.New("нет сети")
	}
	if lat == 0 && lon == 0 {
		return GeoPlace{}, ErrPlaceNotFound
	}
	return GeoPlace{Address: "Тверская улица, 7, Москва", Lat: lat, Lon: lon}, nil
}

func newFakeNominatim(t *testing.T, body string, status int) *NominatimGeocoder {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("Nominatim требует User-Agent")
		}
		if r.URL.Query().Get("accept-language") != "ru" {
			t.Error("адреса нужны на русском")
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	g := NewNominatimGeocoder(srv.URL)
	g.minInterval = 0
	return g
}

func TestNominatimSearch(t *testing.T) {
	body := `[
		{"lat": "55.757", "lon": "37.613", "display_name": "7, Тверская улица, Москва, Россия",
		 "address": {"road": "Тверская улица", "house_number": "7", "city": "Москва"}},
		{"lat": "55.1", "lon": "37.1", "display_name": "Тверская, Подольск, Московская область, Россия",
		 "address": {"town": "Подольск"}},
		{"lat": "плохо", "lon": "37.1", "display_name": "x"}
	]`
	g := newFakeNominatim(t, body, http.StatusOK)

	places, err := g.Search(context.Background(), "Тверская 7")
	if err != nil {
		t.Fatal(err)
	}
	if len(places) != 2 {
		t.Fatalf("строка с плохими координатами должна пропускаться, получили %d", len(places))
	}
	if places[0].Address != "Тверская улица, 7, Москва" || places[0].Lat != 55.757 {
		t.Errorf("неверный первый адрес: %+v", places[0])
	}
	if places[1].Address != "Тверская, Подольск, Московская область" {
		t.Errorf("без улицы берём начало полного адреса, получили %q", places[1].Address)
	}
}

func TestNominatimReverse(t *testing.T) {
	body := `{"lat": "55.7570001", "lon": "37.6130001", "display_name": "7, Тверская улица, Москва",
		"address": {"road": "Тверская улица", "house_number": "7", "city": "Москва"}}`
	g := newFakeNominatim(t, body, http.StatusOK)

	place, err := g.Reverse(context.Background(), 55.757, 37.613)
	if err != nil {
		t.Fatal(err)
	}
	if place.Address != "Тверская улица, 7, Москва" {
		t.Errorf("адрес %q", place.Address)
	}
	if place.Lat != 55.757 || place.Lon != 37.613 {
		t.Error("координаты должны остаться теми, куда нажал пользователь")
	}
}

func TestNominatimErrors(t *testing.T) {
	g := newFakeNominatim(t, `{"error": "Unable to geocode"}`, http.StatusOK)
	if _, err := g.Reverse(context.Background(), 0, 0); !errors.Is(err, ErrPlaceNotFound) {
		t.Errorf("ожидали ErrPlaceNotFound, получили %v", err)
	}

	g = newFakeNominatim(t, `oops`, http.StatusTooManyRequests)
	if _, err := g.Search(context.Background(), "Тверская"); err == nil {
		t.Error("статус 429 должен давать ошибку")
	}

	g = newFakeNominatim(t, `не json`, http.StatusOK)
	if _, err := g.Search(context.Background(), "Тверская"); err == nil {
		t.Error("битый ответ должен давать ошибку")
	}
}
