# Ingestion Service

Go microservice that accepts image uploads, processes them into multiple variants, and stores them in S3 (or locally for development).

## What it does

1. Accepts a multipart image upload (JPEG, HEIC, PNG, and others)
2. Returns a `jobId` immediately (202 Accepted) — processing is async
3. In the background:
   - Converts HEIC to JPEG if needed
   - Extracts EXIF metadata (GPS is never stored)
   - Generates a thumbnail (300px longest side) and web-optimised variant (1920px longest side)
   - Uploads all three variants to S3 (or local directory)
4. Job status is queryable via a status endpoint

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | None | Health check |
| `POST` | `/api/v1/ingest/upload` | JWT | Upload an image |
| `GET` | `/api/v1/ingest/status/:jobId` | JWT | Get job status |

All non-health endpoints require a `Authorization: Bearer <token>` header. Tokens are HS256 JWTs issued by the Node.js backend.

### Upload response (202)

```json
{
  "jobId": "abc123",
  "imageId": "def456",
  "status": "queued"
}
```

### Status response (200)

```json
{
  "jobId": "abc123",
  "imageId": "def456",
  "status": "completed",
  "keys": {
    "thumbnail": "user-001/def456/thumbnail.jpg",
    "web": "user-001/def456/web.jpg",
    "original": "user-001/def456/original.heic"
  },
  "metadata": {
    "width": 4032,
    "height": 3024,
    "cameraMake": "Apple",
    "cameraModel": "iPhone 14 Pro",
    "captureDate": "2024-06-15T10:30:00Z"
  }
}
```

Status values: `queued` → `processing` → `completed` / `failed`

## Running locally

### Prerequisites

- Go 1.21+
- libheif (`brew install libheif` on macOS)

### Start the service

```bash
JWT_SECRET=smoketest STORAGE_BACKEND=local go run .
```

Processed images are written to `./local-storage/{userId}/{imageId}/`.

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | required | HS256 signing secret (must match the Node.js backend) |
| `STORAGE_BACKEND` | `s3` | `s3` or `local` |
| `LOCAL_STORAGE_DIR` | `local-storage` | Output directory when `STORAGE_BACKEND=local` |
| `AWS_REGION` | required if S3 | e.g. `ap-southeast-2` |
| `S3_BUCKET` | required if S3 | S3 bucket name |
| `PORT` | `8080` | HTTP listen port |
| `WORKER_COUNT` | `10` | Max concurrent image processing jobs |
| `MAX_FILE_SIZE_MB` | `15` | Upload size limit in MB |

Copy `.env.example` as a starting point.

### Test a upload with curl

Generate a JWT at [jwt.io](https://jwt.io) with algorithm HS256, secret `smoketest`, and payload:
```json
{ "userId": "user-001", "permissions": ["images.upload"], "exp": 9999999999 }
```

```bash
# Health check
curl http://localhost:8080/health

# Upload
curl -X POST http://localhost:8080/api/v1/ingest/upload \
  -H "Authorization: Bearer <token>" \
  -F "image=@/path/to/photo.jpg"

# Check status
curl http://localhost:8080/api/v1/ingest/status/<jobId> \
  -H "Authorization: Bearer <token>"
```

## Running with Docker

```bash
docker build -t ingestion-service:dev .

docker run --rm -p 8080:8080 \
  -e JWT_SECRET=smoketest \
  -e STORAGE_BACKEND=local \
  -e LOCAL_STORAGE_DIR=/data \
  -v $(pwd)/local-storage:/data \
  ingestion-service:dev
```

## Running tests

```bash
go test ./...
```

Note: The HEIC conversion test is skipped unless a `.heic` test fixture is present at `internal/processor/testdata/sample.heic`.

## S3 storage layout

```
{userId}/{imageId}/thumbnail.jpg    → S3 Standard
{userId}/{imageId}/web.jpg          → S3 Standard
{userId}/{imageId}/original.{ext}   → S3 Intelligent-Tiering
```

Cross-region Glacier Deep Archive backup is configured as a bucket replication rule, not in application code.
