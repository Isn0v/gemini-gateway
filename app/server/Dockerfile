FROM golang:1.24

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /gemini-server

EXPOSE 8080

CMD [ "/gemini-server" ]