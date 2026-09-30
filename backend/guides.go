package main

import (
	"context"
	"net/http"
)

// Android и iPhone устроены по-разному, поэтому шаги для них отдельные.
// Если шаги одинаковые, заполнено только поле Steps
type Guide struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Short     string   `json:"short"`
	Icon      string   `json:"icon"`
	Important bool     `json:"important"`
	Steps     []string `json:"steps,omitempty"`
	Android   []string `json:"android,omitempty"`
	IPhone    []string `json:"iphone,omitempty"`
	Tip       string   `json:"tip,omitempty"`
}

var guides = []Guide{
	{
		ID: "scam", Title: "Как не попасться мошенникам", Icon: "shield", Important: true,
		Short: "Главные правила, которые защитят ваши деньги",
		Steps: []string{
			"Никому не называйте коды из СМС — ни «банку», ни «полиции», ни «Госуслугам». Настоящие сотрудники их никогда не спрашивают",
			"Если звонят из «банка» или «полиции» и торопят — положите трубку. Перезвоните сами по номеру с обратной стороны карты или с официального сайта",
			"Не переводите деньги на «безопасный счёт» — такого не бывает",
			"Если пишут или звонят, что родственник попал в беду, — сначала сами позвоните этому родственнику",
			"Не устанавливайте приложения и не открывайте ссылки по просьбе незнакомцев",
			"Сомневаетесь — позвоните близкому человеку и посоветуйтесь. Торопиться не нужно",
		},
		Tip: "Звонок в банк — по номеру на обратной стороне карты. В полицию — 102 или 112.",
	},
	{
		ID: "call", Title: "Как позвонить", Icon: "phone",
		Short: "Набрать номер и начать разговор",
		Android: []string{
			"Найдите на экране зелёный значок с трубкой «Телефон» и нажмите на него",
			"Нажмите на значок с точками-кнопками внизу — откроются цифры",
			"Наберите номер",
			"Нажмите зелёную кнопку с трубкой",
			"Чтобы закончить разговор, нажмите красную кнопку",
		},
		IPhone: []string{
			"Нажмите зелёный значок «Телефон»",
			"Внизу выберите «Клавиши»",
			"Наберите номер",
			"Нажмите зелёную кнопку с трубкой",
			"Чтобы закончить разговор, нажмите красную кнопку",
		},
	},
	{
		ID: "answer", Title: "Как ответить на звонок", Icon: "phone",
		Short: "Принять или отклонить входящий вызов",
		Android: []string{
			"Когда телефон звонит, на экране появится зелёная и красная кнопки",
			"Чтобы ответить — проведите пальцем по зелёной кнопке вверх или в сторону. На некоторых телефонах достаточно нажать",
			"Чтобы отклонить — сделайте то же с красной кнопкой",
		},
		IPhone: []string{
			"Если экран заблокирован — проведите пальцем по кнопке «Ответить» вправо",
			"Если телефон разблокирован — нажмите зелёную кнопку",
			"Чтобы отклонить, нажмите красную кнопку или два раза кнопку сбоку",
		},
	},
	{
		ID: "font", Title: "Как сделать буквы крупнее", Icon: "search",
		Short: "Увеличить шрифт на всём телефоне",
		Android: []string{
			"Откройте «Настройки» — значок с шестерёнкой",
			"Найдите пункт «Экран» или «Дисплей»",
			"Выберите «Размер шрифта» или «Размер и стиль шрифта»",
			"Передвиньте ползунок вправо, пока буквы не станут удобными",
		},
		IPhone: []string{
			"Откройте «Настройки»",
			"Выберите «Экран и яркость»",
			"Нажмите «Размер текста» и передвиньте ползунок вправо",
			"Ещё крупнее можно сделать в «Универсальный доступ» → «Дисплей и размер текста» → «Увеличенный текст»",
		},
	},
	{
		ID: "flashlight", Title: "Как включить фонарик", Icon: "sun",
		Short: "Посветить в темноте",
		Android: []string{
			"Проведите пальцем от верхнего края экрана вниз",
			"Найдите значок «Фонарик» и нажмите на него",
			"Чтобы выключить, нажмите на значок ещё раз",
		},
		IPhone: []string{
			"Проведите пальцем вниз от правого верхнего угла экрана. На iPhone с кнопкой «Домой» — вверх от нижнего края",
			"Нажмите на значок фонарика",
			"Чтобы выключить, нажмите на него ещё раз",
		},
	},
	{
		ID: "wifi", Title: "Как подключиться к Wi-Fi", Icon: "wifi",
		Short: "Интернет дома без мобильного трафика",
		Android: []string{
			"Откройте «Настройки»",
			"Выберите «Wi-Fi» или «Подключения» → «Wi-Fi»",
			"Включите Wi-Fi, если он выключен",
			"Нажмите на название своей сети — оно обычно написано на наклейке роутера",
			"Введите пароль с той же наклейки и нажмите «Подключить»",
		},
		IPhone: []string{
			"Откройте «Настройки» и выберите «Wi-Fi»",
			"Включите Wi-Fi",
			"Нажмите на название своей сети",
			"Введите пароль с наклейки на роутере и нажмите «Подкл.»",
		},
	},
	{
		ID: "max-photo", Title: "Как отправить фото в MAX", Icon: "image",
		Short: "Поделиться снимком с близкими",
		Steps: []string{
			"Откройте MAX и нажмите на чат с нужным человеком",
			"Нажмите на значок скрепки или «+» рядом с полем для текста",
			"Выберите «Фото» или «Галерея»",
			"Нажмите на нужную фотографию",
			"Нажмите кнопку отправки со стрелкой",
		},
	},
	{
		ID: "max-video", Title: "Как позвонить по видео в MAX", Icon: "video",
		Short: "Видеть собеседника во время разговора",
		Steps: []string{
			"Откройте MAX и нажмите на чат с нужным человеком",
			"Вверху экрана нажмите на значок камеры",
			"Дождитесь, пока собеседник ответит",
			"Чтобы закончить, нажмите красную кнопку",
		},
		Tip: "Звонки в MAX бесплатные, нужен только интернет.",
	},
	{
		ID: "gosuslugi", Title: "Как поставить Госуслуги", Icon: "document",
		Short: "Приложение для заявлений, записи к врачу и справок",
		Android: []string{
			"Откройте магазин приложений RuStore или Google Play",
			"В поиске напишите «Госуслуги»",
			"Проверьте, что разработчик — Минцифры России",
			"Нажмите «Установить», потом «Открыть»",
			"Войдите по номеру телефона или СНИЛС и паролю",
		},
		IPhone: []string{
			"Откройте Safari и зайдите на сайт gosuslugi.ru",
			"Нажмите на кнопку «Поделиться» — квадрат со стрелкой",
			"Выберите «На экран «Домой»»",
			"На экране появится значок Госуслуг — открывайте сайт через него",
		},
		Tip: "Пароль от Госуслуг никому не называйте.",
	},
	{
		ID: "this-app", Title: "Как пользоваться помощником", Icon: "help",
		Short: "Что умеет это приложение",
		Steps: []string{
			"Напишите в поиске на главной, что нужно: например, «парацетамол» или «хлеб»",
			"В «Задачах на сегодня» отмечайте кругом то, что сделали. Кнопка «+» добавляет новое дело",
			"В «Возможностях» — аптеки и магазины рядом, запись к врачу и соцпомощь",
			"В профиле (кружок справа от поиска) укажите адрес — по нему ищем места рядом",
			"В «Документах» — льготы и что нужно для их оформления",
		},
	},
}

func findGuide(id string) (Guide, bool) {
	for _, g := range guides {
		if g.ID == id {
			return g, true
		}
	}
	return Guide{}, false
}

func (s *Store) ReadGuides(ctx context.Context, userID string) (map[string]bool, error) {
	rows, err := s.db.Query(ctx, `SELECT guide_id FROM user_guides WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	read := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		read[id] = true
	}
	return read, rows.Err()
}

func (s *Store) SetGuideRead(ctx context.Context, userID, guideID string, read bool) error {
	if !read {
		_, err := s.db.Exec(ctx, `DELETE FROM user_guides WHERE user_id = $1 AND guide_id = $2`, userID, guideID)
		return err
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO user_guides (user_id, guide_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, guideID)
	return err
}

func (s *Server) listGuides(w http.ResponseWriter, r *http.Request) {
	read, err := s.store.ReadGuides(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}

	type item struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Short     string `json:"short"`
		Icon      string `json:"icon"`
		Important bool   `json:"important"`
		Read      bool   `json:"read"`
	}
	items := []item{}
	for _, g := range guides {
		items = append(items, item{ID: g.ID, Title: g.Title, Short: g.Short, Icon: g.Icon, Important: g.Important, Read: read[g.ID]})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) guideDetails(w http.ResponseWriter, r *http.Request) {
	g, ok := findGuide(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Такой инструкции нет")
		return
	}
	read, err := s.store.ReadGuides(r.Context(), userID(r))
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"guide": g, "read": read[g.ID]})
}

func (s *Server) markGuide(w http.ResponseWriter, r *http.Request) {
	g, ok := findGuide(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Такой инструкции нет")
		return
	}
	var body struct {
		Read *bool `json:"read"`
	}
	if err := decodeBody(w, r, &body); err != nil || body.Read == nil {
		writeError(w, http.StatusBadRequest, "Ожидается {\"read\": true или false}")
		return
	}
	if err := s.store.SetGuideRead(r.Context(), userID(r), g.ID, *body.Read); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"read": *body.Read})
}
