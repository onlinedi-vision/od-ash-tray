FROM golang AS build
COPY ./ ./
RUN go build

FROM scratch
COPY --from=build /go/od-ash-tray ./od-ash-tray
ENTRYPOINT ["./od-ash-tray"]
