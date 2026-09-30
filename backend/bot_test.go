package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type sentMessage struct {
	UserID string
	Msg    maxMessage
}

type fakeSender struct {
	sent    []sentMessage
	answers []string
	fail    bool
}

func (f *fakeSender) Send(ctx context.Context, userID string, msg maxMessage) error {
	if f.fail {
		return errors.New("max недоступен")
	}
	f.sent = append(f.sent, sentMessage{userID, msg})
	return nil
}

func (f *fakeSender) Answer(ctx context.Context, callbackID, notification string) error {
	f.answers = append(f.answers, notification)
	return nil
}

func newTestBot(t *testing.T) (*Bot, *fakeSender, *Store) {
	t.Helper()
	srv, store := newTestServer(t)
	sender := &fakeSender{}
	return &Bot{srv: srv, api: sender, botName: "pomoshnik_bot"}, sender, store
}

func buttons(msg maxMessage) []maxButton {
	list := []maxButton{}
	for _, a := range msg.Attachments {
		rows := a["payload"].(map[string]any)["buttons"].([][]maxButton)
		for _, row := range rows {
			list = append(list, row...)
		}
	}
	return list
}

func findButton(msg maxMessage, kind string) (maxButton, bool) {
	for _, b := range buttons(msg) {
		if b.Type == kind {
			return b, true
		}
	}
	return maxButton{}, false
}

func textUpdate(userID int64, text string) maxUpdate {
	u := maxUpdate{UpdateType: "message_created"}
	json.Unmarshal([]byte(`{"message":{"sender":{"user_id":`+itoa(userID)+`,"first_name":"Анна"},"body":{"text":`+strings.ReplaceAll(`"`+text+`"`, "\n", " ")+`}}}`), &u)
	u.UpdateType = "message_created"
	return u
}

func TestBotWelcome(t *testing.T) {
	bot, sender, store := newTestBot(t)
	u := maxUpdate{UpdateType: "bot_started", User: &maxUser{UserID: 555, FirstName: "Анна"}}
	if err := bot.Handle(context.Background(), u); err != nil {
		t.Fatal(err)
	}

	if len(sender.sent) != 1 || sender.sent[0].UserID != "555" || !strings.Contains(sender.sent[0].Msg.Text, "Здравствуйте, Анна") {
		t.Fatalf("%+v", sender.sent)
	}
	app, ok := findButton(sender.sent[0].Msg, "open_app")
	if !ok || app.WebApp != "pomoshnik_bot" {
		t.Errorf("нужна кнопка открытия миниаппа: %+v", buttons(sender.sent[0].Msg))
	}
	if _, ok := findButton(sender.sent[0].Msg, "request_geo_location"); !ok {
		t.Error("нужна кнопка отправки местоположения")
	}

	var n int
	store.db.QueryRow(context.Background(), `SELECT count(*) FROM bot_users WHERE user_id = '555'`).Scan(&n)
	if n != 1 {
		t.Error("пользователь бота должен сохраниться")
	}

	bot.Handle(context.Background(), textUpdate(555, "/start"))
	if len(sender.sent) != 2 || !strings.Contains(sender.sent[1].Msg.Text, "Здравствуйте") {
		t.Error("/start тоже приветствует")
	}
}

func TestBotMedicineReply(t *testing.T) {
	bot, sender, store := newTestBot(t)
	a := addPlace(t, store, Place{Kind: "pharmacy", Name: "Ригла", Lat: 55.755, Lon: 37.622, Rating: 4.8, Reviews: 100})
	b := addPlace(t, store, Place{Kind: "pharmacy", Name: "36,6", Lat: 55.756, Lon: 37.623, Rating: 4.0, Reviews: 100})
	addMedicinePrice(t, store, a.ID, "paracetamol", 50)
	addMedicinePrice(t, store, b.ID, "paracetamol", 40)

	if err := bot.Handle(context.Background(), textUpdate(555, "где купить парацетомол")); err != nil {
		t.Fatal(err)
	}
	msg := sender.sent[0].Msg
	if !strings.Contains(msg.Text, "Парацетамол в аптеках рядом:") || !strings.Contains(msg.Text, "1. Ригла — 50 ₽") ||
		!strings.Contains(msg.Text, "36,6 — 40 ₽") || !strings.Contains(msg.Text, "дешевле всего") {
		t.Errorf("%s", msg.Text)
	}
	app, _ := findButton(msg, "open_app")
	if app.Payload != "med_paracetamol" || app.Text != "Все аптеки в приложении" {
		t.Errorf("кнопка должна открыть миниапп сразу на лекарстве: %+v", app)
	}
	if !strings.Contains(msg.Text, "Пришлите своё местоположение") {
		t.Error("без адреса предлагаем прислать геопозицию")
	}
}

func TestBotLocationThenNearby(t *testing.T) {
	bot, sender, store := newTestBot(t)
	addPlace(t, store, Place{Kind: "shop", Name: "Магнит", Lat: 55.791, Lon: 49.121, Rating: 4.5, Reviews: 10})

	loc := maxUpdate{UpdateType: "message_created"}
	json.Unmarshal([]byte(`{"message":{"sender":{"user_id":555},"body":{"text":"","attachments":[{"type":"location","latitude":55.79,"longitude":49.12}]}}}`), &loc)
	loc.UpdateType = "message_created"
	if err := bot.Handle(context.Background(), loc); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sender.sent[0].Msg.Text, "Запомнил адрес: Тверская улица, 7, Москва") {
		t.Errorf("%s", sender.sent[0].Msg.Text)
	}
	p, _ := store.GetProfile(context.Background(), "555")
	if !p.HasLocation || p.Lat != 55.79 {
		t.Errorf("адрес из бота сохраняется в тот же профиль, что и в миниаппе: %+v", p)
	}

	bot.Handle(context.Background(), textUpdate(555, "хочу в магазин"))
	msg := sender.sent[1].Msg
	if !strings.Contains(msg.Text, "1. Магнит") || strings.Contains(msg.Text, "центром Москвы") {
		t.Errorf("магазины рядом с присланной точкой: %s", msg.Text)
	}
	if app, _ := findButton(msg, "open_app"); app.Payload != "goods" {
		t.Errorf("%+v", app)
	}
}

func TestBotOtherReplies(t *testing.T) {
	bot, sender, store := newTestBot(t)
	addPlace(t, store, Place{Kind: "social", Name: "ЦСО Мещанский", Phone: "+7 495 123-45-67", Lat: 55.755, Lon: 37.622, Rating: 4.5, Reviews: 10})

	cases := []struct {
		text, want, payload string
	}{
		{"нужна соцзащита", "тел. +7 495 123-45-67", "social"},
		{"записаться к врачу", "поликлинику по прописке", "doctor"},
		{"где мой паспорт", "документы и льготы", "documents"},
		{"расскажи анекдот", "Пока не понял", ""},
		{"молоко", "рядом не нашёл", "prod_milk"},
	}
	for i, c := range cases {
		bot.Handle(context.Background(), textUpdate(555, c.text))
		msg := sender.sent[i].Msg
		if !strings.Contains(msg.Text, c.want) {
			t.Errorf("%q: %s", c.text, msg.Text)
		}
		if app, _ := findButton(msg, "open_app"); app.Payload != c.payload {
			t.Errorf("%q: кнопка %+v", c.text, app)
		}
	}

	bot.srv.ai = &fakeAI{answer: `{"type": "doctor", "target": "neurologist"}`}
	bot.Handle(context.Background(), textUpdate(555, "голова кружится по утрам"))
	last := sender.sent[len(sender.sent)-1].Msg
	if app, _ := findButton(last, "open_app"); !strings.Contains(last.Text, "неврологу") || app.Payload != "doctor_neurologist" {
		t.Errorf("непонятное уходит в ИИ: %s %+v", last.Text, app)
	}
}

func TestBotNewActions(t *testing.T) {
	bot, sender, _ := newTestBot(t)
	cases := []struct {
		text, want, button, payload string
	}{
		{"нужен кардиолог", "к кардиологу", "Записаться к кардиологу", "doctor_cardiologist"},
		{"хочу к врачу по давлению", "к кардиологу", "Записаться к кардиологу", "doctor_cardiologist"},
		{"глазник", "к офтальмологу", "Записаться к офтальмологу", "doctor_ophthalmologist"},
		{"компенсация за капремонт", "Компенсация за капремонт", "Как оформить", "benefit_overhaul"},
		{"как увеличить шрифт", "Как сделать буквы крупнее", "Открыть инструкцию", "guide_font"},
		{"поменять адрес", "профиль", "Открыть профиль", "profile"},
		{"что выпить от давления", "врач", "Открыть лекарства", "medicines"},
		{"купить лекарство", "название лекарства", "Искать лекарство", "medicines"},
	}
	for i, c := range cases {
		bot.Handle(context.Background(), textUpdate(555, c.text))
		msg := sender.sent[i].Msg
		app, _ := findButton(msg, "open_app")
		if !strings.Contains(msg.Text, c.want) || app.Text != c.button || app.Payload != c.payload {
			t.Errorf("%q: %s %+v", c.text, msg.Text, app)
		}
	}
}

func callbackUpdate(userID int64, payload string) maxUpdate {
	cb := maxUpdate{UpdateType: "message_callback"}
	data, _ := json.Marshal(map[string]any{"callback": map[string]any{"callback_id": "c1", "payload": payload, "user": map[string]any{"user_id": userID}}})
	json.Unmarshal(data, &cb)
	cb.UpdateType = "message_callback"
	return cb
}

func TestBotTaskNeedsConfirmation(t *testing.T) {
	bot, sender, store := newTestBot(t)
	bot.srv.now = func() time.Time { return askNow }
	ctx := context.Background()

	bot.Handle(ctx, textUpdate(555, "напомни выпить таблетку в 9"))
	msg := sender.sent[0].Msg
	if !strings.Contains(msg.Text, "«Выпить таблетку» на сегодня в 09:00") {
		t.Errorf("%s", msg.Text)
	}
	tasks, _ := store.TasksByDate(ctx, "555", "2026-09-30")
	if len(tasks) != 0 {
		t.Fatal("без подтверждения задачу не создаём")
	}
	add, ok := findButton(msg, "callback")
	if !ok || add.Text != "Добавить" {
		t.Fatalf("нужна кнопка «Добавить»: %+v", buttons(msg))
	}

	bot.Handle(ctx, callbackUpdate(555, add.Payload))
	tasks, _ = store.TasksByDate(ctx, "555", "2026-09-30")
	if len(tasks) != 1 || tasks[0].Title != "Выпить таблетку" || tasks[0].Time != "09:00" || tasks[0].Kind != "medicine" {
		t.Fatalf("%+v", tasks)
	}
	if sender.answers[0] != "Добавил: Выпить таблетку, сегодня в 09:00" {
		t.Errorf("%q", sender.answers[0])
	}

	bot.Handle(ctx, textUpdate(555, "напомни полить цветы"))
	if last := sender.sent[len(sender.sent)-1].Msg; !strings.Contains(last.Text, "Во сколько напомнить") || len(buttons(last)) != 0 {
		t.Errorf("без времени спрашиваем время: %+v", last)
	}

	for _, payload := range []string{"addtask:2026-09-30 09:00 medicine", "addtask:2026-09-30 9 утра medicine x", "addtask:2026-09-30 09:00 taxi x", "addtask:2020-01-01 09:00 other x"} {
		bot.Handle(ctx, callbackUpdate(555, payload))
	}
	want := []string{"Не понял кнопку", "Не понял кнопку", "Не понял кнопку", "Эта дата уже прошла"}
	got := sender.answers[len(sender.answers)-4:]
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ответ %d: %q, ожидали %q", i, got[i], want[i])
		}
	}
}

func TestBotReminders(t *testing.T) {
	bot, sender, store := newTestBot(t)
	now := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	bot.srv.now = func() time.Time { return now }
	ctx := context.Background()
	store.SaveBotUser(ctx, "555", "Анна")

	pill, _ := store.AddTask(ctx, "555", "2026-09-30", Task{Time: "08:45", Title: "Выпить таблетку", Kind: "medicine"})
	store.AddTask(ctx, "555", "2026-09-30", Task{Time: "07:00", Title: "Давно прошло", Kind: "other"})
	store.AddTask(ctx, "555", "2026-09-30", Task{Time: "10:00", Title: "Ещё рано", Kind: "other"})
	done, _ := store.AddTask(ctx, "555", "2026-09-30", Task{Time: "08:50", Title: "Уже сделано", Kind: "other"})
	store.SetDone(ctx, "555", done.ID, true)
	store.AddTask(ctx, "not-in-bot", "2026-09-30", Task{Time: "08:55", Title: "Не писал боту", Kind: "other"})

	if err := bot.SendReminders(ctx); err != nil {
		t.Fatal(err)
	}
	if len(sender.sent) != 1 || sender.sent[0].Msg.Text != "Пора: Выпить таблетку (08:45)" {
		t.Fatalf("напоминаем только о подходящей задаче: %+v", sender.sent)
	}
	bs := buttons(sender.sent[0].Msg)
	if len(bs) != 2 || bs[0].Payload != "done:"+itoa(int64(pill.ID)) || bs[1].Payload != "snooze:"+itoa(int64(pill.ID)) {
		t.Errorf("%+v", bs)
	}

	bot.SendReminders(ctx)
	if len(sender.sent) != 1 {
		t.Error("повторно не напоминаем")
	}

	cb := maxUpdate{UpdateType: "message_callback"}
	json.Unmarshal([]byte(`{"callback":{"callback_id":"c1","payload":"snooze:`+itoa(int64(pill.ID))+`","user":{"user_id":555}}}`), &cb)
	cb.UpdateType = "message_callback"
	bot.Handle(ctx, cb)
	if sender.answers[0] != "Напомню через 15 минут" {
		t.Errorf("%v", sender.answers)
	}

	now = now.Add(14 * time.Minute)
	bot.SendReminders(ctx)
	if len(sender.sent) != 1 {
		t.Error("через 14 минут ещё рано")
	}
	now = now.Add(2 * time.Minute)
	bot.SendReminders(ctx)
	if len(sender.sent) != 2 {
		t.Fatal("через 16 минут напоминаем снова")
	}

	json.Unmarshal([]byte(`{"callback":{"callback_id":"c2","payload":"done:`+itoa(int64(pill.ID))+`","user":{"user_id":555}}}`), &cb)
	bot.Handle(ctx, cb)
	tasks, _ := store.TasksByDate(ctx, "555", "2026-09-30")
	for _, task := range tasks {
		if task.ID == pill.ID && !task.Done {
			t.Error("«Сделано» отмечает задачу и в миниаппе")
		}
	}

	for _, payload := range []string{"done:99999", "snooze:" + itoa(int64(pill.ID)), "hello", "done:abc"} {
		json.Unmarshal([]byte(`{"callback":{"callback_id":"c3","payload":"`+payload+`","user":{"user_id":555}}}`), &cb)
		bot.Handle(ctx, cb)
	}
	want := []string{"Задача не найдена", "Задача уже выполнена или удалена", "Не понял кнопку", "Не понял кнопку"}
	got := sender.answers[len(sender.answers)-4:]
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ответ %d: %q, ожидали %q", i, got[i], want[i])
		}
	}
}

func TestBotRemindersAfterMidnight(t *testing.T) {
	bot, sender, store := newTestBot(t)
	now := time.Date(2026, 9, 30, 0, 10, 0, 0, time.UTC)
	bot.srv.now = func() time.Time { return now }
	store.SaveBotUser(context.Background(), "555", "")
	store.AddTask(context.Background(), "555", "2026-09-30", Task{Time: "00:05", Title: "Ночная таблетка", Kind: "medicine"})
	store.AddTask(context.Background(), "555", "2026-09-29", Task{Time: "23:55", Title: "Вчерашняя", Kind: "medicine"})

	bot.SendReminders(context.Background())
	if len(sender.sent) != 1 || !strings.Contains(sender.sent[0].Msg.Text, "Ночная таблетка") {
		t.Errorf("%+v", sender.sent)
	}
}

func TestBotSendFailureKeepsReminder(t *testing.T) {
	bot, sender, store := newTestBot(t)
	now := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	bot.srv.now = func() time.Time { return now }
	store.SaveBotUser(context.Background(), "555", "")
	store.AddTask(context.Background(), "555", "2026-09-30", Task{Time: "08:59", Title: "Таблетка", Kind: "medicine"})

	sender.fail = true
	bot.SendReminders(context.Background())
	sender.fail = false
	bot.SendReminders(context.Background())
	if len(sender.sent) != 1 {
		t.Error("если отправка не удалась, напоминание не теряется")
	}
}

func TestFormatKm(t *testing.T) {
	cases := map[float64]string{0: "10 м", 0.34: "340 м", 1: "1,0 км", 2.46: "2,5 км"}
	for km, want := range cases {
		if got := formatKm(km); got != want {
			t.Errorf("formatKm(%v) = %q", km, got)
		}
	}
}

func TestMaxClient(t *testing.T) {
	var gotAuth, gotQuery, gotBody string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.RawQuery
		body := make([]byte, 2000)
		n, _ := r.Body.Read(body)
		gotBody = string(body[:n])
		switch r.URL.Path {
		case "/updates":
			w.Write([]byte(`{"updates":[{"update_type":"bot_started","user":{"user_id":7,"first_name":"Анна"}}],"marker":42}`))
		case "/messages", "/answers":
			w.Write([]byte(`{"success":true}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"bad token"}`))
		}
	}))
	defer fake.Close()

	c, err := NewMaxClient("secret", "")
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = fake.URL

	updates, marker, err := c.Updates(context.Background(), 0)
	if err != nil || marker != 42 || len(updates) != 1 || updates[0].User.UserID != 7 {
		t.Fatalf("%+v %d %v", updates, marker, err)
	}
	if gotAuth != "secret" || !strings.Contains(gotQuery, "types=message_created%2Cmessage_callback%2Cbot_started") || strings.Contains(gotQuery, "marker") {
		t.Errorf("токен в заголовке, без marker в первый раз: %s %s", gotAuth, gotQuery)
	}
	c.Updates(context.Background(), 42)
	if !strings.Contains(gotQuery, "marker=42") {
		t.Errorf("%s", gotQuery)
	}

	msg := maxMessage{Text: "Привет", Attachments: []map[string]any{keyboard([]maxButton{{Type: "open_app", Text: "Открыть", WebApp: "bot", Payload: "med_x"}})}}
	if err := c.Send(context.Background(), "7", msg); err != nil {
		t.Fatal(err)
	}
	if gotQuery != "user_id=7" || !strings.Contains(gotBody, `"type":"inline_keyboard"`) || !strings.Contains(gotBody, `"web_app":"bot"`) || strings.Contains(gotBody, `"url"`) {
		t.Errorf("%s %s", gotQuery, gotBody)
	}

	if err := c.Answer(context.Background(), "cb1", "Готово"); err != nil || gotQuery != "callback_id=cb1" || !strings.Contains(gotBody, `"notification":"Готово"`) {
		t.Errorf("%v %s %s", err, gotQuery, gotBody)
	}

	if err := c.do(context.Background(), "GET", "/me", nil, nil, nil); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("ошибка API должна быть видна: %v", err)
	}
}

func TestMaxClientCAFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := NewMaxClient("t", filepath.Join(dir, "nope.pem")); err == nil {
		t.Error("несуществующий файл сертификата")
	}
	bad := filepath.Join(dir, "bad.pem")
	os.WriteFile(bad, []byte("не сертификат"), 0o600)
	if _, err := NewMaxClient("t", bad); err == nil {
		t.Error("файл без PEM")
	}
}

func TestBotTaskSaveFailureAnswersButton(t *testing.T) {
	bot, sender, _ := newTestBot(t)
	bot.srv.now = func() time.Time { return askNow }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := bot.Handle(ctx, callbackUpdate(555, "addtask:2026-09-30 09:00 medicine Выпить таблетку")); err == nil {
		t.Error("ошибка базы должна вернуться")
	}
	if len(sender.answers) != 1 || sender.answers[0] != "Не удалось добавить задачу, попробуйте ещё раз" {
		t.Errorf("на нажатие нужно ответить, иначе кнопка крутится: %v", sender.answers)
	}
}
