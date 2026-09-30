package main

import (
	"strconv"
	"time"
)

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Region: "" — по всей России, "moscow" или "tatarstan" — только там.
// Auto — назначается без заявления, подавать ничего не нужно
type Benefit struct {
	ID        string   `json:"id"`
	Category  string   `json:"category"`
	Title     string   `json:"title"`
	Short     string   `json:"short"`
	Who       string   `json:"who"`
	Documents []string `json:"documents"`
	Steps     []string `json:"steps"`
	Links     []Link   `json:"links"`
	Phone     string   `json:"phone,omitempty"`
	Deadline  string   `json:"deadline,omitempty"`
	Auto      bool     `json:"auto"`
	Region    string   `json:"region,omitempty"`
	MinAge    int      `json:"-"`
	AgeWindow []int    `json:"-"`
}

const sfrPhone = "8 800 100-00-01"

var benefitCategories = []string{"Документы", "Пенсия", "Льготы", "Досуг"}

var benefits = []Benefit{
	{
		ID: "passport", Category: "Документы",
		Title: "Замена паспорта в 20 и 45 лет",
		Short: "Паспорт нужно заменить в течение 90 дней после дня рождения",
		Who:   "Всем, кому исполнилось 20 или 45 лет. После 45 лет паспорт действует бессрочно.",
		Documents: []string{
			"Старый паспорт",
			"Две фотографии 35×45 мм (если подаёте не через Госуслуги)",
			"Квитанция об оплате госпошлины",
		},
		Steps: []string{
			"Зайдите на Госуслуги заранее, пока старый паспорт ещё действует — после срока войти будет сложнее",
			"Выберите услугу «Замена паспорта в 20 и 45 лет» и заполните заявление",
			"Оплатите госпошлину на Госуслугах — будет скидка",
			"Придите в отделение МВД или МФЦ в назначенный день с оригиналами документов",
			"Заберите новый паспорт — сделать это можно только лично",
		},
		Links: []Link{
			{Title: "Подать на Госуслугах", URL: "https://www.gosuslugi.ru/newsearch/zamena-pasporta-v-20-i-45-let"},
			{Title: "Сроки замены", URL: "https://www.gosuslugi.ru/help/faq/passport/10016712"},
		},
		AgeWindow: []int{20, 45},
	},
	{
		ID: "moscow-card", Category: "Документы", Region: "moscow",
		Title: "Карта москвича",
		Short: "Бесплатный проезд, скидки в магазинах и аптеках, запись в «Московское долголетие»",
		Who:   "Пенсионерам, которые живут в Москве.",
		Documents: []string{
			"Паспорт",
			"СНИЛС",
			"Фотография (если подаёте в центре госуслуг — сфотографируют на месте)",
		},
		Steps: []string{
			"Зайдите на mos.ru под своей учётной записью",
			"Откройте услугу «Оформление карты москвича» и заполните анкету — большинство данных подставится само",
			"Выберите центр госуслуг «Мои документы», где заберёте карту",
			"Через 30 дней заберите карту лично с паспортом",
		},
		Links: []Link{
			{Title: "Оформить на mos.ru", URL: "https://www.mos.ru/pgu2/landing/karta_moskvicha"},
			{Title: "Что даёт карта пенсионеру", URL: "https://www.mos.ru/karta-moskvicha/tipy-derzhataley/pensionery/"},
		},
		MinAge: 55,
	},
	{
		ID: "pension-80", Category: "Пенсия", Auto: true,
		Title: "Прибавка к пенсии после 80 лет",
		Short: "Фиксированная выплата к пенсии удваивается",
		Who:   "Получателям страховой пенсии по старости с 80 лет. На пенсию по инвалидности прибавка не распространяется.",
		Steps: []string{
			"Ничего подавать не нужно — Социальный фонд пересчитает пенсию сам со дня 80-летия",
			"Проверьте новую сумму в выписке на Госуслугах или позвоните в Социальный фонд",
		},
		Links: []Link{
			{Title: "Подробнее на Госуслугах", URL: "https://www.gosuslugi.ru/help/faq/pension_receiving/3047"},
		},
		Phone:  sfrPhone,
		MinAge: 80,
	},
	{
		ID: "social-supplement", Category: "Пенсия", Auto: true,
		Title: "Доплата до прожиточного минимума",
		Short: "Если вся пенсия меньше прожиточного минимума в регионе",
		Who:   "Неработающим пенсионерам, у которых все выплаты вместе меньше прожиточного минимума пенсионера в регионе.",
		Steps: []string{
			"Обычно доплату назначают автоматически",
			"Если кажется, что доплата положена, но её нет — позвоните в Социальный фонд или обратитесь в клиентскую службу",
		},
		Links: []Link{
			{Title: "Социальный фонд: условия", URL: "https://sfr.gov.ru/grazhdanam/pensioneram/socialnaya_doplata_do_prozhitochnogo_minimuma/"},
		},
		Phone:  sfrPhone,
		MinAge: 55,
	},
	{
		ID: "overhaul", Category: "Льготы",
		Title: "Компенсация за капремонт",
		Short: "С 70 лет — 50%, с 80 лет — 100% взноса",
		Who:   "Неработающим пенсионерам-собственникам жилья от 70 лет, которые живут одни или в семье только из неработающих пенсионеров. Компенсируют взнос в пределах нормы площади.",
		Documents: []string{
			"Паспорт",
			"СНИЛС",
			"Документ о праве собственности на жильё",
			"Квитанции об оплате взноса на капремонт",
			"Реквизиты счёта для перечисления",
		},
		Steps: []string{
			"Проверьте, что оплачиваете взнос на капремонт без долгов",
			"Подайте заявление в соцзащиту, МФЦ или на Госуслугах",
			"Дождитесь решения — компенсация будет приходить на счёт",
		},
		Links: []Link{
			{Title: "Льготы по капремонту на Госуслугах", URL: "https://www.gosuslugi.ru/newsearch/lgoty-po-kapremontu"},
			{Title: "Для москвичей: фонд капремонта", URL: "https://fond.mos.ru/regional-system-overhaul/benefits-to-pay-contributions-for-the-overhaul-.php"},
		},
		MinAge: 70,
	},
	{
		ID: "property-tax", Category: "Льготы",
		Title: "Налоговые льготы пенсионерам",
		Short: "Не платить налог за одну квартиру, дом и гараж и за 6 соток земли",
		Who:   "Всем пенсионерам. Льгота на имущество — по одному объекту каждого вида, на землю — не платите за 6 соток одного участка.",
		Documents: []string{
			"Паспорт",
			"Пенсионное удостоверение или справка о пенсии (обычно не нужна — налоговая получает данные сама)",
		},
		Steps: []string{
			"Обычно льготу дают без заявления: налоговая получает данные из Социального фонда",
			"Проверьте налоговое уведомление — если налог начислен, подайте заявление о льготе",
			"Заявление можно подать в личном кабинете налогоплательщика, в налоговой или МФЦ",
		},
		Links: []Link{
			{Title: "Льготы на Госуслугах", URL: "https://www.gosuslugi.ru/life/details/property_tax_benefits_for_retirees"},
			{Title: "Заявление на сайте ФНС", URL: "https://www.nalog.gov.ru/rn77/taxation/taxes/nnifz/5686398/"},
		},
		MinAge: 55,
	},
	{
		ID: "social-services-set", Category: "Льготы",
		Title:    "Набор социальных услуг",
		Short:    "Лекарства, санаторий и проезд — или деньги вместо них",
		Who:      "Тем, кто получает ежемесячную денежную выплату (ЕДВ): людям с инвалидностью, ветеранам и другим федеральным льготникам.",
		Deadline: "Заменить набор деньгами или вернуть его можно до 1 октября — изменения начнут действовать с 1 января",
		Documents: []string{
			"Паспорт",
			"СНИЛС",
		},
		Steps: []string{
			"Решите, что выгоднее: получать лекарства, санаторий и проезд или деньги",
			"До 1 октября подайте заявление в Социальный фонд, МФЦ или на Госуслугах",
			"С 1 января выплата изменится",
		},
		Links: []Link{
			{Title: "Как заменить деньгами — Госуслуги", URL: "https://www.gosuslugi.ru/help/faq/social_services/1200012"},
			{Title: "Социальный фонд: набор услуг", URL: "https://sfr.gov.ru/grazhdanam/socialnaya_podderzhka/federal_beneficiaries/nsu"},
		},
		Phone: sfrPhone,
	},
	{
		ID: "housing-subsidy", Category: "Льготы",
		Title: "Субсидия на оплату ЖКУ",
		Short: "Если коммунальные платежи больше установленной доли дохода",
		Who:   "Тем, у кого расходы на ЖКУ больше установленной доли дохода семьи (по России — 22%, в регионах может быть меньше).",
		Documents: []string{
			"Паспорт",
			"Сведения о доходах всех членов семьи за 6 месяцев",
			"Квитанции за ЖКУ без долгов",
			"Документ о праве пользования жильём",
		},
		Steps: []string{
			"Посчитайте доход семьи за последние 6 месяцев",
			"Подайте заявление на Госуслугах, в МФЦ или соцзащите",
			"Субсидию назначают на 6 месяцев — потом нужно подать заявление снова",
		},
		Links: []Link{
			{Title: "Подать на Госуслугах", URL: "https://www.gosuslugi.ru/newsearch/subsidiya-zhkkh"},
		},
	},
	{
		ID: "labour-veteran", Category: "Льготы",
		Title: "Звание «Ветеран труда»",
		Short: "Даёт региональные льготы: выплаты, проезд, скидки на ЖКУ",
		Who:   "Тем, у кого есть ведомственная награда за труд и стаж не меньше 40 лет у мужчин и 35 лет у женщин. Условия в регионах отличаются.",
		Documents: []string{
			"Паспорт",
			"Документ о награде",
			"Трудовая книжка или справка о стаже",
			"Фотография 3×4",
		},
		Steps: []string{
			"Проверьте условия своего региона",
			"Подайте заявление в соцзащиту, МФЦ или на Госуслугах",
			"После решения получите удостоверение и оформите льготы",
		},
		Links: []Link{
			{Title: "Как получить — Госуслуги", URL: "https://www.gosuslugi.ru/newsearch/kak-poluchit-zvanie-veterana-truda"},
		},
		MinAge: 55,
	},
	{
		ID: "moscow-longevity", Category: "Досуг", Region: "moscow",
		Title: "Московское долголетие",
		Short: "Бесплатные занятия: спорт, языки, рисование, компьютер и ещё 40 направлений",
		Who:   "Пенсионерам и предпенсионерам, у которых есть карта москвича.",
		Documents: []string{
			"Паспорт",
			"Карта москвича",
		},
		Steps: []string{
			"Выберите направление на mos.ru или в центре московского долголетия",
			"Запишитесь онлайн, в центре «Мои документы» или в регистратуре поликлиники",
			"Приходите на первое занятие с картой москвича",
		},
		Links: []Link{
			{Title: "Записаться на mos.ru", URL: "https://www.mos.ru/city/projects/dolgoletie/"},
		},
		MinAge: 55,
	},
}

func findBenefit(id string) (Benefit, bool) {
	for _, b := range benefits {
		if b.ID == id {
			return b, true
		}
	}
	return Benefit{}, false
}

func regionOf(loc Location) string {
	return bookingFor(loc).Region
}

// подходит ли льгота человеку по возрасту и региону. Если дата рождения не указана — не знаем
func benefitFit(b Benefit, birth *time.Time, region string, today time.Time) (bool, string) {
	if b.Region != "" && region != b.Region {
		return false, ""
	}
	if birth == nil {
		return false, ""
	}
	age := ageOn(*birth, today)

	if len(b.AgeWindow) > 0 {
		for _, a := range b.AgeWindow {
			if age == a || age == a-1 {
				return true, "Вам скоро или уже исполнилось " + yearsWord(a)
			}
		}
		return false, ""
	}
	if b.MinAge > 0 && age >= b.MinAge {
		return true, "Подходит по возрасту"
	}
	if b.MinAge > 0 && age >= b.MinAge-1 {
		return true, "Скоро будет положено по возрасту"
	}
	return false, ""
}

func yearsWord(n int) string {
	switch {
	case n%100 >= 11 && n%100 <= 14:
		return strconv.Itoa(n) + " лет"
	case n%10 == 1:
		return strconv.Itoa(n) + " год"
	case n%10 >= 2 && n%10 <= 4:
		return strconv.Itoa(n) + " года"
	}
	return strconv.Itoa(n) + " лет"
}
