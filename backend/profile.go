package main

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Profile struct {
	Name         string  `json:"name"`
	BirthDate    string  `json:"birthDate"`
	Address      string  `json:"address"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	HasLocation  bool    `json:"hasLocation"`
	ContactName  string  `json:"contactName"`
	ContactPhone string  `json:"contactPhone"`
	Health       string  `json:"health"`

	RegAddress      string  `json:"regAddress"`
	RegLat          float64 `json:"regLat"`
	RegLon          float64 `json:"regLon"`
	HasRegistration bool    `json:"hasRegistration"`
	ClinicID        int64   `json:"clinicId"`
}

type Location struct {
	Address   string  `json:"address"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	IsDefault bool    `json:"isDefault"`
}

var defaultLocation = Location{Address: "Москва, центр", Lat: 55.7539, Lon: 37.6208, IsDefault: true}

// адрес прописки нужен для записи к врачу; если его нет, берём адрес проживания
func (p Profile) RegistrationLocation() (Location, bool) {
	if p.HasRegistration {
		return Location{Address: p.RegAddress, Lat: p.RegLat, Lon: p.RegLon}, true
	}
	return Location{}, false
}

func (p Profile) Location() Location {
	if !p.HasLocation {
		return defaultLocation
	}
	return Location{Address: p.Address, Lat: p.Lat, Lon: p.Lon}
}

func cleanPhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "", nil
	}

	digits := []rune{}
	for i, r := range phone {
		switch {
		case r >= '0' && r <= '9':
			digits = append(digits, r)
		case r == '+' && i == 0:
		case r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return "", errors.New("В телефоне могут быть только цифры")
		}
	}
	if len(digits) < 10 || len(digits) > 15 {
		return "", errors.New("Проверьте номер телефона")
	}
	return phone, nil
}

func validateBirthDate(value string, today time.Time) error {
	if value == "" {
		return nil
	}
	d, err := time.Parse("2006-01-02", value)
	if err != nil {
		return errors.New("Такой даты не существует")
	}
	if d.Year() < 1900 {
		return errors.New("Проверьте год рождения")
	}
	if d.After(today) {
		return errors.New("Дата рождения не может быть в будущем")
	}
	return nil
}

// сколько полных лет исполнится к дате today
func ageOn(birth, today time.Time) int {
	age := today.Year() - birth.Year()
	if today.Month() < birth.Month() || (today.Month() == birth.Month() && today.Day() < birth.Day()) {
		age--
	}
	return age
}

func validateProfile(p Profile, today time.Time) (Profile, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Address = strings.TrimSpace(p.Address)
	p.ContactName = strings.TrimSpace(p.ContactName)
	p.Health = strings.TrimSpace(p.Health)

	if utf8.RuneCountInString(p.Name) > 100 || utf8.RuneCountInString(p.ContactName) > 100 {
		return p, errors.New("Имя слишком длинное")
	}
	if err := validateBirthDate(p.BirthDate, today); err != nil {
		return p, err
	}
	if utf8.RuneCountInString(p.Address) > 300 {
		return p, errors.New("Адрес слишком длинный")
	}
	if utf8.RuneCountInString(p.Health) > 1000 {
		return p, errors.New("Текст о здоровье слишком длинный")
	}

	if p.HasLocation {
		if p.Lat < -90 || p.Lat > 90 || p.Lon < -180 || p.Lon > 180 {
			return p, errors.New("Неверные координаты")
		}
		if p.Address == "" {
			return p, errors.New("Не указан адрес")
		}
	} else {
		if p.Address != "" {
			return p, errors.New("Выберите адрес из списка или на карте")
		}
		p.Lat = 0
		p.Lon = 0
	}

	phone, err := cleanPhone(p.ContactPhone)
	if err != nil {
		return p, err
	}
	p.ContactPhone = phone
	return p, nil
}
