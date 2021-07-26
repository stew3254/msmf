# build web server
FROM golang:latest AS go-build

WORKDIR /go/src/

COPY ./backend/src/go.mod ./
RUN go mod download

COPY ./backend/src ./
RUN go build -o server

# build static files
#FROM node:16 AS node-build
#
#WORKDIR /node/src/
#
#RUN npm add -g pnpm
#
##COPY frontend/package.json frontend/pnpm-lock.yml ./
#COPY ./frontend/package.json ./
#RUN pnpm install
#
#COPY ./frontend/src ./src
#RUN pnpm run build

# copy server and static files to clean alpine image
FROM debian:latest

WORKDIR /srv/website
EXPOSE 5000

RUN apt-get update -y && apt-get upgrade -y && apt-get dist-upgrade -y && apt-get install curl -y
RUN curl -sSL https://get.docker.com/ | sh

COPY --from=go-build /go/src/ ./
#COPY ./static ./static
#COPY --from=node-build /node/dist/ /static


CMD ["./server"]