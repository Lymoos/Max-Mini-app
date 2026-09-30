# Помощник — миниапп для MAX

Фронтенд: React + Vite (`frontend`). Бэкенд: Go + PostgreSQL (`backend`).

## Запуск

```bash
docker compose up -d
```

```bash
cd backend && cp .env.example .env
```

Заполнить в `backend/.env` ключ `YANDEX_API_KEY` и `YANDEX_FOLDER_ID` (без них всё работает, но без YandexGPT).

Один раз загрузить аптеки, магазины, поликлиники и соцпомощь Москвы и Казани из OpenStreetMap (несколько минут, можно указать только нужные виды: `go run . import clinic social`):

```bash
cd backend && go run . import
```

```bash
cd backend && go run .
```

```bash
cd frontend && npm install && npm run dev
```

Перед показом можно обновить акции, чтобы даты были свежими: `go run . promos`.

## Тесты

```bash
cd backend && go test ./...
```

```bash
cd frontend && npm test
```

Тесты бэкенда используют базу `maxapp_test` из docker compose (порт 5440).

## Данные

Координаты, названия и адреса аптек и магазинов — настоящие, из OpenStreetMap.
Телефоны организаций соцпомощи и клиник — из OpenStreetMap, где они указаны.
Рейтинги, цены, акции, скидки в день рождения и врачи в поликлиниках — демонстрационные (в базе `source = 'demo'`),
их нужно заменить на реальные источники перед запуском для пользователей.

Записаться к врачу прямо в приложении нельзя: у ЕМИАС и Госуслуг нет открытого API.
Приложение показывает поликлинику по прописке, врачей и открывает официальный портал записи своего региона.
