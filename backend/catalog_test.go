package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCatalogSizes(t *testing.T) {
	if len(medicines) < 150 {
		t.Errorf("лекарств %d, нужно не меньше 150", len(medicines))
	}
	if len(products) < 150 {
		t.Errorf("товаров %d, нужно не меньше 150", len(products))
	}
}

// одно и то же название у двух позиций путает поиск
func TestCatalogNamesDoNotRepeat(t *testing.T) {
	for name, items := range map[string][]CatalogItem{"лекарства": medicineItems(), "товары": productItems()} {
		owner := map[string]string{}
		for _, item := range items {
			for _, n := range append([]string{item.Name}, item.Aliases...) {
				key := string(normalize(n))
				if prev, ok := owner[key]; ok && prev != item.ID {
					t.Errorf("%s: %q есть и у %s, и у %s", name, n, prev, item.ID)
				}
				owner[key] = item.ID
			}
		}
	}
}

// выкидываем букву из середины названия — позиция всё равно должна найтись.
// В коротких словах вроде «йод» опечатки не прощаем, их не проверяем
func withTypo(name string) string {
	r := normalize(name)
	i := len(r) / 2
	return string(r[:i]) + string(r[i+1:])
}

func TestCatalogFoundWithTypos(t *testing.T) {
	for name, items := range map[string][]CatalogItem{"лекарства": medicineItems(), "товары": productItems()} {
		for _, item := range items {
			queries := []string{item.Name}
			if len(normalize(item.Name)) >= 5 {
				queries = append(queries, withTypo(item.Name))
			}
			for _, q := range queries {
				found := false
				for _, m := range fuzzySuggest(items, q, 5) {
					if m.ID == item.ID {
						found = true
					}
				}
				if !found {
					t.Errorf("%s: по запросу %q не нашли %s", name, q, item.ID)
				}
			}
		}
	}
}

func TestNewCatalogItemsWithTypos(t *testing.T) {
	tests := []struct {
		query  string
		items  []CatalogItem
		wantID string
	}{
		{"лизинопрл", medicineItems(), "lizinopril"},
		{"конкорр", medicineItems(), "bisoprolol"},
		{"диабетон", medicineItems(), "gliklazid"},
		{"тауфон", medicineItems(), "taufon"},
		{"левомеколь", medicineItems(), "levomekol"},
		{"вольтарен эмульгель", medicineItems(), "diclofenac-gel"},
		{"ренни", medicineItems(), "rennie"},
		{"rhtcnjh", medicineItems(), "rosuvastatin"},
		{"тонометр", productItems(), "tonometer"},
		{"тАнометр", productItems(), "tonometer"},
		{"глюкомер", productItems(), "glucometer"},
		{"трость", productItems(), "cane"},
		{"памперсы для взрослых", productItems(), "adult-diapers"},
		{"ряженка", productItems(), "ryazhenka"},
		{"тушонка", productItems(), "stew"},
		{"gtkmvtyb", productItems(), "dumplings"},
	}
	for _, tt := range tests {
		got := fuzzySuggest(tt.items, tt.query, 5)
		if len(got) == 0 || got[0].ID != tt.wantID {
			t.Errorf("%q: получили %+v, ожидали %s", tt.query, got, tt.wantID)
		}
	}
}

// у каждой позиции справочника после импорта есть цена рядом с центром Москвы и Казани
func TestEveryCatalogItemHasPricesInBothCities(t *testing.T) {
	s := testStore(t)
	centers := map[string][2]float64{"Москва": {55.7558, 37.6173}, "Казань": {55.7963, 49.1088}}

	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		query := r.Form.Get("data")
		c, offset := centers["Москва"], 0
		if strings.Contains(query, "49.") {
			c, offset = centers["Казань"], 1000
		}
		shop := strings.Contains(query, "supermarket")
		elements := []string{}
		for i := 0; i < 12; i++ {
			tags := `{"name": "Аптека"}`
			if shop {
				tags = `{"name": "Магазин", "shop": "supermarket"}`
			}
			elements = append(elements, fmt.Sprintf(`{"type": "node", "id": %d, "lat": %f, "lon": %f, "tags": %s}`,
				offset+i+1, c[0]+float64(i)*0.001, c[1]+float64(i)*0.001, tags))
		}
		w.Write([]byte(`{"elements": [` + strings.Join(elements, ",") + `]}`))
	}))
	defer fake.Close()
	old, oldPause := overpassURL, overpassPause
	overpassURL, overpassPause = fake.URL, 0
	defer func() { overpassURL, overpassPause = old, oldPause }()

	ctx := context.Background()
	for _, c := range importCities {
		for _, kind := range []string{"pharmacy", "shop"} {
			if _, err := importPlaces(ctx, s, kind, c); err != nil {
				t.Fatal(err)
			}
		}
	}

	for cityName, c := range centers {
		for _, m := range medicines {
			offers, err := s.MedicineOffers(ctx, m.ID, c[0], c[1], searchKm)
			if err != nil || len(offers) == 0 {
				t.Errorf("%s: нет цен на %s (%v)", cityName, m.ID, err)
			}
		}
		for _, p := range products {
			offers, err := s.ProductOffers(ctx, p.ID, "2026-09-30", c[0], c[1], searchKm)
			if err != nil || len(offers) == 0 {
				t.Errorf("%s: нет цен на %s (%v)", cityName, p.ID, err)
			}
		}
	}
}
