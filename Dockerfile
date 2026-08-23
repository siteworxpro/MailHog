FROM golang:1.27-alpine AS builder

RUN apk --no-cache add --virtual build-dependencies git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /out/MailHog .

FROM alpine:3

RUN adduser -D -u 1000 mailhog

COPY --from=builder /out/MailHog /usr/local/bin/

USER mailhog

ADD LICENSE.md .

WORKDIR /home/mailhog

ENTRYPOINT ["MailHog"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025
