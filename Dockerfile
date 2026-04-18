FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY webapp/go.mod webapp/go.sum ./
RUN go mod download

# buffalo and buffalo-pop are deployment tooling only (see comment in runtime stage).
RUN CGO_ENABLED=0 go install github.com/gobuffalo/cli/cmd/buffalo@v0.18.14 && \
    CGO_ENABLED=0 go install github.com/gobuffalo/buffalo-pop/v3@v3.0.7

COPY webapp/ .

RUN CGO_ENABLED=0 go build -o bin/app ./cmd/app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/bin/app ./bin/app
COPY --from=builder /build/database.yml ./database.yml
COPY db/migrations ./db/migrations

# buffalo and buffalo-pop are deployment tooling, not app logic: they exist here
# solely for `buffalo pop migrate` during `mise run deploy`. At single-VPS scale,
# co-packaging avoids a separate migration image at the cost of ~15 MB.
# If this ever moves to multi-instance deployment, extract into a dedicated migrate image.
COPY --from=builder /go/bin/buffalo /usr/local/bin/buffalo
COPY --from=builder /go/bin/buffalo-pop /usr/local/bin/buffalo-pop

EXPOSE 3000

CMD ["./bin/app"]
