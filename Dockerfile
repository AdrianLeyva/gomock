# ---- Build stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/gomock ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /out/gomock ./gomock
COPY data ./data

ENV PORT=8080
ENV DATA_DIR=./data

EXPOSE 8080

ENTRYPOINT ["./gomock"]
