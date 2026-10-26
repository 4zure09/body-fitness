# Body Fitness

Ứng dụng hỗ trợ người dùng xây dựng và theo dõi mục tiêu hình thể thông qua bài tập, chế độ ăn, dữ liệu sinh học cơ bản và các nội dung hướng dẫn.

> Sản phẩm chỉ cung cấp thông tin và gợi ý sức khỏe/lối sống phổ thông. Sản phẩm không chẩn đoán, điều trị hoặc thay thế tư vấn của bác sĩ, chuyên gia dinh dưỡng hay chuyên gia y tế.

## Tài liệu dự án

- [Hướng dẫn cài đặt môi trường](docs/INSTALLATION.md)
- [Roadmap MVP theo phase và sprint](docs/MVP_ROADMAP.md)
- [Kiến trúc và quy tắc tổ chức code](docs/ARCHITECTURE.md)

## Định hướng kỹ thuật

- Backend: Go
- Database: PostgreSQL
- Kiến trúc giai đoạn MVP: modular monolith, REST API
- Message queue: chưa sử dụng trong MVP

## Chạy local

Yêu cầu: Git, Go 1.26.x, Docker Desktop, `sqlc` và Goose. Xem [hướng dẫn cài đặt chi tiết](docs/INSTALLATION.md).

```bash
docker compose up postgres -d
go run ./cmd/api
```

Các endpoint nền tảng:

- API docs bằng Scalar: `http://localhost:8080/docs`
- OpenAPI schema: `http://localhost:8080/openapi.yaml`
- Liveness: `http://localhost:8080/health/live`
- Readiness: `http://localhost:8080/health/ready`

Authentication API:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me` — yêu cầu access token
- `GET /api/v1/admin/access-check` — yêu cầu role `admin`

Chạy toàn bộ stack bằng container:

```bash
docker compose up --build
```

Database schema được quản lý bằng Goose và query báo cáo được sinh bằng `sqlc`:

```bash
goose -dir db/migrations postgres "$DATABASE_DSN" up
sqlc generate
```

Hai CLI này là development tools, không được nhúng vào runtime image. File migration và `sqlc.yaml` đã được đặt sẵn trong repo.

Để cấp role `admin` cho một tài khoản đã đăng ký:

```bash
go run ./cmd/admin --email admin@example.com
```

Role mới có hiệu lực trên access token được cấp sau lần login/refresh tiếp theo.

## Cấu trúc chính

```text
cmd/api/                  # Entrypoint
internal/bootstrap/       # Dependency wiring và lifecycle
internal/modules/         # Các module nghiệp vụ
internal/platform/        # Database, HTTP, config, logging, OpenAPI
internal/shared/          # Helper kỹ thuật nhỏ, không mang nghiệp vụ
db/migrations/            # Goose migrations
db/queries/reports/       # Query báo cáo dùng sqlc
```

