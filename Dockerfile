FROM golang:1.13-rc

WORKDIR /go/src/app/
COPY . .

RUN go get -d -v ./...
RUN go install server.go

EXPOSE 80

CMD ["server"]