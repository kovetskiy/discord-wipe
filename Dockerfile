FROM alpine:edge

RUN apk --update --no-cache add ca-certificates

COPY /discord-wipe /discord-wipe

CMD ["/discord-wipe"]
