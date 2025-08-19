FROM golang

COPY ./ ./
RUN go build

ENTRYPOINT ["./od-ash-tray"]
