FROM golang:1.26.5 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server .

FROM gcr.io/distroless/base

COPY --from=builder /app/server /server

CMD ["/server"]

