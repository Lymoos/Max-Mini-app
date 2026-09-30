package main

import "testing"

func TestAnswerQuestion(t *testing.T) {
	tests := []struct {
		text       string
		wantType   string
		wantTarget string
	}{
		{"Где ближайшая аптека?", "feature", "pharmacy"},
		{"АПТЕКА", "feature", "pharmacy"},
		{"хочу записаться к врачу", "feature", "doctor"},
		{"вызови такси", "unknown", ""},
		{"нужен волонтёр", "feature", "social"},
		{"где соцзащита", "feature", "social"},
		{"купить хлеб в магазине", "product", "bread-white"},
		{"хочу в магазин", "feature", "goods"},
		{"где купить нурофен", "medicine", "ibuprofen"},
		{"нужна гречка", "product", "buckwheat"},
		{"когда пить таблетки", "tab", "medicines"},
		{"где мой паспорт", "tab", "documents"},
		{"мне плохо", "tab", "help"},
		{"какая сегодня погода", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := answerQuestion(tt.text)
			if got.Type != tt.wantType || got.Target != tt.wantTarget {
				t.Errorf("получили %s/%s, ожидали %s/%s", got.Type, got.Target, tt.wantType, tt.wantTarget)
			}
			if got.Message == "" {
				t.Error("сообщение не должно быть пустым")
			}
			if got.Type == "medicine" && (got.Medicine == nil || got.Medicine.ID != tt.wantTarget) {
				t.Errorf("нет объекта лекарства: %+v", got)
			}
			if got.Type == "product" && (got.Product == nil || got.Product.ID != tt.wantTarget) {
				t.Errorf("нет объекта товара: %+v", got)
			}
			if got.Type == "feature" && (got.Feature == nil || got.Feature.ID != tt.wantTarget) {
				t.Errorf("для возможности должен приходить объект плитки, получили %+v", got.Feature)
			}
		})
	}
}

func TestAskRulesPointToExistingTargets(t *testing.T) {
	for _, rule := range askRules {
		if rule.kind == "feature" {
			if _, ok := findFeature(rule.target); !ok {
				t.Errorf("правило ссылается на несуществующую возможность %q", rule.target)
			}
		} else if tabTitles[rule.target] == "" {
			t.Errorf("правило ссылается на несуществующую вкладку %q", rule.target)
		}
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
