package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	botTopLimit     = 3
	remindEvery     = 30 * time.Second
	remindLateLimit = 30 * time.Minute
	snoozeFor       = 15 * time.Minute
)

type BotSender interface {
	Send(ctx context.Context, userID string, msg maxMessage) error
	Answer(ctx context.Context, callbackID, notification string) error
}

type Bot struct {
	srv     *Server
	api     BotSender
	botName string
}

type dueTask struct {
	ID     int
	UserID string
	Time   string
	Title  string
}

func (s *Store) SaveBotUser(ctx context.Context, userID, firstName string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO bot_users (user_id, first_name) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET first_name = EXCLUDED.first_name`,
		userID, firstName,
	)
	return err
}

// задачи, о которых пора напомнить: время наступило (но не больше получаса назад, чтобы после
// перезапуска сервера не прислать пачку старых) или истекла отсрочка «через 15 минут»
func (s *Store) DueReminders(ctx context.Context, date, from, to string, now time.Time) ([]dueTask, error) {
	rows, err := s.db.Query(ctx,
		`SELECT t.id, t.user_id, t.time, t.title FROM tasks t
		 JOIN bot_users b ON b.user_id = t.user_id
		 WHERE t.date = $1 AND NOT t.done AND (
		   (t.reminded_at IS NULL AND t.time >= $2 AND t.time <= $3)
		   OR (t.remind_after IS NOT NULL AND t.remind_after <= $4))
		 ORDER BY t.time, t.id`,
		date, from, to, now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []dueTask{}
	for rows.Next() {
		var t dueTask
		if err := rows.Scan(&t.ID, &t.UserID, &t.Time, &t.Title); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Store) MarkReminded(ctx context.Context, taskID int, now time.Time) error {
	_, err := s.db.Exec(ctx, `UPDATE tasks SET reminded_at = $2, remind_after = NULL WHERE id = $1`, taskID, now)
	return err
}

func (s *Store) Snooze(ctx context.Context, userID string, taskID int, until time.Time) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE tasks SET remind_after = $3 WHERE id = $1 AND user_id = $2 AND NOT done`, taskID, userID, until)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *Bot) openApp(text, payload string) maxButton {
	return maxButton{Type: "open_app", Text: text, WebApp: b.botName, Payload: payload}
}

func geoButton() maxButton {
	return maxButton{Type: "request_geo_location", Text: "Отправить моё местоположение"}
}

func (b *Bot) Run(ctx context.Context, api *MaxClient) {
	go b.remindLoop(ctx)

	var marker int64
	for ctx.Err() == nil {
		updates, next, err := api.Updates(ctx, marker)
		if err != nil {
			log.Printf("бот: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		marker = next
		for _, u := range updates {
			if err := b.Handle(ctx, u); err != nil {
				log.Printf("бот, событие %s: %v", u.UpdateType, err)
			}
		}
	}
}

func (b *Bot) Handle(ctx context.Context, u maxUpdate) error {
	switch u.UpdateType {
	case "bot_started":
		if u.User == nil {
			return nil
		}
		return b.welcome(ctx, u.User)
	case "message_created":
		if u.Message == nil {
			return nil
		}
		return b.onMessage(ctx, u)
	case "message_callback":
		if u.Callback == nil {
			return nil
		}
		return b.onCallback(ctx, u)
	}
	return nil
}

func (b *Bot) welcome(ctx context.Context, user *maxUser) error {
	id := strconv.FormatInt(user.UserID, 10)
	if err := b.srv.store.SaveBotUser(ctx, id, user.FirstName); err != nil {
		return err
	}
	name := ""
	if user.FirstName != "" {
		name = ", " + user.FirstName
	}
	text := "Здравствуйте" + name + "! Я помощник.\n\n" +
		"Напишите название лекарства или продукта — подскажу, где купить рядом и сколько стоит.\n" +
		"Помогу записаться к врачу, найти соцпомощь, льготы и инструкции для телефона.\n" +
		"Напишите «напомни выпить таблетку в 9» — добавлю задачу и напомню вовремя."
	return b.api.Send(ctx, id, maxMessage{Text: text, Attachments: []map[string]any{
		keyboard([]maxButton{b.openApp("Открыть помощника", "")}, []maxButton{geoButton()}),
	}})
}

func (b *Bot) onMessage(ctx context.Context, u maxUpdate) error {
	sender := u.Message.Sender
	id := strconv.FormatInt(sender.UserID, 10)
	if err := b.srv.store.SaveBotUser(ctx, id, sender.FirstName); err != nil {
		return err
	}

	for _, a := range u.Message.Body.Attachments {
		if a.Type == "location" {
			return b.onLocation(ctx, id, a.Latitude, a.Longitude)
		}
	}

	text := strings.TrimSpace(u.Message.Body.Text)
	if text == "" || text == "/start" {
		return b.welcome(ctx, &sender)
	}
	return b.api.Send(ctx, id, b.reply(ctx, id, text))
}

func (b *Bot) onLocation(ctx context.Context, userID string, lat, lon float64) error {
	address := "Точка на карте"
	if place, err := b.srv.geocoder.Reverse(ctx, lat, lon); err == nil && place.Address != "" {
		address = place.Address
	}
	if err := b.srv.store.SaveAddress(ctx, userID, address, &lat, &lon); err != nil {
		return err
	}
	return b.api.Send(ctx, userID, maxMessage{
		Text: "Запомнил адрес: " + address + ".\nТеперь ищу аптеки и магазины рядом с вами. Напишите, что нужно найти.",
	})
}

func (b *Bot) reply(ctx context.Context, userID, text string) maxMessage {
	answer := b.srv.understand(ctx, text)

	p, err := b.srv.store.GetProfile(ctx, userID)
	if err != nil {
		return b.sorry()
	}
	loc := p.Location()

	var msg maxMessage
	switch answer.Type {
	case "medicine":
		offers, err := b.srv.store.MedicineOffers(ctx, answer.Target, loc.Lat, loc.Lon, searchKm)
		if err != nil {
			return b.sorry()
		}
		msg = b.offersMessage(answer.Medicine.Name, "аптеках", sortOffers(offers, botTopLimit), "Все аптеки в приложении", answer.payload())
	case "product":
		offers, err := b.srv.store.ProductOffers(ctx, answer.Target, b.srv.today(), loc.Lat, loc.Lon, searchKm)
		if err != nil {
			return b.sorry()
		}
		msg = b.offersMessage(answer.Product.Name, "магазинах", sortOffers(offers, botTopLimit), "Все магазины в приложении", answer.payload())
	case "feature":
		msg = b.placesMessage(ctx, answer.Target, loc)
	case "task":
		msg = taskMessage(answer)
	case "unknown":
		msg = maxMessage{
			Text:        "Пока не понял. Напишите, например: «парацетамол», «молоко», «нужен кардиолог» или «напомни выпить таблетку в 9».",
			Attachments: []map[string]any{keyboard([]maxButton{b.openApp("Открыть помощника", "")})},
		}
	default:
		msg = maxMessage{Text: answer.Message, Attachments: []map[string]any{keyboard([]maxButton{b.openApp(answer.Button, answer.payload())})}}
	}

	if loc.IsDefault && (answer.Type == "medicine" || answer.Type == "product" || answer.Type == "feature") {
		msg.Text += "\n\nИщу рядом с центром Москвы. Пришлите своё местоположение — найду рядом с вами."
		msg.Attachments = append(msg.Attachments, keyboard([]maxButton{geoButton()}))
	}
	return msg
}

// задачу создаём только после нажатия «Добавить», данные задачи лежат в кнопке
func taskMessage(answer AskAnswer) maxMessage {
	t := answer.Task
	if t.Time == "" {
		return maxMessage{Text: "Во сколько напомнить? Напишите, например: «напомни " + strings.ToLower(t.Title) + " в 9:00»."}
	}
	payload := "addtask:" + t.Date + " " + t.Time + " " + t.Kind + " " + t.Title
	return maxMessage{
		Text:        answer.Message,
		Attachments: []map[string]any{keyboard([]maxButton{{Type: "callback", Text: "Добавить", Payload: payload}})},
	}
}

func (b *Bot) addTaskFromButton(ctx context.Context, userID, callbackID, data string) error {
	parts := strings.SplitN(data, " ", 4)
	if len(parts) != 4 || !taskKinds[parts[2]] || !clockOnlyRe.MatchString(parts[1]) || parts[3] == "" {
		return b.api.Answer(ctx, callbackID, "Не понял кнопку")
	}
	if _, err := time.Parse("2006-01-02", parts[0]); err != nil || parts[0] < b.srv.today() {
		return b.api.Answer(ctx, callbackID, "Эта дата уже прошла")
	}
	if _, err := b.srv.store.AddTask(ctx, userID, parts[0], Task{Time: parts[1], Title: parts[3], Kind: parts[2]}); err != nil {
		// без ответа на нажатие кнопка в MAX так и будет крутиться
		b.api.Answer(ctx, callbackID, "Не удалось добавить задачу, попробуйте ещё раз")
		return err
	}
	return b.api.Answer(ctx, callbackID, "Добавил: "+parts[3]+", "+humanDate(parts[0], b.srv.now())+" в "+parts[1])
}

func (b *Bot) sorry() maxMessage {
	return maxMessage{Text: "Что-то пошло не так. Попробуйте ещё раз чуть позже."}
}

func formatKm(km float64) string {
	if km < 1 {
		return strconv.Itoa(max(10, int(km*100+0.5)*10)) + " м"
	}
	return strings.Replace(strconv.FormatFloat(km, 'f', 1, 64), ".", ",", 1) + " км"
}

func (b *Bot) offersMessage(name, where string, offers []Offer, buttonText, payload string) maxMessage {
	if len(offers) == 0 {
		return maxMessage{
			Text:        name + ": рядом не нашёл. Попробуйте посмотреть в приложении или указать другой адрес.",
			Attachments: []map[string]any{keyboard([]maxButton{b.openApp(buttonText, payload)})},
		}
	}

	lines := []string{name + " в " + where + " рядом:"}
	for i, o := range offers {
		price := strconv.Itoa(o.Price) + " ₽"
		if o.OldPrice > 0 {
			price += " (скидка, было " + strconv.Itoa(o.OldPrice) + " ₽)"
		}
		line := fmt.Sprintf("%d. %s — %s, %s", i+1, o.Name, price, formatKm(o.DistanceKm))
		if o.Cheapest {
			line += " — дешевле всего"
		}
		lines = append(lines, line)
	}
	return maxMessage{
		Text:        strings.Join(lines, "\n"),
		Attachments: []map[string]any{keyboard([]maxButton{b.openApp(buttonText, payload)})},
	}
}

func (b *Bot) placesMessage(ctx context.Context, kind string, loc Location) maxMessage {
	storeKind := map[string]string{"pharmacy": "pharmacy", "goods": "shop", "social": "social"}[kind]
	titles := map[string]string{"pharmacy": "Лучшие аптеки рядом:", "goods": "Лучшие магазины рядом:", "social": "Соцпомощь рядом:"}

	places, err := b.srv.store.PlacesNear(ctx, storeKind, loc.Lat, loc.Lon, searchKm)
	if err != nil {
		return b.sorry()
	}
	ranked := rankPlaces(places, loc.Lat, loc.Lon, searchKm)
	if len(ranked) > botTopLimit {
		ranked = ranked[:botTopLimit]
	}

	lines := []string{titles[kind]}
	for i, p := range ranked {
		line := fmt.Sprintf("%d. %s — %s", i+1, p.Name, formatKm(p.DistanceKm))
		if kind == "social" && p.Phone != "" {
			line += ", тел. " + p.Phone
		}
		lines = append(lines, line)
	}
	if len(ranked) == 0 {
		lines = []string{"Рядом ничего не нашёл."}
	}
	return maxMessage{
		Text:        strings.Join(lines, "\n"),
		Attachments: []map[string]any{keyboard([]maxButton{b.openApp("Открыть в приложении", kind)})},
	}
}

func (b *Bot) onCallback(ctx context.Context, u maxUpdate) error {
	c := u.Callback
	userID := strconv.FormatInt(c.User.UserID, 10)
	action, idText, _ := strings.Cut(c.Payload, ":")
	if action == "addtask" {
		return b.addTaskFromButton(ctx, userID, c.CallbackID, idText)
	}
	taskID, err := strconv.Atoi(idText)
	if err != nil {
		return b.api.Answer(ctx, c.CallbackID, "Не понял кнопку")
	}

	switch action {
	case "done":
		if _, err := b.srv.store.SetDone(ctx, userID, taskID, true); err != nil {
			return b.api.Answer(ctx, c.CallbackID, "Задача не найдена")
		}
		return b.api.Answer(ctx, c.CallbackID, "Отмечено. Молодцы!")
	case "snooze":
		if err := b.srv.store.Snooze(ctx, userID, taskID, b.srv.now().Add(snoozeFor)); err != nil {
			return b.api.Answer(ctx, c.CallbackID, "Задача уже выполнена или удалена")
		}
		return b.api.Answer(ctx, c.CallbackID, "Напомню через 15 минут")
	}
	return b.api.Answer(ctx, c.CallbackID, "Не понял кнопку")
}

func (b *Bot) remindLoop(ctx context.Context) {
	ticker := time.NewTicker(remindEvery)
	defer ticker.Stop()
	for {
		if err := b.SendReminders(ctx); err != nil {
			log.Printf("напоминания: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (b *Bot) SendReminders(ctx context.Context) error {
	now := b.srv.now()
	from := now.Add(-remindLateLimit)
	if from.Day() != now.Day() {
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	due, err := b.srv.store.DueReminders(ctx, now.Format("2006-01-02"), from.Format("15:04"), now.Format("15:04"), now)
	if err != nil {
		return err
	}
	for _, t := range due {
		msg := maxMessage{
			Text: "Пора: " + t.Title + " (" + t.Time + ")",
			Attachments: []map[string]any{keyboard(
				[]maxButton{{Type: "callback", Text: "Сделано", Payload: "done:" + strconv.Itoa(t.ID)}},
				[]maxButton{{Type: "callback", Text: "Напомнить через 15 минут", Payload: "snooze:" + strconv.Itoa(t.ID)}},
			)},
		}
		if err := b.api.Send(ctx, t.UserID, msg); err != nil {
			log.Printf("напоминание %d: %v", t.ID, err)
			continue
		}
		if err := b.srv.store.MarkReminded(ctx, t.ID, now); err != nil {
			return err
		}
	}
	return nil
}
