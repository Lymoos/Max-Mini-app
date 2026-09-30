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
	{ID: "adult-diapers", Name: "Подгузники для взрослых", Unit: "10 шт", Aliases: []string{"Памперсы для взрослых", "Подгузники"}, BasePrice: 780},
	{ID: "cat-food", Name: "Корм для кошек", Unit: "85 г", Aliases: []string{"Кошачий корм"}, BasePrice: 40},

	// молочные продукты
	{ID: "ryazhenka", Name: "Ряженка 4%", Unit: "500 мл", Aliases: []string{"Ряженка"}, BasePrice: 85},
	{ID: "prostokvasha", Name: "Простокваша 2,5%", Unit: "900 мл", Aliases: []string{"Простокваша"}, BasePrice: 90},
	{ID: "cream", Name: "Сливки 10%", Unit: "500 мл", Aliases: []string{"Сливки"}, BasePrice: 150},
	{ID: "cottage-cheese-grain", Name: "Творог зернёный", Unit: "130 г", Aliases: []string{"Зернёный творог"}, BasePrice: 110},
	{ID: "processed-cheese", Name: "Сыр плавленый", Unit: "200 г", Aliases: []string{"Плавленый сырок"}, BasePrice: 140},
	{ID: "condensed-milk", Name: "Молоко сгущённое", Unit: "380 г", Aliases: []string{"Сгущёнка"}, BasePrice: 120},
	{ID: "milk-lactose-free", Name: "Молоко безлактозное", Unit: "1 л", Aliases: []string{"Безлактозное молоко"}, BasePrice: 130},

	// хлеб и выпечка
	{ID: "bread-rye", Name: "Хлеб ржаной", Unit: "500 г", Aliases: []string{"Ржаной хлеб"}, BasePrice: 60},
	{ID: "bread-bran", Name: "Хлеб с отрубями", Unit: "300 г", Aliases: []string{"Хлеб отрубной"}, BasePrice: 70},
	{ID: "crispbread", Name: "Хлебцы цельнозерновые", Unit: "100 г", Aliases: []string{"Хлебцы"}, BasePrice: 90},
	{ID: "sushki", Name: "Сушки", Unit: "300 г", Aliases: []string{"Баранки"}, BasePrice: 70},
	{ID: "gingerbread", Name: "Пряники", Unit: "300 г", Aliases: []string{"Пряник"}, BasePrice: 90},
	{ID: "crackers", Name: "Сухари", Unit: "250 г", Aliases: []string{"Сухарики"}, BasePrice: 60},

	// крупы и бакалея
	{ID: "millet", Name: "Пшено", Unit: "900 г", Aliases: []string{"Пшённая крупа"}, BasePrice: 60},
	{ID: "semolina", Name: "Манная крупа", Unit: "700 г", Aliases: []string{"Манка"}, BasePrice: 60},
	{ID: "pearl-barley", Name: "Перловка", Unit: "900 г", Aliases: []string{"Перловая крупа"}, BasePrice: 55},
	{ID: "peas", Name: "Горох колотый", Unit: "800 г", Aliases: []string{"Горох"}, BasePrice: 60},
	{ID: "lentils", Name: "Чечевица", Unit: "450 г", BasePrice: 110},
	{ID: "beans-canned", Name: "Фасоль консервированная", Unit: "400 г", Aliases: []string{"Фасоль"}, BasePrice: 110},
	{ID: "green-peas-canned", Name: "Горошек зелёный", Unit: "400 г", Aliases: []string{"Зелёный горошек"}, BasePrice: 100},
	{ID: "corn-canned", Name: "Кукуруза консервированная", Unit: "340 г", Aliases: []string{"Кукуруза"}, BasePrice: 100},
	{ID: "sugar-substitute", Name: "Сахарозаменитель", Unit: "650 таблеток", Aliases: []string{"Заменитель сахара", "Подсластитель"}, BasePrice: 200},
	{ID: "starch", Name: "Крахмал картофельный", Unit: "400 г", Aliases: []string{"Крахмал"}, BasePrice: 70},
	{ID: "yeast", Name: "Дрожжи сухие", Unit: "11 г", Aliases: []string{"Дрожжи"}, BasePrice: 30},
	{ID: "soda", Name: "Сода пищевая", Unit: "500 г", Aliases: []string{"Сода"}, BasePrice: 50},
	{ID: "vinegar", Name: "Уксус столовый 9%", Unit: "500 мл", Aliases: []string{"Уксус"}, BasePrice: 40},
	{ID: "mayonnaise", Name: "Майонез", Unit: "400 г", BasePrice: 120},
	{ID: "ketchup", Name: "Кетчуп", Unit: "350 г", BasePrice: 110},
	{ID: "tomato-paste", Name: "Томатная паста", Unit: "270 г", BasePrice: 90},
	{ID: "olive-oil", Name: "Масло оливковое", Unit: "500 мл", Aliases: []string{"Оливковое масло"}, BasePrice: 650},
	{ID: "black-pepper", Name: "Перец чёрный молотый", Unit: "50 г", Aliases: []string{"Перец молотый"}, BasePrice: 60},
	{ID: "bay-leaf", Name: "Лавровый лист", Unit: "10 г", BasePrice: 30},
	{ID: "canned-fish", Name: "Сайра консервированная", Unit: "250 г", Aliases: []string{"Рыбные консервы", "Сайра"}, BasePrice: 150},
	{ID: "sprats", Name: "Шпроты", Unit: "160 г", BasePrice: 140},
	{ID: "stew", Name: "Тушёнка говяжья", Unit: "325 г", Aliases: []string{"Тушёнка", "Тушенка"}, BasePrice: 250},
	{ID: "jam", Name: "Варенье клубничное", Unit: "350 г", Aliases: []string{"Варенье", "Джем"}, BasePrice: 180},
	{ID: "dried-fruits", Name: "Сухофрукты", Unit: "500 г", Aliases: []string{"Компотная смесь", "Чернослив", "Курага"}, BasePrice: 250},
	{ID: "walnuts", Name: "Грецкие орехи", Unit: "200 г", Aliases: []string{"Орехи"}, BasePrice: 300},
	{ID: "raisins", Name: "Изюм", Unit: "200 г", BasePrice: 120},

	// мясо и рыба
	{ID: "pork", Name: "Свинина лопатка", Unit: "1 кг", Aliases: []string{"Свинина"}, BasePrice: 450},
	{ID: "beef", Name: "Говядина", Unit: "1 кг", Aliases: []string{"Говядина мякоть"}, BasePrice: 750},
	{ID: "chicken-legs", Name: "Окорочка куриные", Unit: "1 кг", Aliases: []string{"Окорочка"}, BasePrice: 260},
	{ID: "turkey", Name: "Филе индейки", Unit: "600 г", Aliases: []string{"Индейка"}, BasePrice: 450},
	{ID: "liver", Name: "Печень куриная", Unit: "500 г", Aliases: []string{"Печень", "Печёнка"}, BasePrice: 170},
	{ID: "dumplings", Name: "Пельмени", Unit: "800 г", BasePrice: 350},
	{ID: "cutlets", Name: "Котлеты замороженные", Unit: "600 г", Aliases: []string{"Котлеты"}, BasePrice: 300},
	{ID: "herring", Name: "Сельдь слабосолёная", Unit: "300 г", Aliases: []string{"Селёдка", "Сельдь"}, BasePrice: 220},
	{ID: "salmon", Name: "Горбуша замороженная", Unit: "1 кг", Aliases: []string{"Горбуша", "Красная рыба"}, BasePrice: 480},
	{ID: "cod", Name: "Треска филе", Unit: "500 г", Aliases: []string{"Треска"}, BasePrice: 400},

	// овощи, фрукты, зелень
	{ID: "beet", Name: "Свёкла", Unit: "1 кг", Aliases: []string{"Свекла", "Бурак"}, BasePrice: 45},
	{ID: "garlic", Name: "Чеснок", Unit: "100 г", BasePrice: 40},
	{ID: "pepper-bell", Name: "Перец болгарский", Unit: "1 кг", Aliases: []string{"Сладкий перец"}, BasePrice: 300},
	{ID: "zucchini", Name: "Кабачки", Unit: "1 кг", Aliases: []string{"Кабачок"}, BasePrice: 120},
	{ID: "greens", Name: "Укроп", Unit: "50 г", Aliases: []string{"Зелень", "Петрушка"}, BasePrice: 50},
	{ID: "pear", Name: "Груши", Unit: "1 кг", Aliases: []string{"Груша"}, BasePrice: 220},
	{ID: "mandarin", Name: "Мандарины", Unit: "1 кг", Aliases: []string{"Мандарин"}, BasePrice: 200},
	{ID: "grapes", Name: "Виноград", Unit: "1 кг", BasePrice: 300},
	{ID: "frozen-vegetables", Name: "Овощная смесь замороженная", Unit: "400 г", Aliases: []string{"Замороженные овощи"}, BasePrice: 150},
	{ID: "frozen-berries", Name: "Ягоды замороженные", Unit: "300 г", Aliases: []string{"Замороженные ягоды", "Клюква"}, BasePrice: 220},
	{ID: "sauerkraut", Name: "Капуста квашеная", Unit: "500 г", Aliases: []string{"Квашеная капуста"}, BasePrice: 120},
	{ID: "pickles", Name: "Огурцы маринованные", Unit: "680 г", Aliases: []string{"Солёные огурцы"}, BasePrice: 160},

	// напитки и сладкое
	{ID: "tea-green", Name: "Чай зелёный", Unit: "25 пакетиков", Aliases: []string{"Зелёный чай"}, BasePrice: 120},
	{ID: "tea-herbal", Name: "Чай травяной", Unit: "20 пакетиков", Aliases: []string{"Травяной чай", "Ромашковый чай"}, BasePrice: 130},
	{ID: "coffee-ground", Name: "Кофе молотый", Unit: "250 г", Aliases: []string{"Молотый кофе"}, BasePrice: 400},
	{ID: "chicory", Name: "Цикорий растворимый", Unit: "100 г", Aliases: []string{"Цикорий"}, BasePrice: 150},
	{ID: "cocoa", Name: "Какао-порошок", Unit: "100 г", Aliases: []string{"Какао"}, BasePrice: 150},
	{ID: "mineral-water", Name: "Вода минеральная", Unit: "1,5 л", Aliases: []string{"Минералка", "Ессентуки", "Боржоми"}, BasePrice: 90},
	{ID: "kissel", Name: "Кисель", Unit: "220 г", BasePrice: 50},
	{ID: "compote", Name: "Морс клюквенный", Unit: "1 л", Aliases: []string{"Морс", "Компот"}, BasePrice: 120},
	{ID: "marshmallow", Name: "Зефир", Unit: "250 г", BasePrice: 130},
	{ID: "candies", Name: "Конфеты шоколадные", Unit: "250 г", Aliases: []string{"Конфеты"}, BasePrice: 250},
	{ID: "diabetic-sweets", Name: "Сладости для диабетиков", Unit: "100 г", Aliases: []string{"Диабетические конфеты", "Шоколад без сахара"}, BasePrice: 150},
	{ID: "halva", Name: "Халва подсолнечная", Unit: "350 г", Aliases: []string{"Халва"}, BasePrice: 130},

	// бытовая химия
	{ID: "washing-gel", Name: "Гель для стирки", Unit: "1,3 л", Aliases: []string{"Жидкий порошок"}, BasePrice: 450},
	{ID: "fabric-softener", Name: "Кондиционер для белья", Unit: "1 л", Aliases: []string{"Ополаскиватель для белья"}, BasePrice: 250},
	{ID: "bleach", Name: "Отбеливатель", Unit: "1 л", Aliases: []string{"Белизна"}, BasePrice: 70},
	{ID: "stain-remover", Name: "Пятновыводитель", Unit: "450 мл", BasePrice: 300},
	{ID: "toilet-cleaner", Name: "Средство для унитаза", Unit: "750 мл", Aliases: []string{"Утёнок", "Доместос"}, BasePrice: 200},
	{ID: "glass-cleaner", Name: "Средство для стёкол", Unit: "500 мл", Aliases: []string{"Мистер Мускул для стёкол"}, BasePrice: 180},
	{ID: "floor-cleaner", Name: "Средство для мытья полов", Unit: "1 л", BasePrice: 200},
	{ID: "scouring-powder", Name: "Чистящее средство", Unit: "500 мл", Aliases: []string{"Пемолюкс", "Комет"}, BasePrice: 120},
	{ID: "sponges", Name: "Губки для посуды", Unit: "5 шт", Aliases: []string{"Губки"}, BasePrice: 70},
	{ID: "rags", Name: "Салфетки для уборки", Unit: "3 шт", Aliases: []string{"Тряпки", "Микрофибра"}, BasePrice: 150},
	{ID: "rubber-gloves", Name: "Перчатки резиновые", Unit: "1 пара", Aliases: []string{"Хозяйственные перчатки"}, BasePrice: 90},
	{ID: "garbage-bags", Name: "Мешки для мусора", Unit: "30 шт", Aliases: []string{"Пакеты для мусора"}, BasePrice: 100},

	// гигиена и уход
	{ID: "toothbrush", Name: "Зубная щётка", Unit: "1 шт", Aliases: []string{"Щётка зубная"}, BasePrice: 120},
	{ID: "denture-cream", Name: "Крем для фиксации зубных протезов", Unit: "40 г", Aliases: []string{"Корега", "Лакалют фиксатор"}, BasePrice: 450},
	{ID: "denture-tablets", Name: "Таблетки для очистки протезов", Unit: "30 шт", Aliases: []string{"Очиститель протезов"}, BasePrice: 400},
	{ID: "mouthwash", Name: "Ополаскиватель для рта", Unit: "250 мл", BasePrice: 200},
	{ID: "shower-gel", Name: "Гель для душа", Unit: "250 мл", BasePrice: 200},
	{ID: "liquid-soap", Name: "Мыло жидкое", Unit: "500 мл", Aliases: []string{"Жидкое мыло"}, BasePrice: 150},
	{ID: "household-soap", Name: "Мыло хозяйственное", Unit: "200 г", Aliases: []string{"Хозяйственное мыло"}, BasePrice: 50},
	{ID: "hand-cream", Name: "Крем для рук", Unit: "75 мл", BasePrice: 120},
	{ID: "foot-cream", Name: "Крем для ног", Unit: "75 мл", Aliases: []string{"Крем для пяток"}, BasePrice: 180},
	{ID: "face-cream", Name: "Крем для лица", Unit: "50 мл", BasePrice: 300},
	{ID: "deodorant", Name: "Дезодорант", Unit: "150 мл", BasePrice: 250},
	{ID: "shaving-razor", Name: "Станок для бритья", Unit: "1 шт", Aliases: []string{"Бритва", "Станок"}, BasePrice: 300},
	{ID: "shaving-foam", Name: "Пена для бритья", Unit: "200 мл", BasePrice: 200},
	{ID: "cotton-pads", Name: "Ватные диски", Unit: "100 шт", BasePrice: 80},
	{ID: "cotton-buds", Name: "Ватные палочки", Unit: "200 шт", BasePrice: 70},
	{ID: "wet-wipes", Name: "Влажные салфетки", Unit: "72 шт", BasePrice: 120},
	{ID: "paper-napkins", Name: "Салфетки бумажные", Unit: "100 шт", Aliases: []string{"Салфетки"}, BasePrice: 60},
	{ID: "handkerchiefs", Name: "Платочки бумажные", Unit: "10 пачек", Aliases: []string{"Носовые платки"}, BasePrice: 90},
	{ID: "sanitary-pads", Name: "Прокладки урологические", Unit: "10 шт", Aliases: []string{"Урологические прокладки"}, BasePrice: 350},
	{ID: "comb", Name: "Расчёска", Unit: "1 шт", BasePrice: 100},
	{ID: "nail-clippers", Name: "Кусачки для ногтей", Unit: "1 шт", Aliases: []string{"Ножницы маникюрные", "Щипчики для ногтей"}, BasePrice: 200},

	// товары для здоровья
	{ID: "tonometer", Name: "Тонометр автоматический", Unit: "1 шт", Aliases: []string{"Тонометр", "Прибор для давления"}, BasePrice: 2500},
	{ID: "glucometer", Name: "Глюкометр", Unit: "1 шт", Aliases: []string{"Прибор для сахара"}, BasePrice: 1300},
	{ID: "thermometer", Name: "Термометр электронный", Unit: "1 шт", Aliases: []string{"Градусник", "Термометр"}, BasePrice: 300},
	{ID: "pulse-oximeter", Name: "Пульсоксиметр", Unit: "1 шт", Aliases: []string{"Оксиметр"}, BasePrice: 1200},
	{ID: "cane", Name: "Трость опорная", Unit: "1 шт", Aliases: []string{"Трость", "Палочка для ходьбы"}, BasePrice: 900},
	{ID: "walker", Name: "Ходунки", Unit: "1 шт", BasePrice: 3500},
	{ID: "underpads", Name: "Пелёнки впитывающие", Unit: "30 шт", Aliases: []string{"Одноразовые пелёнки", "Пеленки"}, BasePrice: 550},
	{ID: "adult-pants", Name: "Трусы-подгузники для взрослых", Unit: "10 шт", Aliases: []string{"Впитывающие трусы"}, BasePrice: 850},
	{ID: "compression-stockings", Name: "Компрессионные чулки", Unit: "1 пара", Aliases: []string{"Компрессионные гольфы", "Чулки от варикоза"}, BasePrice: 1800},
	{ID: "knee-brace", Name: "Бандаж на колено", Unit: "1 шт", Aliases: []string{"Наколенник"}, BasePrice: 900},
	{ID: "pill-box", Name: "Таблетница на неделю", Unit: "1 шт", Aliases: []string{"Таблетница", "Коробка для таблеток"}, BasePrice: 250},
	{ID: "hot-water-bottle", Name: "Грелка резиновая", Unit: "1 шт", Aliases: []string{"Грелка"}, BasePrice: 300},
	{ID: "hearing-aid-batteries", Name: "Батарейки для слухового аппарата", Unit: "6 шт", Aliases: []string{"Батарейки для слухового"}, BasePrice: 350},
	{ID: "reading-glasses", Name: "Очки для чтения", Unit: "1 шт", Aliases: []string{"Очки"}, BasePrice: 400},
	{ID: "masks", Name: "Маски медицинские", Unit: "50 шт", Aliases: []string{"Маски"}, BasePrice: 150},
	{ID: "antiseptic", Name: "Антисептик для рук", Unit: "100 мл", Aliases: []string{"Санитайзер"}, BasePrice: 100},
	{ID: "anti-slip-mat", Name: "Коврик для ванной противоскользящий", Unit: "1 шт", Aliases: []string{"Коврик для ванной"}, BasePrice: 500},
	{ID: "bath-handrail", Name: "Поручень для ванной", Unit: "1 шт", Aliases: []string{"Поручень"}, BasePrice: 800},

	// товары для дома
	{ID: "batteries-aa", Name: "Батарейки АА", Unit: "4 шт", Aliases: []string{"Батарейки", "Пальчиковые батарейки"}, BasePrice: 200},
	{ID: "light-bulb", Name: "Лампочка светодиодная", Unit: "1 шт", Aliases: []string{"Лампочка", "Лампа"}, BasePrice: 120},
	{ID: "flashlight", Name: "Фонарик", Unit: "1 шт", BasePrice: 350},
	{ID: "matches", Name: "Спички", Unit: "10 коробков", BasePrice: 40},
	{ID: "candles", Name: "Свечи хозяйственные", Unit: "4 шт", Aliases: []string{"Свечи"}, BasePrice: 100},
	{ID: "extension-cord", Name: "Удлинитель", Unit: "3 м", BasePrice: 400},
	{ID: "foil", Name: "Фольга пищевая", Unit: "10 м", Aliases: []string{"Фольга"}, BasePrice: 120},
	{ID: "cling-film", Name: "Плёнка пищевая", Unit: "20 м", Aliases: []string{"Пищевая плёнка"}, BasePrice: 100},
	{ID: "food-bags", Name: "Пакеты для продуктов", Unit: "100 шт", Aliases: []string{"Пакеты фасовочные"}, BasePrice: 70},
	{ID: "baking-paper", Name: "Бумага для выпечки", Unit: "8 м", Aliases: []string{"Пергамент"}, BasePrice: 90},
	{ID: "towel", Name: "Полотенце махровое", Unit: "1 шт", Aliases: []string{"Полотенце"}, BasePrice: 500},
	{ID: "slippers", Name: "Тапочки домашние", Unit: "1 пара", Aliases: []string{"Тапочки", "Тапки"}, BasePrice: 400},
	{ID: "socks", Name: "Носки хлопковые", Unit: "1 пара", Aliases: []string{"Носки"}, BasePrice: 150},
	{ID: "dog-food", Name: "Корм для собак", Unit: "1 кг", Aliases: []string{"Собачий корм"}, BasePrice: 350},
	{ID: "cat-litter", Name: "Наполнитель для кошачьего туалета", Unit: "5 кг", Aliases: []string{"Наполнитель"}, BasePrice: 300},
	{ID: "seeds", Name: "Семена овощей", Unit: "1 пакетик", Aliases: []string{"Семена", "Рассада"}, BasePrice: 40},
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
