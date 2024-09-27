FROM golang:1.23-alpine

RUN apk add --no-cache bash

# Install CA certificates
RUN apk update && \
    apk add --no-cache ca-certificates

WORKDIR /app

COPY .env .

COPY env_out.sh .

RUN bash ./env_out.sh

RUN cp /logs/.exported_vars.env /app/

COPY go.mod go.sum .

RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "cmd/main.go"]
