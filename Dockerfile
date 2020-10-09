FROM golang:1.14.1 as builder

ENV PROJECT_DIR="/app"

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# final stage
FROM scratch

ENV PROJECT_DIR=/

COPY --from=builder /app/githubTop /app/
COPY --from=builder /app/config /config
ENTRYPOINT ["/app/githubTop", "http"]