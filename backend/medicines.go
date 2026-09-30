package main

type Medicine struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Form      string   `json:"form"`
	Aliases   []string `json:"-"`
	BasePrice int      `json:"-"`
}

var medicines = []Medicine{
	// обезболивающие и жаропонижающие
	{ID: "paracetamol", Name: "Парацетамол", Form: "таблетки 500 мг, 20 шт", Aliases: []string{"Панадол", "Эффералган", "Калпол"}, BasePrice: 45},
	{ID: "ibuprofen", Name: "Ибупрофен", Form: "таблетки 200 мг, 20 шт", Aliases: []string{"Нурофен", "МИГ"}, BasePrice: 70},
	{ID: "aspirin", Name: "Аспирин", Form: "таблетки 500 мг, 10 шт", Aliases: []string{"Ацетилсалициловая кислота"}, BasePrice: 60},
	{ID: "citramon", Name: "Цитрамон", Form: "таблетки, 10 шт", BasePrice: 35},
	{ID: "drotaverin", Name: "Дротаверин", Form: "таблетки 40 мг, 20 шт", Aliases: []string{"Но-шпа", "Спазмол"}, BasePrice: 55},
	{ID: "metamizol", Name: "Анальгин", Form: "таблетки 500 мг, 10 шт", Aliases: []string{"Метамизол натрия"}, BasePrice: 40},
	{ID: "pentalgin", Name: "Пенталгин", Form: "таблетки, 12 шт", BasePrice: 190},
	{ID: "ketorolac", Name: "Кеторолак", Form: "таблетки 10 мг, 20 шт", Aliases: []string{"Кетанов", "Кеторол"}, BasePrice: 60},
	{ID: "nimesulid", Name: "Нимесулид", Form: "таблетки 100 мг, 20 шт", Aliases: []string{"Найз", "Нимесил"}, BasePrice: 120},
	{ID: "diclofenac-tab", Name: "Диклофенак таблетки", Form: "таблетки 50 мг, 20 шт", Aliases: []string{"Вольтарен таблетки"}, BasePrice: 45},
	{ID: "meloxicam", Name: "Мелоксикам", Form: "таблетки 15 мг, 20 шт", Aliases: []string{"Мовалис", "Амелотекс"}, BasePrice: 110},
	{ID: "ketoprofen", Name: "Кетопрофен", Form: "таблетки 100 мг, 20 шт", Aliases: []string{"Кетонал"}, BasePrice: 210},

	// давление
	{ID: "enalapril", Name: "Эналаприл", Form: "таблетки 10 мг, 20 шт", Aliases: []string{"Энап", "Ренитек"}, BasePrice: 65},
	{ID: "lozartan", Name: "Лозартан", Form: "таблетки 50 мг, 30 шт", Aliases: []string{"Лозап", "Козаар"}, BasePrice: 180},
	{ID: "amlodipin", Name: "Амлодипин", Form: "таблетки 5 мг, 30 шт", Aliases: []string{"Норваск"}, BasePrice: 75},
	{ID: "bisoprolol", Name: "Бисопролол", Form: "таблетки 5 мг, 30 шт", Aliases: []string{"Конкор"}, BasePrice: 120},
	{ID: "kaptopril", Name: "Каптоприл", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Капотен"}, BasePrice: 40},
	{ID: "lizinopril", Name: "Лизиноприл", Form: "таблетки 10 мг, 30 шт", Aliases: []string{"Диротон", "Лизорил"}, BasePrice: 110},
	{ID: "perindopril", Name: "Периндоприл", Form: "таблетки 5 мг, 30 шт", Aliases: []string{"Престариум", "Престанс"}, BasePrice: 390},
	{ID: "valsartan", Name: "Валсартан", Form: "таблетки 80 мг, 28 шт", Aliases: []string{"Диован", "Вальсакор"}, BasePrice: 320},
	{ID: "telmisartan", Name: "Телмисартан", Form: "таблетки 40 мг, 28 шт", Aliases: []string{"Микардис", "Телзап"}, BasePrice: 350},
	{ID: "indapamid", Name: "Индапамид", Form: "таблетки 2,5 мг, 30 шт", Aliases: []string{"Арифон", "Равел"}, BasePrice: 90},
	{ID: "hydrochlorothiazide", Name: "Гидрохлоротиазид", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Гипотиазид"}, BasePrice: 100},
	{ID: "metoprolol", Name: "Метопролол", Form: "таблетки 50 мг, 60 шт", Aliases: []string{"Эгилок", "Беталок"}, BasePrice: 140},
	{ID: "moxonidin", Name: "Моксонидин", Form: "таблетки 0,4 мг, 28 шт", Aliases: []string{"Физиотенз"}, BasePrice: 330},
	{ID: "nifedipin", Name: "Нифедипин", Form: "таблетки 10 мг, 50 шт", Aliases: []string{"Коринфар", "Кордафлекс"}, BasePrice: 70},
	{ID: "spironolakton", Name: "Спиронолактон", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Верошпирон"}, BasePrice: 110},
	{ID: "furosemid", Name: "Фуросемид", Form: "таблетки 40 мг, 50 шт", BasePrice: 40},
	{ID: "torasemid", Name: "Торасемид", Form: "таблетки 10 мг, 30 шт", Aliases: []string{"Диувер", "Бритомар"}, BasePrice: 370},

	// сердце и сосуды
	{ID: "atorvastatin", Name: "Аторвастатин", Form: "таблетки 20 мг, 30 шт", Aliases: []string{"Аторис", "Липримар"}, BasePrice: 250},
	{ID: "rosuvastatin", Name: "Розувастатин", Form: "таблетки 10 мг, 30 шт", Aliases: []string{"Крестор", "Мертенил", "Розукард"}, BasePrice: 390},
	{ID: "simvastatin", Name: "Симвастатин", Form: "таблетки 20 мг, 30 шт", Aliases: []string{"Вазилип", "Зокор"}, BasePrice: 180},
	{ID: "cardiomagnil", Name: "Кардиомагнил", Form: "таблетки 75 мг, 100 шт", Aliases: []string{"Тромбо АСС"}, BasePrice: 270},
	{ID: "validol", Name: "Валидол", Form: "таблетки 60 мг, 10 шт", BasePrice: 45},
	{ID: "corvalol", Name: "Корвалол", Form: "капли, 25 мл", Aliases: []string{"Валокордин"}, BasePrice: 50},
	{ID: "nitroglycerin", Name: "Нитроглицерин", Form: "таблетки 0,5 мг, 40 шт", Aliases: []string{"Нитроминт"}, BasePrice: 70},
	{ID: "panangin", Name: "Панангин", Form: "таблетки, 50 шт", Aliases: []string{"Аспаркам"}, BasePrice: 170},
	{ID: "warfarin", Name: "Варфарин", Form: "таблетки 2,5 мг, 100 шт", BasePrice: 150},
	{ID: "rivaroxaban", Name: "Ксарелто", Form: "таблетки 20 мг, 28 шт", Aliases: []string{"Ривароксабан"}, BasePrice: 2900},
	{ID: "apixaban", Name: "Эликвис", Form: "таблетки 5 мг, 60 шт", Aliases: []string{"Апиксабан"}, BasePrice: 2800},
	{ID: "clopidogrel", Name: "Клопидогрел", Form: "таблетки 75 мг, 28 шт", Aliases: []string{"Плавикс", "Зилт"}, BasePrice: 300},
	{ID: "trimetazidin", Name: "Триметазидин", Form: "таблетки 35 мг, 60 шт", Aliases: []string{"Предуктал", "Милдронат"}, BasePrice: 320},
	{ID: "isosorbid", Name: "Изосорбида мононитрат", Form: "таблетки 40 мг, 30 шт", Aliases: []string{"Моночинкве", "Пектрол"}, BasePrice: 200},
	{ID: "digoxin", Name: "Дигоксин", Form: "таблетки 0,25 мг, 50 шт", BasePrice: 60},
	{ID: "motherwort", Name: "Пустырник", Form: "таблетки, 50 шт", Aliases: []string{"Пустырника экстракт"}, BasePrice: 60},
	{ID: "valerian", Name: "Валериана", Form: "таблетки 20 мг, 50 шт", Aliases: []string{"Валерианы экстракт"}, BasePrice: 55},
	{ID: "persen", Name: "Персен", Form: "таблетки, 40 шт", Aliases: []string{"Ново-пассит"}, BasePrice: 450},

	// диабет
	{ID: "metformin", Name: "Метформин", Form: "таблетки 850 мг, 60 шт", Aliases: []string{"Глюкофаж", "Сиофор"}, BasePrice: 160},
	{ID: "gliklazid", Name: "Гликлазид", Form: "таблетки 60 мг, 30 шт", Aliases: []string{"Диабетон", "Диабефарм"}, BasePrice: 250},
	{ID: "glimepirid", Name: "Глимепирид", Form: "таблетки 2 мг, 30 шт", Aliases: []string{"Амарил"}, BasePrice: 280},
	{ID: "vildagliptin", Name: "Вилдаглиптин", Form: "таблетки 50 мг, 28 шт", Aliases: []string{"Галвус"}, BasePrice: 800},
	{ID: "sitagliptin", Name: "Ситаглиптин", Form: "таблетки 100 мг, 28 шт", Aliases: []string{"Янувия"}, BasePrice: 1300},
	{ID: "dapagliflozin", Name: "Дапаглифлозин", Form: "таблетки 10 мг, 30 шт", Aliases: []string{"Форсига"}, BasePrice: 2300},
	{ID: "test-strips", Name: "Тест-полоски для глюкометра", Form: "50 шт", Aliases: []string{"Тест полоски", "Полоски для сахара"}, BasePrice: 900},
	{ID: "lancets", Name: "Ланцеты для глюкометра", Form: "100 шт", Aliases: []string{"Ланцеты"}, BasePrice: 450},
	{ID: "thioctic-acid", Name: "Тиоктовая кислота", Form: "таблетки 600 мг, 30 шт", Aliases: []string{"Тиогамма", "Берлитион", "Октолипен"}, BasePrice: 650},

	// щитовидная железа
	{ID: "levothyroxine", Name: "Левотироксин", Form: "таблетки 50 мкг, 100 шт", Aliases: []string{"Эутирокс", "L-тироксин"}, BasePrice: 150},
	{ID: "iodomarin", Name: "Йодомарин", Form: "таблетки 100 мкг, 100 шт", Aliases: []string{"Калия йодид"}, BasePrice: 170},

	// суставы и спина
	{ID: "diclofenac-gel", Name: "Диклофенак гель", Form: "гель 1%, 50 г", Aliases: []string{"Вольтарен эмульгель", "Диклак"}, BasePrice: 120},
	{ID: "ibuprofen-gel", Name: "Ибупрофен гель", Form: "гель 5%, 50 г", Aliases: []string{"Долгит", "Нурофен экспресс гель"}, BasePrice: 150},
	{ID: "ketoprofen-gel", Name: "Кетопрофен гель", Form: "гель 2,5%, 50 г", Aliases: []string{"Фастум гель", "Быструм гель"}, BasePrice: 280},
	{ID: "chondroitin", Name: "Хондроитин", Form: "капсулы 250 мг, 50 шт", Aliases: []string{"Хондроксид", "Структум"}, BasePrice: 450},
	{ID: "teraflex", Name: "Терафлекс", Form: "капсулы, 60 шт", Aliases: []string{"Глюкозамин", "Артра"}, BasePrice: 1400},
	{ID: "kapsikam", Name: "Капсикам", Form: "мазь, 50 г", Aliases: []string{"Финалгон"}, BasePrice: 380},
	{ID: "menovazin", Name: "Меновазин", Form: "раствор, 40 мл", BasePrice: 60},
	{ID: "mydocalm", Name: "Мидокалм", Form: "таблетки 150 мг, 30 шт", Aliases: []string{"Толперизон"}, BasePrice: 520},
	{ID: "milgamma", Name: "Мильгамма", Form: "таблетки, 30 шт", Aliases: []string{"Комбилипен", "Нейромультивит"}, BasePrice: 700},
	{ID: "allopurinol", Name: "Аллопуринол", Form: "таблетки 100 мг, 50 шт", BasePrice: 100},
	{ID: "calcium-d3", Name: "Кальций Д3 Никомед", Form: "таблетки жевательные, 60 шт", Aliases: []string{"Кальций с витамином Д3", "Компливит Кальций Д3"}, BasePrice: 480},
	{ID: "alendronat", Name: "Алендронат", Form: "таблетки 70 мг, 4 шт", Aliases: []string{"Фосамакс", "Остерепар"}, BasePrice: 450},

	// желудок и кишечник
	{ID: "omeprazol", Name: "Омепразол", Form: "капсулы 20 мг, 30 шт", Aliases: []string{"Омез", "Ультоп"}, BasePrice: 95},
	{ID: "pantoprazol", Name: "Пантопразол", Form: "таблетки 40 мг, 28 шт", Aliases: []string{"Нольпаза", "Контролок"}, BasePrice: 300},
	{ID: "rabeprazol", Name: "Рабепразол", Form: "капсулы 20 мг, 28 шт", Aliases: []string{"Париет", "Разо"}, BasePrice: 450},
	{ID: "famotidin", Name: "Фамотидин", Form: "таблетки 20 мг, 20 шт", Aliases: []string{"Квамател"}, BasePrice: 70},
	{ID: "pankreatin", Name: "Панкреатин", Form: "таблетки, 20 шт", Aliases: []string{"Мезим", "Креон"}, BasePrice: 80},
	{ID: "smekta", Name: "Смекта", Form: "порошок, 10 пакетиков", Aliases: []string{"Диосмектит"}, BasePrice: 190},
	{ID: "coal", Name: "Активированный уголь", Form: "таблетки 250 мг, 10 шт", Aliases: []string{"Уголь"}, BasePrice: 20},
	{ID: "loperamid", Name: "Лоперамид", Form: "капсулы 2 мг, 20 шт", Aliases: []string{"Имодиум"}, BasePrice: 50},
	{ID: "almagel", Name: "Алмагель", Form: "суспензия, 170 мл", Aliases: []string{"Маалокс", "Фосфалюгель"}, BasePrice: 280},
	{ID: "rennie", Name: "Ренни", Form: "таблетки жевательные, 24 шт", Aliases: []string{"Гевискон"}, BasePrice: 300},
	{ID: "motilium", Name: "Домперидон", Form: "таблетки 10 мг, 30 шт", Aliases: []string{"Мотилиум", "Мотилак"}, BasePrice: 400},
	{ID: "trimebutin", Name: "Тримедат", Form: "таблетки 200 мг, 30 шт", Aliases: []string{"Тримебутин"}, BasePrice: 450},
	{ID: "duspatalin", Name: "Дюспаталин", Form: "капсулы 200 мг, 30 шт", Aliases: []string{"Мебеверин", "Спарекс"}, BasePrice: 600},
	{ID: "espumizan", Name: "Эспумизан", Form: "капсулы 40 мг, 50 шт", Aliases: []string{"Симетикон"}, BasePrice: 450},
	{ID: "lactulose", Name: "Лактулоза", Form: "сироп, 200 мл", Aliases: []string{"Дюфалак", "Нормазе"}, BasePrice: 350},
	{ID: "senade", Name: "Сенаде", Form: "таблетки 13,5 мг, 20 шт", Aliases: []string{"Сенна", "Бисакодил"}, BasePrice: 90},
	{ID: "forlax", Name: "Форлакс", Form: "порошок, 20 пакетиков", Aliases: []string{"Макрогол"}, BasePrice: 400},
	{ID: "linex", Name: "Линекс", Form: "капсулы, 32 шт", Aliases: []string{"Бифиформ", "Аципол"}, BasePrice: 650},
	{ID: "enterofuril", Name: "Энтерофурил", Form: "капсулы 200 мг, 16 шт", Aliases: []string{"Нифуроксазид"}, BasePrice: 450},
	{ID: "enterosgel", Name: "Энтеросгель", Form: "паста, 225 г", Aliases: []string{"Полисорб"}, BasePrice: 550},
	{ID: "ursosan", Name: "Урсосан", Form: "капсулы 250 мг, 50 шт", Aliases: []string{"Урсофальк", "Урсодезоксихолевая кислота"}, BasePrice: 900},
	{ID: "essentiale", Name: "Эссенциале форте", Form: "капсулы 300 мг, 30 шт", Aliases: []string{"Эссенциале", "Фосфоглив"}, BasePrice: 700},
	{ID: "heptral", Name: "Гептрал", Form: "таблетки 400 мг, 20 шт", Aliases: []string{"Адеметионин"}, BasePrice: 1700},
	{ID: "rehydron", Name: "Регидрон", Form: "порошок, 20 пакетиков", BasePrice: 400},

	// простуда, горло, кашель, нос
	{ID: "amoxicillin", Name: "Амоксициллин", Form: "капсулы 500 мг, 16 шт", Aliases: []string{"Флемоксин", "Амосин"}, BasePrice: 110},
	{ID: "azithromycin", Name: "Азитромицин", Form: "таблетки 500 мг, 3 шт", Aliases: []string{"Сумамед", "Азитрал"}, BasePrice: 150},
	{ID: "nazivin", Name: "Називин", Form: "спрей назальный, 10 мл", Aliases: []string{"Оксиметазолин"}, BasePrice: 200},
	{ID: "ksilometazolin", Name: "Ксилометазолин", Form: "спрей назальный 0,1%, 10 мл", Aliases: []string{"Отривин", "Ксимелин", "Галазолин"}, BasePrice: 120},
	{ID: "aquamaris", Name: "Аквамарис", Form: "спрей назальный, 30 мл", Aliases: []string{"Аквалор", "Морская вода для носа"}, BasePrice: 350},
	{ID: "ambroxol", Name: "Амброксол", Form: "таблетки 30 мг, 20 шт", Aliases: []string{"Лазолван", "Амбробене"}, BasePrice: 55},
	{ID: "acc", Name: "АЦЦ", Form: "таблетки шипучие 600 мг, 10 шт", Aliases: []string{"Ацетилцистеин", "Флуимуцил"}, BasePrice: 350},
	{ID: "bromhexin", Name: "Бромгексин", Form: "таблетки 8 мг, 20 шт", BasePrice: 50},
	{ID: "gerbion", Name: "Гербион", Form: "сироп подорожника, 150 мл", Aliases: []string{"Сироп от кашля", "Доктор Мом"}, BasePrice: 400},
	{ID: "sinekod", Name: "Синекод", Form: "капли, 20 мл", Aliases: []string{"Бутамират", "Омнитус"}, BasePrice: 450},
	{ID: "strepsils", Name: "Стрепсилс", Form: "таблетки для рассасывания, 24 шт", Aliases: []string{"Граммидин", "Фарингосепт"}, BasePrice: 380},
	{ID: "tantum-verde", Name: "Тантум Верде", Form: "спрей для горла, 30 мл", Aliases: []string{"Гексорал", "Ингалипт"}, BasePrice: 420},
	{ID: "miramistin", Name: "Мирамистин", Form: "раствор со спреем, 150 мл", BasePrice: 450},
	{ID: "chlorhexidine", Name: "Хлоргексидин", Form: "раствор 0,05%, 100 мл", BasePrice: 30},
	{ID: "theraflu", Name: "Терафлю", Form: "порошок, 10 пакетиков", Aliases: []string{"Фервекс", "Колдрекс"}, BasePrice: 500},
	{ID: "arbidol", Name: "Арбидол", Form: "капсулы 100 мг, 20 шт", Aliases: []string{"Умифеновир"}, BasePrice: 500},
	{ID: "ingavirin", Name: "Ингавирин", Form: "капсулы 90 мг, 7 шт", BasePrice: 550},
	{ID: "kagocel", Name: "Кагоцел", Form: "таблетки 12 мг, 20 шт", BasePrice: 480},
	{ID: "ergoferon", Name: "Эргоферон", Form: "таблетки для рассасывания, 20 шт", Aliases: []string{"Анаферон"}, BasePrice: 450},
	{ID: "otipax", Name: "Отипакс", Form: "капли ушные, 16 г", Aliases: []string{"Отинум"}, BasePrice: 380},

	// аллергия
	{ID: "loratadin", Name: "Лоратадин", Form: "таблетки 10 мг, 10 шт", Aliases: []string{"Кларитин", "Кларотадин"}, BasePrice: 90},
	{ID: "suprastin", Name: "Супрастин", Form: "таблетки 25 мг, 20 шт", Aliases: []string{"Хлоропирамин"}, BasePrice: 150},
	{ID: "cetirizin", Name: "Цетиризин", Form: "таблетки 10 мг, 10 шт", Aliases: []string{"Зиртек", "Зодак"}, BasePrice: 110},
	{ID: "desloratadin", Name: "Дезлоратадин", Form: "таблетки 5 мг, 10 шт", Aliases: []string{"Эриус", "Эзлор"}, BasePrice: 250},
	{ID: "levocetirizin", Name: "Левоцетиризин", Form: "таблетки 5 мг, 10 шт", Aliases: []string{"Супрастинекс", "Ксизал"}, BasePrice: 300},
	{ID: "fenistil", Name: "Фенистил", Form: "гель 0,1%, 30 г", Aliases: []string{"Фенистил гель"}, BasePrice: 450},
	{ID: "avamys", Name: "Авамис", Form: "спрей назальный, 120 доз", Aliases: []string{"Назонекс", "Мометазон"}, BasePrice: 700},

	// витамины и минералы
	{ID: "vitamin-d3", Name: "Витамин D3", Form: "капли, 10 мл", Aliases: []string{"Витамин Д3", "Аквадетрим", "Колекальциферол"}, BasePrice: 220},
	{ID: "magne-b6", Name: "Магне B6", Form: "таблетки, 60 шт", Aliases: []string{"Магнелис", "Магний B6"}, BasePrice: 600},
	{ID: "complivit", Name: "Компливит", Form: "таблетки, 60 шт", Aliases: []string{"Мультивитамины"}, BasePrice: 250},
	{ID: "centrum-silver", Name: "Центрум Сильвер 50+", Form: "таблетки, 30 шт", Aliases: []string{"Центрум", "Витамины 50 плюс"}, BasePrice: 750},
	{ID: "vitamin-c", Name: "Витамин C", Form: "драже 50 мг, 200 шт", Aliases: []string{"Аскорбинка", "Аскорбиновая кислота"}, BasePrice: 70},
	{ID: "omega-3", Name: "Омега-3", Form: "капсулы 1000 мг, 30 шт", Aliases: []string{"Рыбий жир"}, BasePrice: 400},
	{ID: "folic-acid", Name: "Фолиевая кислота", Form: "таблетки 1 мг, 50 шт", BasePrice: 60},
	{ID: "ferrum", Name: "Сорбифер Дурулес", Form: "таблетки 100 мг, 50 шт", Aliases: []string{"Железо", "Мальтофер", "Тотема"}, BasePrice: 550},
	{ID: "vitamin-b12", Name: "Витамин B12", Form: "таблетки, 60 шт", Aliases: []string{"Цианокобаламин"}, BasePrice: 300},
	{ID: "glycine", Name: "Глицин", Form: "таблетки 100 мг, 50 шт", BasePrice: 45},

	// память, нервы, сон
	{ID: "piracetam", Name: "Пирацетам", Form: "таблетки 400 мг, 60 шт", Aliases: []string{"Ноотропил", "Луцетам"}, BasePrice: 120},
	{ID: "mexidol", Name: "Мексидол", Form: "таблетки 125 мг, 50 шт", Aliases: []string{"Мексиприм", "Этилметилгидроксипиридина сукцинат"}, BasePrice: 450},
	{ID: "betahistine", Name: "Бетагистин", Form: "таблетки 24 мг, 30 шт", Aliases: []string{"Бетасерк", "Вестибо"}, BasePrice: 450},
	{ID: "cinnarizin", Name: "Циннаризин", Form: "таблетки 25 мг, 50 шт", BasePrice: 60},
	{ID: "vinpocetin", Name: "Винпоцетин", Form: "таблетки 5 мг, 50 шт", Aliases: []string{"Кавинтон"}, BasePrice: 150},
	{ID: "tanakan", Name: "Танакан", Form: "таблетки 40 мг, 30 шт", Aliases: []string{"Гинкго билоба", "Билобил"}, BasePrice: 600},
	{ID: "melatonin", Name: "Мелатонин", Form: "таблетки 3 мг, 30 шт", Aliases: []string{"Мелаксен"}, BasePrice: 500},
	{ID: "donormil", Name: "Донормил", Form: "таблетки 15 мг, 30 шт", Aliases: []string{"Доксиламин"}, BasePrice: 400},
	{ID: "afobazol", Name: "Афобазол", Form: "таблетки 10 мг, 60 шт", BasePrice: 450},
	{ID: "tenoten", Name: "Тенотен", Form: "таблетки для рассасывания, 40 шт", BasePrice: 350},

	// мочеполовая система
	{ID: "tamsulosin", Name: "Тамсулозин", Form: "капсулы 0,4 мг, 30 шт", Aliases: []string{"Омник", "Фокусин"}, BasePrice: 450},
	{ID: "prostamol", Name: "Простамол Уно", Form: "капсулы 320 мг, 30 шт", Aliases: []string{"Простамол"}, BasePrice: 900},
	{ID: "canephron", Name: "Канефрон", Form: "драже, 60 шт", Aliases: []string{"Цистон"}, BasePrice: 550},
	{ID: "monural", Name: "Монурал", Form: "гранулы 3 г, 1 пакетик", Aliases: []string{"Фосфомицин"}, BasePrice: 500},

	// глаза
	{ID: "taufon", Name: "Тауфон", Form: "капли глазные 4%, 10 мл", Aliases: []string{"Таурин"}, BasePrice: 150},
	{ID: "vizin", Name: "Визин", Form: "капли глазные, 15 мл", Aliases: []string{"Визин классический"}, BasePrice: 450},
	{ID: "artificial-tears", Name: "Искусственная слеза", Form: "капли глазные, 10 мл", Aliases: []string{"Хилокомод", "Систейн", "Слеза натуральная"}, BasePrice: 350},
	{ID: "timolol", Name: "Тимолол", Form: "капли глазные 0,5%, 5 мл", Aliases: []string{"Арутимол"}, BasePrice: 80},
	{ID: "xalatan", Name: "Ксалатан", Form: "капли глазные, 2,5 мл", Aliases: []string{"Латанопрост", "Глаупрост"}, BasePrice: 700},
	{ID: "levomycetin-drops", Name: "Левомицетин капли", Form: "капли глазные 0,25%, 10 мл", Aliases: []string{"Тобрекс", "Альбуцид"}, BasePrice: 50},
	{ID: "quinax", Name: "Квинакс", Form: "капли глазные, 15 мл", Aliases: []string{"Катахром"}, BasePrice: 800},
	{ID: "lutein", Name: "Лютеин комплекс", Form: "таблетки, 60 шт", Aliases: []string{"Черника форте", "Окувайт"}, BasePrice: 600},

	// мази, кожа, раны
	{ID: "levomekol", Name: "Левомеколь", Form: "мазь, 40 г", BasePrice: 150},
	{ID: "bepanten", Name: "Бепантен", Form: "крем 5%, 30 г", Aliases: []string{"Декспантенол", "Пантенол"}, BasePrice: 500},
	{ID: "troxevasin", Name: "Троксевазин", Form: "гель 2%, 40 г", Aliases: []string{"Троксерутин"}, BasePrice: 250},
	{ID: "heparin-ointment", Name: "Гепариновая мазь", Form: "мазь, 25 г", Aliases: []string{"Лиотон", "Гепатромбин"}, BasePrice: 90},
	{ID: "detralex", Name: "Детралекс", Form: "таблетки 1000 мг, 30 шт", Aliases: []string{"Флебодиа", "Венарус"}, BasePrice: 1500},
	{ID: "relief", Name: "Релиф", Form: "свечи, 12 шт", Aliases: []string{"Проктозан", "Постеризан"}, BasePrice: 550},
	{ID: "clotrimazol", Name: "Клотримазол", Form: "крем 1%, 20 г", Aliases: []string{"Кандид"}, BasePrice: 70},
	{ID: "terbinafin", Name: "Тербинафин", Form: "крем 1%, 15 г", Aliases: []string{"Ламизил", "Экзифин"}, BasePrice: 120},
	{ID: "hydrocortisone", Name: "Гидрокортизоновая мазь", Form: "мазь 1%, 10 г", Aliases: []string{"Адвантан", "Акридерм"}, BasePrice: 60},
	{ID: "zinc-ointment", Name: "Цинковая мазь", Form: "мазь 10%, 25 г", Aliases: []string{"Судокрем"}, BasePrice: 45},
	{ID: "vishnevsky", Name: "Мазь Вишневского", Form: "линимент, 30 г", Aliases: []string{"Вишневского"}, BasePrice: 80},
	{ID: "rescuer", Name: "Спасатель", Form: "бальзам, 30 г", Aliases: []string{"Бальзам Спасатель"}, BasePrice: 200},
	{ID: "iodine", Name: "Йод", Form: "раствор 5%, 10 мл", BasePrice: 30},
	{ID: "brilliant-green", Name: "Зелёнка", Form: "раствор 1%, 10 мл", Aliases: []string{"Бриллиантовый зелёный"}, BasePrice: 30},
	{ID: "hydrogen-peroxide", Name: "Перекись водорода", Form: "раствор 3%, 100 мл", BasePrice: 25},
	{ID: "patch", Name: "Лейкопластырь", Form: "набор, 20 шт", Aliases: []string{"Пластырь"}, BasePrice: 90},
	{ID: "bandage", Name: "Бинт стерильный", Form: "7 м × 14 см", Aliases: []string{"Бинт"}, BasePrice: 40},
	{ID: "elastic-bandage", Name: "Бинт эластичный", Form: "1,5 м × 8 см", Aliases: []string{"Эластичный бинт"}, BasePrice: 150},
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
