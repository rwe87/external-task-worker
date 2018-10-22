FROM golang:1.11


COPY . /go/src/external-task-worker
WORKDIR /go/src/external-task-worker

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./external-task-worker