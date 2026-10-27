# API change backlog cho Frontend

File này là hàng đợi phối hợp Backend → Frontend. Mọi API mới, thay đổi contract,
deprecation hoặc removal phải có một entry tại đây trong cùng commit/PR với code Backend.

## Quy ước trạng thái

- `[ ] FE pending`: Backend đã sẵn sàng, FE chưa xác nhận.
- `[~] FE in progress`: FE đang tích hợp hoặc kiểm tra regression.
- `[x] FE synced`: FE đã cập nhật và xác nhận.
- `[!] Blocked`: cần quyết định hoặc dependency khác.
- `[N/A]`: thay đổi không ảnh hưởng contract FE.

Mỗi entry phải ghi:

- Mã thay đổi duy nhất.
- Ngày và commit/PR Backend.
- Loại: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Breaking`.
- Endpoint/schema liên quan.
- Tác động và hành động cụ thể cho FE.
- Trạng thái FE.

Không xóa entry đã hoàn thành; chuyển trạng thái sang `[x]` để giữ lịch sử.

## FE review backlog

### API-2026-10-27-04 — CORS allowlist cho Frontend

- Trạng thái: `[!] Blocked`
- Backend commit/PR: Chưa có.
- Loại: `Added`
- Breaking: Không.
- Phạm vi: Tất cả browser request từ origin khác API.
- Blocker:
  - FE cần cung cấp origin local, preview và production chính xác.
- Backend cần làm:
  - Thêm CORS middleware theo allowlist; không dùng wildcard khi có credential/token.
  - Cho phép các method/header thực tế và xử lý preflight `OPTIONS`.
  - Thêm test cho allowed origin và rejected origin.
- FE cần làm:
  - Trong lúc chờ, dùng dev proxy hoặc reverse proxy same-origin.
  - Không dùng `mode: "no-cors"`.
- Cách kiểm tra:
  - Browser request từ từng origin được phép không còn lỗi CORS.
  - Origin ngoài allowlist không nhận CORS allow headers.

### API-2026-10-27-02 — Exercise và nutrition catalog

- Trạng thái: `[ ] FE pending`
- Backend commit: `19989aa`
- Loại: `Added`
- Breaking: Không.
- Endpoint:
  - `/api/v1/muscle-groups`
  - `/api/v1/equipment`
  - `/api/v1/exercises`
  - `/api/v1/foods`
  - `/api/v1/foods/{foodID}/servings`
- FE cần làm:
  - Tạo API client và type cho muscle group, equipment, exercise, food, serving.
  - Tích hợp phân trang `page`, `page_size`, `total` và tìm kiếm `q`.
  - Tích hợp filter bài tập và brand thực phẩm.
  - Chỉ hiển thị hành động create/update/delete cho role `admin`.
  - Xử lý `409` khi catalog đang được sử dụng và `422` khi reference không hợp lệ.
  - Không tự tính nutrition của serving; dùng field `nutrition` từ response.
- Kiểm tra contract: `docs/API_INTEGRATION.md`, Scalar `/docs`.

### API-2026-10-27-01 — Authentication, refresh rotation và RBAC

- Trạng thái: `[ ] FE pending`
- Backend commit: `dbc5eca`
- Loại: `Added`
- Breaking: Không.
- Endpoint:
  - `/api/v1/auth/register`
  - `/api/v1/auth/login`
  - `/api/v1/auth/refresh`
  - `/api/v1/auth/logout`
  - `/api/v1/auth/me`
  - `/api/v1/admin/access-check`
- FE cần làm:
  - Lưu và thay thế token pair sau login/register/refresh.
  - Gắn Bearer access token cho endpoint yêu cầu xác thực.
  - Refresh đúng một lần khi token hết hạn; logout nếu refresh trả `401`.
  - Route/action guard theo `roles`, nhưng vẫn coi Backend là nơi quyết định quyền cuối cùng.
  - Register không cho chọn role admin.
- Kiểm tra contract: `docs/API_INTEGRATION.md`, Scalar `/docs`.

## Đã đồng bộ hoặc không cần FE thay đổi

### API-2026-10-27-03 — OpenAPI dùng current origin

- Trạng thái: `[N/A]`
- Backend commit: `19989aa`
- Loại: `Fixed`
- Breaking: Không.
- Thay đổi: OpenAPI server chuyển từ hardcoded localhost sang `/` để Scalar gọi đúng origin.
- FE cần làm: Không; base URL của FE vẫn lấy từ environment.

## Template cho thay đổi mới

Copy block dưới đây lên đầu mục `FE review backlog`:

```markdown
### API-YYYY-MM-DD-NN — Tên thay đổi

- Trạng thái: `[ ] FE pending`
- Backend commit/PR: `<link-or-sha>`
- Loại: `Added | Changed | Deprecated | Removed | Fixed | Breaking`
- Breaking: `Có/Không` — nếu có, ghi migration window hoặc ngày loại bỏ contract cũ.
- Endpoint/schema:
  - `<METHOD /path hoặc Schema.field>`
- Backend thay đổi:
  - `<mô tả contract, status code, auth hoặc behavior>`
- FE cần làm:
  - `<hành động cụ thể>`
- Cách kiểm tra:
  - `<Scalar/OpenAPI/test case>`
```

## Quy trình đóng backlog

1. Backend cập nhật code, OpenAPI, `API_INTEGRATION.md` và thêm entry `[ ] FE pending`.
2. FE đọc entry, chuyển sang `[~] FE in progress` trong PR tích hợp nếu cần.
3. FE kiểm tra happy path, validation, auth/RBAC và error states.
4. FE chuyển entry sang `[x] FE synced`, ghi commit/PR FE cạnh entry.
5. Breaking change chỉ được remove contract cũ sau khi FE synced hoặc có xác nhận release riêng.
