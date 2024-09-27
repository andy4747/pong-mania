FROM golang:1.23.0-alpine AS builder
RUN go install github.com/a-h/templ/cmd/templ@latest
ARG VERSION="4.17.1"
RUN set -x \
    && apk add --no-cache git \
    && git clone --branch "v${VERSION}" --depth 1 --single-branch https://github.com/golang-migrate/migrate /tmp/go-migrate
WORKDIR /tmp/go-migrate
RUN set -x \
    && CGO_ENABLED=0 go build -tags 'postgres' -ldflags="-s -w" -o ./migrate ./cmd/migrate \
    && ./migrate -version

RUN cp /tmp/go-migrate/migrate /usr/bin/migrate
RUN apk add --no-cache bash
# Install CA certificates
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY . .
# script to build the env vars
RUN bash ./env_out.sh
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate ./views
#build
RUN go build -o main ./cmd/main.go
RUN ls -lh main

# starting a new stage
FROM scratch
WORKDIR /app/
COPY --from=builder /logs/.exported_vars.env .
COPY --from=builder /app/main .
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /usr/bin/migrate /usr/bin/migrate
# Copy CA certificates from the Alpine stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY static ./static
ENV GOENV=production
EXPOSE 8080
CMD ["./main"]
