# Этап 1: Сборка приложения
FROM golang:1.25-alpine as builder

WORKDIR /app

# Копируем файлы зависимостей и загружаем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Убедимся, что .env файл существует. Если нет, сборка упадет.
# Это предотвращает сборку некорректно сконфигурированного образа.
RUN if [ ! -f .env ]; then echo "Error: .env file not found. Please create it from .env.example before building."; exit 1; fi

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth ./cmd/api/main.go

# Этап 2: Создание минимального образа для запуска
FROM alpine:latest

WORKDIR /app

# Копируем только скомпилированный бинарник из этапа сборки
COPY --from=builder /auth /app/auth

# Копируем .env файл. Приложение подхватит его при старте.
COPY --from=builder /app/.env .

# Копируем папку с ключами
COPY --from=builder /app/keys ./keys

# Открываем порт, который приложение будет слушать внутри контейнера.
# Это значение должно совпадать с переменной PORT в вашем .env файле.
EXPOSE 8081

# Команда для запуска приложения
CMD ["/app/auth"]
