FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux
RUN go build -o app -tags netgo -ldflags '-extldflags "-static"' ./cmd/main.go


FROM alpine AS runner

WORKDIR /app

COPY --from=builder /app/app ./
COPY .env .env
COPY ./docs ./docs
COPY ./migrations ./migrations

CMD ["./app"]