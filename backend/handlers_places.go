package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxQueryLength = 200
	suggestLimit   = 5
	birthdayLimit  = 5
)

type Suggestion struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Form      string `json:"form,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Matched   string `json:"matched"`
	Corrected bool   `json:"corrected"`
}

func (s *Server) userLocation(ctx context.Context, r *http.Request) (Location, error) {
	p, err := s.store.GetProfile(ctx, userID(r))
	if err != nil {
		return Location{}, err
	}
	return p.Location(), nil
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProfile(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) saveProfile(w http.ResponseWriter, r *http.Request) {
	var p Profile
	if err := decodeBody(w, r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	p, err := validateProfile(p, s.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.store.SaveProfile(r.Context(), userID(r), p); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) saveAddress(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Address string   `json:"address"`
		Lat     *float64 `json:"lat"`
		Lon     *float64 `json:"lon"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	address := strings.TrimSpace(body.Address)
	if address == "" {
		if err := s.store.SaveAddress(r.Context(), userID(r), "", nil, nil); err != nil {
			serverError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, defaultLocation)
		return
	}

	if utf8.RuneCountInString(address) > 300 {
		writeError(w, http.StatusBadRequest, "Адрес слишком длинный")
		return
	}
	if body.Lat == nil || body.Lon == nil || *body.Lat < -90 || *body.Lat > 90 || *body.Lon < -180 || *body.Lon > 180 {
		writeError(w, http.StatusBadRequest, "Выберите адрес из списка или на карте")
		return
	}

	if err := s.store.SaveAddress(r.Context(), userID(r), address, body.Lat, body.Lon); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, Location{Address: address, Lat: *body.Lat, Lon: *body.Lon})
}

func (s *Server) forYou(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProfile(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}

	items := []ForYouItem{}
	birth, err := time.Parse("2006-01-02", p.BirthDate)
	if err != nil {
		writeJSON(w, http.StatusOK, items)
		return
	}
	today := s.now()

	if item, ok := birthdayItem(birth, today); ok {
		loc := p.Location()
		places := []Place{}
		for _, kind := range []string{"shop", "pharmacy"} {
			list, err := s.store.PlacesNear(r.Context(), kind, loc.Lat, loc.Lon, searchKm)
			if err != nil {
				serverError(w, r, err)
				return
			}
			for _, pl := range list {
				if pl.BirthdayDiscount > 0 {
					places = append(places, pl)
				}
			}
		}
		item.Places = rankPlaces(places, loc.Lat, loc.Lon, searchKm)
		if len(item.Places) > birthdayLimit {
			item.Places = item.Places[:birthdayLimit]
		}
		items = append(items, item)
	}
	if item, ok := passportItem(birth, today); ok {
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) geoSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) < 3 {
		writeError(w, http.StatusBadRequest, "Напишите адрес подробнее")
		return
	}
	if utf8.RuneCountInString(q) > maxQueryLength {
		writeError(w, http.StatusBadRequest, "Адрес слишком длинный")
		return
	}

	places, err := s.geocoder.Search(r.Context(), q)
	if err != nil {
		log.Printf("geo search: %v", err)
		writeError(w, http.StatusBadGateway, "Сервис адресов сейчас недоступен")
		return
	}
	writeJSON(w, http.StatusOK, places)
}

func parseCoord(value string, limit float64) (float64, bool) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil || v < -limit || v > limit {
		return 0, false
	}
	return v, true
}

func (s *Server) geoReverse(w http.ResponseWriter, r *http.Request) {
	lat, ok1 := parseCoord(r.URL.Query().Get("lat"), 90)
	lon, ok2 := parseCoord(r.URL.Query().Get("lon"), 180)
	if !ok1 || !ok2 {
		writeError(w, http.StatusBadRequest, "Неверные координаты")
		return
	}

	place, err := s.geocoder.Reverse(r.Context(), lat, lon)
	if errors.Is(err, ErrPlaceNotFound) {
		writeError(w, http.StatusNotFound, "По этой точке адрес не найден")
		return
	}
	if err != nil {
		log.Printf("geo reverse: %v", err)
		writeError(w, http.StatusBadGateway, "Сервис адресов сейчас недоступен")
		return
	}
	writeJSON(w, http.StatusOK, place)
}

// сначала ищем сами, а если ничего не нашли — просим ИИ понять, что имелось в виду
func (s *Server) suggest(w http.ResponseWriter, r *http.Request, kind string, items []CatalogItem, toSuggestion func(id string) Suggestion) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) > maxQueryLength {
		writeError(w, http.StatusBadRequest, "Запрос слишком длинный")
		return
	}

	result := []Suggestion{}
	for _, m := range fuzzySuggest(items, q, suggestLimit) {
		sg := toSuggestion(m.ID)
		sg.Matched = m.Matched
		sg.Corrected = m.Corrected
		result = append(result, sg)
	}

	if len(result) == 0 && len(normalize(q)) >= 4 {
		if item, ok := s.aiCorrect(r.Context(), kind, items, q); ok {
			sg := toSuggestion(item.ID)
			sg.Matched = item.Name
			sg.Corrected = true
			result = append(result, sg)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) medicineSuggest(w http.ResponseWriter, r *http.Request) {
	s.suggest(w, r, "medicine", medicineItems(), func(id string) Suggestion {
		m, _ := findMedicine(id)
		return Suggestion{ID: m.ID, Name: m.Name, Form: m.Form}
	})
}

func (s *Server) productSuggest(w http.ResponseWriter, r *http.Request) {
	s.suggest(w, r, "product", productItems(), func(id string) Suggestion {
		p, _ := findProduct(id)
		return Suggestion{ID: p.ID, Name: p.Name, Unit: p.Unit}
	})
}

func (s *Server) nearby(w http.ResponseWriter, r *http.Request, kind, listName string) {
	loc, err := s.userLocation(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	places, err := s.store.PlacesNear(r.Context(), kind, loc.Lat, loc.Lon, searchKm)
	if err != nil {
		serverError(w, r, err)
		return
	}

	list := rankPlaces(places, loc.Lat, loc.Lon, searchKm)
	if len(list) > nearbyLimit {
		list = list[:nearbyLimit]
	}
	writeJSON(w, http.StatusOK, map[string]any{"location": loc, listName: list})
}

func (s *Server) nearbyPharmacies(w http.ResponseWriter, r *http.Request) {
	s.nearby(w, r, "pharmacy", "pharmacies")
}

func (s *Server) nearbyShops(w http.ResponseWriter, r *http.Request) {
	s.nearby(w, r, "shop", "shops")
}

func (s *Server) medicineOffers(w http.ResponseWriter, r *http.Request) {
	m, ok := findMedicine(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Лекарство не найдено")
		return
	}

	loc, err := s.userLocation(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	offers, err := s.store.MedicineOffers(r.Context(), m.ID, loc.Lat, loc.Lon, searchKm)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"medicine": m, "location": loc, "offers": sortOffers(offers, offersLimit)})
}

func (s *Server) productOffers(w http.ResponseWriter, r *http.Request) {
	p, ok := findProduct(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Товар не найден")
		return
	}

	loc, err := s.userLocation(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	offers, err := s.store.ProductOffers(r.Context(), p.ID, s.today(), loc.Lat, loc.Lon, searchKm)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"product": p, "location": loc, "offers": sortOffers(offers, offersLimit)})
}

func (s *Server) shopDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Неверный id магазина")
		return
	}

	shop, err := s.store.GetPlace(r.Context(), "shop", id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Магазин не найден")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	loc, err := s.userLocation(r.Context(), r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	promos, err := s.store.ShopPromos(r.Context(), id, s.today())
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"shop": rankPlace(shop, loc.Lat, loc.Lon), "promos": promos})
}
