FROM golang:1.24-alpine AS builder

RUN apk add --no-cache make git

COPY . /src
WORKDIR /src

RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000

CMD ["./employee-service", "-conf", "/data/conf/config.yaml"]
