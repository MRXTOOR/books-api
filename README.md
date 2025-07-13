# Books API

REST API для магазина книг на Go с поддержкой PostgreSQL, Kafka, Docker, автотестами и миграциями.

## Возможности

- CRUD для книг (`/api/v1/books`)
- CRUD для подборок (`/api/v1/collections`)
- PostgreSQL (без ORM, только SQL и миграции)
- Kafka (event producer)
- Docker и docker-compose для локального и интеграционного запуска
- Интеграционные и unit-тесты

## Структура проекта

```
books-api/
├── cmd/                # main.go (точка входа)
├── internal/
│   ├── books/          # обработчики и логика книг
│   ├── collections/    # обработчики и логика подборок
│   ├── db/             # работа с БД, транзакции
│   ├── kafka/          # интеграция с Kafka
│   ├── integration_test/ # интеграционные тесты
│   └── ...
├── migrations/         # SQL-миграции
├── docker-compose.yml  # запуск сервисов
├── Dockerfile          # билд приложения
└── README.md           # этот файл
```

## Быстрый старт

### 1. Запуск через Docker Compose

```sh
docker-compose up --build
```

- Приложение будет доступно на `http://localhost:8080`
- Postgres: `localhost:5432`, пользователь/пароль: `books`
- Kafka: `localhost:9092`

### 2. Локальный запуск

1. Запустите Postgres и Kafka (можно через docker-compose)
2. Примените миграции из папки `migrations/` к базе
3. Запустите приложение:
   ```sh
   go run ./cmd/main.go
   ```

### 3. Тесты

- Unit-тесты:
  ```sh
  go test ./internal/...
  ```
- Интеграционные тесты (в Docker):
  ```sh
  docker-compose run --rm test
  ```
- Тесты для базы данных автоматически применяют миграции перед запуском.

## Миграции

- Все миграции — обычные SQL-файлы в папке `migrations/`.
- Для применения вручную используйте любой инструмент (например, [golang-migrate](https://github.com/golang-migrate/migrate)) или выполните SQL-файлы вручную.

## Технологии

- Go 1.24+
- PostgreSQL
- Kafka (segmentio/kafka-go)
- Chi router
- Docker, docker-compose
- Unit и интеграционные тесты
