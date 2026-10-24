# Business modules

Mỗi module sở hữu nghiệp vụ và dữ liệu của chính nó. Cấu trúc chuẩn khi bắt đầu triển khai một module:

```text
<module>/
├── domain/          # Entity, value object, domain rule; không phụ thuộc framework
├── application/     # Use case; điều phối một yêu cầu nghiệp vụ
├── port/            # Interface repository/gateway do application cần
├── infrastructure/  # GORM repository, sqlc report adapter, external provider
└── transport/http/  # Request DTO, handler, route registration
```

Các module dự kiến:

- `auth`: đăng ký, đăng nhập, refresh/revoke token và phân quyền.
- `users`: hồ sơ và dữ liệu sinh học cơ bản.
- `exercises`: bài tập, nhóm cơ và dụng cụ.
- `nutrition`: thực phẩm, khẩu phần, calo và macro.
- `recipes`: công thức, nguyên liệu và hướng dẫn chế biến.
- `programs`: chế độ mẫu và chế độ cá nhân.
- `tracking`: nhật ký tập luyện, ăn uống và số đo.
- `goals`: mục tiêu, milestone và tiến độ.
- `ranking`: point event, thành tích và bảng xếp hạng.
- `advice`: tính chỉ số và tư vấn rule-based có guardrail.

Không import trực tiếp infrastructure của module khác. Giao tiếp xuyên module thông qua application service hoặc port nhỏ, có mục đích rõ ràng.

