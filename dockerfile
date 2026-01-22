# Стадия сборки
FROM golang:1.25-alpine AS builder

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata git

WORKDIR /app

# Копируем зависимости для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Получаем версию, коммит и время сборки
ARG VERSION
ARG COMMIT
ARG BUILDTIME

# Если аргументы не переданы — определяем их внутри контейнера
RUN if [ -z "$VERSION" ]; then \
        VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
    fi && \
    if [ -z "$COMMIT" ]; then \
        COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown"); \
    fi && \
    if [ -z "$BUILDTIME" ]; then \
        BUILDTIME=$(date -u +%Y-%m-%dT%H:%M:%SZ); \
    fi && \
    echo "Version: $VERSION" && \
    echo "Commit: $COMMIT" && \
    echo "BuildTime: $BUILDTIME" && \
    CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "-s -w \
            -X 'main.version=$VERSION' \
            -X 'main.commit=$COMMIT' \
            -X 'main.buildTime=$BUILDTIME'" \
        -o /fileserver ./cmd/fileserver

# Финальная стадия
FROM scratch

# Копируем сертификаты и таймзону
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Безопасность: непривилегированный пользователь
USER 65532:65532

EXPOSE 8080
COPY --from=builder /fileserver /fileserver

ENTRYPOINT ["/fileserver"]