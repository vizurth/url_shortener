# URL Shortener

HTTP-сервис для сокращения ссылок. Принимает длинный URL, возвращает короткий код из 10 символов, редиректит по нему.

## Как работает

```
POST /       принимает URL, возвращает короткий код
GET  /{code} редиректит на оригинальный URL (302)
```

Короткий код — 10 символов из `[a-zA-Z0-9_]`, генерируется случайно через `crypto/rand`. Один оригинальный URL всегда получает один и тот же код (повторный запрос вернёт существующий).

## Запуск

**Docker (postgres):**
```bash
make up
```

**Остановить:**
```bash
make down
```

## Makefile

| Команда | Описание |
|---|---|
| `make up` | поднять docker-compose (сборка + запуск) |
| `make down` | остановить и удалить контейнеры и volumes |
| `make test` | unit-тесты с race detector |
| `make test-integration` | интеграционные тесты (нужен postgres) |
| `make lint` | запустить golangci-lint |
| `make mock` | перегенерировать моки через mockery |

## API

**Сократить ссылку:**
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/path"}'
```
```json
{"short_url":"http://localhost:8080/aB3xY9kLmQ","short_code":"aB3xY9kLmQ"}
```
- `201 Created` — новая ссылка
- `200 OK` — ссылка уже существовала
- `400 Bad Request` — невалидный URL

**Перейти по короткой ссылке:**
```bash
curl -L http://localhost:8080/aB3xY9kLmQ
```
- `302 Found` + `Location` header
- `404 Not Found` — код не найден

## Конфигурация

Через переменные окружения:

| Переменная | По умолчанию | Описание |
|---|---|---|
| `STORAGE_TYPE` | `memory` | `postgres` или `memory` |
| `QUERY_HTTP_ADDR` | `0.0.0.0:8080` | адрес сервера |
| `QUERY_SHORT_URL_BASE` | `http://localhost:8080` | база для коротких ссылок |
| `POSTGRES_HOST` | `localhost` | хост БД |
| `POSTGRES_PORT` | `5432` | порт БД |
| `POSTGRES_USERNAME` | `postgres` | пользователь |
| `POSTGRES_PASSWORD` | `postgres` | пароль |
| `POSTGRES_DATABASE` | `postgres` | база данных |

Или через `configs/config.yaml`.