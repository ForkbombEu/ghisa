# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY server.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/ghisa server.go

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/ghisa /ghisa
EXPOSE 5552
ENTRYPOINT ["/ghisa"]
