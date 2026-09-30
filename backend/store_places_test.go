package main

import (
	"context"
	"errors"
	"testing"
)

var center = Location{Lat: 55.7539, Lon: 37.6208}

func TestPlacesNear(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	addPlace(t, s, Place{Kind: "pharmacy", Name: "рядом", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 10})
	addPlace(t, s, Place{Kind: "shop", Name: "магазин рядом", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 10})
	addPlace(t, s, Place{Kind: "pharmacy", Name: "Казань", City: "Казань", Lat: 55.79, Lon: 49.12, Rating: 5, Reviews: 100})

	got, err := s.PlacesNear(ctx, "pharmacy", center.Lat, center.Lon, searchKm)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "рядом" || got[0].Kind != "pharmacy" {
		t.Errorf("должна найтись одна аптека рядом: %+v", got)
	}
	if got[0].Rating != 4.5 {
		t.Errorf("рейтинг прочитался неверно: %v", got[0].Rating)
	}
}

func TestGetPlace(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	shop := addPlace(t, s, Place{Kind: "shop", Name: "Магнит", Lat: 55.75, Lon: 37.62, Rating: 4, Reviews: 1})

	got, err := s.GetPlace(ctx, "shop", shop.ID)
	if err != nil || got.Name != "Магнит" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := s.GetPlace(ctx, "pharmacy", shop.ID); !errors.Is(err, ErrNotFound) {
		t.Error("магазин не должен находиться как аптека")
	}
	if _, err := s.GetPlace(ctx, "shop", 99999); !errors.Is(err, ErrNotFound) {
		t.Error("несуществующий магазин")
	}
}

func TestMedicineOffersFromDB(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := addPlace(t, s, Place{Kind: "pharmacy", Name: "А", Lat: 55.755, Lon: 37.622, Rating: 4.8, Reviews: 300})
	b := addPlace(t, s, Place{Kind: "pharmacy", Name: "Б", Lat: 55.760, Lon: 37.630, Rating: 4.0, Reviews: 300})
	far := addPlace(t, s, Place{Kind: "pharmacy", Name: "далеко", Lat: 55.90, Lon: 37.90, Rating: 5, Reviews: 300})
	addMedicinePrice(t, s, a.ID, "paracetamol", 50)
	addMedicinePrice(t, s, b.ID, "paracetamol", 40)
	addMedicinePrice(t, s, far.ID, "paracetamol", 10)
	addMedicinePrice(t, s, a.ID, "ibuprofen", 70)

	offers, err := s.MedicineOffers(ctx, "paracetamol", center.Lat, center.Lon, searchKm)
	if err != nil {
		t.Fatal(err)
	}
	offers = sortOffers(offers, offersLimit)
	if len(offers) != 2 {
		t.Fatalf("дальняя аптека должна отсеяться: %+v", offers)
	}
	if offers[0].Name != "А" || offers[0].Price != 50 {
		t.Errorf("первой должна быть аптека с лучшим рейтингом: %+v", offers[0])
	}
	if !offers[1].Cheapest || offers[0].Cheapest {
		t.Error("дешевле всего — аптека Б")
	}
}

func TestProductOffersAndPromos(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	shop := addPlace(t, s, Place{Kind: "shop", Name: "Пятёрочка", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 100})
	addProductPrice(t, s, shop.ID, "milk", 100, 70, "2026-10-05")
	addProductPrice(t, s, shop.ID, "bread-white", 50, 45, "2026-09-20")
	addProductPrice(t, s, shop.ID, "eggs", 120, 0, "")

	offers, err := s.ProductOffers(ctx, "milk", "2026-09-30", center.Lat, center.Lon, searchKm)
	if err != nil || len(offers) != 1 {
		t.Fatalf("%+v %v", offers, err)
	}
	if offers[0].Price != 70 || offers[0].OldPrice != 100 || offers[0].PromoUntil != "2026-10-05" {
		t.Errorf("акционная цена не применилась: %+v", offers[0])
	}

	offers, _ = s.ProductOffers(ctx, "bread-white", "2026-09-30", center.Lat, center.Lon, searchKm)
	if offers[0].Price != 50 || offers[0].OldPrice != 0 {
		t.Errorf("закончившаяся акция не должна действовать: %+v", offers[0])
	}

	promos, err := s.ShopPromos(ctx, shop.ID, "2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	if len(promos) != 1 || promos[0].ProductID != "milk" || promos[0].Discount != 30 || promos[0].Unit != "1 л" {
		t.Errorf("должна быть одна действующая акция на молоко: %+v", promos)
	}

	promos, _ = s.ShopPromos(ctx, shop.ID, "2026-10-06")
	if len(promos) != 0 {
		t.Error("после окончания акций список пустой")
	}
}

func TestRefreshPromos(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		shop := addPlace(t, s, Place{Kind: "shop", Name: "Магнит", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 100})
		for _, p := range products {
			addProductPrice(t, s, shop.ID, p.ID, 100, 0, "")
		}
	}

	if err := refreshPromos(ctx, s); err != nil {
		t.Fatal(err)
	}

	var bad, total int
	s.db.QueryRow(ctx, `SELECT count(*) FILTER (WHERE promo_price >= price OR promo_price < price * 0.6 OR promo_until < current_date),
		count(*) FILTER (WHERE promo_price IS NOT NULL) FROM product_prices`).Scan(&bad, &total)
	if bad != 0 {
		t.Errorf("у %d акций неверная цена или дата", bad)
	}
	if total == 0 {
		t.Error("после обновления должны появиться акции")
	}
}

func TestCountPlaces(t *testing.T) {
	s := testStore(t)
	addPlace(t, s, Place{Kind: "shop", Name: "м", Lat: 55.75, Lon: 37.62, Rating: 4, Reviews: 1})
	counts, err := s.CountPlaces(context.Background())
	if err != nil || counts["shop"] != 1 || counts["pharmacy"] != 0 {
		t.Errorf("%v %v", counts, err)
	}
}
