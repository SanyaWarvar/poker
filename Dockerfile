FROM golang:alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN go build -o main ./cmd/app/main.go


RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


FROM alpine:latest

WORKDIR /app


COPY --from=builder /app/main .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate


COPY cmd/db/schema /app/schema
COPY .env .env


RUN mkdir -p /app/user_data/profile_pictures


RUN apk add --no-cache netcat-openbsd


EXPOSE 80


CMD sh -c "while ! nc -z $POSTGRES_HOST $POSTGRES_PORT; do sleep 1; done; \
           while ! nc -z $REDIS_HOST $REDIS_PORT; do sleep 1; done; \
           migrate -path /app/schema -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable up && \
           ./main"