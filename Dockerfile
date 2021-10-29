FROM golang:alpine as build-env
COPY . /src
WORKDIR /src
RUN go build -o gowon-lastfm

FROM alpine:3.14.2
RUN mkdir /data
ENV GOWON_LASTFM_KV_PATH /data/kv.db
WORKDIR /app
COPY --from=build-env /src/gowon-lastfm /app/
ENTRYPOINT ["./gowon-lastfm"]
