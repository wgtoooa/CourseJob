FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server


FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/server /app/server
COPY --from=builder /src/migrations /app/migrations

ENV HTTP_ADDR=:8080 \
    DATABASE_URL= \
    DATABASE_USER=postgres \
    DATABASE_PASSWORD=postgres \
    DATABASE_HOST=postgres \
    DATABASE_PORT=5432 \
    DATABASE_NAME=coursejob

EXPOSE 8080

CMD ["/app/server"]
