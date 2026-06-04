FROM golang:1.25.5 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o server ./cmd/server/

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
