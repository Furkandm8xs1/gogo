FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server .

FROM alpine:3.22
RUN adduser -D -H appuser
USER appuser
COPY --from=build /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
