# Kiến trúc dự án

## 1. Kiểu kiến trúc

Dự án dùng **modular monolith**. Toàn bộ API được build thành một binary, nhưng nghiệp vụ được chia theo feature để mỗi module có ownership và ranh giới phụ thuộc rõ ràng.

```text
cmd/api                     Composition root và process lifecycle
internal/bootstrap          Khởi tạo và nối dependency
internal/modules            Nghiệp vụ theo feature
internal/platform           Database, HTTP server, config, logging, OpenAPI
internal/shared             Primitive kỹ thuật nhỏ, không chứa nghiệp vụ
internal/generated          Code do tool sinh; không sửa thủ công
db/migrations               Goose migrations
db/queries/reports          SQL phức tạp dùng với sqlc
```

## 2. Luồng phụ thuộc

```text
HTTP handler
    ↓
Application use case
    ↓
Port/interface
    ↓
Infrastructure adapter (GORM hoặc sqlc)
    ↓
PostgreSQL
```

`domain` không phụ thuộc HTTP, GORM, PostgreSQL hay framework. `application` có thể phụ thuộc `domain` và interface trong `port`. Adapter phụ thuộc các interface mà nó hiện thực.

## 3. Trách nhiệm từng lớp

### Domain

- Bảo vệ invariant và biểu diễn khái niệm nghiệp vụ.
- Không parse JSON, đọc biến môi trường hoặc thực thi SQL.
- Method thay đổi entity chỉ thực hiện một hành vi nghiệp vụ rõ ràng.

### Application

- Mỗi use case biểu diễn một yêu cầu: `RegisterUser`, `LogWorkout`, `CreateGoal`.
- Điều phối repository, transaction, clock, password hasher hoặc external gateway.
- Không chứa chi tiết HTTP hay câu SQL.

### Port

- Interface nhỏ, được định nghĩa gần nơi sử dụng.
- Không tạo một repository khổng lồ cho toàn module.
- Tách read port và write port khi nhu cầu truy vấn khác nhau.

### Infrastructure

- GORM cho CRUD/query đơn giản.
- `sqlc` cho báo cáo, join, aggregate, CTE và truy vấn cần kiểm soát SQL.
- Goose là nguồn duy nhất thay đổi schema production.
- Không để GORM model lan vào domain hoặc response DTO.

### Transport

- Decode và validate request.
- Gọi đúng một application use case.
- Chuyển kết quả/application error thành HTTP response.
- Không chứa nghiệp vụ, không gọi database trực tiếp.

## 4. Quy tắc single responsibility

“Mỗi hàm chỉ làm một việc” được đánh giá theo mức trừu tượng, không phải số dòng:

- Handler chỉ xử lý HTTP và gọi use case.
- Use case chỉ điều phối một hành vi nghiệp vụ.
- Repository method chỉ thực hiện một ý định lưu/đọc dữ liệu.
- Hàm config chỉ đọc hoặc parse một loại cấu hình.
- Hàm bootstrap được phép nối dependency vì đó là trách nhiệm duy nhất của composition root.

Tách hàm khi một đoạn code có lý do thay đổi khác, cần test độc lập hoặc có tên mô tả tốt hơn comment. Không tách các hàm một dòng nếu việc đó làm luồng xử lý khó hiểu hơn.

## 5. Quy tắc truy cập database

- CRUD đơn giản thuộc write model dùng GORM.
- Báo cáo/read model phức tạp dùng SQL trong `db/queries/reports` và sinh code bằng `sqlc`.
- Không nối input người dùng vào chuỗi SQL.
- Dynamic sort/filter identifier phải ánh xạ qua allowlist do server kiểm soát.
- Không trộn GORM và `sqlc` trong cùng transaction nếu chưa có transaction boundary chung.
- Tránh transaction cho báo cáo read-only; ghi dữ liệu nghiệp vụ qua application use case và GORM repository.

## 6. Thêm module mới

1. Xác định entity, invariant và input/output của use case.
2. Viết port nhỏ mà use case cần.
3. Viết use case và unit test bằng fake/stub port.
4. Viết infrastructure adapter và integration test PostgreSQL.
5. Viết HTTP DTO/handler, đăng ký route tại module registrar.
6. Nối module tại `internal/bootstrap`.
7. Cập nhật OpenAPI và migration nếu cần.

## 7. Các quyết định nền tảng

- API: REST JSON dưới `/api/v1`.
- API contract: OpenAPI 3.1, hiển thị bằng Scalar tại `/docs`.
- ID public: UUID sinh tại application/domain boundary.
- Thời gian: lưu UTC; timezone người dùng dùng khi tổng hợp theo ngày.
- Logging: structured JSON bằng `log/slog`; mỗi request có request ID.
- Shutdown: graceful shutdown khi nhận `SIGINT` hoặc `SIGTERM`.
- Health: `/health/live` kiểm tra process, `/health/ready` kiểm tra PostgreSQL.

## 8. Quản trị API contract với Frontend

OpenAPI tại `internal/platform/openapi/openapi.yaml` là contract máy đọc. Tài liệu
`docs/API_INTEGRATION.md` là hướng dẫn tích hợp cho FE và `docs/API_CHANGELOG.md` là backlog
phối hợp BE → FE.

Mọi thay đổi thêm/sửa/deprecate/xóa endpoint, request, response, enum, auth, status code hoặc
error code phải thực hiện trong cùng commit/PR:

1. Cập nhật handler và test.
2. Cập nhật OpenAPI.
3. Cập nhật API integration guide nếu cách tích hợp bị ảnh hưởng.
4. Thêm entry vào API change backlog với trạng thái `[ ] FE pending`.
5. Ghi rõ breaking change và thời gian chuyển đổi nếu có.

Không merge API change khi thiếu một trong các phần trên. FE chuyển backlog sang `[x] FE synced`
sau khi tích hợp và kiểm tra regression; không xóa lịch sử thay đổi đã hoàn thành.

