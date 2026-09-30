package main

import (
	"math"
	"sort"
)

const (
	priorRating  = 4.0
	priorVotes   = 20.0
	ratingDistKm = 1.0
	decayDistKm  = 5.0
	searchKm     = 3.0
	nearbyLimit  = 10
	offersLimit  = 20
)

type Place struct {
	ID               int64   `json:"id"`
	Kind             string  `json:"kind"`
	Name             string  `json:"name"`
	Address          string  `json:"address"`
	City             string  `json:"city"`
	Lat              float64 `json:"lat"`
	Lon              float64 `json:"lon"`
	Rating           float64 `json:"rating"`
	Reviews          int     `json:"reviews"`
	BirthdayDiscount int     `json:"birthdayDiscount"`
	Phone            string  `json:"phone"`
	IsState          bool    `json:"isState"`
}

type RankedPlace struct {
	Place
	DistanceKm float64 `json:"distanceKm"`
	Score      float64 `json:"score"`
}

type Offer struct {
	RankedPlace
	Price      int    `json:"price"`
	OldPrice   int    `json:"oldPrice,omitempty"`
	PromoUntil string `json:"promoUntil,omitempty"`
	Cheapest   bool   `json:"cheapest"`
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusKm * math.Asin(math.Sqrt(a))
}

// прямоугольник вокруг точки, чтобы база быстро отсекла дальние места по индексу
func boundingBox(lat, lon, km float64) (minLat, maxLat, minLon, maxLon float64) {
	dLat := km / 111.0
	dLon := km / (111.0 * math.Cos(lat*math.Pi/180))
	return lat - dLat, lat + dLat, lon - dLon, lon + dLon
}

// рейтинг с поправкой на число отзывов: у места с одной пятёркой он ближе к среднему
func bayesRating(rating float64, reviews int) float64 {
	v := float64(reviews)
	return (v*rating + priorVotes*priorRating) / (v + priorVotes)
}

// главное — оценка. Степень растёт с расстоянием, поэтому чем дальше место,
// тем сильнее низкий рейтинг тянет его вниз. Делитель слегка поднимает ближние
func placeScore(rating float64, reviews int, distKm float64) float64 {
	r := bayesRating(rating, reviews) / 5
	return math.Pow(r, 1+distKm/ratingDistKm) / (1 + distKm/decayDistKm)
}

func rankPlace(p Place, lat, lon float64) RankedPlace {
	d := haversineKm(lat, lon, p.Lat, p.Lon)
	return RankedPlace{
		Place:      p,
		DistanceKm: math.Round(d*100) / 100,
		Score:      math.Round(placeScore(p.Rating, p.Reviews, d)*1000) / 1000,
	}
}

func rankPlaces(list []Place, lat, lon, maxKm float64) []RankedPlace {
	result := []RankedPlace{}
	for _, p := range list {
		r := rankPlace(p, lat, lon)
		if r.DistanceKm > maxKm {
			continue
		}
		result = append(result, r)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

// предложения уже с посчитанным расстоянием и оценкой: сортируем и помечаем самую низкую цену
func sortOffers(offers []Offer, limit int) []Offer {
	sort.SliceStable(offers, func(i, j int) bool {
		return offers[i].Score > offers[j].Score
	})
	if len(offers) > limit {
		offers = offers[:limit]
	}

	cheapest := -1
	for i := range offers {
		offers[i].Cheapest = false
		if cheapest == -1 || offers[i].Price < offers[cheapest].Price {
			cheapest = i
		}
	}
	if cheapest != -1 {
		offers[cheapest].Cheapest = true
	}
	return offers
}
