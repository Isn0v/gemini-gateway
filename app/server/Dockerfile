FROM golang:1.24

ENV RUNNING_PLATFORM=k8s
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /gemini-server 

EXPOSE 8080

CMD [ "/gemini-server" ]