## Backend (Go + Gin) — HTTP API и Telegram‑сервис

- `app/` — основной HTTP API для пользователей, слотов, записей, ролей, уведомлений и админки.
- `telegram/` — Telegram‑бот и вспомогательный HTTP‑слой, работающий поверх основного API.

Оба сервиса написаны на Go и демонстрируют слоистую архитектуру, собственный graceful shutdown, работу с БД через GORM и интеграцию с внешними API.

---

## Порты и документация

- **HTTP API**: `http://localhost:8090`
- **Swagger UI**: `http://localhost:8090/swagger/index.html#/`
- Telegram‑сервис поднимает собственный HTTP/бот‑сервер (порт и URL настраиваются через env).

---

## Быстрый старт (локально)

### API‑сервис (`backend/app`)

```bash
cd backend/app
go mod download
go run .
```

### Telegram‑сервис (`backend/telegram`)

```bash
cd backend/telegram
go mod download
go run ./cmd/run.go
```

> Для полноценной работы нужен запущенный API‑сервис и настроенные переменные окружения (см. ниже).

---

## Технологический стек

- **Язык и Web‑фреймворк**: Go, Gin
- **ORM и БД**: GORM (PostgreSQL‑совместимая база)
- **Документация API**: Swagger (swaggo)
- **Логирование**: logrus + обертка `internal/logger`
- **Фоновые задачи**: `context`, `time.Ticker` (планировщик напоминаний)
- **Telegram‑интеграция**: Telegram Bot API в отдельном сервисе `telegram/`

---

## Архитектура API‑сервиса (`app/`)

Главная идея — **четко разделить уровни ответственности** (HTTP → usecase → repository → models) и обеспечить контроль завершения работы сервиса.

### Верхний уровень

```text
backend/
  app/
    main.go          # точка входа HTTP API
    docs/            # сгенерированные Swagger‑артефакты
    internal/        # инфраструктура (БД, логгер, scheduler)
    http/            # HTTP‑слой (контроллеры, middleware, router, usecase, repository)
    pkg/             # общий код (models, closer и др.)
```

### Важные каталоги

- `internal/database`

  - `database.go` — инициализация GORM‑подключения, ретраи при старте.
  - Экспортирует объект БД, используемый во всех репозиториях.
- `internal/logger`

  - Обертка над `logrus.Logger` с единым форматом логов для сервиса.
- `internal/scheduler`

  - Планировщик напоминаний о записях за 1 час до начала.
  - Работает через `context.Context` и `time.Ticker`, отсылает уведомления в Telegram и в базу (in‑app нотификации).
- `pkg/closer`

  - Собственный **менеджер graceful shutdown**:
    - Подписывается на `SIGINT`, `SIGTERM`.
    - Поддерживает два типа сущностей: `Graceful` (имеют `Shutdown(ctx)`) и `Closer` (имеют `Close(ctx)`).
    - Управляет порядком завершения (сначала HTTP‑сервер и фоновые задачи, затем БД).
  - Основная точка входа — `Manager.WaitForSignal()` и `Manager.Shutdown(ctx)`.
- `pkg/models`

  - GORM‑модели: `User`, `Slot`, `Record`, `Notification`, `Service`, `AdClicks` и т.п.
  - Соответствуют схемам таблиц БД, используются во всех репозиториях.

### HTTP‑слой (`http/`)

```text
http/
  controller/     # HTTP‑хэндлеры, только разбор/валидация запросов и ответы
  middleware/     # auth, session, rate limiting, sanitize, admin‑guard
  repository/     # слой доступа к БД (DAO)
  usecase/        # бизнес‑логика (сервисы)
  router/         # конфигурация Gin‑роутов и запуск HTTP‑сервера
  sender/         # отправка уведомлений (Telegram, др. каналы)
  utils/          # вспомогательные функции для auth и т.п.
```

- `controller/*`

  - `user/`, `slot/`, `record/`, `role/`, `service/`, `notification/`, `admin/`, `metrics/`.
  - Каждый контроллер:
    - Принимает `gin.Context`.
    - Парсит входные данные (`JSON`, `path`, `query`).
    - Делегирует работу в соответствующий usecase‑сервис.
- `usecase/*`

  - Инкапсулируют **бизнес‑правила**:
    - Создание/отмена слотов.
    - Создание и подтверждение записей.
    - Работа с ролями, правами и админ‑операциями.
    - Отправка уведомлений и планирование напоминаний.
  - Не знают о HTTP — только входные структуры и доменные модели.
- `repository/*`

  - Обертка над GORM, отвечает только за SQL‑уровень:
    - `user.Repository`, `slot.Repository`, `record.Repository`, `notification.Repository` и т.д.
  - Все запросы к БД (включая сложные JOIN/WHERE) инкапсулированы здесь.
- `router/`

  - Файлы вида `user.go`, `slot.go`, `record.go`, `admin.go` и т.п. группируют хэндлеры по префиксам (`/user`, `/slot`, `/record`, `/admin`, `/metrics`).
  - `runner.go` создает и запускает `http.Server`, регистрирует middleware и интегрируется с `pkg/closer` для graceful shutdown.

---

## Telegram‑сервис (`telegram/`)

Telegram‑служба — отдельный бинарь, который:

- общается с пользователем через Telegram Bot API;
- ходит в основной API через HTTP‑клиент `internal/adapter/backendapi`;
- реализует сценарии:
  - регистрация/подтверждение входа;
  - просмотр расписания;
  - управление слотами и записями;
  - отправка уведомлений.

### Структура

```text
telegram/
  cmd/run.go           # точка входа, запуск бота
  internal/
    adapter/
      backendapi/      # HTTP‑клиент к API (user, slot, record и т.д.)
      crypto/          # шифрование/токены для Telegram
    app/
      login/           # логика авторизации через бота
      slots/           # логика работы со слотами
      formatter/       # форматирование расписания в сообщения
    bot/
      middleware.go    # middleware бота (rate limit, логирование)
    config/
      config.go        # загрузка config + env
    domain/
      slot.go, ...     # доменные сущности, используемые в боте
    handlers/
      start/, login/, slot/, record/, timezone/, info/  # реакция на команды
    logger/
      logger.go        # единый логгер для сервиса
    transport/
      bot/             # запуск long polling / webhook бота
      http/            # HTTP‑эндпоинты для связи с другими сервисами
  pkg/
    closer/            # graceful shutdown для бота
    encrypt/           # шифрование токенов
    models/            # транспортные модели (record, user и т.д.)
```

Ключевая идея — **не дублировать бизнес‑логику**, а использовать основной API как единственный источник истины.

---

## Основные группы API‑эндпоинтов

Подробные сигнатуры и примеры — в Swagger, здесь только обзор.

- **Пользователь** `/user`

  - `POST /user/register`
  - `POST /user/login`
  - `POST /user/confirm-login/:telegram_id`
  - `GET /user/check/:telegram_id`
  - `POST /user/logout`
  - `DELETE /user/clear`
- **Слот** `/slot`

  - `POST /slot/master/create`
  - `GET /slot/:master_id` (ожидает `telegram_id`/`master_id` в зависимости от контекста)
  - `DELETE /slot/master/:master_id`
- **Запись** `/record`

  - `POST /record/master/create-book`
  - `GET /record/:client_id`
  - `GET /record/master/:slot_id`
  - `DELETE /record/master/:id`
- **Роли и админка** `/admin`, `/role`

  - управление ролями пользователей;
  - просмотр статистики, слотов, записей, услуг;
  - операции очистки/удаления данных.
- **Метрики и уведомления** `/metrics`, `/notification`

  - клики по рекламе;
  - создание и чтение in‑app уведомлений.

---

## Идентификаторы и согласование

- В БД `slots.master_id` → FK на `users.id`.
- В публичных и Telegram‑сценариях может использоваться:
  - `user.id` (UUID) как master_id;
  - `telegram_id` как внешний идентификатор для поиска пользователя.
- Для фронтенда и Telegram‑бота предусмотрены адаптеры, которые скрывают детали этой логики.

---

## Переменные окружения

Рекомендуется вынести все чувствительные значения в `.env`/`local.env`. Для удобства можно завести `.env.example` на базе рекомендаций из `PROFESSIONAL_IMPROVEMENTS.md`.

Примеры переменных:

- **База данных**

  - `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- **HTTP API**

  - `GIN_MODE` (`release`/`debug`)
  - `PORT` (по умолчанию `8090`)
- **Безопасность**

  - `JWT_SECRET`
  - `ADMIN_PASSWORD`
  - internal‑токены для взаимодействия сервисов
- **CORS и фронтенд**

  - `ALLOWED_ORIGINS` — список доменов фронтенда
- **Telegram**

  - `BOT_TOKEN` — токен бота (обязательно задавать только через env)
  - `BACKEND_BASE_URL` — URL HTTP API
