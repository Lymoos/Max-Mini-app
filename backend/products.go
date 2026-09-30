package main

type Product struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Unit      string   `json:"unit"`
	Aliases   []string `json:"-"`
	BasePrice int      `json:"-"`
}

var products = []Product{
	{ID: "milk", Name: "Молоко 2,5%", Unit: "1 л", Aliases: []string{"Молоко"}, BasePrice: 90},
	{ID: "kefir", Name: "Кефир 1%", Unit: "1 л", Aliases: []string{"Кефир"}, BasePrice: 95},
	{ID: "smetana", Name: "Сметана 15%", Unit: "300 г", Aliases: []string{"Сметана"}, BasePrice: 110},
	{ID: "tvorog", Name: "Творог 5%", Unit: "200 г", Aliases: []string{"Творог"}, BasePrice: 120},
	{ID: "butter", Name: "Масло сливочное 82%", Unit: "180 г", Aliases: []string{"Сливочное масло"}, BasePrice: 230},
	{ID: "cheese", Name: "Сыр Российский", Unit: "200 г", Aliases: []string{"Сыр"}, BasePrice: 220},
	{ID: "yogurt", Name: "Йогурт питьевой", Unit: "290 г", Aliases: []string{"Йогурт"}, BasePrice: 75},
	{ID: "eggs", Name: "Яйца куриные С1", Unit: "10 шт", Aliases: []string{"Яйца", "Яйцо"}, BasePrice: 120},
	{ID: "bread-borodinsky", Name: "Хлеб Бородинский", Unit: "400 г", Aliases: []string{"Бородинский", "Чёрный хлеб"}, BasePrice: 65},
	{ID: "bread-white", Name: "Хлеб белый", Unit: "400 г", Aliases: []string{"Хлеб"}, BasePrice: 55},
	{ID: "baton", Name: "Батон нарезной", Unit: "400 г", Aliases: []string{"Батон"}, BasePrice: 60},
	{ID: "buckwheat", Name: "Гречка", Unit: "900 г", Aliases: []string{"Гречневая крупа"}, BasePrice: 110},
	{ID: "rice", Name: "Рис круглозёрный", Unit: "900 г", Aliases: []string{"Рис"}, BasePrice: 120},
	{ID: "pasta", Name: "Макароны", Unit: "450 г", Aliases: []string{"Спагетти", "Вермишель"}, BasePrice: 80},
	{ID: "oats", Name: "Овсяные хлопья", Unit: "500 г", Aliases: []string{"Геркулес", "Овсянка"}, BasePrice: 85},
	{ID: "flour", Name: "Мука пшеничная", Unit: "2 кг", Aliases: []string{"Мука"}, BasePrice: 120},
	{ID: "sugar", Name: "Сахар", Unit: "1 кг", Aliases: []string{"Сахар-песок"}, BasePrice: 85},
	{ID: "salt", Name: "Соль поваренная", Unit: "1 кг", Aliases: []string{"Соль"}, BasePrice: 30},
	{ID: "sunflower-oil", Name: "Масло подсолнечное", Unit: "1 л", Aliases: []string{"Подсолнечное масло", "Растительное масло"}, BasePrice: 150},
	{ID: "tea", Name: "Чай чёрный", Unit: "100 пакетиков", Aliases: []string{"Чай"}, BasePrice: 180},
	{ID: "coffee", Name: "Кофе растворимый", Unit: "95 г", Aliases: []string{"Кофе"}, BasePrice: 320},
	{ID: "chicken", Name: "Филе куриное", Unit: "1 кг", Aliases: []string{"Курица", "Куриное филе"}, BasePrice: 420},
	{ID: "mince", Name: "Фарш говяжий", Unit: "400 г", Aliases: []string{"Фарш"}, BasePrice: 290},
	{ID: "sausages", Name: "Сосиски молочные", Unit: "450 г", Aliases: []string{"Сосиски"}, BasePrice: 260},
	{ID: "bologna", Name: "Колбаса варёная", Unit: "400 г", Aliases: []string{"Колбаса", "Докторская"}, BasePrice: 280},
	{ID: "fish", Name: "Минтай замороженный", Unit: "1 кг", Aliases: []string{"Рыба", "Минтай"}, BasePrice: 330},
	{ID: "potato", Name: "Картофель", Unit: "1 кг", Aliases: []string{"Картошка"}, BasePrice: 45},
	{ID: "carrot", Name: "Морковь", Unit: "1 кг", Aliases: []string{"Морковка"}, BasePrice: 50},
	{ID: "onion", Name: "Лук репчатый", Unit: "1 кг", Aliases: []string{"Лук"}, BasePrice: 40},
	{ID: "cabbage", Name: "Капуста белокочанная", Unit: "1 кг", Aliases: []string{"Капуста"}, BasePrice: 35},
	{ID: "tomato", Name: "Помидоры", Unit: "1 кг", Aliases: []string{"Томаты"}, BasePrice: 220},
	{ID: "cucumber", Name: "Огурцы", Unit: "1 кг", Aliases: []string{"Огурец"}, BasePrice: 180},
	{ID: "apple", Name: "Яблоки", Unit: "1 кг", Aliases: []string{"Яблоко"}, BasePrice: 140},
	{ID: "banana", Name: "Бананы", Unit: "1 кг", Aliases: []string{"Банан"}, BasePrice: 150},
	{ID: "orange", Name: "Апельсины", Unit: "1 кг", Aliases: []string{"Апельсин"}, BasePrice: 170},
	{ID: "lemon", Name: "Лимоны", Unit: "1 кг", Aliases: []string{"Лимон"}, BasePrice: 230},
	{ID: "juice", Name: "Сок яблочный", Unit: "1 л", Aliases: []string{"Сок"}, BasePrice: 130},
	{ID: "water", Name: "Вода питьевая", Unit: "5 л", Aliases: []string{"Вода"}, BasePrice: 110},
	{ID: "cookies", Name: "Печенье овсяное", Unit: "300 г", Aliases: []string{"Печенье"}, BasePrice: 110},
	{ID: "chocolate", Name: "Шоколад молочный", Unit: "90 г", Aliases: []string{"Шоколад"}, BasePrice: 100},
	{ID: "honey", Name: "Мёд цветочный", Unit: "250 г", Aliases: []string{"Мёд"}, BasePrice: 290},
	{ID: "toilet-paper", Name: "Туалетная бумага", Unit: "4 рулона", BasePrice: 150},
	{ID: "toothpaste", Name: "Зубная паста", Unit: "100 мл", BasePrice: 160},
	{ID: "soap", Name: "Мыло туалетное", Unit: "90 г", Aliases: []string{"Мыло"}, BasePrice: 60},
	{ID: "shampoo", Name: "Шампунь", Unit: "400 мл", BasePrice: 320},
	{ID: "washing-powder", Name: "Стиральный порошок", Unit: "3 кг", Aliases: []string{"Порошок"}, BasePrice: 520},
	{ID: "dish-soap", Name: "Средство для мытья посуды", Unit: "500 мл", Aliases: []string{"Средство для посуды"}, BasePrice: 150},
	{ID: "paper-towels", Name: "Бумажные полотенца", Unit: "2 рулона", BasePrice: 140},
	{ID: "adult-diapers", Name: "Подгузники для взрослых", Unit: "10 шт", Aliases: []string{"Памперсы для взрослых"}, BasePrice: 780},
	{ID: "cat-food", Name: "Корм для кошек", Unit: "85 г", Aliases: []string{"Кошачий корм"}, BasePrice: 40},
}

func findProduct(id string) (Product, bool) {
	for _, p := range products {
		if p.ID == id {
			return p, true
		}
	}
	return Product{}, false
}

func productItems() []CatalogItem {
	items := []CatalogItem{}
	for _, p := range products {
		items = append(items, CatalogItem{ID: p.ID, Name: p.Name, Aliases: p.Aliases})
	}
	return items
}
