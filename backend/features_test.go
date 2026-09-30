package main

import (
	"strings"
	"testing"
	"time"
)

var askNow = time.Date(2026, 9, 30, 8, 0, 0, 0, time.UTC)

type askCase struct {
	text       string
	wantType   string
	wantTarget string
}

func checkAnswers(t *testing.T, tests []askCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := answerQuestion(tt.text, askNow)
			if got.Type != tt.wantType || got.Target != tt.wantTarget {
				t.Errorf("получили %s/%s, ожидали %s/%s", got.Type, got.Target, tt.wantType, tt.wantTarget)
			}
			if got.Message == "" {
				t.Error("нет текста «что можно сделать»")
			}
			if got.Type != "unknown" && got.Button == "" {
				t.Error("нет надписи на кнопке")
			}
			switch got.Type {
			case "medicine":
				if got.Medicine == nil || got.Medicine.ID != tt.wantTarget {
					t.Errorf("нет объекта лекарства: %+v", got)
				}
			case "product":
				if got.Product == nil || got.Product.ID != tt.wantTarget {
					t.Errorf("нет объекта товара: %+v", got)
				}
			case "feature":
				if got.Feature == nil || got.Feature.ID != tt.wantTarget {
					t.Errorf("нет объекта плитки: %+v", got.Feature)
				}
			case "doctor":
				if tt.wantTarget != "" && (got.Specialty == nil || got.Specialty.ID != tt.wantTarget) {
					t.Errorf("нет специальности: %+v", got.Specialty)
				}
			case "benefit", "guide":
				if got.Title == "" {
					t.Error("нет названия")
				}
			}
		})
	}
}

func TestAskMedicine(t *testing.T) {
	checkAnswers(t, []askCase{
		{"где купить нурофен", "medicine", "ibuprofen"},
		{"нурафен", "medicine", "ibuprofen"},
		{"yehjaty", "medicine", "ibuprofen"},
		{"нужен парацетомол срочно", "medicine", "paracetamol"},
		{"купить лекарство конкор", "medicine", "bisoprolol"},
		{"сколько стоит лизиноприл", "medicine", "lizinopril"},
		{"ищу тауфон", "medicine", "taufon"},
		{"диабетон где дешевле", "medicine", "gliklazid"},
		{"есть ли кардиомагнил в аптеке", "medicine", "cardiomagnil"},
		{"мазь левомиколь", "medicine", "levomekol"},
		{"можно ли выпить нурофен", "medicine", "ibuprofen"},
		{"витамин d3", "medicine", "vitamin-d3"},
		{"но шпа", "medicine", "drotaverin"},
		{"омепразол", "medicine", "omeprazol"},
	})
}

// лечение не назначаем: ни правила, ни ИИ не называют препарат, если его не назвал человек
func TestAskTreatmentIsNotPrescribed(t *testing.T) {
	checkAnswers(t, []askCase{
		{"что выпить от давления", "tab", "medicines"},
		{"что принять от сердца", "tab", "medicines"},
		{"что пить от боли в спине", "tab", "medicines"},
		{"чем лечить кашель", "tab", "medicines"},
		{"таблетки от давления", "tab", "medicines"},
		{"какое лекарство от головы", "tab", "medicines"},
		{"посоветуй что-нибудь от бессонницы", "tab", "medicines"},
		{"чем сбить температуру", "tab", "medicines"},
		{"что помогает от изжоги", "tab", "medicines"},
		{"средство от аллергии", "tab", "medicines"},
	})
	for _, text := range []string{"что выпить от давления", "что принять от сердца", "что пить от боли"} {
		got := answerQuestion(text, askNow)
		if got.Medicine != nil || !strings.Contains(got.Message, "врач") {
			t.Errorf("%q: нельзя советовать препарат, нужно отправить к врачу: %+v", text, got)
		}
	}
}

func TestAskMedicinesSection(t *testing.T) {
	checkAnswers(t, []askCase{
		{"купить лекарство", "tab", "medicines"},
		{"купить лекарства", "tab", "medicines"},
		{"кепить лекарство", "tab", "medicines"},
		{"когда пить таблетки", "tab", "medicines"},
		{"нужны таблетки", "tab", "medicines"},
		{"лекарства", "tab", "medicines"},
		{"лекарства", "tab", "medicines"},
		{"лекрства подешевле", "tab", "medicines"},
		{"где взять препарат", "tab", "medicines"},
		{"витамины для пожилых", "tab", "medicines"},
		{"глазные капли", "tab", "medicines"},
	})
}

func TestAskProduct(t *testing.T) {
	checkAnswers(t, []askCase{
		{"купить хлеб в магазине", "product", "bread-white"},
		{"нужна гречка", "product", "buckwheat"},
		{"где купить молоко", "product", "milk"},
		{"vjkjrj", "product", "milk"},
		{"малако", "product", "milk"},
		{"нужен тонометр", "product", "tonometer"},
		{"где купить трость", "product", "cane"},
		{"подгузники для взрослых", "product", "adult-diapers"},
		{"глюкометр подешевле", "product", "glucometer"},
		{"купить фонарик", "product", "flashlight"},
		{"картошка", "product", "potato"},
		{"стиральный порошок", "product", "washing-powder"},
	})
}

func TestAskPharmacies(t *testing.T) {
	checkAnswers(t, []askCase{
		{"Где ближайшая аптека?", "feature", "pharmacy"},
		{"АПТЕКА", "feature", "pharmacy"},
		{"аптеки рядом", "feature", "pharmacy"},
		{"аптка", "feature", "pharmacy"},
		{"fgntrf", "feature", "pharmacy"},
		{"какая аптека открыта", "feature", "pharmacy"},
		{"купить лекарство в аптеке", "feature", "pharmacy"},
		{"адрес аптеки", "feature", "pharmacy"},
		{"дежурная аптека", "feature", "pharmacy"},
		{"хорошая аптека недалеко", "feature", "pharmacy"},
	})
}

func TestAskShops(t *testing.T) {
	checkAnswers(t, []askCase{
		{"хочу в магазин", "feature", "goods"},
		{"магазины рядом", "feature", "goods"},
		{"vfufpby", "feature", "goods"},
		{"магозин", "feature", "goods"},
		{"купить продукты", "feature", "goods"},
		{"где скидки", "feature", "goods"},
		{"акции в магазинах", "feature", "goods"},
		{"ближайший супермаркет", "feature", "goods"},
		{"товары рядом", "feature", "goods"},
		{"нужно купить поесть", "feature", "goods"},
	})
}

func TestAskDoctor(t *testing.T) {
	checkAnswers(t, []askCase{
		{"нужен кардиолог", "doctor", "cardiologist"},
		{"хочу к врачу по давлению", "doctor", "cardiologist"},
		{"кордиолог", "doctor", "cardiologist"},
		{"rfhlbjkju", "doctor", "cardiologist"},
		{"сердечный врач", "doctor", "cardiologist"},
		{"глазник", "doctor", "ophthalmologist"},
		{"записаться к окулисту", "doctor", "ophthalmologist"},
		{"врач по зрению", "doctor", "ophthalmologist"},
		{"нужен лор", "doctor", "ent"},
		{"записаться к ухогорлонос", "doctor", "ent"},
		{"врач болит горло", "doctor", "ent"},
		{"эндокринолог", "doctor", "endocrinologist"},
		{"к врачу по сахару", "doctor", "endocrinologist"},
		{"невролог", "doctor", "neurologist"},
		{"записаться к терапевту", "doctor", "therapist"},
		{"к какому врачу если болит желудок", "doctor", "gastroenterologist"},
		{"записаться к кожнику", "doctor", "dermatologist"},
		{"травматолог рядом", "doctor", "traumatologist"},
	})
}

func TestAskDoctorGeneral(t *testing.T) {
	checkAnswers(t, []askCase{
		{"хочу записаться к врачу", "doctor", ""},
		{"врач", "doctor", ""},
		{"врачь", "doctor", ""},
		{"dhfx", "doctor", ""},
		{"доктор", "doctor", ""},
		{"моя поликлиника", "doctor", ""},
		{"поликлинника по прописке", "doctor", ""},
		{"запиши меня на прием", "doctor", ""},
		{"взять талон", "doctor", ""},
		{"нужен специалист", "doctor", ""},
	})
}

func TestAskSocial(t *testing.T) {
	checkAnswers(t, []askCase{
		{"где соцзащита", "feature", "social"},
		{"cjwpfobnf", "feature", "social"},
		{"соцзащта", "feature", "social"},
		{"нужен волонтёр", "feature", "social"},
		{"соцработник", "feature", "social"},
		{"социальный работник", "feature", "social"},
		{"нужна сиделка", "feature", "social"},
		{"где собес", "feature", "social"},
		{"соцпомощь рядом", "feature", "social"},
		{"центр социального обслуживания", "feature", "social"},
		{"помощь на дому", "feature", "social"},
	})
}

func TestAskBenefit(t *testing.T) {
	checkAnswers(t, []askCase{
		{"компенсация за капремонт", "benefit", "overhaul"},
		{"rfghtvjyn", "benefit", "overhaul"},
		{"капремнт льгота", "benefit", "overhaul"},
		{"не платить взносы на ремонт", "benefit", "overhaul"},
		{"как получить карту москвича", "benefit", "moscow-card"},
		{"субсидия на квартплату", "benefit", "housing-subsidy"},
		{"помощь с коммуналкой", "benefit", "housing-subsidy"},
		{"льгота по налогу на квартиру", "benefit", "property-tax"},
		{"ветеран труда", "benefit", "labour-veteran"},
		{"доплата к пенсии", "benefit", "social-supplement"},
		{"прибавка к пенсии в 80 лет", "benefit", "pension-80"},
		{"московское долголетие", "benefit", "moscow-longevity"},
		{"бесплатные лекарства", "benefit", "social-services-set"},
		{"поменять паспорт в 45", "benefit", "passport"},
	})
}

func TestAskGuide(t *testing.T) {
	checkAnswers(t, []askCase{
		{"как увеличить шрифт", "guide", "font"},
		{"ihban", "guide", "font"},
		{"шрфит крупнее", "guide", "font"},
		{"увеличить шривт", "guide", "font"},
		{"сделать буквы больше", "guide", "font"},
		{"мелкие буквы на телефоне", "guide", "font"},
		{"как включить фонарик", "guide", "flashlight"},
		{"как подключить wi-fi", "guide", "wifi"},
		{"нет интернета", "guide", "wifi"},
		{"как позвонить по видео", "guide", "max-video"},
		{"как отправить фото внуку", "guide", "max-photo"},
		{"звонят из банка просят код", "guide", "scam"},
		{"мошенники", "guide", "scam"},
		{"как ответить на звонок", "guide", "answer"},
		{"как позвонить", "guide", "call"},
		{"как поставить госуслуги", "guide", "gosuslugi"},
		{"как пользоваться приложением", "guide", "this-app"},
	})
}

func TestAskProfile(t *testing.T) {
	checkAnswers(t, []askCase{
		{"профиль", "profile", ""},
		{"ghjabkm", "profile", ""},
		{"прафиль", "profile", ""},
		{"мои данные", "profile", ""},
		{"поменять адрес", "profile", ""},
		{"указать прописку", "profile", ""},
		{"изменить мой адрес", "profile", ""},
		{"день рождения", "profile", ""},
		{"настройки", "profile", ""},
		{"где мой профиль", "profile", ""},
	})
}

func TestAskTabs(t *testing.T) {
	checkAnswers(t, []askCase{
		{"где мой паспорт", "tab", "documents"},
		{"документы", "tab", "documents"},
		{"снилс", "tab", "documents"},
		{"какие льготы мне положены", "tab", "documents"},
		{"мне плохо", "tab", "help"},
		{"помогите", "tab", "help"},
		{"инструкции", "tab", "help"},
	})
}

func TestAskUnknown(t *testing.T) {
	checkAnswers(t, []askCase{
		{"вызови такси", "unknown", ""},
		{"какая сегодня погода", "unknown", ""},
		{"привет", "unknown", ""},
		{"расскажи анекдот", "unknown", ""},
		{"zzzz", "unknown", ""},
		{"12345", "unknown", ""},
		{"!!!", "unknown", ""},
		{"кто выиграл матч", "unknown", ""},
		{"сколько времени", "unknown", ""},
		{"спасибо", "unknown", ""},
	})
}

func TestAskTask(t *testing.T) {
	tests := []struct {
		text  string
		title string
		time  string
		date  string
		kind  string
	}{
		{"напомни выпить таблетку в 9", "Выпить таблетку", "09:00", "2026-09-30", "medicine"},
		{"yfgjvyb dsgbnm nf,ktnre d 9", "Выпить таблетку", "09:00", "2026-09-30", "medicine"},
		{"напомни мне в 7 утра выпить таблетки", "Выпить таблетки", "07:00", "2026-10-01", "medicine"},
		{"напомни завтра в 10:30 позвонить дочке", "Позвонить дочке", "10:30", "2026-10-01", "call"},
		{"напомните в 7 вечера измерить давление", "Измерить давление", "19:00", "2026-09-30", "medicine"},
		{"напомни 15 октября в 11 сходить к врачу", "Сходить к врачу", "11:00", "2026-10-15", "doctor"},
		{"напомни в пятницу в 12:00 оплатить квартиру", "Оплатить квартиру", "12:00", "2026-10-02", "other"},
		{"напомни через 30 минут выключить плиту", "Выключить плиту", "08:30", "2026-09-30", "other"},
		{"напомни через час полить цветы", "Полить цветы", "09:00", "2026-09-30", "other"},
		{"напомни купить хлеба", "Купить хлеба", "", "2026-09-30", "other"},
		{"не забыть в полдень принять капли", "Принять капли", "12:00", "2026-09-30", "medicine"},
		{"добавь задачу послезавтра в 9.15 анализы", "Анализы", "09:15", "2026-10-02", "doctor"},
		{"напомнить про лекарство в 21:00", "Лекарство", "21:00", "2026-09-30", "medicine"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := answerQuestion(tt.text, askNow)
			if got.Type != "task" || got.Task == nil {
				t.Fatalf("ожидали задачу, получили %+v", got)
			}
			want := TaskDraft{Title: tt.title, Time: tt.time, Date: tt.date, Kind: tt.kind}
			if *got.Task != want {
				t.Errorf("получили %+v, ожидали %+v", *got.Task, want)
			}
			if got.Message == "" || got.Button == "" {
				t.Error("нужен вопрос-подтверждение и кнопка")
			}
		})
	}
}

func TestAskTaskMessage(t *testing.T) {
	got := answerQuestion("напомни выпить таблетку в 9", askNow)
	if got.Message != "Добавить задачу «Выпить таблетку» на сегодня в 09:00? Напомню в это время." {
		t.Errorf("%q", got.Message)
	}
	got = answerQuestion("напомни купить хлеба", askNow)
	if !strings.Contains(got.Message, "Выберите время") {
		t.Errorf("без времени просим выбрать время: %q", got.Message)
	}
}

func TestAnswerPayload(t *testing.T) {
	tests := map[string]string{
		"нурофен":                     "med_ibuprofen",
		"молоко":                      "prod_milk",
		"нужен кардиолог":             "doctor_cardiologist",
		"хочу к врачу":                "doctor",
		"компенсация за капремонт":    "benefit_overhaul",
		"как увеличить шрифт":         "guide_font",
		"где соцзащита":               "social",
		"аптека":                      "pharmacy",
		"профиль":                     "profile",
		"документы":                   "documents",
		"что выпить от давления":      "medicines",
		"напомни выпить таблетку в 9": "",
	}
	for text, want := range tests {
		if got := answerQuestion(text, askNow).payload(); got != want {
			t.Errorf("%q: %q, ожидали %q", text, got, want)
		}
	}
}

func TestRulesPointToExistingTargets(t *testing.T) {
	for _, r := range guideRules {
		if _, ok := findGuide(r.target); !ok {
			t.Errorf("нет инструкции %q", r.target)
		}
	}
	for _, r := range benefitRules {
		if _, ok := findBenefit(r.target); !ok {
			t.Errorf("нет льготы %q", r.target)
		}
	}
	for _, name := range specialties {
		if specialtyIDByName(name) == "" {
			t.Errorf("у врача специальность %q, которой нет в справочнике", name)
		}
	}
}

func TestSpecialtiesFilled(t *testing.T) {
	seen := map[string]bool{}
	for _, sp := range doctorSpecialties {
		if sp.ID == "" || sp.Name == "" || sp.Dative == "" || len(sp.Aliases) == 0 || seen[sp.ID] {
			t.Errorf("плохая специальность %+v", sp)
		}
		seen[sp.ID] = true
	}
}

func TestFeaturesUniqueAndFilled(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range features {
		if f.ID == "" || f.Title == "" || f.Icon == "" || f.Color == "" {
			t.Errorf("пустые поля у возможности %+v", f)
		}
		if seen[f.ID] {
			t.Errorf("повторяется id %q", f.ID)
		}
		seen[f.ID] = true
	}
}
