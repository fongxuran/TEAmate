FROM golang:1.26.1-alpine AS builder

WORKDIR /workspace

COPY api/go.mod api/go.sum ./api/
WORKDIR /workspace/api
RUN go mod download

COPY api/ /workspace/api/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /workspace/bin/server ./cmd/serverd

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /workspace/bin/server /app/server

USER app
EXPOSE 8080

CMD ["/app/server"]
