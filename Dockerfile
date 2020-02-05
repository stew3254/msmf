# build web server
FROM golang:1.13-rc AS go-build

WORKDIR /go/src/
COPY ./src .

RUN go get -d -v ./...
RUN go install -v ./...

# build static files
FROM node:10-alpine AS node-build

RUN npm add -g pnpm

WORKDIR /node/src/
COPY ./web .

RUN pnpm install && pnpm run build

# copy server and static files to clean alpine image
FROM alpine:latest

WORKDIR /opt/msmf
COPY --from=go-build /go/src/ .
COPY --from=node-build /node/dist/ /static

EXPOSE 80

CMD ["./server"]