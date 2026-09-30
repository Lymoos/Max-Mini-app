package main

import (
	"strings"
	"testing"
	"time"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestAgeOn(t *testing.T) {
	tests := []struct {
		birth, today string
		want         int
	}{
		{"1956-03-12", "2026-03-11", 69},
		{"1956-03-12", "2026-03-12", 70},
		{"1956-03-12", "2026-09-30", 70},
		{"2000-02-29", "2026-02-28", 25},
		{"2000-02-29", "2026-03-01", 26},
	}
	for _, tt := range tests {
		if got := ageOn(date(tt.birth), date(tt.today)); got != tt.want {
			t.Errorf("ageOn(%s, %s) = %d, ожидали %d", tt.birth, tt.today, got, tt.want)
		}
	}
}

func TestNextBirthday(t *testing.T) {
	tests := []struct {
		birth, today, want string
	}{
		{"1956-10-05", "2026-09-30", "2026-10-05"},
		{"1956-09-30", "2026-09-30", "2026-09-30"},
		{"1956-09-29", "2026-09-30", "2027-09-29"},
		{"1956-02-29", "2026-09-30", "2027-02-28"},
		{"1956-02-29", "2027-09-30", "2028-02-29"},
		{"1956-01-10", "2026-12-31", "2027-01-10"},
	}
	for _, tt := range tests {
		got := nextBirthday(date(tt.birth), date(tt.today))
		if got.Format("2006-01-02") != tt.want {
			t.Errorf("nextBirthday(%s, %s) = %s, ожидали %s", tt.birth, tt.today, got.Format("2006-01-02"), tt.want)
		}
	}
}

func TestBirthdayItem(t *testing.T) {
	tests := []struct {
		birth, today string
		wantOK       bool
		wantTitle    string
	}{
		{"1956-09-30", "2026-09-30", true, "С днём рождения!"},
		{"1956-10-01", "2026-09-30", true, "До дня рождения 1 день"},
		{"1956-10-03", "2026-09-30", true, "До дня рождения 3 дня"},
		{"1956-10-14", "2026-09-30", true, "До дня рождения 14 дней"},
		{"1956-10-15", "2026-09-30", false, ""},
		{"1956-09-29", "2026-09-30", false, ""},
		{"1956-01-05", "2026-12-25", true, "До дня рождения 11 дней"},
	}
	for _, tt := range tests {
		item, ok := birthdayItem(date(tt.birth), date(tt.today))
		if ok != tt.wantOK || item.Title != tt.wantTitle {
			t.Errorf("birthdayItem(%s, %s) = %q %v", tt.birth, tt.today, item.Title, ok)
		}
	}
}

func TestPassportItem(t *testing.T) {
	tests := []struct {
		name, birth, today string
		wantOK             bool
		wantText           string
	}{
		{"за месяц до 45", "1981-10-20", "2026-09-30", true, "20 октября 2026 вам исполнится 45"},
		{"в день 45-летия", "1981-09-30", "2026-09-30", true, "20 октября 2026 вам исполнится 45"},
		{"через 10 дней после 45", "1981-09-20", "2026-09-30", true, "осталось 80 дней"},
		{"последний день", "1981-07-02", "2026-09-30", true, "осталось 0 дней"},
		{"срок прошёл", "1981-07-01", "2026-09-30", false, ""},
		{"за 31 день до 45", "1981-10-31", "2026-09-30", false, ""},
		{"20 лет", "2006-10-10", "2026-09-30", true, "исполнится 20"},
		{"обычный возраст", "1956-03-12", "2026-09-30", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, ok := passportItem(date(tt.birth), date(tt.today))
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, текст %q", ok, item.Text)
			}
			if ok && tt.name == "в день 45-летия" {
				tt.wantText = "Вам исполнилось 45"
			}
			if ok && !strings.Contains(item.Text, tt.wantText) {
				t.Errorf("в тексте %q нет %q", item.Text, tt.wantText)
			}
		})
	}
}

func TestDaysWord(t *testing.T) {
	want := map[int]string{0: "0 дней", 1: "1 день", 2: "2 дня", 5: "5 дней", 11: "11 дней", 21: "21 день", 22: "22 дня", 112: "112 дней"}
	for n, w := range want {
		if got := daysWord(n); got != w {
			t.Errorf("daysWord(%d) = %q", n, got)
		}
	}
}

func TestValidateBirthDate(t *testing.T) {
	today := date("2026-09-30")
	for _, ok := range []string{"", "1956-03-12", "2000-02-29", "2026-09-30", "1900-01-01"} {
		if err := validateBirthDate(ok, today); err != nil {
			t.Errorf("%q должна проходить: %v", ok, err)
		}
	}
	for _, bad := range []string{"2026-10-01", "1899-12-31", "2001-02-29", "1956-04-31", "12.03.1956", "1956-3-12", "abc"} {
		if err := validateBirthDate(bad, today); err == nil {
			t.Errorf("%q не должна проходить", bad)
		}
	}
}
