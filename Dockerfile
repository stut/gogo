FROM golang:1.18.0-alpine3.15 AS build

WORKDIR /app

RUN adduser -D scratchuser

COPY go.* ./
RUN go mod download

COPY src/*.go ./

RUN CGO_ENABLED=0 go build -o /gogo -ldflags="-s -w"

FROM scratch

WORKDIR /www

USER scratchuser

COPY --from=0 /etc/passwd /etc/passwd
COPY --from=build /gogo /gogo

ENTRYPOINT ["/gogo"]

