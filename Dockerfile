FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum  /app/

COPY cmd  /app/cmd

COPY internal  /app/internal

COPY pkg  /app/pkg

RUN go mod download

# RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOAARCH=amd64 go build -o /gokeeper ./cmd/server/main.go

EXPOSE 8080

CMD [ "/gokeeper" ]