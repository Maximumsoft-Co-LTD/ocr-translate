# ---- Build Stage ----
FROM --platform=linux/amd64 docker.mirror.hashicorp.services/golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /ocr-translate .

# ---- Runtime Stage ----
FROM --platform=linux/amd64 docker.mirror.hashicorp.services/alpine:3.20

RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng ca-certificates tzdata
RUN tesseract --version
RUN tesseract --list-langs

ENV TESSDATA_PREFIX=/usr/share/tessdata
ENV OMP_THREAD_LIMIT=1

WORKDIR /app
COPY --from=builder /ocr-translate .

EXPOSE 9090

CMD ["./ocr-translate"]