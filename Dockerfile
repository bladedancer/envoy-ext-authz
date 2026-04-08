# golang:version is the latest debian builder image
FROM golang:latest AS builder

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  envuser

WORKDIR /build

COPY . .

RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/extauthzdemo ./main.go

# final container stage
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /build/bin/extauthzdemo .

USER envuser:envuser

EXPOSE 10001

CMD ["./extauthzdemo", "--port", "10001"]
