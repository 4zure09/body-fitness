# Exercises module

Sở hữu bài tập, nhóm cơ, dụng cụ, độ khó và quan hệ nhóm cơ chính/phụ. Tìm kiếm phức tạp được tách khỏi write repository.

- Các endpoint đọc là công khai, hỗ trợ phân trang và tìm kiếm.
- Create/update/delete yêu cầu access token có role `admin`.
- Nhóm cơ và dụng cụ đang được bài tập active sử dụng không thể xóa.
- Xóa catalog dùng soft delete để giữ khả năng tham chiếu lịch sử trong các module tracking sau này.

