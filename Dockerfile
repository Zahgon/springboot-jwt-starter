FROM golang:1.27-alpine AS build-container

# WORKDIR creates the directory; the alpine base has no /usr/src to mkdir into.
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/src/app/springboot-jwt-starter .

FROM alpine:3.20
RUN adduser -Dh /home/bfwg bfwg
WORKDIR /app
COPY --from=build-container /usr/src/app/springboot-jwt-starter .
USER bfwg
ENTRYPOINT ["/app/springboot-jwt-starter"]
