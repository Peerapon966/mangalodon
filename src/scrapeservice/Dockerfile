FROM golang:1.25-alpine AS builder

ARG ENV=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mv internal/crawler/config/configs.${ENV}.json.enc internal/crawler/config/configs.json.enc

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o /bin/scraper ./cmd/scraper

FROM public.ecr.aws/lambda/provided:al2023

WORKDIR /var/runtime

COPY --from=builder /bin/scraper ./bootstrap
RUN chmod +x ./bootstrap

CMD ["./bootstrap"]