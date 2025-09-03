FROM golang AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /od-ash-tray .

FROM scratch
COPY --from=build /od-ash-tray /od-ash-tray

# bandaid fix
# we'll have to fix this later...
COPY ./fullchain.pem fullchain.pem
COPY ./privkey.pem privkey.pem
COPY ./ash.yaml ./ash.yaml
CMD ["/od-ash-tray"]
