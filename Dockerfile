FROM golang:1.15.5-alpine as base
RUN apk add --update --no-cache ca-certificates git
WORKDIR /app

FROM base as dev
RUN go get github.com/githubnemo/CompileDaemon

FROM base as builder
ADD . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o funssest-slip-telegram main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/funssest-slip-telegram /app/funssest-slip-telegram
ENTRYPOINT ["/app/funssest-slip-telegram"]

