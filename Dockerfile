FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go mod download
# CGO_ENABLED=0 for scratch
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

FROM gcr.io/distroless/base-debian13:nonroot
COPY --from=builder /app/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]