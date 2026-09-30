package main

import (
	"fmt"
	"strconv"
	"time"
)

const (
	birthdaySoonDays   = 14
	passportBeforeDays = 30
	passportAfterDays  = 90
)

type ForYouItem struct {
	Type   string        `json:"type"`
	Title  string        `json:"title"`
	Text   string        `json:"text"`
	Places []RankedPlace `json:"places,omitempty"`
}

var monthsGenitive = []string{"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря"}

func formatDateRu(d time.Time) string {
	return fmt.Sprintf("%d %s %d", d.Day(), monthsGenitive[d.Month()-1], d.Year())
}

func daysWord(n int) string {
	if n%100 >= 11 && n%100 <= 14 {
		return strconv.Itoa(n) + " дней"
	}
	switch n % 10 {
	case 1:
		return strconv.Itoa(n) + " день"
	case 2, 3, 4:
		return strconv.Itoa(n) + " дня"
	}
	return strconv.Itoa(n) + " дней"
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(dateOnly(to).Sub(dateOnly(from)).Hours() / 24)
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// кто родился 29 февраля, в невисокосный год отмечает 28-го
func birthdayInYear(birth time.Time, year int) time.Time {
	if birth.Month() == time.February && birth.Day() == 29 && !isLeap(year) {
		return time.Date(year, time.February, 28, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(year, birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
}

func nextBirthday(birth, today time.Time) time.Time {
	d := birthdayInYear(birth, today.Year())
	if d.Before(dateOnly(today)) {
		d = birthdayInYear(birth, today.Year()+1)
	}
	return d
}

func birthdayItem(birth, today time.Time) (ForYouItem, bool) {
	days := daysBetween(today, nextBirthday(birth, today))
	if days > birthdaySoonDays {
		return ForYouItem{}, false
	}

	title := "До дня рождения " + daysWord(days)
	if days == 0 {
		title = "С днём рождения!"
	}
	return ForYouItem{
		Type:  "birthday",
		Title: title,
		Text:  "В этих местах рядом скидка именинникам. Обычно она действует несколько дней до и после праздника — возьмите с собой паспорт.",
	}, true
}

// паспорт меняют в 20 и 45 лет, на это даётся 90 дней после дня рождения
func passportItem(birth, today time.Time) (ForYouItem, bool) {
	for _, age := range []int{20, 45} {
		day := birthdayInYear(birth, birth.Year()+age)
		deadline := day.AddDate(0, 0, passportAfterDays)
		before := daysBetween(today, day)
		after := daysBetween(day, today)

		if before > 0 && before <= passportBeforeDays {
			return ForYouItem{
				Type:  "passport",
				Title: "Скоро менять паспорт",
				Text: fmt.Sprintf("%s вам исполнится %d. Паспорт нужно будет заменить до %s. Заявление можно подать на Госуслугах или в МФЦ.",
					formatDateRu(day), age, formatDateRu(deadline)),
			}, true
		}
		if after >= 0 && after <= passportAfterDays {
			return ForYouItem{
				Type:  "passport",
				Title: "Пора заменить паспорт",
				Text: fmt.Sprintf("Вам исполнилось %d. Паспорт нужно заменить до %s — осталось %s. Заявление можно подать на Госуслугах или в МФЦ.",
					age, formatDateRu(deadline), daysWord(passportAfterDays-after)),
			}, true
		}
	}
	return ForYouItem{}, false
}
