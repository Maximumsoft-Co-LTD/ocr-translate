# OCR Translate

Tesseract OCR HTTP microservice - แปลงรูปภาพเป็นข้อความผ่าน REST API

## Architecture

```
POST /solve (image) → [Tesseract OCR] → "A3kZ" (text)
```

## Prerequisites

- **Go** >= 1.24.0 (สำหรับ build)
- **Tesseract** (สำหรับ runtime)
  - macOS: `brew install tesseract`
  - Ubuntu: `apt-get install tesseract-ocr`
  - Alpine: `apk add tesseract-ocr`

## Run

### Local

```bash
go run main.go
# → OCR Translate Service starting on port 9090
```

### Docker

```bash
docker build -t ocr-translate .
docker run -d -p 9090:9090 --name ocr-translate ocr-translate
```

## API

### GET /health

Health check

```bash
curl http://localhost:9090/health
```

```json
{"status":"ok","engine":"tesseract","version":"tesseract 5.5.2"}
```

### POST /solve

แปลงรูปภาพเป็นข้อความ

**Request:** multipart/form-data

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | file | **Yes** | - | ไฟล์รูปภาพ (png, jpg, etc.) |
| `psm` | string | No | `7` | Tesseract Page Segmentation Mode |
| `whitelist` | string | No | `0-9A-Za-z` | ตัวอักษรที่อนุญาต |

**Response:** `text/plain` - ข้อความที่อ่านได้จากรูป

**Example:**

```bash
curl -X POST http://localhost:9090/solve \
  -F "image=@captcha.png"
# → A3kZ
```

**ปรับ OCR options:**

```bash
# psm=8 (single word), เฉพาะตัวเลข
curl -X POST http://localhost:9090/solve \
  -F "image=@numbers.png" \
  -F "psm=8" \
  -F "whitelist=0123456789"
# → 4829
```

## PSM Modes

| Mode | ความหมาย |
|------|----------|
| `6` | Uniform block of text |
| `7` | Single text line (default) |
| `8` | Single word |
| `10` | Single character |
| `13` | Raw line |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9090` | Port ที่ server listen |

## Integration

ใช้ร่วมกับ service อื่นโดยชี้ URL มาที่ `/solve`:

```bash
# ตัวอย่าง: 1key project
CAPTCHA_SOLVER=external
EXTERNAL_SOLVER_URL=http://ocr-translate:9090/solve
```
