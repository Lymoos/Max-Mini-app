package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrPlaceNotFound = errors.New("place not found")

type GeoPlace struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

type Geocoder interface {
	Search(ctx context.Context, query string) ([]GeoPlace, error)
	Reverse(ctx context.Context, lat, lon float64) (GeoPlace, error)
}

type NominatimGeocoder struct {
	baseURL     string
	client      *http.Client
	minInterval time.Duration
	mu          sync.Mutex
	lastCall    time.Time
}

func NewNominatimGeocoder(baseURL string) *NominatimGeocoder {
	return &NominatimGeocoder{
		baseURL:     baseURL,
		client:      &http.Client{Timeout: 8 * time.Second},
		minInterval: time.Second,
	}
}

type nominatimPlace struct {
	Lat         string            `json:"lat"`
	Lon         string            `json:"lon"`
	DisplayName string            `json:"display_name"`
	Address     map[string]string `json:"address"`
	Error       string            `json:"error"`
}

// правила Nominatim: не чаще одного запроса в секунду
func (g *NominatimGeocoder) waitTurn() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if wait := g.minInterval - time.Since(g.lastCall); wait > 0 {
		time.Sleep(wait)
	}
	g.lastCall = time.Now()
}

func (g *NominatimGeocoder) get(ctx context.Context, path string, params url.Values, dst any) error {
	params.Set("format", "jsonv2")
	params.Set("addressdetails", "1")
	params.Set("accept-language", "ru")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "max-miniapp-hackathon/0.1")

	g.waitTurn()
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nominatim: статус %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func shortAddress(p nominatimPlace) string {
	a := p.Address
	street := a["road"]
	if street == "" {
		street = a["pedestrian"]
	}
	city := a["city"]
	if city == "" {
		city = a["town"]
	}
	if city == "" {
		city = a["village"]
	}

	if street == "" {
		parts := strings.Split(p.DisplayName, ", ")
		if len(parts) > 3 {
			parts = parts[:3]
		}
		return strings.Join(parts, ", ")
	}

	parts := []string{street}
	if a["house_number"] != "" {
		parts = append(parts, a["house_number"])
	}
	if city != "" {
		parts = append(parts, city)
	}
	return strings.Join(parts, ", ")
}

func toPlace(p nominatimPlace) (GeoPlace, error) {
	lat, err1 := strconv.ParseFloat(p.Lat, 64)
	lon, err2 := strconv.ParseFloat(p.Lon, 64)
	if err1 != nil || err2 != nil {
		return GeoPlace{}, fmt.Errorf("nominatim: неверные координаты %q %q", p.Lat, p.Lon)
	}
	return GeoPlace{Address: shortAddress(p), Lat: lat, Lon: lon}, nil
}

func (g *NominatimGeocoder) Search(ctx context.Context, query string) ([]GeoPlace, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("countrycodes", "ru")
	params.Set("limit", "5")

	var raw []nominatimPlace
	if err := g.get(ctx, "/search", params, &raw); err != nil {
		return nil, err
	}

	places := []GeoPlace{}
	for _, p := range raw {
		place, err := toPlace(p)
		if err != nil {
			continue
		}
		places = append(places, place)
	}
	return places, nil
}

func (g *NominatimGeocoder) Reverse(ctx context.Context, lat, lon float64) (GeoPlace, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', 6, 64))
	params.Set("zoom", "18")

	var raw nominatimPlace
	if err := g.get(ctx, "/reverse", params, &raw); err != nil {
		return GeoPlace{}, err
	}
	if raw.Error != "" {
		return GeoPlace{}, ErrPlaceNotFound
	}

	place, err := toPlace(raw)
	if err != nil {
		return GeoPlace{}, err
	}
	place.Lat = lat
	place.Lon = lon
	return place, nil
}
