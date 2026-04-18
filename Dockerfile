FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies separately — only rebuilds when go.mod/go.sum change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/shortener

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /service /app/service
COPY configs/ /app/configs/
COPY migrations/ /app/migrations/

ENTRYPOINT ["/app/service"]
