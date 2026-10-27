# Hướng dẫn tích hợp API cho Frontend

Tài liệu này là bản hướng dẫn tích hợp nhanh cho Frontend. Contract máy đọc và đầy đủ nhất
vẫn là OpenAPI; khi tài liệu này và OpenAPI khác nhau, dừng tích hợp và xác nhận lại với
Backend trước khi tự suy đoán.

## 1. Điểm truy cập

- Production API: `https://body-fitness-production.up.railway.app`
- Local API: `http://localhost:8080`
- Scalar production: `https://body-fitness-production.up.railway.app/docs`
- OpenAPI production: `https://body-fitness-production.up.railway.app/openapi.yaml`
- OpenAPI trong source: `internal/platform/openapi/openapi.yaml`
- Backlog thay đổi API: `docs/API_CHANGELOG.md`

Frontend nên lấy base URL từ biến môi trường, không hardcode trong source. Ví dụ:

```env
VITE_API_BASE_URL=https://body-fitness-production.up.railway.app
```

Không thêm dấu `/` ở cuối base URL để tránh tạo URL có hai dấu `//`.

## 2. Quy ước chung

- REST JSON, prefix nghiệp vụ: `/api/v1`.
- Request và response dùng `snake_case`.
- `Content-Type`: `application/json`.
- ID public là UUID string.
- Thời gian là ISO 8601/RFC 3339 theo UTC.
- Các field không có giá trị có thể bị bỏ khỏi JSON nếu schema ghi optional.
- Request có field lạ bị từ chối với HTTP `400`.
- Request body tối đa 1 MiB.
- `DELETE` thành công trả `204 No Content`, không parse JSON body.

## 3. Authentication và RBAC

Các endpoint đọc catalog là public. Endpoint tạo, sửa và xóa catalog yêu cầu access token
có role `admin`.

Gửi access token bằng header:

```http
Authorization: Bearer <access_token>
```

### Luồng đăng nhập phía FE

1. Gọi register hoặc login để nhận `user` và `tokens`.
2. Gắn `access_token` vào các request cần đăng nhập.
3. Khi access token hết hạn, gọi refresh đúng một lần bằng refresh token mới nhất.
4. Refresh token được rotate; thay thế đồng thời cả access token và refresh token cũ.
5. Nếu refresh trả `401`, xóa session phía FE và chuyển người dùng về màn hình login.
6. Khi logout, gửi refresh token mới nhất rồi xóa session phía FE dù response là `204`.

Register luôn tạo role `user`. Không có API công khai để tự đăng ký role `admin`.

### Browser origin và CORS

Backend hiện chưa cấu hình CORS cho frontend chạy khác origin. Scalar hoạt động vì `/docs`
và API cùng origin. Nếu FE dev chạy tại `http://localhost:5173` hoặc FE production dùng domain
khác API, browser sẽ chặn request cho đến khi Backend thêm origin đó vào allowlist.

Trong thời gian chờ:

- Dùng dev proxy của FE để proxy `/api` tới Backend; hoặc
- Phục vụ FE và API cùng origin qua reverse proxy.

Không dùng `mode: "no-cors"`; chế độ đó không giải quyết authentication và FE không đọc được
response. Theo dõi việc mở CORS tại entry `API-2026-10-27-04` trong API change backlog.

### User schema

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "Nguyen Van A",
  "roles": ["user"]
}
```

Role hợp lệ:

- `user`
- `admin`

### TokenPair schema

```json
{
  "access_token": "jwt",
  "refresh_token": "refresh-jwt",
  "token_type": "Bearer",
  "access_expires_at": "2026-10-27T08:30:00Z",
  "refresh_expires_at": "2026-11-26T08:15:00Z"
}
```

## 4. Phân trang, tìm kiếm và lỗi

Các endpoint danh sách dùng:

- `page`: số nguyên dương, mặc định `1`.
- `page_size`: từ `1` đến `100`, mặc định `20`.
- `q`: tìm kiếm tên không phân biệt hoa thường.

Response phân trang:

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

API chưa trả `total_pages`; FE tự tính `Math.ceil(total / page_size)`.

Mọi lỗi JSON có dạng:

```json
{
  "error": {
    "code": "invalid_input",
    "message": "Nội dung mô tả lỗi"
  }
}
```

Các status FE cần xử lý:

- `400`: request, validation, UUID, filter hoặc phân trang không hợp lệ.
- `401`: thiếu/sai token, token hết hạn hoặc refresh token không còn hợp lệ.
- `403`: tài khoản không active hoặc không có role cần thiết.
- `404`: resource không tồn tại hoặc đã bị xóa mềm.
- `409`: trùng dữ liệu hoặc catalog đang được resource khác sử dụng.
- `422`: UUID tham chiếu hợp lệ về định dạng nhưng resource không tồn tại/không active.
- `500`: lỗi không dự kiến; FE hiển thị thông báo chung và không lộ chi tiết kỹ thuật.
- `503`: dependency của API chưa sẵn sàng.

FE nên dùng `error.code` cho logic và i18n; không phụ thuộc nguyên văn `error.message`.

## 5. Danh sách endpoint

### Health

- `GET /health/live` — public, kiểm tra process.
- `GET /health/ready` — public, kiểm tra process và PostgreSQL.

### Authentication

- `POST /api/v1/auth/register` — public, tạo user và trả token pair.
- `POST /api/v1/auth/login` — public, trả user và token pair.
- `POST /api/v1/auth/refresh` — public, rotate refresh token.
- `POST /api/v1/auth/logout` — public, revoke refresh token.
- `GET /api/v1/auth/me` — yêu cầu Bearer token.
- `GET /api/v1/admin/access-check` — yêu cầu role `admin`.

### Exercise catalog

- `GET /api/v1/muscle-groups` — public.
- `POST /api/v1/muscle-groups` — admin.
- `GET /api/v1/muscle-groups/{id}` — public.
- `PUT /api/v1/muscle-groups/{id}` — admin.
- `DELETE /api/v1/muscle-groups/{id}` — admin.
- `GET /api/v1/equipment` — public.
- `POST /api/v1/equipment` — admin.
- `GET /api/v1/equipment/{id}` — public.
- `PUT /api/v1/equipment/{id}` — admin.
- `DELETE /api/v1/equipment/{id}` — admin.
- `GET /api/v1/exercises` — public.
- `POST /api/v1/exercises` — admin.
- `GET /api/v1/exercises/{id}` — public.
- `PUT /api/v1/exercises/{id}` — admin.
- `DELETE /api/v1/exercises/{id}` — admin.

Filter riêng của danh sách bài tập:

- `difficulty`: `beginner`, `intermediate`, `advanced`.
- `equipment_id`: UUID.
- `muscle_group_id`: UUID.

### Nutrition catalog

- `GET /api/v1/foods` — public.
- `POST /api/v1/foods` — admin.
- `GET /api/v1/foods/{foodID}` — public.
- `PUT /api/v1/foods/{foodID}` — admin.
- `DELETE /api/v1/foods/{foodID}` — admin.
- `GET /api/v1/foods/{foodID}/servings` — public.
- `POST /api/v1/foods/{foodID}/servings` — admin.
- `GET /api/v1/foods/{foodID}/servings/{servingID}` — public.
- `PUT /api/v1/foods/{foodID}/servings/{servingID}` — admin.
- `DELETE /api/v1/foods/{foodID}/servings/{servingID}` — admin.

Filter riêng của danh sách thực phẩm:

- `brand`: so khớp brand không phân biệt hoa thường.

## 6. Request và response schema

### Register

```http
POST /api/v1/auth/register
```

```json
{
  "email": "user@example.com",
  "password": "Password@123",
  "display_name": "Nguyen Van A"
}
```

Validation chính:

- `email`: email hợp lệ, tối đa 254 ký tự.
- `password`: từ 8 đến 128 ký tự.
- `display_name`: từ 2 đến 80 ký tự.

Response `201`:

```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Nguyen Van A",
    "roles": ["user"]
  },
  "tokens": {
    "access_token": "jwt",
    "refresh_token": "refresh-jwt",
    "token_type": "Bearer",
    "access_expires_at": "2026-10-27T08:30:00Z",
    "refresh_expires_at": "2026-11-26T08:15:00Z"
  }
}
```

### Login

```json
{
  "email": "user@example.com",
  "password": "Password@123"
}
```

Response `200` giống register.

### Refresh và logout

```json
{
  "refresh_token": "refresh-jwt-moi-nhat"
}
```

- Refresh trả `200` cùng schema với login.
- Logout trả `204`, không có body.

### MuscleGroup

Body tạo/cập nhật:

```json
{
  "name": "Chest",
  "body_region": "upper_body",
  "description": "Nhóm cơ ngực"
}
```

`body_region` hợp lệ:

- `upper_body`
- `lower_body`
- `core`
- `full_body`

Response:

```json
{
  "id": "uuid",
  "name": "Chest",
  "body_region": "upper_body",
  "description": "Nhóm cơ ngực",
  "created_at": "2026-10-27T08:15:00Z",
  "updated_at": "2026-10-27T08:15:00Z"
}
```

### Equipment

Body tạo/cập nhật:

```json
{
  "name": "Barbell",
  "description": "Đòn tạ"
}
```

Response thêm `id`, `created_at`, `updated_at` tương tự MuscleGroup.

### Exercise

Body tạo/cập nhật:

```json
{
  "name": "Bench Press",
  "description": "Bài đẩy ngực",
  "instructions": "Giữ lưng ổn định và đẩy tạ có kiểm soát.",
  "difficulty": "beginner",
  "equipment_id": "uuid-or-null",
  "primary_muscle_ids": ["uuid"],
  "secondary_muscle_ids": ["uuid"]
}
```

Quy tắc:

- `difficulty`: `beginner`, `intermediate`, `advanced`.
- `equipment_id`: optional/null.
- Cần ít nhất một `primary_muscle_id`.
- Một muscle group không được xuất hiện đồng thời ở primary và secondary.
- FE nên tải muscle groups và equipment trước khi render form bài tập.

Response:

```json
{
  "id": "uuid",
  "name": "Bench Press",
  "description": "Bài đẩy ngực",
  "instructions": "Giữ lưng ổn định và đẩy tạ có kiểm soát.",
  "difficulty": "beginner",
  "equipment": {
    "id": "uuid",
    "name": "Barbell",
    "description": "Đòn tạ",
    "created_at": "2026-10-27T08:15:00Z",
    "updated_at": "2026-10-27T08:15:00Z"
  },
  "muscles": [
    {
      "muscle_group": {
        "id": "uuid",
        "name": "Chest",
        "body_region": "upper_body",
        "description": "Nhóm cơ ngực",
        "created_at": "2026-10-27T08:15:00Z",
        "updated_at": "2026-10-27T08:15:00Z"
      },
      "involvement": "primary"
    }
  ],
  "created_at": "2026-10-27T08:15:00Z",
  "updated_at": "2026-10-27T08:15:00Z"
}
```

`equipment` có thể không xuất hiện khi bài tập không dùng dụng cụ. `involvement` là
`primary` hoặc `secondary`.

### Food và nutrients

Body tạo/cập nhật:

```json
{
  "name": "Chicken Breast",
  "brand": null,
  "description": "Ức gà chín",
  "reference_amount": 100,
  "reference_unit": "g",
  "nutrients": {
    "calories_kcal": 165,
    "protein_g": 31.2,
    "carbohydrates_g": 0,
    "fat_g": 3.6,
    "fiber_g": 0
  }
}
```

Quy tắc:

- `reference_unit`: `g`, `ml`, `unit`.
- `reference_amount` lớn hơn `0`.
- Mọi nutrient không âm.
- Nutrients mô tả lượng dinh dưỡng trên `reference_amount`, không mặc định luôn là 100g.

Response thêm `id`, `created_at`, `updated_at`. `brand` có thể không xuất hiện nếu null.

### Serving

Body tạo/cập nhật:

```json
{
  "name": "Một phần 150g",
  "amount": 150
}
```

Response:

```json
{
  "id": "uuid",
  "food_id": "uuid",
  "name": "Một phần 150g",
  "amount": 150,
  "unit": "g",
  "nutrition": {
    "calories_kcal": 247.5,
    "protein_g": 46.8,
    "carbohydrates_g": 0,
    "fat_g": 5.4,
    "fiber_g": 0
  },
  "created_at": "2026-10-27T08:15:00Z",
  "updated_at": "2026-10-27T08:15:00Z"
}
```

`nutrition` đã được Backend quy đổi theo `amount`; FE không tự tính lại.

## 7. Gợi ý tổ chức API client phía FE

Tách client theo domain thay vì gọi HTTP trực tiếp trong component:

```text
api/http-client
api/auth
api/exercises
api/nutrition
```

HTTP client chung nên chịu trách nhiệm:

- Gắn base URL và `Content-Type`.
- Gắn Bearer token.
- Parse error envelope.
- Refresh token một lần và retry request phù hợp.
- Không parse body cho response `204`.
- Hủy request khi component/query không còn cần kết quả.

Mỗi khi bắt đầu tích hợp hoặc nâng version, FE kiểm tra `docs/API_CHANGELOG.md` trước,
sau đó đối chiếu OpenAPI/Scalar để lấy contract mới nhất.
