package main

import (
	"strconv"
	"strings"
	"time"
)

type Feature struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

var features = []Feature{
	{ID: "pharmacy", Title: "Аптеки рядом", Icon: "pill", Color: "green"},
	{ID: "goods", Title: "Товары рядом", Icon: "cart", Color: "blue"},
	{ID: "doctor", Title: "Запись к врачу", Icon: "stethoscope", Color: "rose"},
	{ID: "social", Title: "Соцпомощь", Icon: "heart", Color: "orange"},
}

func findFeature(id string) (Feature, bool) {
	for _, f := range features {
		if f.ID == id {
			return f, true
		}
	}
	return Feature{}, false
}

// Message — что можно сделать, Button — надпись на кнопке перехода
type AskAnswer struct {
	Type      string     `json:"type"`
	Target    string     `json:"target,omitempty"`
	Message   string     `json:"message"`
	Button    string     `json:"button,omitempty"`
	Title     string     `json:"title,omitempty"`
	Feature   *Feature   `json:"feature,omitempty"`
	Medicine  *Medicine  `json:"medicine,omitempty"`
	Product   *Product   `json:"product,omitempty"`
	Specialty *Specialty `json:"specialty,omitempty"`
	Task      *TaskDraft `json:"task,omitempty"`
}

// данные для кнопки бота «открыть приложение», тот же формат понимает frontend/src/startParam.js
func (a AskAnswer) payload() string {
	switch a.Type {
	case "medicine":
		return "med_" + a.Target
	case "product":
		return "prod_" + a.Target
	case "doctor":
		if a.Target != "" {
			return "doctor_" + a.Target
		}
		return "doctor"
	case "benefit":
		return "benefit_" + a.Target
	case "guide":
		return "guide_" + a.Target
	case "feature", "tab":
		return a.Target
	case "profile":
		return "profile"
	}
	return ""
}

func medicineAnswer(id string) AskAnswer {
	m, _ := findMedicine(id)
	return AskAnswer{Type: "medicine", Target: m.ID, Medicine: &m, Button: "Найти в аптеках",
		Message: "Покажу, в каких аптеках рядом есть «" + m.Name + "» и где дешевле."}
}

func productAnswer(id string) AskAnswer {
	p, _ := findProduct(id)
	return AskAnswer{Type: "product", Target: p.ID, Product: &p, Button: "Найти в магазинах",
		Message: "Покажу магазины рядом, где есть «" + p.Name + "», с ценами и скидками."}
}

func featureAnswer(id string) AskAnswer {
	f, _ := findFeature(id)
	a := AskAnswer{Type: "feature", Target: f.ID, Feature: &f}
	switch id {
	case "pharmacy":
		a.Message, a.Button = "Покажу аптеки рядом с вами: адрес, рейтинг и как дойти.", "Открыть аптеки"
	case "goods":
		a.Message, a.Button = "Покажу магазины рядом и товары со скидкой. Можно сразу найти нужный продукт.", "Открыть магазины"
	case "social":
		a.Message, a.Button = "Покажу центры соцпомощи рядом с телефонами. Там помогают с сиделкой, продуктами и документами.", "Открыть соцпомощь"
	}
	return a
}

func doctorAnswer(specialtyID string) AskAnswer {
	f, _ := findFeature("doctor")
	sp, ok := findSpecialty(specialtyID)
	if !ok {
		return AskAnswer{Type: "doctor", Feature: &f, Button: "Записаться к врачу",
			Message: "Помогу записаться к врачу: покажу вашу поликлинику по прописке и врачей в ней."}
	}
	return AskAnswer{Type: "doctor", Target: sp.ID, Feature: &f, Specialty: &sp, Button: "Записаться к " + sp.Dative,
		Message: "Помогу записаться к " + sp.Dative + ": открою вашу поликлинику и сразу покажу врачей этой специальности."}
}

func benefitAnswer(id string) AskAnswer {
	b, _ := findBenefit(id)
	return AskAnswer{Type: "benefit", Target: b.ID, Title: b.Title, Button: "Как оформить",
		Message: "«" + b.Title + "». " + strings.TrimSuffix(b.Short, ".") + ". Расскажу, кому положено и как оформить."}
}

func guideAnswer(id string) AskAnswer {
	g, _ := findGuide(id)
	return AskAnswer{Type: "guide", Target: g.ID, Title: g.Title, Button: "Открыть инструкцию",
		Message: "Есть инструкция «" + g.Title + "» — по шагам, отдельно для Android и iPhone."}
}

var tabAnswers = map[string]AskAnswer{
	"medicines": {Type: "tab", Target: "medicines", Button: "Искать лекарство",
		Message: "Напишите название лекарства — покажу, в какой аптеке рядом оно есть и где дешевле."},
	"documents": {Type: "tab", Target: "documents", Button: "Открыть документы",
		Message: "Покажу документы и льготы, которые вам положены, и как их оформить."},
	"help": {Type: "tab", Target: "help", Button: "Открыть инструкции",
		Message: "Вот инструкции: как позвонить, сделать буквы крупнее, не попасться мошенникам. Если очень плохо — звоните 103 или 112."},
}

// ИИ и правила не назначают лечение: только раздел лекарств и совет сходить к врачу
func treatmentAnswer() AskAnswer {
	return AskAnswer{Type: "tab", Target: "medicines", Button: "Открыть лекарства",
		Message: "Какое лекарство принимать, подскажет только врач. Если препарат вам уже назначили, найду его в аптеках рядом."}
}

func profileAnswer() AskAnswer {
	return AskAnswer{Type: "profile", Button: "Открыть профиль",
		Message: "Откроем профиль: там имя, адрес, прописка и телефон близкого человека."}
}

func unknownAnswer() AskAnswer {
	return AskAnswer{Type: "unknown",
		Message: "Не понял запрос. Напишите, например: «аптека рядом», «нужен кардиолог» или «напомни выпить таблетку в 9»."}
}

var monthNames = []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}

func humanDate(date string, now time.Time) string {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	switch date {
	case now.Format("2006-01-02"):
		return "сегодня"
	case now.AddDate(0, 0, 1).Format("2006-01-02"):
		return "завтра"
	}
	return strconv.Itoa(d.Day()) + " " + monthNames[d.Month()-1]
}

func taskAnswer(t TaskDraft, now time.Time) AskAnswer {
	a := AskAnswer{Type: "task", Task: &t, Button: "Добавить"}
	when := humanDate(t.Date, now)
	if t.Time == "" {
		a.Message = "Добавлю задачу «" + t.Title + "» на " + when + ". Выберите время — и я напомню."
	} else {
		a.Message = "Добавить задачу «" + t.Title + "» на " + when + " в " + t.Time + "? Напомню в это время."
	}
	return a
}

// слова запроса в нижнем регистре. Слово, набранное латиницей, переводим в русскую раскладку: fgntrf -> аптека
type askText struct {
	words  []string
	joined string
}

var latinWords = map[string]bool{"wifi": true, "wi-fi": true, "max": true, "sms": true, "ok": true}

func prepareAsk(text string) askText {
	text = strings.ReplaceAll(strings.ToLower(text), "ё", "е")
	words := []string{}
	for _, token := range strings.Fields(text) {
		if isLatin(token) && !latinWords[strings.Trim(token, ",.!?")] {
			token, _ = fixLayout(token)
		}
		parts := strings.FieldsFunc(token, func(r rune) bool {
			return r == ',' || r == '!' || r == '?' || r == '"' || r == '«' || r == '»' || r == '(' || r == ')'
		})
		for _, w := range parts {
			if w = strings.Trim(w, ".:;-"); w != "" {
				words = append(words, w)
			}
		}
	}
	return askText{words: words, joined: " " + strings.Join(words, " ") + " "}
}

// слово набрано в английской раскладке: только латиница и знаки, под которыми на клавиатуре русские буквы
func isLatin(w string) bool {
	hasLetter := false
	for _, r := range w {
		switch {
		case r >= 'a' && r <= 'z':
			hasLetter = true
		case strings.ContainsRune(",.;'[]`-", r):
		default:
			return false
		}
	}
	return hasLetter
}

func (q askText) String() string {
	return strings.TrimSpace(q.joined)
}

// основа совпадает с началом слова. В основах от 5 букв прощаем одну опечатку: «аптка», «кордиолог».
// Для коротких основ только в коротких словах, иначе «кружится» станет «кружком»
func (q askText) hasStem(stem string) bool {
	s := []rune(stem)
	for _, w := range q.words {
		if strings.HasPrefix(w, stem) {
			return true
		}
		r := []rune(w)
		if len(s) < 5 || len(r) < len(s)-1 || r[0] != s[0] || (len(s) == 5 && len(r) > 6) {
			continue
		}
		for _, n := range []int{len(s) - 1, len(s), len(s) + 1} {
			if n <= len(r) && typoDistance(r[:n], s) <= 1 {
				return true
			}
		}
	}
	return false
}

// фразы из нескольких слов ищем точно, с начала слова
func (q askText) has(stem string) bool {
	if strings.Contains(stem, " ") {
		return strings.Contains(q.joined, " "+stem)
	}
	return q.hasStem(stem)
}

func (q askText) hasAny(stems ...string) bool {
	for _, s := range stems {
		if q.has(s) {
			return true
		}
	}
	return false
}

func (q askText) hasAll(stems []string) bool {
	for _, s := range stems {
		if !q.has(s) {
			return false
		}
	}
	return true
}

// правило срабатывает, если в запросе есть все основы хотя бы из одной группы
type phraseRule struct {
	target string
	groups [][]string
}

func matchRules(q askText, rules []phraseRule) (string, bool) {
	for _, r := range rules {
		for _, g := range r.groups {
			if q.hasAll(g) {
				return r.target, true
			}
		}
	}
	return "", false
}

// порядок важен: видеозвонок раньше обычного звонка
var guideRules = []phraseRule{
	{"max-video", [][]string{{"видеозвон"}, {"видео", "звон"}, {"по видео"}}},
	{"max-photo", [][]string{{"фото", "отправ"}, {"фотограф", "отправ"}, {"фото", "max"}, {"фото", "макс"}, {"картинк", "отправ"}, {"снимок", "отправ"}}},
	{"scam", [][]string{{"мошен"}, {"обманыва"}, {"развод", "звон"}, {"безопасный счет"}, {"код из смс"}, {"из банка", "звон"}}},
	{"answer", [][]string{{"ответить", "звон"}, {"принять звон"}, {"взять трубк"}, {"поднять трубк"}, {"отклонить", "звон"}}},
	{"call", [][]string{{"как позвонить"}, {"как звонить"}, {"набрать номер"}, {"как набрать"}}},
	{"font", [][]string{{"шрифт"}, {"буквы", "крупн"}, {"буквы", "больш"}, {"буквы", "мелк"}, {"текст", "крупн"}, {"увеличить текст"}, {"плохо видно", "экран"}, {"мелко"}}},
	{"flashlight", [][]string{{"фонарик", "включ"}, {"фонарик", "как"}, {"фонарик", "телефон"}, {"фонарь", "включ"}}},
	{"wifi", [][]string{{"wifi"}, {"wi-fi"}, {"вайфай"}, {"вай фай"}, {"вай-фай"}, {"подключ", "интернет"}, {"нет интернета"}}},
	{"gosuslugi", [][]string{{"госуслуг", "постав"}, {"госуслуг", "установ"}, {"госуслуг", "скача"}, {"госуслуг", "приложен"}, {"госуслуг", "зарегистр"}, {"госуслуг", "как"}}},
	{"this-app", [][]string{{"как пользоваться"}, {"что ты умеешь"}, {"что умеет"}, {"как работает приложение"}}},
}

var benefitRules = []phraseRule{
	{"overhaul", [][]string{{"капремонт"}, {"капитальн", "ремонт"}, {"взнос", "ремонт"}}},
	{"moscow-card", [][]string{{"карта москвича"}, {"карту москвича"}, {"карты москвича"}, {"социальная карта"}, {"социальную карту"}, {"соцкарт"}}},
	{"pension-80", [][]string{{"80 лет"}, {"восемьдесят"}, {"пенси", "80"}}},
	{"social-supplement", [][]string{{"прожиточн"}, {"доплат", "пенси"}, {"маленькая пенсия"}, {"пенсия маленькая"}, {"низкая пенсия"}}},
	{"property-tax", [][]string{{"налог"}}},
	{"social-services-set", [][]string{{"набор социальных услуг"}, {"нсу"}, {"бесплатн", "лекарств"}, {"бесплатн", "путевк"}, {"санатор"}}},
	{"housing-subsidy", [][]string{{"субсиди"}, {"жку"}, {"жкх"}, {"коммунал"}, {"квартплат"}}},
	{"labour-veteran", [][]string{{"ветеран труда"}, {"ветерана труда"}}},
	{"moscow-longevity", [][]string{{"долголет"}, {"кружк"}, {"бесплатн", "заняти"}}},
	{"passport", [][]string{{"паспорт", "замен"}, {"паспорт", "поменя"}, {"паспорт", "45"}, {"паспорт", "срок"}, {"паспорт", "просроч"}}},
}

var doctorStems = []string{"врач", "доктор", "поликлиник", "запис", "на прием", "к кому", "специалист", "талон"}

func specialtyInText(q askText) (string, bool) {
	// «глазные капли» — это лекарство, а не глазник
	if q.hasAny(medWords...) && !q.hasAny(doctorStems...) {
		return "", false
	}
	for _, sp := range doctorSpecialties {
		for _, a := range sp.Aliases {
			if q.has(a) {
				return sp.ID, true
			}
		}
	}
	if !q.hasAny(doctorStems...) {
		return "", false
	}
	for _, sp := range doctorSpecialties {
		for _, s := range sp.Symptoms {
			if q.has(s) {
				return sp.ID, true
			}
		}
	}
	return "", false
}

var treatVerbs = []string{"что выпить", "что пить", "что попить", "что можно выпить", "что принять", "что принимать",
	"чем лечить", "чем лечиться", "чем сбить", "чем снизить", "чем понизить", "чем помочь", "что помогает", "что поможет",
	"посоветуй", "подскажи лекарств", "подскажите лекарств", "какое лекарство", "какие лекарства", "какие таблетки",
	"какую таблетку", "какие капли", "какую мазь"}
var medWords = []string{"таблет", "лекарств", "средств", "препарат", "капли", "мазь", "пилюл"}
var symptomWords = []string{"давлен", "сердц", "бол", "кашл", "кашел", "температур", "простуд", "аллерг", "изжог",
	"запор", "понос", "бессонниц", "насморк", "головы", "живот", "желуд", "сустав", "спины", "нервов"}

func isTreatmentQuestion(q askText) bool {
	if q.hasAny(treatVerbs...) {
		return true
	}
	return strings.Contains(q.joined, " от ") && q.hasAny(append(medWords, symptomWords...)...)
}

type askRule struct {
	stems []string
	pick  func() AskAnswer
}

// порядок важен: «купить лекарство» — это лекарства, а не товары
var askRules = []askRule{
	{[]string{"аптек"}, func() AskAnswer { return featureAnswer("pharmacy") }},
	{[]string{"лекарств", "таблет", "препарат", "витамин", "капли", "мазь", "пилюл"}, func() AskAnswer { return tabAnswers["medicines"] }},
	{[]string{"волонт", "соцпомощ", "соцзащит", "соцработ", "социальн", "собес", "сиделк", "кцсон", "тцсо", "помощь на дому"}, func() AskAnswer { return featureAnswer("social") }},
	{[]string{"магазин", "продукт", "товар", "купить", "покупк", "супермаркет", "скидк", "акци"}, func() AskAnswer { return featureAnswer("goods") }},
	{[]string{"документ", "снилс", "полис", "справк", "льгот", "пенси", "выплат", "пособи", "паспорт"}, func() AskAnswer { return tabAnswers["documents"] }},
	{[]string{"профил", "мои данные", "мой адрес", "адрес", "прописк", "день рождения", "мое имя", "настройк"}, func() AskAnswer { return profileAnswer() }},
	{[]string{"помощ", "помоги", "плохо", "инструкц", "не получается", "не могу"}, func() AskAnswer { return tabAnswers["help"] }},
}

func answerQuestion(text string, now time.Time) AskAnswer {
	q := prepareAsk(text)
	if q.hasAny(remindStems...) {
		return taskAnswer(parseTask(q.String(), now), now)
	}

	med, medNamed := findInText(medicineItems(), q.String())
	if medNamed {
		return medicineAnswer(med.ID)
	}
	if isTreatmentQuestion(q) {
		return treatmentAnswer()
	}
	if id, ok := matchRules(q, guideRules); ok {
		return guideAnswer(id)
	}
	if id, ok := matchRules(q, benefitRules); ok {
		return benefitAnswer(id)
	}
	if id, ok := specialtyInText(q); ok {
		return doctorAnswer(id)
	}
	if q.hasAny(doctorStems...) {
		return doctorAnswer("")
	}
	if item, ok := findInText(productItems(), q.String()); ok {
		return productAnswer(item.ID)
	}
	for _, rule := range askRules {
		if q.hasAny(rule.stems...) {
			return rule.pick()
		}
	}
	return unknownAnswer()
}
