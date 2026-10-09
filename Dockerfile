FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY web/ web/

RUN CGO_ENABLED=0 go build -o /passport ./cmd/server

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /passport /passport

EXPOSE 8080

CMD ["/passport"]
