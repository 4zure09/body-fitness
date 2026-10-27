# Nutrition module

Sở hữu thực phẩm, khẩu phần, đơn vị, calo, macro và nguồn dữ liệu. Các phép quy đổi khẩu phần thuộc domain service của module.

- Dinh dưỡng gốc lưu theo `reference_amount` và `reference_unit` (`g`, `ml`, `unit`).
- Khẩu phần lưu lượng tương ứng trong cùng đơn vị; calo và macro được tính theo tỷ lệ khi trả response.
- PostgreSQL lưu dữ liệu dinh dưỡng bằng `NUMERIC`; domain dùng decimal, chỉ chuyển thành JSON number ở transport.
- Các endpoint đọc là công khai; create/update/delete yêu cầu role `admin`.

