# Subscription Aggregator API

REST API на Go для учета пользовательских онлайн-подписок и подсчета их суммарной стоимости за выбранный период.

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white)
![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?style=flat-square&logo=swagger&logoColor=black)

## О проекте

Сервис реализует CRUDL-операции для записей о подписках пользователей и предоставляет ручку для агрегации расходов. Проект сделан в рамках тестового задания Junior Golang Developer.

Каждая подписка содержит:

- название сервиса;
- стоимость месячной подписки в рублях;
- ID пользователя в формате UUID;
- дату начала подписки в формате `MM-YYYY`;
- опциональную дату окончания подписки.

## Возможности

- Создание, получение, обновление и удаление подписок.
- Получение списка всех подписок.
- Подсчет суммарной стоимости подписок с фильтрацией по пользователю и названию сервиса.
- PostgreSQL как основное хранилище данных.
- SQL-миграции для инициализации схемы.
- Конфигурация через `.env`.
- HTTP-логирование запросов.
- Swagger UI для документации API.
- Запуск окружения через Docker Compose.

## Стек

- Go 1.22
- Gin
- pgx
- PostgreSQL 15
- golang-migrate
- Docker / Docker Compose
- Swagger

## Структура проекта

```text
.
|-- cmd/server              # Точка входа приложения
|-- internal/config         # Загрузка конфигурации
|-- internal/handlers       # HTTP-обработчики
|-- internal/middleware     # Middleware
|-- internal/models         # DTO и доменные модели
|-- internal/repository     # Работа с PostgreSQL
|-- internal/service        # Бизнес-логика
|-- migrations              # SQL-миграции
|-- pkg/database            # Подключение к БД
|-- docker-compose.yml
|-- Dockerfile
`-- README.md
```

## Быстрый старт

Склонируйте репозиторий:

```bash
git clone https://github.com/t0fox/subscription-aggregator-api.git
cd subscription-aggregator-api
```

Создайте `.env` на основе примера:

```bash
cp .env.example .env
```

Запустите сервис и PostgreSQL:

```bash
docker compose up --build
```

API будет доступно по адресу:

```text
http://localhost:8080/api/v1
```

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

## Конфигурация

Переменные окружения:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=subscription_user
DB_PASSWORD=subscription_password
DB_NAME=subscription_db
SERVER_PORT=8080
LOG_LEVEL=info
```

## API

| Метод | Путь | Описание |
| --- | --- | --- |
| `POST` | `/api/v1/subscriptions` | Создать подписку |
| `GET` | `/api/v1/subscriptions` | Получить список подписок |
| `GET` | `/api/v1/subscriptions/{id}` | Получить подписку по ID |
| `PUT` | `/api/v1/subscriptions/{id}` | Обновить подписку |
| `DELETE` | `/api/v1/subscriptions/{id}` | Удалить подписку |
| `POST` | `/api/v1/subscriptions/sum` | Посчитать сумму подписок |

## Примеры запросов

Создание подписки:

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Получение списка:

```bash
curl http://localhost:8080/api/v1/subscriptions
```

Подсчет суммы:

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/sum \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "service_name": "Yandex Plus",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

Ответ:

```json
{
  "total": 400
}
```

## Миграции

SQL-миграции лежат в директории `migrations`.

```text
migrations/
|-- 000001_init_schema.up.sql
`-- 000001_init_schema.down.sql
```

## Модель данных

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Автор

Kirill - [t0fox](https://github.com/t0fox)
