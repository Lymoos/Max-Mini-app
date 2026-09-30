package main

import "testing"

func TestSuggestMedicines(t *testing.T) {
	tests := []struct {
		query         string
		wantID        string
		wantCorrected bool
	}{
		{"Парацетамол", "paracetamol", false},
		{"пара", "paracetamol", false},
		{"парацетомол", "paracetamol", true},
		{"gfhfwtnfvjk", "paracetamol", true},
		{"нурофен", "ibuprofen", false},
		{"нурафен", "ibuprofen", true},
		{"но-шпа", "drotaverin", false},
		{"НО ШПА", "drotaverin", false},
		{"витамин д", "vitamin-d3", false},
		{"витамин d", "vitamin-d3", false},
		{"витамен д3", "vitamin-d3", true},
		{"кардиамагнил", "cardiomagnil", true},
		{"глюкофаж", "metformin", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := fuzzySuggest(medicineItems(), tt.query, 5)
			if len(got) == 0 {
				t.Fatal("ничего не нашли")
			}
			if got[0].ID != tt.wantID || got[0].Corrected != tt.wantCorrected || got[0].Matched == "" {
				t.Errorf("получили %+v, ожидали %s corrected=%v", got[0], tt.wantID, tt.wantCorrected)
			}
		})
	}
}

func TestSuggestProducts(t *testing.T) {
	tests := []struct {
		query  string
		wantID string
	}{
		{"молоко", "milk"},
		{"малако", "milk"},
		{"картошка", "potato"},
		{"картофель", "potato"},
		{"гречка", "buckwheat"},
		{"греча", "buckwheat"},
		{"яйца", "eggs"},
		{"курица", "chicken"},
		{"vjkjrj", "milk"},
		{"туалетная", "toilet-paper"},
		{"подгузники", "adult-diapers"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := fuzzySuggest(productItems(), tt.query, 5)
			if len(got) == 0 || got[0].ID != tt.wantID {
				t.Errorf("получили %+v, ожидали %s", got, tt.wantID)
			}
		})
	}
}

func TestSuggestNothing(t *testing.T) {
	for _, q := range []string{"", " ", "а", "zzzz", "компьютер", "12345", "!!!"} {
		if got := fuzzySuggest(medicineItems(), q, 5); len(got) != 0 {
			t.Errorf("лекарства для %q: %v", q, got)
		}
		if got := fuzzySuggest(productItems(), q, 5); len(got) != 0 {
			t.Errorf("товары для %q: %v", q, got)
		}
	}
}

func TestSuggestLimitAndUnique(t *testing.T) {
	got := fuzzySuggest(productItems(), "ма", 3)
	if len(got) > 3 {
		t.Errorf("лимит не соблюдён: %d", len(got))
	}
	seen := map[string]bool{}
	for _, m := range got {
		if seen[m.ID] {
			t.Errorf("%s повторяется", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestFindInText(t *testing.T) {
	tests := []struct {
		text   string
		items  []CatalogItem
		wantID string
	}{
		{"где купить нурофен", medicineItems(), "ibuprofen"},
		{"нужен парацетомол срочно", medicineItems(), "paracetamol"},
		{"хочу купить нурофена", medicineItems(), "ibuprofen"},
		{"где аптека", medicineItems(), ""},
		{"этап работ", medicineItems(), ""},
		{"мигать фарами", medicineItems(), ""},
		{"где купить молоко", productItems(), "milk"},
		{"нужна картошка", productItems(), "potato"},
		{"купить гречку", productItems(), "buckwheat"},
		{"вызови такси", productItems(), ""},
		{"", productItems(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			item, ok := findInText(tt.items, tt.text)
			if tt.wantID == "" {
				if ok {
					t.Errorf("не ожидали находки, нашли %s", item.ID)
				}
				return
			}
			if !ok || item.ID != tt.wantID {
				t.Errorf("нашли %q (%v), ожидали %s", item.ID, ok, tt.wantID)
			}
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"кот", "", 3},
		{"кот", "кит", 1},
		{"кот", "крот", 1},
		{"парацетамол", "парацетомол", 1},
	}
	for _, tt := range tests {
		if got := levenshtein([]rune(tt.a), []rune(tt.b)); got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, ожидали %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCatalogs(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range medicines {
		if m.ID == "" || m.Name == "" || m.Form == "" || m.BasePrice <= 0 || seen[m.ID] {
			t.Errorf("плохое лекарство %+v", m)
		}
		seen[m.ID] = true
	}
	seen = map[string]bool{}
	for _, p := range products {
		if p.ID == "" || p.Name == "" || p.Unit == "" || p.BasePrice <= 0 || seen[p.ID] {
			t.Errorf("плохой товар %+v", p)
		}
		seen[p.ID] = true
	}
}
