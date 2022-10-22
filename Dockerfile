FROM golang:alpine AS builder

WORKDIR /build/

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" main.go

# -

FROM alpine:3.13 AS certificates

RUN apk --no-cache add ca-certificates

# -

FROM scratch

WORKDIR /api/
ENV PATH=/api/bin/:$PATH

COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /build/main ./bin/main
COPY --from=builder /build/.env/ .
COPY --from=builder /build/data/ ./data/


EXPOSE 8080 8080

CMD [ "main" ]
