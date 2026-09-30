package main

import "strings"

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
	{ID: "taxi", Title: "Вызвать такси", Icon: "car", Color: "amber"},
	{ID: "call", Title: "Связаться с близкими", Icon: "phone", Color: "violet"},
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

type AskAnswer struct {
	Type     string    `json:"type"`
	Target   string    `json:"target,omitempty"`
	Message  string    `json:"message"`
	Feature  *Feature  `json:"feature,omitempty"`
	Medicine *Medicine `json:"medicine,omitempty"`
	Product  *Product  `json:"product,omitempty"`
}

func medicineAnswer(id string) AskAnswer {
	m, _ := findMedicine(id)
	return AskAnswer{Type: "medicine", Target: m.ID, Message: "Ищем «" + m.Name + "» в аптеках рядом", Medicine: &m}
}

func productAnswer(id string) AskAnswer {
	p, _ := findProduct(id)
	return AskAnswer{Type: "product", Target: p.ID, Message: "Ищем «" + p.Name + "» в магазинах рядом", Product: &p}
}

type askRule struct {
	words  []string
	kind   string
	target string
}

var tabTitles = map[string]string{
	"medicines": "Лекарства",
	"documents": "Документы",
	"help":      "Помощь",
}

// порядок важен: «аптека» должна сработать раньше, чем «лекарства»
var askRules = []askRule{
	{words: []string{"аптек"}, kind: "feature", target: "pharmacy"},
	{words: []string{"врач", "доктор", "поликлиник", "терапевт", "запис"}, kind: "feature", target: "doctor"},
	{words: []string{"такси", "поехать", "доехать"}, kind: "feature", target: "taxi"},
	{words: []string{"волонт", "соцпомощ", "соцзащит", "соцработ", "социальн"}, kind: "feature", target: "social"},
	{words: []string{"позвон", "дочер", "сын", "внук", "родн", "близк"}, kind: "feature", target: "call"},
	{words: []string{"купить", "товар", "магазин", "продукт"}, kind: "feature", target: "goods"},
	{words: []string{"таблет", "лекарств", "витамин"}, kind: "tab", target: "medicines"},
	{words: []string{"документ", "паспорт", "снилс", "полис", "справк"}, kind: "tab", target: "documents"},
	{words: []string{"помощ", "помоги", "плохо"}, kind: "tab", target: "help"},
}

func answerQuestion(text string) AskAnswer {
	if item, ok := findInText(medicineItems(), text); ok {
		return medicineAnswer(item.ID)
	}
	if item, ok := findInText(productItems(), text); ok {
		return productAnswer(item.ID)
	}

	text = strings.ToLower(text)
	for _, rule := range askRules {
		for _, w := range rule.words {
			if !strings.Contains(text, w) {
				continue
			}
			if rule.kind == "tab" {
				return AskAnswer{Type: "tab", Target: rule.target, Message: "Открываю раздел «" + tabTitles[rule.target] + "»"}
			}
			f, _ := findFeature(rule.target)
			return AskAnswer{Type: "feature", Target: f.ID, Message: "Открываю «" + f.Title + "»", Feature: &f}
		}
	}
	return AskAnswer{Type: "unknown", Message: "Не поняли запрос. Попробуйте написать, например: «аптека», «врач» или «такси»"}
}
