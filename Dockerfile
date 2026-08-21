FROM golang AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /od-ash-tray .

FROM scratch AS runtime
COPY --from=build /od-ash-tray /od-ash-tray

CMD ["/od-ash-tray"]
