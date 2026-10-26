# Cài đặt môi trường phát triển

Tài liệu này dành cho Windows/PowerShell và tách rõ phần bắt buộc với công cụ hỗ trợ phát triển.

## 1. Phần mềm bắt buộc

### 1.1. Git

Dùng để clone, commit và đồng bộ source code.

- Tải từ: <https://git-scm.com/download/win>
- Kiểm tra sau khi cài:

```powershell
git --version
```

### 1.2. Go 1.26.x

Project khai báo `go 1.26.0` trong `go.mod` và Docker build bằng Go 1.26. Nên cài bản patch mới nhất thuộc dòng 1.26.

- Tải từ: <https://go.dev/dl/>
- Hướng dẫn chính thức: <https://go.dev/doc/install>
- Kiểm tra:

```powershell
go version
```

Kết quả cần bắt đầu bằng `go version go1.26`.

### 1.3. Docker Desktop

Docker Desktop cung cấp cả Docker Engine và lệnh `docker compose`. Project dùng container PostgreSQL 17, vì vậy không cần cài PostgreSQL trực tiếp lên Windows.

- Hướng dẫn chính thức: <https://docs.docker.com/desktop/setup/install/windows-install/>
- Khuyến nghị dùng backend WSL 2 và Linux containers.
- Kiểm tra:

```powershell
docker version
docker compose version
docker run --rm hello-world
```

Nếu máy chưa có WSL 2, mở PowerShell bằng quyền Administrator rồi chạy:

```powershell
wsl --install
wsl --update
```

Khởi động lại Windows nếu trình cài đặt yêu cầu.

## 2. Development CLI cần cài

Các lệnh sau cài binary vào thư mục `$(go env GOPATH)\bin`.

### 2.1. sqlc 1.31.1

`sqlc` đọc SQL trong `db/queries/reports` và sinh mã Go type-safe vào `internal/generated/reportquery`.

```powershell
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
sqlc version
```

Tài liệu chính thức: <https://docs.sqlc.dev/en/latest/overview/install.html>

### 2.2. Goose 3.28.0

Goose chạy các migration trong `db/migrations`.

```powershell
go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
goose --version
```

Tài liệu chính thức: <https://pressly.github.io/goose/installation/>

### 2.3. Thêm Go binary directory vào PATH

Nếu PowerShell báo không tìm thấy `sqlc` hoặc `goose`, kiểm tra thư mục binary:

```powershell
go env GOPATH
Get-ChildItem "$(go env GOPATH)\bin"
```

Thêm tạm cho terminal hiện tại:

```powershell
$env:Path += ";$(go env GOPATH)\bin"
```

Để dùng lâu dài, thêm giá trị `$(go env GOPATH)\bin` thực tế, thường là `%USERPROFILE%\go\bin`, vào biến môi trường `Path` của user trong Windows.

## 3. Không cần cài thủ công

- **PostgreSQL:** Docker Compose dùng image `postgres:17-alpine`.
- **Scalar:** trang tài liệu API tải Scalar ở runtime; không cần Node.js hoặc npm.
- **Chi, GORM, pgx:** Go tự tải theo `go.mod` khi chạy `go mod download`, `go test` hoặc `go run`.
- **Message queue:** chưa dùng trong MVP.
- **Node.js/npm:** backend hiện tại không cần.

## 4. Kiểm tra toàn bộ môi trường

Chạy tại thư mục project:

```powershell
git --version
go version
docker version
docker compose version
sqlc version
goose --version
go mod download
go test ./...
go vet ./...
docker compose config --quiet
```

## 5. Khởi tạo project lần đầu

### Bước 1: chạy PostgreSQL

```powershell
docker compose up postgres -d
docker compose ps
```

### Bước 2: khai báo chuỗi kết nối cho Goose

API có giá trị mặc định cho môi trường local, nhưng Goose cần nhận DSN từ câu lệnh hoặc biến môi trường:

```powershell
$env:DATABASE_DSN = "postgres://body_fitness:body_fitness@localhost:5432/body_fitness?sslmode=disable"
$env:JWT_SECRET = "replace-with-a-random-secret-containing-at-least-32-characters"
```

### Bước 3: chạy migration

```powershell
goose -dir db/migrations postgres $env:DATABASE_DSN up
goose -dir db/migrations postgres $env:DATABASE_DSN status
```

### Bước 4: sinh code query báo cáo

```powershell
sqlc generate
```

Chạy lại lệnh này khi schema hoặc SQL trong `db/queries/reports` thay đổi.

### Bước 5: chạy API

```powershell
go run ./cmd/api
```

Kiểm tra:

- Scalar API docs: <http://localhost:8080/docs>
- OpenAPI schema: <http://localhost:8080/openapi.yaml>
- Liveness: <http://localhost:8080/health/live>
- Readiness: <http://localhost:8080/health/ready>

### Cấp quyền admin ban đầu

Đăng ký tài khoản trước, sau đó chạy management command:

```powershell
go run ./cmd/admin --email admin@example.com
```

Đăng nhập hoặc refresh lại để access token mới chứa role `admin`. Không đặt tài khoản admin mặc định hoặc mật khẩu admin trong migration/source code.

## 6. Cách chạy chỉ với Docker

Nếu chỉ muốn chạy ứng dụng mà không sửa hoặc test mã Go:

```powershell
docker compose up --build
```

Lưu ý: cấu hình Compose hiện chưa tự chạy Goose migration. Khi database mới được tạo, vẫn cần chạy migration bằng Goose hoặc bổ sung một migration service trong Compose/CI trước khi dùng API nghiệp vụ.

## 7. Checklist kết quả mong đợi

- `go version` là Go 1.26.x.
- `docker compose ps` hiển thị PostgreSQL ở trạng thái healthy.
- `go test ./...` chạy thành công.
- `sqlc generate` sinh package `internal/generated/reportquery`.
- `goose ... status` hiển thị migration `000001` đã được áp dụng.
- `/health/live` và `/health/ready` trả HTTP 200.

