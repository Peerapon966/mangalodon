FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/scheduler ./cmd/scheduler

FROM gcr.io/distroless/base-debian13:nonroot

COPY --from=builder /go/bin/scheduler /

CMD ["/scheduler"]