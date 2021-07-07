FROM golang:1.16.5

WORKDIR /app/alarm-digester

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG env
ENV ENV=$env

RUN go build -o ./out/alarm-digester .

CMD ["./out/alarm-digester"]
