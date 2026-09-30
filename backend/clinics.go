package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

const myClinicKm = 5.0

type Doctor struct {
	ID          int64   `json:"id"`
	ClinicID    int64   `json:"clinicId"`
	Name        string  `json:"name"`
	Specialty   string  `json:"specialty"`
	SpecialtyID string  `json:"specialtyId"`
	Experience  int     `json:"experience"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	Reviews     int     `json:"reviews"`
}

type Booking struct {
	Region string `json:"region"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Phone  string `json:"phone"`
}

var stateClinicRe = regexp.MustCompile(`(?i)поликлиник|больниц|гбуз|гауз|гуз|муз|црб|медико-санитарн|госпитал`)
var notAdultClinicRe = regexp.MustCompile(`(?i)детск|стомат|женск|ветерин`)

// в списке для записи не показываем детские и ветеринарные клиники и лаборатории, где врачей нет
var skipClinicRe = regexp.MustCompile(`(?i)детск|ветерин|инвитро|invitro|гемотест|labquest|kdl|кдл|хеликс|helix|cmd|ситилаб|citilab|лаборатор|анализ|грот|массаж|салон`)

func isStateClinic(name, operator string) bool {
	return stateClinicRe.MatchString(name + " " + operator)
}

// «ваша» поликлиника — государственная взрослая поликлиника
func isAdultPolyclinic(p Place) bool {
	name := strings.ToLower(p.Name)
	return p.IsState && strings.Contains(name, "поликлиник") && !notAdultClinicRe.MatchString(name)
}

// ближайшая взрослая поликлиника к адресу прописки. Настоящую зону прикрепления узнать неоткуда,
// поэтому пользователь может выбрать свою поликлинику вручную
func nearestPolyclinic(list []Place, lat, lon float64) (Place, bool) {
	best := -1
	bestDist := 0.0
	for i, p := range list {
		if !isAdultPolyclinic(p) {
			continue
		}
		d := haversineKm(lat, lon, p.Lat, p.Lon)
		if d > myClinicKm {
			continue
		}
		if best == -1 || d < bestDist {
			best = i
			bestDist = d
		}
	}
	if best == -1 {
		return Place{}, false
	}
	return list[best], true
}

func inBox(lat, lon, south, west, north, east float64) bool {
	return lat >= south && lat <= north && lon >= west && lon <= east
}

// запись идёт через портал региона, где человек прикреплён по полису
func bookingFor(loc Location) Booking {
	switch {
	case inBox(loc.Lat, loc.Lon, 55.14, 36.80, 56.02, 37.97):
		return Booking{Region: "moscow", Title: "Записаться в ЕМИАС", URL: "https://emias.info", Phone: "122"}
	case inBox(loc.Lat, loc.Lon, 53.97, 47.25, 56.68, 54.27):
		return Booking{Region: "tatarstan", Title: "Записаться на Госуслугах Татарстана", URL: "https://uslugi.tatarstan.ru/mis/tatarstan/init", Phone: "122"}
	}
	return Booking{Region: "other", Title: "Записаться на Госуслугах", URL: "https://www.gosuslugi.ru/10066/1", Phone: "122"}
}

var specialties = []string{
	"Терапевт", "Терапевт", "Терапевт", "Хирург", "Кардиолог", "Невролог", "Офтальмолог",
	"Оториноларинголог (ЛОР)", "Эндокринолог", "Уролог", "Гастроэнтеролог", "Травматолог-ортопед",
	"Дерматолог", "Пульмонолог", "Ревматолог", "Гериатр",
}

var maleSurnames = []string{"Иванов", "Смирнов", "Кузнецов", "Попов", "Васильев", "Петров", "Соколов", "Михайлов",
	"Новиков", "Фёдоров", "Морозов", "Волков", "Алексеев", "Лебедев", "Семёнов", "Егоров", "Павлов", "Козлов",
	"Степанов", "Николаев", "Орлов", "Андреев", "Макаров", "Никитин", "Захаров", "Зайцев", "Соловьёв", "Борисов",
	"Яковлев", "Григорьев", "Романов", "Воробьёв", "Сергеев", "Фролов", "Александров", "Гусев", "Титов", "Кудрявцев",
	"Баранов", "Куликов", "Хабибуллин", "Галиев", "Сафин", "Закиров", "Гарипов", "Шарипов"}
var maleNames = []string{"Александр", "Сергей", "Дмитрий", "Андрей", "Алексей", "Максим", "Евгений", "Игорь",
	"Михаил", "Владимир", "Николай", "Олег", "Павел", "Роман", "Виктор", "Ильдар", "Рустам", "Айдар"}
var femaleNames = []string{"Елена", "Ольга", "Наталья", "Татьяна", "Ирина", "Светлана", "Анна", "Мария",
	"Екатерина", "Юлия", "Марина", "Галина", "Людмила", "Надежда", "Гульнара", "Альбина", "Лилия", "Эльвира"}
var fatherNames = []string{"Александр", "Сергей", "Владимир", "Николай", "Иван", "Михаил", "Андрей", "Виктор",
	"Пётр", "Анатолий", "Ренат", "Рашид"}

func patronymic(father string, female bool) string {
	stem := strings.TrimSuffix(father, "й")
	if stem != father {
		if female {
			return stem + "евна"
		}
		return stem + "евич"
	}
	if strings.HasSuffix(father, "ётр") {
		stem = strings.TrimSuffix(father, "ётр") + "етр"
	} else {
		stem = father
	}
	if female {
		return stem + "овна"
	}
	return stem + "ович"
}

func femaleSurname(male string) string {
	if strings.HasSuffix(male, "ий") {
		return strings.TrimSuffix(male, "ий") + "ая"
	}
	return male + "а"
}

func demoDoctorName(r *rand.Rand) string {
	surname := maleSurnames[r.Intn(len(maleSurnames))]
	father := fatherNames[r.Intn(len(fatherNames))]
	if r.Intn(2) == 0 {
		return fmt.Sprintf("%s %s %s", femaleSurname(surname), femaleNames[r.Intn(len(femaleNames))], patronymic(father, true))
	}
	return fmt.Sprintf("%s %s %s", surname, maleNames[r.Intn(len(maleNames))], patronymic(father, false))
}

func demoDoctors(r *rand.Rand, placeID int64) []Doctor {
	count := 6 + r.Intn(9)
	list := []Doctor{}
	for i := 0; i < count; i++ {
		experience := 2 + r.Intn(38)
		category := ""
		switch {
		case experience >= 15 && r.Float64() < 0.7:
			category = "Высшая категория"
		case experience >= 7 && r.Float64() < 0.6:
			category = "Первая категория"
		case experience >= 3 && r.Float64() < 0.4:
			category = "Вторая категория"
		}
		list = append(list, Doctor{
			ClinicID:   placeID,
			Name:       demoDoctorName(r),
			Specialty:  specialties[r.Intn(len(specialties))],
			Experience: experience,
			Category:   category,
			Rating:     math.Round((4.0+r.Float64())*10) / 10,
			Reviews:    r.Intn(250),
		})
	}
	return list
}

func (s *Store) Doctors(ctx context.Context, placeID int64) ([]Doctor, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, place_id, name, specialty, experience, category, rating, reviews
		 FROM doctors WHERE place_id = $1 ORDER BY specialty, rating DESC, name`,
		placeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Doctor{}
	for rows.Next() {
		var d Doctor
		var rating float32
		if err := rows.Scan(&d.ID, &d.ClinicID, &d.Name, &d.Specialty, &d.Experience, &d.Category, &rating, &d.Reviews); err != nil {
			return nil, err
		}
		d.Rating = float64(rating)
		d.SpecialtyID = specialtyIDByName(d.Specialty)
		list = append(list, d)
	}
	return list, rows.Err()
}

func (s *Store) GetDoctor(ctx context.Context, id int64) (Doctor, error) {
	var d Doctor
	var rating float32
	err := s.db.QueryRow(ctx,
		`SELECT id, place_id, name, specialty, experience, category, rating, reviews FROM doctors WHERE id = $1`,
		id,
	).Scan(&d.ID, &d.ClinicID, &d.Name, &d.Specialty, &d.Experience, &d.Category, &rating, &d.Reviews)
	if errors.Is(err, pgx.ErrNoRows) {
		return Doctor{}, ErrNotFound
	}
	d.Rating = float64(rating)
	d.SpecialtyID = specialtyIDByName(d.Specialty)
	return d, err
}
