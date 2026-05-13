# Log Parser Server

Задание написано с использованием чистой архитектуры.
Слои:

- REST adapters
- usecase
- repository
- entity

Данные сохраняются в таблицы:

- `logs`
- `nodes`
- `ports`
- `nodes_info`

Использование sqlx драйвера

Миграции применяются автоматически при старте приложения.

Есть логи на уровне middleware для отслеживания curl

---

## Запуск

Запуск через Docker Compose

Из корня проекта:

docker compose up --build -d

## файлы для парсинга

Файлы для парсинга лежат в:

server/data/

## Примеры curl

- Парсинг файлов:

curl -X POST http://localhost:28082/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{"path":"data"}'

- Информация о логе:

curl http://localhost:28082/api/v1/log/1

- Получить топологию:

curl http://localhost:28082/api/v1/topology/1

- Получить узел: 

curl http://localhost:28082/api/v1/node/2

- Получить порты узла:

curl http://localhost:28082/api/v1/port/2

- Несуществующий лог:

curl -i http://localhost:28082/api/v1/log/999

- Несуществующий узел:

curl -i http://localhost:28082/api/v1/node/1000

- Неправильный путь:

curl -i -X POST http://localhost:28082/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{"path":"data/not_exists"}'



