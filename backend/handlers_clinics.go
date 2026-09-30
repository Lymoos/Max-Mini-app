package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (s *Server) saveRegistration(w http.ResponseWriter, r *http.Request) {
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
		if err := s.store.SaveRegistration(r.Context(), userID(r), "", nil, nil); err != nil {
			serverError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"hasRegistration": false})
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

	if err := s.store.SaveRegistration(r.Context(), userID(r), address, body.Lat, body.Lon); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"hasRegistration": true, "regAddress": address})
}

func (s *Server) saveClinic(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClinicID int64 `json:"clinicId"`
	}
	if err := decodeBody(w, r, &body); err != nil || body.ClinicID < 0 {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	var clinicID *int64
	if body.ClinicID > 0 {
		if _, err := s.store.GetPlace(r.Context(), "clinic", body.ClinicID); err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "Поликлиника не найдена")
				return
			}
			serverError(w, r, err)
			return
		}
		clinicID = &body.ClinicID
	}

	if err := s.store.SaveClinic(r.Context(), userID(r), clinicID); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"clinicId": body.ClinicID})
}

// сначала «ваша» поликлиника (выбранная вручную или ближайшая к прописке), потом клиники рядом с домом
func (s *Server) listClinics(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProfile(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}
	loc := p.Location()
	reg, hasReg := p.RegistrationLocation()

	var my *RankedPlace
	chosen := false
	if p.ClinicID > 0 {
		clinic, err := s.store.GetPlace(r.Context(), "clinic", p.ClinicID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			serverError(w, r, err)
			return
		}
		if err == nil {
			ranked := rankPlace(clinic, loc.Lat, loc.Lon)
			my = &ranked
			chosen = true
		}
	}
	if my == nil && hasReg {
		near, err := s.store.PlacesNear(r.Context(), "clinic", reg.Lat, reg.Lon, myClinicKm)
		if err != nil {
			serverError(w, r, err)
			return
		}
		if clinic, ok := nearestPolyclinic(near, reg.Lat, reg.Lon); ok {
			ranked := rankPlace(clinic, loc.Lat, loc.Lon)
			my = &ranked
		}
	}

	places, err := s.store.PlacesNear(r.Context(), "clinic", loc.Lat, loc.Lon, searchKm)
	if err != nil {
		serverError(w, r, err)
		return
	}
	nearby := []RankedPlace{}
	for _, c := range rankPlaces(places, loc.Lat, loc.Lon, searchKm) {
		if my != nil && c.ID == my.ID || skipClinicRe.MatchString(c.Name) {
			continue
		}
		nearby = append(nearby, c)
		if len(nearby) == nearbyLimit {
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"location":        loc,
		"hasRegistration": hasReg,
		"regAddress":      p.RegAddress,
		"myClinic":        my,
		"myClinicChosen":  chosen,
		"nearby":          nearby,
	})
}

func parseID(value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0
}

func (s *Server) clinicDetails(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "Неверный id поликлиники")
		return
	}

	clinic, err := s.store.GetPlace(r.Context(), "clinic", id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Поликлиника не найдена")
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
	doctors, err := s.store.Doctors(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}

	specialtyList := []string{}
	seen := map[string]bool{}
	for _, d := range doctors {
		if !seen[d.Specialty] {
			seen[d.Specialty] = true
			specialtyList = append(specialtyList, d.Specialty)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"clinic":      rankPlace(clinic, loc.Lat, loc.Lon),
		"doctors":     doctors,
		"specialties": specialtyList,
	})
}

func (s *Server) doctorDetails(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "Неверный id врача")
		return
	}

	doctor, err := s.store.GetDoctor(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Врач не найден")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	clinic, err := s.store.GetPlace(r.Context(), "clinic", doctor.ClinicID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	p, err := s.store.GetProfile(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}

	loc := p.Location()
	regionLoc, ok := p.RegistrationLocation()
	if !ok {
		regionLoc = Location{Lat: clinic.Lat, Lon: clinic.Lon}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"doctor":  doctor,
		"clinic":  rankPlace(clinic, loc.Lat, loc.Lon),
		"booking": bookingFor(regionLoc),
	})
}

func (s *Server) nearbySocial(w http.ResponseWriter, r *http.Request) {
	s.nearby(w, r, "social", "places")
}
