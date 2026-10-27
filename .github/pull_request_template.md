## Nội dung

Mô tả ngắn thay đổi và lý do.

## Kiểm tra

- [ ] Đã chạy test phù hợp.
- [ ] Không commit credential, token hoặc connection string thật.

## API contract

- [ ] Đã xác định PR này có thay đổi API contract hay không.

Nếu có thêm/sửa/deprecate/xóa API:

- [ ] Đã cập nhật `internal/platform/openapi/openapi.yaml`.
- [ ] Đã cập nhật `docs/API_INTEGRATION.md`.
- [ ] Đã thêm entry vào `docs/API_CHANGELOG.md` với trạng thái FE.
- [ ] Đã ghi rõ breaking change, migration window và endpoint/schema bị ảnh hưởng.
- [ ] Đã thêm/cập nhật test cho contract và permission liên quan.

Nếu không thay đổi API contract, ghi `N/A` ở phần mô tả PR.
