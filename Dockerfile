# build web server
FROM golang:alpine AS go-build

WORKDIR /go/src/

COPY go.mod ./
RUN go mod download

COPY ./src ./
RUN go build -o server

# build static files
# FROM node:alpine AS node-build

# WORKDIR /node/src/

# RUN npm add -g pnpm

# COPY package.json pnpm-lock.yml ./
# RUN pnpm install

# COPY ./static ./
# RUN pnpm run build

# copy server and static files to clean alpine image
FROM alpine:latest

WORKDIR /opt/msmf
COPY --from=go-build /go/src/ ./
COPY ./static ./static
# COPY --from=node-build /node/dist/ /static

EXPOSE 80

CMD ["./server"]