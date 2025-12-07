# Link Checker

Сервис проверки доступности ссылок с генерацией PDF-отчётов.

## Быстрый старт

### Запуск

```bash
go mod download
go build -o server.exe ./server
.\server.exe
```

Сервер доступен на `http://localhost:8000`

**Примечание:** При первом запуске удалите папку `data/` если она существует, чтобы начать с чистого состояния.
- Timeout для проверки каждой ссылки: 7 секунд (DNS + TCP + request)
- Timeout ответа на запрос `/api/links`: 10 секунд

## API

**POST** `/api/links` — создать задачу проверки ссылок

```json
{"links": ["google.com", "malformedlink.gg"]}
```

Ответ:
```json
{"links": {"google.com": "available", "malformedlink.gg": "not available"}, "links_num": 1}
```

**POST** `/api/report` — получить PDF-отчёт

```json
{"links_list": [1, 2]}
```

Ответ: PDF с информацией по всем ссылкам из задач 1 и 2

## Примеры

```bash
# Создать задачу с тестовыми ссылками
curl -X POST http://localhost:8000/api/links \
  -H "Content-Type: application/json" \
  -d '{"links":["google.com","malformedlink.gg"]}'

# Создать ещё одну задачу
curl -X POST http://localhost:8000/api/links \
  -H "Content-Type: application/json" \
  -d '{"links":["gg.c","yandex.ru"]}'

# Получить PDF-отчёт по обеим задачам
curl -X POST http://localhost:8000/api/report \
  -H "Content-Type: application/json" \
  -d '{"links_list":[1,2]}' -o report.pdf
```
