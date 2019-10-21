FROM golang:1.13.1-alpine

LABEL maintainer="Dawid Biały <dawid_bialy@tvn.pl>"

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

WORKDIR /app

COPY . . 

RUN go mod download
