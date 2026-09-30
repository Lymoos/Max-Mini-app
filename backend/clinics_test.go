package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestIsStateClinic(t *testing.T) {
	yes := []struct{ name, operator string }{
		{"Городская поликлиника № 5", ""},
		{"Поликлиника", ""},
		{"Филиал №2", "ГБУЗ ГП № 62 ДЗМ"},
		{"Городская клиническая больница № 1", ""},
		{"ГАУЗ Городская поликлиника № 21", ""},
	}
	for _, c := range yes {
		if !isStateClinic(c.name, c.operator) {
			t.Errorf("%q / %q должна считаться государственной", c.name, c.operator)
		}
	}
	for _, name := range []string{"Медси", "СМ-Клиника", "Клиника Здоровье", "Инвитро"} {
		if isStateClinic(name, "") {
			t.Errorf("%q — частная клиника", name)
		}
	}
}

func TestIsAdultPolyclinic(t *testing.T) {
	cases := map[string]bool{
		"Городская поликлиника № 5":            true,
		"Детская городская поликлиника № 32":   false,
		"Стоматологическая поликлиника № 9":    false,
		"Городская клиническая больница № 1":   false,
		"Женская консультация при поликлинике": false,
	}
	for name, want := range cases {
		if got := isAdultPolyclinic(Place{Name: name, IsState: true}); got != want {
			t.Errorf("%q: %v, ожидали %v", name, got, want)
		}
	}
	if isAdultPolyclinic(Place{Name: "Поликлиника", IsState: false}) {
		t.Error("частная поликлиника не подходит")
	}
}

func TestNearestPolyclinic(t *testing.T) {
	list := []Place{
		{ID: 1, Name: "Городская поликлиника № 1", IsState: true, Lat: 55.760, Lon: 37.620},
		{ID: 2, Name: "Городская поликлиника № 2", IsState: true, Lat: 55.752, Lon: 37.621},
		{ID: 3, Name: "Детская поликлиника № 3", IsState: true, Lat: 55.7539, Lon: 37.6208},
		{ID: 4, Name: "Медси", IsState: false, Lat: 55.7539, Lon: 37.6208},
	}
	got, ok := nearestPolyclinic(list, 55.7539, 37.6208)
	if !ok || got.ID != 2 {
		t.Errorf("ближайшая взрослая — №2, получили %+v", got)
	}

	if _, ok := nearestPolyclinic(list, 59.93, 30.33); ok {
		t.Error("дальше 5 км свою поликлинику не выбираем")
	}
	if _, ok := nearestPolyclinic(nil, 55.75, 37.62); ok {
		t.Error("пустой список")
	}
}

func TestBookingFor(t *testing.T) {
	tests := []struct {
		name   string
		loc    Location
		region string
		url    string
	}{
		{"Москва", Location{Lat: 55.75, Lon: 37.62}, "moscow", "https://emias.info"},
		{"Зеленоград", Location{Lat: 55.99, Lon: 37.19}, "moscow", "https://emias.info"},
		{"Казань", Location{Lat: 55.79, Lon: 49.12}, "tatarstan", "https://uslugi.tatarstan.ru/mis/tatarstan/init"},
		{"Набережные Челны", Location{Lat: 55.74, Lon: 52.40}, "tatarstan", "https://uslugi.tatarstan.ru/mis/tatarstan/init"},
		{"Петербург", Location{Lat: 59.93, Lon: 30.33}, "other", "https://www.gosuslugi.ru/10066/1"},
		{"Самара", Location{Lat: 53.20, Lon: 50.15}, "other", "https://www.gosuslugi.ru/10066/1"},
	}
	for _, tt := range tests {
		got := bookingFor(tt.loc)
		if got.Region != tt.region || got.URL != tt.url || got.Phone != "122" || got.Title == "" {
			t.Errorf("%s: %+v", tt.name, got)
		}
	}
}

func TestDoctorNames(t *testing.T) {
	cases := map[string][2]string{
		"Николай":   {"Николаевич", "Николаевна"},
		"Андрей":    {"Андреевич", "Андреевна"},
		"Сергей":    {"Сергеевич", "Сергеевна"},
		"Иван":      {"Иванович", "Ивановна"},
		"Пётр":      {"Петрович", "Петровна"},
		"Анатолий":  {"Анатолиевич", "Анатолиевна"},
		"Александр": {"Александрович", "Александровна"},
	}
	for father, want := range cases {
		if got := patronymic(father, false); got != want[0] {
			t.Errorf("%s → %s", father, got)
		}
		if got := patronymic(father, true); got != want[1] {
			t.Errorf("%s → %s", father, got)
		}
	}
	if femaleSurname("Иванов") != "Иванова" || femaleSurname("Хабибуллин") != "Хабибуллина" {
		t.Error("женская фамилия")
	}
}

func TestDemoDoctors(t *testing.T) {
	list := demoDoctors(placeRand("node/1"), 7)
	again := demoDoctors(placeRand("node/1"), 7)
	if len(list) < 6 || len(list) > 14 {
		t.Fatalf("врачей %d", len(list))
	}
	for i, d := range list {
		if d != again[i] {
			t.Error("при повторном импорте врачи должны совпадать")
		}
		if d.ClinicID != 7 || len(strings.Fields(d.Name)) != 3 || d.Specialty == "" {
			t.Errorf("плохой врач %+v", d)
		}
		if d.Experience < 2 || d.Experience > 39 || d.Rating < 4 || d.Rating > 5 {
			t.Errorf("стаж или рейтинг вне границ: %+v", d)
		}
		if d.Category == "Высшая категория" && d.Experience < 15 {
			t.Errorf("высшая категория при стаже %d", d.Experience)
		}
	}
}

func TestDoctorsStore(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	clinic := addPlace(t, s, Place{Kind: "clinic", Name: "ГП №5", Lat: 55.75, Lon: 37.62, Rating: 4, Reviews: 1})
	addDoctor(t, s, Doctor{ClinicID: clinic.ID, Name: "Б", Specialty: "Терапевт", Experience: 10, Rating: 4.5, Reviews: 1})
	a := addDoctor(t, s, Doctor{ClinicID: clinic.ID, Name: "А", Specialty: "Хирург", Experience: 20, Category: "Высшая категория", Rating: 4.9, Reviews: 3})

	list, err := s.Doctors(ctx, clinic.ID)
	if err != nil || len(list) != 2 || list[0].Specialty != "Терапевт" {
		t.Fatalf("%+v %v", list, err)
	}

	got, err := s.GetDoctor(ctx, a.ID)
	if err != nil || got.Name != "А" || got.Category != "Высшая категория" || got.Rating != float64(float32(4.9)) {
		t.Errorf("%+v %v", got, err)
	}
	if _, err := s.GetDoctor(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Error("несуществующий врач")
	}
}

func TestRegistrationAndClinicStore(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	s.SaveProfile(ctx, "u1", Profile{Name: "Анна"})
	clinic := addPlace(t, s, Place{Kind: "clinic", Name: "ГП №5", Lat: 55.75, Lon: 37.62, Rating: 4, Reviews: 1})

	lat, lon := 55.76, 37.63
	s.SaveRegistration(ctx, "u1", "Мясницкая, 20", &lat, &lon)
	s.SaveClinic(ctx, "u1", &clinic.ID)

	p, _ := s.GetProfile(ctx, "u1")
	if !p.HasRegistration || p.RegAddress != "Мясницкая, 20" || p.ClinicID != clinic.ID || p.Name != "Анна" {
		t.Errorf("%+v", p)
	}

	s.SaveProfile(ctx, "u1", Profile{Name: "Анна Петровна"})
	p, _ = s.GetProfile(ctx, "u1")
	if !p.HasRegistration || p.ClinicID != clinic.ID {
		t.Error("сохранение профиля не должно стирать прописку и поликлинику")
	}

	s.SaveClinic(ctx, "u1", nil)
	s.SaveRegistration(ctx, "u1", "", nil, nil)
	p, _ = s.GetProfile(ctx, "u1")
	if p.HasRegistration || p.ClinicID != 0 {
		t.Errorf("должно сброситься: %+v", p)
	}
}

func TestSkipClinic(t *testing.T) {
	skip := []string{"Инвитро", "Гемотест", "LabQuest", "KDL", "Хеликс", "Детская поликлиника № 1", "Ветеринарная клиника", "Лаборатория анализов", "Ситилаб", "Соляной грот"}
	keep := []string{"Городская поликлиника № 2", "СМ-Клиника", "Медси", "Гинекологическая больница № 5"}
	for _, name := range skip {
		if !skipClinicRe.MatchString(name) {
			t.Errorf("%q не должна быть в списке для записи", name)
		}
	}
	for _, name := range keep {
		if skipClinicRe.MatchString(name) {
			t.Errorf("%q должна остаться", name)
		}
	}
}
