FROM golang:1.21-alpine3.17 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN go build -o /server ./cmd/server

FROM gcr.io/distroless/static

COPY --from=builder /server /

CMD ["/server"]