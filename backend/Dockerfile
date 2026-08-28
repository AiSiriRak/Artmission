# --- Stage 1: Build ---
FROM golang:1.27 AS build

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/app .

# --- Stage 2: Runtime ---
FROM gcr.io/distroless/static-debian12 AS prod

COPY --from=build /go/bin/app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
CMD ["serve"]
