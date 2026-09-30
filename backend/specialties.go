package main

// Aliases ищем в запросе всегда: «глазник», «нужен лор».
// Symptoms — только когда человек просит врача: «к врачу по давлению»
type Specialty struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Dative   string   `json:"-"`
	Aliases  []string `json:"-"`
	Symptoms []string `json:"-"`
}

var doctorSpecialties = []Specialty{
	{ID: "therapist", Name: "Терапевт", Dative: "терапевту",
		Aliases:  []string{"терапевт", "участковый врач", "участкового врача", "участковому", "семейный врач", "общий врач"},
		Symptoms: []string{"простуд", "температур", "анализ", "справк", "диспансериз", "осмотр"}},
	{ID: "cardiologist", Name: "Кардиолог", Dative: "кардиологу",
		Aliases:  []string{"кардиолог", "сердечный врач", "сердечного врача", "сердечному врачу", "врач по сердцу", "врачу по сердцу"},
		Symptoms: []string{"давлен", "сердц", "сердеч", "пульс", "аритми", "гипертон", "тахикард"}},
	{ID: "neurologist", Name: "Невролог", Dative: "неврологу",
		Aliases:  []string{"невролог", "невропатолог", "нервный врач"},
		Symptoms: []string{"головокруж", "голова", "головн", "спина", "спину", "спиной", "поясниц", "онемен", "памят", "бессонниц", "инсульт", "нерв"}},
	{ID: "ophthalmologist", Name: "Офтальмолог", Dative: "офтальмологу",
		Aliases:  []string{"офтальмолог", "окулист", "глазник", "глазной врач", "глазного врача", "глазному врачу"},
		Symptoms: []string{"глаз", "зрени", "очки", "очков", "катаракт", "глауком"}},
	{ID: "ent", Name: "Оториноларинголог (ЛОР)", Dative: "ЛОР-врачу",
		Aliases:  []string{"лор", "лору", "отоларинголог", "оториноларинголог", "ухогорлонос", "ухо-горло-нос"},
		Symptoms: []string{"уши", "ушах", "ухо", "горл", "носом", "нос", "слух", "насморк", "гайморит"}},
	{ID: "endocrinologist", Name: "Эндокринолог", Dative: "эндокринологу",
		Aliases:  []string{"эндокринолог"},
		Symptoms: []string{"сахар", "диабет", "щитовид", "гормон"}},
	{ID: "urologist", Name: "Уролог", Dative: "урологу",
		Aliases:  []string{"уролог"},
		Symptoms: []string{"простат", "почк", "мочев", "мочеисп"}},
	{ID: "gastroenterologist", Name: "Гастроэнтеролог", Dative: "гастроэнтерологу",
		Aliases:  []string{"гастроэнтеролог", "гастролог"},
		Symptoms: []string{"желуд", "живот", "изжог", "кишеч", "печен", "гастрит"}},
	{ID: "traumatologist", Name: "Травматолог-ортопед", Dative: "травматологу",
		Aliases:  []string{"травматолог", "ортопед", "травмпункт"},
		Symptoms: []string{"перелом", "травм", "ушиб", "вывих", "колен", "упал"}},
	{ID: "dermatologist", Name: "Дерматолог", Dative: "дерматологу",
		Aliases:  []string{"дерматолог", "кожник", "кожный врач", "кожного врача", "кожному врачу"},
		Symptoms: []string{"кожа", "коже", "кожи", "сыпь", "родинк", "зуд", "экзем", "лишай"}},
	{ID: "pulmonologist", Name: "Пульмонолог", Dative: "пульмонологу",
		Aliases:  []string{"пульмонолог"},
		Symptoms: []string{"легк", "одышк", "астм", "бронх", "кашел", "кашл"}},
	{ID: "rheumatologist", Name: "Ревматолог", Dative: "ревматологу",
		Aliases:  []string{"ревматолог"},
		Symptoms: []string{"сустав", "артрит", "артроз", "подагр"}},
	{ID: "surgeon", Name: "Хирург", Dative: "хирургу",
		Aliases:  []string{"хирург"},
		Symptoms: []string{"грыж", "нарыв", "операц", "рана", "рану"}},
	{ID: "geriatrician", Name: "Гериатр", Dative: "гериатру",
		Aliases:  []string{"гериатр", "врач для пожилых", "врачу для пожилых", "врача для пожилых"},
		Symptoms: []string{"старост", "пожил"}},
}

func findSpecialty(id string) (Specialty, bool) {
	for _, s := range doctorSpecialties {
		if s.ID == id {
			return s, true
		}
	}
	return Specialty{}, false
}

func specialtyIDByName(name string) string {
	for _, s := range doctorSpecialties {
		if s.Name == name {
			return s.ID
		}
	}
	return ""
}
