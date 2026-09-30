package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TaskDraft struct {
	Title string `json:"title"`
	Time  string `json:"time"`
	Date  string `json:"date"`
	Kind  string `json:"kind"`
}

var remindStems = []string{"напомн", "напомин", "не забыть", "не забудь", "добавь задач", "добавить задач",
	"добавьте задач", "поставь задач", "создай задач", "запланир", "будильник"}

var months = map[string]time.Month{
	"января": time.January, "февраля": time.February, "марта": time.March, "апреля": time.April,
	"мая": time.May, "июня": time.June, "июля": time.July, "августа": time.August,
	"сентября": time.September, "октября": time.October, "ноября": time.November, "декабря": time.December,
}

var weekdays = map[string]time.Weekday{
	"понедельник": time.Monday, "вторник": time.Tuesday, "среду": time.Wednesday, "четверг": time.Thursday,
	"пятницу": time.Friday, "субботу": time.Saturday, "воскресенье": time.Sunday,
}

// в Go \b не понимает русские буквы, поэтому границу слова пишем через (^|\s)
var (
	clockRe   = regexp.MustCompile(`(^|\s)(?:в|к|на)?\s*(\d{1,2})[:.\-](\d{2})(\s+(?:утра|дня|вечера|ночи))?`)
	hourRe    = regexp.MustCompile(`(^|\s)(?:в|к)\s+(\d{1,2})(\s+час(?:а|ов)?)?(\s+(?:утра|дня|вечера|ночи))?`)
	noonRe    = regexp.MustCompile(`(^|\s)в\s+полдень`)
	monthRe   = regexp.MustCompile(`(^|\s)(?:на\s+)?(\d{1,2})(?:-?го)?\s+(января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря)`)
	weekdayRe = regexp.MustCompile(`(^|\s)(?:в|во)\s+(понедельник|вторник|среду|четверг|пятницу|субботу|воскресенье)`)
	laterRe   = regexp.MustCompile(`(^|\s)через\s+(?:(\d{1,3})\s+)?(минуту|минуты|минут|часа|часов|час|полчаса)`)
	dayWordRe = regexp.MustCompile(`(^|\s)(?:на\s+)?(послезавтра|завтра|сегодня)`)
)

var taskFillers = map[string]bool{
	"напомни": true, "напомните": true, "напомнить": true, "напоминание": true, "мне": true, "меня": true,
	"пожалуйста": true, "добавь": true, "добавьте": true, "добавить": true, "поставь": true, "создай": true,
	"задачу": true, "запланируй": true, "запланировать": true, "не": true, "забыть": true, "забудь": true,
	"чтобы": true, "что": true, "про": true, "о": true, "об": true, "том": true, "нужно": true, "надо": true,
	"будильник": true, "на": true,
}

// «надо в среду в 10 сходить в поликлинику» — тоже задача, если названо время
func mentionsTime(text string) bool {
	text = strings.ReplaceAll(strings.ToLower(text), "ё", "е")
	return clockRe.MatchString(text) || hourRe.MatchString(text) || laterRe.MatchString(text)
}

func hourOfDay(hour int, part string) int {
	part = strings.TrimSpace(part)
	if (part == "дня" || part == "вечера") && hour < 12 {
		return hour + 12
	}
	if part == "ночи" && hour == 12 {
		return 0
	}
	return hour
}

func parseTaskTime(text string) (string, string) {
	if m := clockRe.FindStringSubmatchIndex(text); m != nil {
		h, _ := strconv.Atoi(text[m[4]:m[5]])
		min, _ := strconv.Atoi(text[m[6]:m[7]])
		part := ""
		if m[8] != -1 {
			part = text[m[8]:m[9]]
		}
		h = hourOfDay(h, part)
		if h < 24 && min < 60 {
			return formatClock(h, min), text[:m[0]] + " " + text[m[1]:]
		}
	}
	if m := hourRe.FindStringSubmatchIndex(text); m != nil {
		h, _ := strconv.Atoi(text[m[4]:m[5]])
		part := ""
		if m[8] != -1 {
			part = text[m[8]:m[9]]
		}
		h = hourOfDay(h, part)
		if h < 24 {
			return formatClock(h, 0), text[:m[0]] + " " + text[m[1]:]
		}
	}
	if loc := noonRe.FindStringIndex(text); loc != nil {
		return "12:00", text[:loc[0]] + " " + text[loc[1]:]
	}
	return "", text
}

func formatClock(h, m int) string {
	return strconv.Itoa(h/10) + strconv.Itoa(h%10) + ":" + strconv.Itoa(m/10) + strconv.Itoa(m%10)
}

// возвращает дату задачи и текст без слов про дату. Пустая дата — день не назван
func parseTaskDate(text string, now time.Time) (time.Time, bool, string) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if m := monthRe.FindStringSubmatchIndex(text); m != nil {
		day, _ := strconv.Atoi(text[m[4]:m[5]])
		month := months[text[m[6]:m[7]]]
		d := time.Date(now.Year(), month, day, 0, 0, 0, 0, now.Location())
		if d.Before(today) {
			d = d.AddDate(1, 0, 0)
		}
		if d.Day() == day {
			return d, true, text[:m[0]] + " " + text[m[1]:]
		}
	}
	if m := weekdayRe.FindStringSubmatchIndex(text); m != nil {
		want := weekdays[text[m[4]:m[5]]]
		days := (int(want) - int(today.Weekday()) + 7) % 7
		if days == 0 {
			days = 7
		}
		return today.AddDate(0, 0, days), true, text[:m[0]] + " " + text[m[1]:]
	}
	if m := dayWordRe.FindStringSubmatchIndex(text); m != nil {
		add := map[string]int{"сегодня": 0, "завтра": 1, "послезавтра": 2}[text[m[4]:m[5]]]
		return today.AddDate(0, 0, add), true, text[:m[0]] + " " + text[m[1]:]
	}
	return today, false, text
}

func taskKind(title string) string {
	switch {
	case containsAny(title, "таблет", "лекарств", "капл", "укол", "витамин", "давлен", "сахар", "лекарств"):
		return "medicine"
	case containsAny(title, "врач", "поликлиник", "анализ", "прием", "приём", "доктор"):
		return "doctor"
	case containsAny(title, "позвон", "звонок", "набрать"):
		return "call"
	}
	return "other"
}

func containsAny(text string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// «напомни завтра в 9 выпить таблетку» -> «Выпить таблетку», завтра, 09:00
func parseTask(text string, now time.Time) TaskDraft {
	text = strings.ReplaceAll(strings.ToLower(text), "ё", "е")
	date, dateSet, rest := parseTaskDate(text, now)
	clock, rest := parseTaskTime(rest)
	if m := laterRe.FindStringSubmatchIndex(rest); m != nil {
		n := 1
		if m[4] != -1 {
			n, _ = strconv.Atoi(rest[m[4]:m[5]])
		}
		delta := time.Duration(n) * time.Hour
		switch unit := rest[m[6]:m[7]]; {
		case unit == "полчаса":
			delta = 30 * time.Minute
		case strings.HasPrefix(unit, "минут"):
			delta = time.Duration(n) * time.Minute
		}
		at := now.Add(delta)
		clock, date, dateSet = at.Format("15:04"), time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location()), true
		rest = rest[:m[0]] + " " + rest[m[1]:]
	}

	words := strings.FieldsFunc(rest, func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == '!' || r == '?' || r == ':' || r == ';' || r == '"' || r == '«' || r == '»'
	})
	for len(words) > 0 && taskFillers[words[0]] {
		words = words[1:]
	}
	for len(words) > 0 && (taskFillers[words[len(words)-1]] || words[len(words)-1] == "в") {
		words = words[:len(words)-1]
	}
	title := strings.Join(words, " ")
	if title == "" {
		title = "Напоминание"
	}
	r := []rune(title)
	title = strings.ToUpper(string(r[0])) + string(r[1:])
	if len(r) > maxTitleLen {
		title = string([]rune(title)[:maxTitleLen])
	}

	// время уже прошло, а день не назван — значит, завтра
	if !dateSet && clock != "" && clock <= now.Format("15:04") {
		date = date.AddDate(0, 0, 1)
	}
	return TaskDraft{Title: title, Time: clock, Date: date.Format("2006-01-02"), Kind: taskKind(strings.ToLower(title))}
}
