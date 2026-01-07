FROM golang:1.24.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o telegram ./

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /telegram
COPY --from=builder /src/telegram /telegram/telegram

USER nonroot:nonroot
ENTRYPOINT ["/telegram/telegram"]