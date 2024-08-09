FROM golang:1.22 AS build-env

RUN mkdir -p /build/5stars

WORKDIR /build/5stars
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
        -o /build/5stars/5stars \
        main.go

# -------------
# Image creation stage

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
USER app:app

COPY --from=build-env /build/5stars/5stars /app/
CMD /app/5stars
