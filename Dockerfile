# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /ocr-translate .

# ---- Runtime Stage ----
FROM alpine:3.20

RUN apk add --no-cache tesseract-ocr ca-certificates tzdata
RUN tesseract --version

WORKDIR /app
COPY --from=builder /ocr-translate .

EXPOSE 9090

CMD ["./ocr-translate"]
