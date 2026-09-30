package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/joho/godotenv"
)

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("не задан DATABASE_URL, см. .env.example")
	}

	if u := os.Getenv("OVERPASS_URL"); u != "" {
		overpassURL = u
	}

	ctx := context.Background()
	pool, err := connectDB(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}
	store := NewStore(pool)
	if err := store.SeedCatalogs(ctx); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "import":
			if err := runImport(ctx, store, os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			log.Print("импорт закончен")
		case "promos":
			if err := refreshPromos(ctx, store); err != nil {
				log.Fatal(err)
			}
			log.Print("акции обновлены")
		default:
			log.Fatalf("неизвестная команда %q, есть: import, promos", os.Args[1])
		}
		return
	}

	counts, err := store.CountPlaces(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, kind := range importKinds {
		if counts[kind] == 0 {
			log.Printf("в базе нет мест вида %s — запустите `go run . import %s`", kind, kind)
		}
	}
	// задачи и напоминания живут по московскому времени, даже если сервер в UTC
	zone := os.Getenv("APP_TIMEZONE")
	if zone == "" {
		zone = "Europe/Moscow"
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		log.Fatalf("неверный APP_TIMEZONE: %v", err)
	}
	now := func() time.Time { return time.Now().In(loc) }

	if err := seedDemoTasks(ctx, store, defaultUserID, now().Format("2006-01-02")); err != nil {
		log.Fatal(err)
	}

	var ai AI
	if os.Getenv("YANDEX_API_KEY") != "" && os.Getenv("YANDEX_FOLDER_ID") != "" {
		ai = NewYandexGPT(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER_ID"))
	} else {
		log.Print("YandexGPT не настроен, работаем без него")
	}

	srv := NewServer(store, NewNominatimGeocoder("https://nominatim.openstreetmap.org"), ai)
	srv.readers = recipeReaders(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER_ID"))
	srv.now = now

	if token := os.Getenv("MAX_BOT_TOKEN"); token != "" {
		api, err := NewMaxClient(token, os.Getenv("MAX_CA_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		srv.botToken = token
		bot := &Bot{srv: srv, api: api, botName: os.Getenv("MAX_BOT_NAME")}
		go bot.Run(ctx, api)
		log.Print("бот MAX запущен")
	} else {
		log.Print("MAX_BOT_TOKEN не задан — бот выключен, вход без проверки MAX")
	}
	handler := srv.routes()
	if dir := os.Getenv("STATIC_DIR"); dir != "" {
		handler = withStatic(handler, dir)
		log.Printf("отдаём фронт из %s", dir)
	}

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      logRequests(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	log.Printf("сервер запущен на :%s", port)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
