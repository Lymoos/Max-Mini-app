package main

type Medicine struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Form      string   `json:"form"`
	Aliases   []string `json:"-"`
	BasePrice int      `json:"-"`
}

var medicines = []Medicine{
	{ID: "paracetamol", Name: "Парацетамол", Form: "таблетки 500 мг, 20 шт", Aliases: []string{"Панадол", "Эффералган", "Калпол"}, BasePrice: 45},
	{ID: "ibuprofen", Name: "Ибупрофен", Form: "таблетки 200 мг, 20 шт", Aliases: []string{"Нурофен", "МИГ"}, BasePrice: 70},
	{ID: "aspirin", Name: "Аспирин", Form: "таблетки 500 мг, 10 шт", Aliases: []string{"Ацетилсалициловая кислота"}, BasePrice: 60},
	{ID: "citramon", Name: "Цитрамон", Form: "таблетки, 10 шт", BasePrice: 35},
	{ID: "drotaverin", Name: "Дротаверин", Form: "таблетки 40 мг, 20 шт", Aliases: []string{"Но-шпа", "Спазмол"}, BasePrice: 55},
	{ID: "metamizol", Name: "Анальгин", Form: "таблетки 500 мг, 10 шт", Aliases: []string{"Метамизол натрия"}, BasePrice: 40},
	{ID: "loratadin", Name: "Лоратадин", Form: "таблетки 10 мг, 10 шт", Aliases: []string{"Кларитин", "Кларотадин"}, BasePrice: 90},
	{ID: "suprastin", Name: "Супрастин", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Хлоропирамин"}, BasePrice: 150},
	{ID: "cetirizin", Name: "Цетиризин", Form: "таблетки 10 мг, 10 шт", Aliases: []string{"Зиртек", "Зодак"}, BasePrice: 110},
	{ID: "omeprazol", Name: "Омепразол", Form: "капсулы 20 мг, 30 шт", Aliases: []string{"Омез", "Ультоп"}, BasePrice: 95},
	{ID: "pankreatin", Name: "Панкреатин", Form: "таблетки, 20 шт", Aliases: []string{"Мезим", "Креон"}, BasePrice: 80},
	{ID: "smekta", Name: "Смекта", Form: "порошок, 10 пакетиков", Aliases: []string{"Диосмектит"}, BasePrice: 190},
	{ID: "coal", Name: "Активированный уголь", Form: "таблетки 250 мг, 10 шт", Aliases: []string{"Уголь"}, BasePrice: 20},
	{ID: "loperamid", Name: "Лоперамид", Form: "капсулы 2 мг, 20 шт", Aliases: []string{"Имодиум"}, BasePrice: 50},
	{ID: "enalapril", Name: "Эналаприл", Form: "таблетки 10 мг, 20 шт", Aliases: []string{"Энап", "Ренитек"}, BasePrice: 65},
	{ID: "lozartan", Name: "Лозартан", Form: "таблетки 50 мг, 30 шт", Aliases: []string{"Лозап", "Козаар"}, BasePrice: 180},
	{ID: "amlodipin", Name: "Амлодипин", Form: "таблетки 5 мг, 30 шт", Aliases: []string{"Норваск"}, BasePrice: 75},
	{ID: "bisoprolol", Name: "Бисопролол", Form: "таблетки 5 мг, 30 шт", Aliases: []string{"Конкор"}, BasePrice: 120},
	{ID: "kaptopril", Name: "Каптоприл", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Капотен"}, BasePrice: 40},
	{ID: "atorvastatin", Name: "Аторвастатин", Form: "таблетки 20 мг, 30 шт", Aliases: []string{"Аторис", "Липримар"}, BasePrice: 250},
	{ID: "metformin", Name: "Метформин", Form: "таблетки 850 мг, 60 шт", Aliases: []string{"Глюкофаж", "Сиофор"}, BasePrice: 160},
	{ID: "cardiomagnil", Name: "Кардиомагнил", Form: "таблетки 75 мг, 100 шт", Aliases: []string{"Тромбо АСС"}, BasePrice: 270},
	{ID: "validol", Name: "Валидол", Form: "таблетки 60 мг, 10 шт", BasePrice: 45},
	{ID: "corvalol", Name: "Корвалол", Form: "капли, 25 мл", Aliases: []string{"Валокордин"}, BasePrice: 50},
	{ID: "nitroglycerin", Name: "Нитроглицерин", Form: "таблетки 0,5 мг, 40 шт", Aliases: []string{"Нитроминт"}, BasePrice: 70},
	{ID: "glycine", Name: "Глицин", Form: "таблетки 100 мг, 50 шт", BasePrice: 45},
	{ID: "vitamin-d3", Name: "Витамин D3", Form: "капли, 10 мл", Aliases: []string{"Аквадетрим", "Колекальциферол"}, BasePrice: 220},
	{ID: "amoxicillin", Name: "Амоксициллин", Form: "капсулы 500 мг, 16 шт", Aliases: []string{"Флемоксин", "Амосин"}, BasePrice: 110},
	{ID: "nazivin", Name: "Називин", Form: "спрей назальный, 10 мл", Aliases: []string{"Оксиметазолин"}, BasePrice: 200},
	{ID: "ambroxol", Name: "Амброксол", Form: "таблетки 30 мг, 20 шт", Aliases: []string{"Лазолван", "Амбробене"}, BasePrice: 55},
}

func findMedicine(id string) (Medicine, bool) {
	for _, m := range medicines {
		if m.ID == id {
			return m, true
		}
	}
	return Medicine{}, false
}

func medicineItems() []CatalogItem {
	items := []CatalogItem{}
	for _, m := range medicines {
		items = append(items, CatalogItem{ID: m.ID, Name: m.Name, Aliases: m.Aliases})
	}
	return items
}
