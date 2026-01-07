# Frontend - React + Vite

Клиентское приложение для управления тайм-слотами.

## Технологии

- React 19
- Vite 7
- React Router 7
- TanStack Query (React Query) 5
- Axios
- MobX
- React Icons
- Swiper

## Быстрый старт

### Установка зависимостей
```bash
npm install
```

### Разработка
```bash
npm run dev
```
Приложение будет доступно на `http://localhost:5173`

### Сборка для production
```bash
npm run build
```

### Предпросмотр production сборки
```bash
npm run preview
```

## Переменные окружения

Скопируйте `.env.example` в `.env` и заполните необходимые значения:

- `VITE_API_URL` - URL бэкенд API
- `VITE_PUBLIC_SITE_URL` - Публичный URL сайта
- `VITE_INTERNAL_TOKEN` - Внутренний токен для API
- `VITE_FRONTEND_SECRET` - Секрет для фронтенда
- `VITE_DEV_MODE` - Режим разработки (true/false)

## Структура проекта

```
src/
  components/     # Переиспользуемые компоненты
  pages/         # Страницы приложения
  utils/         # Утилиты
  config/        # Конфигурация
```

## Линтинг

```bash
npm run lint
```

## Дополнительная информация

См. основной README.md в корне проекта.
