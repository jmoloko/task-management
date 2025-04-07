# Task Management API

RESTful API-сервис для управления задачами, разработанный с использованием Go 1.24 и принципов чистой архитектуры. Сервис предоставляет полный набор инструментов для управления задачами, включая аутентификацию, CRUD операции, аналитику и мониторинг.

## 🚀 Особенности

- **REST API** с использованием стандартной библиотеки Go 1.24
- **Аутентификация** на основе JWT
- **Управление задачами** (CRUD операции)
- **Фильтрация и поиск** задач
- **Импорт и экспорт** задач в JSON формате
- **Аналитика** по задачам с кэшированием
- **Swagger** документация API
- **Метрики** Prometheus для мониторинга
- **Логирование** с поддержкой структурированных логов
- **Фоновые задачи** для очистки и аналитики
- **Docker** поддержка для легкого развертывания

## 📋 Требования

- Go 1.24
- PostgreSQL 16
- Redis 7
- Docker и Docker Compose (опционально)

## 🛠 Установка и запуск

### Локальная установка

1. Клонируйте репозиторий:
   ```bash
   git clone github.com/jmoloko/task-management
   cd task-management
   ```

2. Создайте файл .env на основе примера:
   ```bash
   cp .env.example .env
   ```

3. Настройте переменные окружения в .env:
   ```env
   DB_HOST=localhost
   DB_PORT=5433
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=taskmanager
   DB_SSLMODE=disable

   REDIS_HOST=localhost
   REDIS_PORT=6380
   REDIS_DB=0

   JWT_SECRET=your-secret-key
   SERVER_PORT=8080
   LOG_LEVEL=info
   ```

4. Создайте базу данных:
   ```bash
   createdb taskmanager
   ```

5. Примените миграции:
   ```bash
   psql -d taskmanager -f migrations/001_init.sql
   ```

6. Соберите и запустите приложение:
   ```bash
   go build -o taskmanager ./cmd/app
   ./taskmanager
   ```

### Docker Compose

1. Соберите и запустите все сервисы:
   ```bash
   docker-compose up -d
   ```

2. Проверьте статус контейнеров:
   ```bash
   docker-compose ps
   ```

## 🌐 Доступные сервисы

После запуска доступны следующие сервисы:

- **API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000
- **PostgreSQL**: localhost:5433
- **Redis**: localhost:6380

## 📊 Мониторинг

### Prometheus

Сервис экспортирует следующие метрики:

- `taskmanager_http_requests_total` - количество HTTP запросов
- `taskmanager_http_request_duration_seconds` - длительность HTTP запросов
- `taskmanager_tasks_created_total` - количество созданных задач
- `taskmanager_tasks_completed_total` - количество завершенных задач

### Grafana

Для визуализации метрик:
1. Откройте http://localhost:3000
2. Логин: admin, пароль: admin
3. Добавьте Prometheus как источник данных
4. Создайте дашборды для мониторинга

## 🔐 API Endpoints

### Аутентификация

#### Регистрация
```http
POST /api/auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

#### Логин
```http
POST /api/auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

### Задачи

#### Создание задачи
```http
POST /api/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
    "title": "Task Title",
    "description": "Task Description",
    "status": "pending",
    "priority": "high",
    "due_date": "2024-04-10T15:04:05Z"
}
```

#### Получение списка задач
```http
GET /api/tasks
Authorization: Bearer <token>
```

#### Получение задачи по ID
```http
GET /api/tasks/{id}
Authorization: Bearer <token>
```

#### Обновление задачи
```http
PUT /api/tasks/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "title": "Updated Title",
    "status": "in_progress"
}
```

#### Удаление задачи
```http
DELETE /api/tasks/{id}
Authorization: Bearer <token>
```

### Импорт/Экспорт

#### Экспорт задач
```http
GET /api/tasks/export
Authorization: Bearer <token>
```

#### Импорт задач
```http
POST /api/tasks/import
Authorization: Bearer <token>
Content-Type: application/json

[
    {
        "title": "Task 1",
        "description": "Description 1",
        "status": "pending",
        "priority": "high",
        "due_date": "2024-04-10T15:04:05Z"
    }
]
```

### Аналитика

#### Получение аналитики
```http
GET /api/tasks/analytics?period=week
Authorization: Bearer <token>
```

## 🏗 Архитектура

Проект следует принципам чистой архитектуры:

### Слои приложения

- **Домен** (domain)
  - Бизнес-модели
  - Интерфейсы репозиториев
  - Бизнес-правила

- **Репозитории** (repository)
  - PostgreSQL реализация
  - Redis кэширование
  - Абстракция доступа к данным

- **Сервисы** (service)
  - Бизнес-логика
  - Валидация
  - Транзакции

- **Обработчики** (handler)
  - HTTP маршрутизация
  - Валидация запросов
  - Форматирование ответов

- **Middleware**
  - Аутентификация
  - Логирование
  - Метрики
  - Обработка ошибок

### Принципы

- **Разделение на слои**
- **Инверсия зависимостей**
- **Принцип разделения интерфейсов**
- **Единая ответственность**
- **Открытость/закрытость**

## 📁 Структура проекта

```
.
├── cmd/                 # Точки входа
│   └── app/             # Основное приложение
├── internal/            # Внутренний код
│   ├── cache/           # Кэширование (Redis)
│   ├── config/          # Конфигурация
│   ├── domain/          # Бизнес-модели
│   ├── handler/         # HTTP обработчики
│   ├── logger/          # Логирование
│   ├── metrics/         # Prometheus метрики
│   ├── middleware/      # HTTP middleware
│   ├── repository/      # Репозитории
│   ├── server/          # HTTP сервер
│   ├── service/         # Бизнес-логика
│   └── worker/          # Фоновые задачи
├── migrations/          # SQL миграции
├── docs/               # Документация
├── tests/              # Тесты
├── docker-compose.yml  # Docker конфигурация
├── Dockerfile          # Docker сборка
└── README.md           # Документация
```

## 🧪 Тестирование

### Unit тесты
```bash
go test ./internal/...
```

### Интеграционные тесты
```bash
go test ./tests/...
```

### Запуск всех тестов
```bash
go test ./...
```

## 📚 Swagger документация

Swagger документация доступна по адресу `/swagger/index.html` после запуска приложения.

Документация генерируется из аннотаций в коде и доступна в форматах:
- YAML: `docs/swagger.yaml`
- JSON: `docs/swagger.json`

## 🔄 Фоновые задачи

Сервис включает следующие фоновые задачи:

1. **Очистка устаревших задач**
   - Запускается каждые 24 часа
   - Удаляет задачи старше 7 дней

2. **Генерация аналитики**
   - Запускается каждые 6 часов
   - Кэширует результаты в Redis

## 📈 Метрики и мониторинг

### HTTP метрики
- Количество запросов
- Длительность запросов
- Коды ответов

### Бизнес метрики
- Количество созданных задач
- Количество завершенных задач
- Статистика по приоритетам
- Статистика по статусам

## 🔒 Безопасность

- **JWT аутентификация** с 15-минутным сроком действия токена
- **Bcrypt** для безопасного хеширования паролей
- Валидация входных данных
- Защита от SQL инъекций
- Rate limiting
- Безопасные заголовки HTTP
- Защита от CSRF атак
- Логирование попыток входа
- Ограничение длины пароля (минимум 6 символов)
- Валидация формата email
- Защита от брутфорса (rate limiting)
- Безопасное хранение секретов в переменных окружения

## 📦 Зависимости

### Основные
- Go 1.24
- PostgreSQL 16
- Redis 7

### Мониторинг
- Prometheus
- Grafana
- Node Exporter
- PostgreSQL Exporter
- Redis Exporter
