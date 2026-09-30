package main

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	d := haversineKm(55.7539, 37.6208, 59.9343, 30.3351)
	if math.Abs(d-634) > 5 {
		t.Errorf("Москва — Петербург: %.1f км", d)
	}
	if haversineKm(55.75, 37.62, 55.75, 37.62) != 0 {
		t.Error("до той же точки 0 км")
	}
}

func TestBoundingBox(t *testing.T) {
	minLat, maxLat, minLon, maxLon := boundingBox(55.75, 37.62, 3)
	for _, p := range [][2]float64{{minLat, 37.62}, {maxLat, 37.62}, {55.75, minLon}, {55.75, maxLon}} {
		if d := haversineKm(55.75, 37.62, p[0], p[1]); math.Abs(d-3) > 0.1 {
			t.Errorf("край прямоугольника в %.2f км, ожидали 3", d)
		}
	}
}

func TestBayesRating(t *testing.T) {
	if bayesRating(5.0, 3) >= bayesRating(4.8, 300) {
		t.Error("три пятёрки не должны перевешивать сотни оценок 4.8")
	}
	if bayesRating(4.0, 0) != priorRating {
		t.Error("без отзывов рейтинг равен среднему")
	}
}

func TestPlaceScore(t *testing.T) {
	if placeScore(4.8, 200, 1) <= placeScore(4.2, 200, 1) {
		t.Error("при одинаковом расстоянии выше рейтинг — выше место")
	}
	if placeScore(4.5, 200, 0.3) <= placeScore(4.5, 200, 2) {
		t.Error("при одинаковом рейтинге ближнее выше")
	}
	near := placeScore(4.8, 200, 0.2) / placeScore(4.0, 200, 0.2)
	far := placeScore(4.8, 200, 3) / placeScore(4.0, 200, 3)
	if far <= near {
		t.Errorf("вдали рейтинг должен решать больше: %.2f против %.2f", far, near)
	}
	if placeScore(4.8, 300, 0.8) <= placeScore(3.5, 300, 0.3) {
		t.Error("хорошее место чуть дальше обгоняет плохое рядом")
	}
	if placeScore(4.9, 300, 8) >= placeScore(4.5, 300, 0.3) {
		t.Error("отличное место за 8 км уступает хорошему в 300 м")
	}
}

func TestRankPlaces(t *testing.T) {
	list := []Place{
		{ID: 1, Rating: 3.5, Reviews: 300, Lat: 55.7540, Lon: 37.6210},
		{ID: 2, Rating: 4.8, Reviews: 300, Lat: 55.7560, Lon: 37.6220},
		{ID: 3, Rating: 5.0, Reviews: 300, Lat: 59.9343, Lon: 30.3351},
	}
	got := rankPlaces(list, 55.7539, 37.6208, searchKm)
	if len(got) != 2 || got[0].ID != 2 {
		t.Errorf("неверный порядок или не отсеяли дальнее: %+v", got)
	}
}

func TestSortOffers(t *testing.T) {
	offers := []Offer{
		{RankedPlace: RankedPlace{Score: 0.5}, Price: 30},
		{RankedPlace: RankedPlace{Score: 0.9}, Price: 50},
		{RankedPlace: RankedPlace{Score: 0.7}, Price: 40, Cheapest: true},
	}
	got := sortOffers(offers, 2)
	if len(got) != 2 || got[0].Score != 0.9 || got[1].Score != 0.7 {
		t.Fatalf("неверная сортировка или лимит: %+v", got)
	}
	if !got[1].Cheapest || got[0].Cheapest {
		t.Error("дешевле всего среди показанных — 40 ₽")
	}
	if len(sortOffers([]Offer{}, 5)) != 0 {
		t.Error("пустой список")
	}
}
