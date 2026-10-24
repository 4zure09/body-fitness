# Body Fitness — Product Roadmap và kế hoạch Sprint cho MVP

## 1. Mục tiêu sản phẩm

Body Fitness giúp người dùng xác định hình thể mong muốn và theo dõi quá trình thực hiện thông qua:

- danh mục bài tập và nhóm cơ;
- chế độ tập luyện, ăn uống và thực phẩm/calo;
- kế hoạch cá nhân và nhật ký thực hiện;
- hướng dẫn sơ chế, chế biến món ăn;
- mục tiêu, tiến độ, thành tích và xếp hạng;
- tư vấn cơ bản dựa trên dữ liệu người dùng tự cung cấp.

### Ranh giới an toàn của tính năng tư vấn

Tính năng tư vấn chỉ được phép:

- tính và giải thích các chỉ số phổ thông như BMI, BMR, TDEE và mức calo tham khảo;
- gợi ý thói quen vận động, dinh dưỡng và nghỉ ngơi ở mức thông tin chung;
- gợi ý nội dung, bài tập hoặc chế độ có sẵn trong hệ thống;
- cảnh báo người dùng tìm chuyên gia khi dữ liệu nằm ngoài ngưỡng an toàn hoặc có triệu chứng bất thường.

Tính năng không được phép chẩn đoán bệnh, kê đơn, đề xuất điều trị, thay đổi thuốc, xử lý cấp cứu hoặc khẳng định kết luận y khoa. Mọi màn hình tư vấn phải hiển thị tuyên bố miễn trừ trách nhiệm phù hợp.

## 2. Phạm vi MVP

### 2.1. Vai trò

**Khách:** xem nội dung công khai, đăng ký và đăng nhập.

**Người dùng:** quản lý hồ sơ, mục tiêu, chế độ cá nhân, nhật ký tập/ăn, tiến độ, nhận gợi ý cơ bản và xem hướng dẫn món ăn.

**Quản trị viên:** quản lý danh mục bài tập, nhóm cơ, thực phẩm, món ăn, bài viết và mẫu chế độ.

### 2.2. Module chức năng

1. **Authentication & Account**
   - Đăng ký, đăng nhập, đăng xuất.
   - Access token và refresh token.
   - Xác thực email có thể triển khai sau MVP; quên mật khẩu thuộc nhóm ưu tiên cao sau khi luồng đăng nhập ổn định.
   - Hồ sơ gồm ngày sinh, giới tính sinh học tùy chọn, chiều cao, cân nặng, mức vận động và đơn vị đo.

2. **Danh mục bài tập và nhóm cơ**
   - CRUD nhóm cơ, bài tập, dụng cụ, độ khó và hướng dẫn thực hiện.
   - Một bài tập có thể tác động nhiều nhóm cơ, gồm nhóm chính và nhóm phụ.
   - Tìm kiếm/lọc theo nhóm cơ, dụng cụ, độ khó và mục tiêu.

3. **Danh mục dinh dưỡng**
   - CRUD thực phẩm, khẩu phần, đơn vị, calo và macro: protein, carbohydrate, fat.
   - Danh mục món ăn/công thức gồm nguyên liệu, khẩu phần và tổng dinh dưỡng tính toán.
   - Cho phép đánh dấu số liệu ước tính và lưu nguồn dữ liệu.

4. **Hướng dẫn sơ chế và chế biến**
   - Bài viết có tiêu đề, ảnh đại diện, nội dung, nguyên liệu, bước làm, thời gian, khẩu phần và tag.
   - Liên kết bài viết với món ăn/thực phẩm liên quan.
   - Trạng thái nháp, đã xuất bản và ẩn.

5. **Chế độ (Regimen/Program)**
   - Mẫu chế độ tập, chế độ ăn hoặc chế độ kết hợp.
   - Lịch theo ngày/tuần, thời lượng, mục tiêu, độ khó và điều kiện áp dụng.
   - Người dùng có thể sao chép mẫu thành chế độ riêng rồi điều chỉnh.

6. **Tracking chế độ riêng**
   - Ghi nhận buổi tập, set, rep, thời lượng, mức tạ và trạng thái hoàn thành.
   - Ghi nhận bữa ăn, thực phẩm, khẩu phần, calo và macro.
   - Ghi nhận cân nặng, số đo cơ thể và ghi chú theo ngày.
   - Tổng hợp theo ngày/tuần và so sánh với kế hoạch.

7. **Mục tiêu, tiến độ và xếp hạng**
   - Mục tiêu về cân nặng, số đo, số buổi tập, thời lượng hoặc mức độ tuân thủ.
   - Biểu đồ tiến độ và mốc hoàn thành.
   - Điểm/xếp hạng dựa trên tính đều đặn và hoàn thành kế hoạch, không dựa trên ngoại hình hoặc cân nặng tuyệt đối.
   - Chống gian lận cơ bản: giới hạn số lần ghi nhận, log thay đổi và không cộng điểm trùng.

8. **Tư vấn cơ bản**
   - Tính BMI, BMR, TDEE bằng công thức có phiên bản và giải thích giả định.
   - Gợi ý mục tiêu calo/mức vận động trong khoảng an toàn do hệ thống cấu hình.
   - Gợi ý bài tập, chế độ và bài viết hiện có theo mục tiêu người dùng.
   - Mỗi kết quả lưu input, phiên bản rule, output và cảnh báo để có thể truy vết.

## 3. Kiến trúc kỹ thuật đề xuất

### 3.1. Kiến trúc tổng thể

MVP nên dùng **modular monolith**: một service Go, một PostgreSQL, các module tách rõ trong code. Cách này giảm chi phí vận hành nhưng vẫn cho phép tách service sau này nếu tải hoặc đội ngũ tăng.

Các lớp chính:

- `transport/http`: router, handler, middleware, request/response DTO;
- `application`: use case và transaction boundary;
- `domain`: entity, value object, business rule;
- `infrastructure`: PostgreSQL, token, file storage, clock và external provider;
- `migrations`: schema migration;
- `cmd/api`: entrypoint HTTP server.

Module dự kiến: `auth`, `users`, `exercises`, `nutrition`, `recipes`, `programs`, `tracking`, `goals`, `progress`, `ranking`, `advice`, `articles` và `admin`.

### 3.2. Thư viện Go đề xuất

| Nhu cầu | Thư viện | Lý do |
| --- | --- | --- |
| HTTP router | `github.com/go-chi/chi/v5` | Nhẹ, gần chuẩn `net/http`, middleware rõ ràng |
| PostgreSQL driver/pool | `github.com/jackc/pgx/v5` | Driver PostgreSQL hiệu năng tốt; GORM PostgreSQL driver cũng sử dụng pgx |
| CRUD/ORM | `gorm.io/gorm` + `gorm.io/driver/postgres` | Dùng cho CRUD, filter và quan hệ đơn giản để giảm code lặp |
| Báo cáo/truy vấn phức tạp | `github.com/sqlc-dev/sqlc` | Dùng SQL tường minh cho join, aggregate, CTE và truy vấn cần tối ưu; sinh mã Go type-safe |
| Migration | `github.com/pressly/goose/v3` | Migration đơn giản, dễ chạy trong CI/CD |
| Validate input | `github.com/go-playground/validator/v10` | Validate DTO phổ biến |
| JWT | `github.com/golang-jwt/jwt/v5` | Access/refresh token chuẩn JWT |
| Password | `golang.org/x/crypto/argon2` | Băm mật khẩu Argon2id; cần lưu salt và tham số |
| Config | `github.com/caarlos0/env/v11` | Map biến môi trường sang struct có type |
| API docs | Scalar + OpenAPI | Hiển thị tài liệu API bằng Scalar; đặc tả OpenAPI được quản lý trong repo |
| Logging | `log/slog` | Có sẵn trong Go, structured logging |
| Metrics | `github.com/prometheus/client_golang` | Metrics cho vận hành khi triển khai |
| Integration test | `github.com/testcontainers/testcontainers-go` | Test thật với PostgreSQL cô lập |
| Assertions | `github.com/stretchr/testify` | Viết test ngắn gọn, dễ đọc |

Tầng truy cập dữ liệu sử dụng kết hợp GORM và `sqlc` theo ranh giới rõ ràng:

- **GORM:** CRUD và query đơn giản như lấy theo ID, tạo/cập nhật bản ghi, phân trang cơ bản và tải quan hệ ít phức tạp.
- **`sqlc`:** báo cáo, dashboard, nhiều `JOIN`, aggregate, CTE, window function hoặc truy vấn cần kiểm soát execution plan.
- Một repository/use case chỉ nên có một cách truy cập dữ liệu chính. Không trộn GORM và `sqlc` trong cùng một transaction nếu chưa có thiết kế transaction chung rõ ràng; các truy vấn báo cáo bằng `sqlc` nên ưu tiên read-only.
- Schema production chỉ được thay đổi bằng Goose migration; không dùng GORM `AutoMigrate` ngoài thử nghiệm cục bộ.

`sqlc` không phải ORM. Lập trình viên viết câu SQL trong các file `.sql`; `sqlc` kiểm tra câu lệnh dựa trên schema rồi sinh struct, tham số và method Go tương ứng. Ứng dụng vẫn kiểm soát trực tiếp SQL và cách tối ưu truy vấn, nhưng giảm code lặp để `Scan` dữ liệu và phát hiện nhiều lỗi sai kiểu ngay lúc build.

Cả GORM và `sqlc` đều có thể chống SQL injection khi giá trị đầu vào được truyền qua placeholder/bound parameter. Không được nối input người dùng vào chuỗi SQL. Những thành phần không thể parameter hóa như tên cột sắp xếp, hướng sắp xếp hoặc tên bảng phải dùng allowlist do server định nghĩa; tuyệt đối không truyền thẳng input người dùng vào `Order`, `Select`, `Group`, `Table`, raw SQL hoặc câu query động.

Scalar là giao diện hiển thị và thử API từ đặc tả OpenAPI, không thay thế bản thân file OpenAPI. Đặc tả nên được lưu và review cùng source code, sau đó API Go phục vụ cả file schema lẫn trang Scalar ở môi trường phù hợp.

Chỉ thêm dependency khi sprint thực sự cần. Không cần message queue trong MVP. Tác vụ nền nhỏ có thể chạy đồng bộ hoặc qua scheduled job có khóa trong PostgreSQL; nếu sau này cần gửi email/báo cáo quy mô lớn mới đánh giá MQ.

### 3.3. Thành phần hạ tầng

- PostgreSQL 16+.
- Docker Compose cho môi trường local: API + PostgreSQL.
- Object storage tương thích S3 cho ảnh; môi trường local có thể dùng MinIO hoặc file storage adapter.
- Reverse proxy/TLS do nền tảng triển khai đảm nhiệm.
- CI chạy format, lint, unit test, integration test, migration check và build image.

### 3.4. Quy ước API và bảo mật

- REST JSON dưới prefix `/api/v1`.
- UUID cho ID public; timestamp lưu UTC.
- Phân trang cursor hoặc page/limit nhất quán; filter/sort được whitelist.
- Password không bao giờ ghi log; refresh token được hash trong database và có cơ chế revoke.
- Rate limit cho login/register/advice; RBAC cho endpoint quản trị.
- Mọi thay đổi danh mục quan trọng có audit log.
- Không thu thập dữ liệu y tế không cần thiết; dữ liệu sinh học nhạy cảm phải có mục đích rõ, quyền xóa và thời hạn lưu.

## 4. Mô hình dữ liệu mức cao

Các bảng chính dự kiến:

- Người dùng: `users`, `user_profiles`, `roles`, `user_roles`, `refresh_tokens`.
- Bài tập: `muscle_groups`, `exercises`, `exercise_muscles`, `equipment`.
- Dinh dưỡng: `foods`, `food_servings`, `recipes`, `recipe_ingredients`, `recipe_steps`.
- Nội dung: `articles`, `article_tags`, `tags`.
- Chế độ: `programs`, `program_days`, `program_exercises`, `program_meals`, `user_programs`.
- Tracking: `workout_logs`, `workout_exercise_logs`, `meal_logs`, `meal_log_items`, `body_measurements`.
- Mục tiêu: `goals`, `goal_milestones`, `progress_snapshots`.
- Gamification: `point_events`, `achievements`, `user_achievements`, `leaderboard_snapshots`.
- Tư vấn: `advice_requests`, `advice_results`, `advice_rule_versions`.
- Quản trị: `audit_logs`.

Các trường dinh dưỡng nên lưu theo đơn vị chuẩn trên 100 g/ml và theo khẩu phần. Decimal được ưu tiên cho số đo/calo; không dùng floating point cho dữ liệu cần tổng hợp chính xác.

## 5. Kế hoạch phase và sprint

Giả định mỗi sprint kéo dài **2 tuần**, đội hình tối thiểu gồm 1 backend, 1 frontend/mobile, 1 QA kiêm phân tích nghiệp vụ và hỗ trợ part-time từ UI/UX hoặc DevOps. Với ít nhân sự hơn, giữ nguyên thứ tự nhưng tăng thời lượng.

### Phase 0 — Discovery và nền móng (Sprint 0, 1 tuần)

**Mục tiêu:** thống nhất phạm vi, luồng chính, tiêu chí dữ liệu và môi trường phát triển.

- Chốt persona, user journey và wireframe tối thiểu.
- Chốt công thức BMI/BMR/TDEE, disclaimer và rule escalation sang chuyên gia.
- Chốt glossary: chế độ, kế hoạch, buổi tập, khẩu phần, mục tiêu, điểm.
- Khởi tạo Go module, cấu trúc module, Docker Compose và PostgreSQL.
- Thiết lập migration, config, logging, health/readiness endpoint và CI.
- Viết OpenAPI skeleton, error format và coding convention.

**Đầu ra:** API chạy local, kết nối DB, migration hoạt động, CI xanh và backlog MVP được ưu tiên.

### Phase 1 — Nền tảng người dùng và danh mục (Sprint 1–2)

#### Sprint 1 — Authentication và hồ sơ

- Register/login/logout/refresh token.
- Hash mật khẩu Argon2id, rate limit, validation và audit sự kiện đăng nhập.
- CRUD hồ sơ và dữ liệu sinh học cơ bản.
- RBAC `user`/`admin` và middleware bảo vệ API.
- Unit/integration test cho auth và authorization.

**Demo:** người dùng tạo tài khoản, đăng nhập, cập nhật hồ sơ; admin truy cập được endpoint quản trị.

#### Sprint 2 — Bài tập, nhóm cơ và thực phẩm

- CRUD nhóm cơ, dụng cụ, bài tập và liên kết nhóm cơ.
- Danh sách, tìm kiếm, filter và phân trang bài tập.
- CRUD thực phẩm, khẩu phần, calo và macro.
- Seed dữ liệu mẫu có nguồn và trạng thái kiểm duyệt.
- Màn hình/API quản trị tối thiểu cho các danh mục.

**Demo:** admin quản lý dữ liệu; người dùng duyệt và lọc bài tập/thực phẩm.

### Phase 2 — Nội dung và chế độ mẫu (Sprint 3–4)

#### Sprint 3 — Công thức và bài viết hướng dẫn

- CRUD công thức, nguyên liệu, các bước chế biến và ảnh.
- Tính tổng calo/macro theo nguyên liệu và khẩu phần.
- CRUD bài viết sơ chế/chế biến, tag, draft/publish/hide.
- Danh sách, tìm kiếm và trang chi tiết nội dung.

**Demo:** admin xuất bản hướng dẫn; người dùng xem nguyên liệu, các bước và dinh dưỡng ước tính.

#### Sprint 4 — Chế độ và kế hoạch cá nhân

- CRUD mẫu chế độ tập/ăn/kết hợp.
- Lịch ngày/tuần, bài tập, bữa ăn, mục tiêu và độ khó.
- Người dùng chọn mẫu, đặt ngày bắt đầu và sao chép thành chế độ riêng.
- Cho phép chỉnh sửa nội dung riêng mà không làm thay đổi mẫu gốc.

**Demo:** người dùng chọn một chế độ mẫu và nhận lịch cá nhân có thể điều chỉnh.

### Phase 3 — Tracking và tiến độ (Sprint 5–6)

#### Sprint 5 — Nhật ký tập luyện và ăn uống

- Ghi nhận buổi tập, set/rep/tạ/thời gian và mức hoàn thành.
- Ghi nhận bữa ăn, khẩu phần, calo và macro.
- Dashboard ngày: kế hoạch so với thực tế.
- Quy tắc timezone, sửa/xóa log và audit tối thiểu.

**Demo:** người dùng hoàn thành lịch trong ngày và xem tổng tập luyện/dinh dưỡng.

#### Sprint 6 — Mục tiêu và tiến độ

- Tạo mục tiêu và milestone theo loại mục tiêu hỗ trợ.
- Ghi cân nặng/số đo và theo dõi lịch sử.
- Tổng hợp theo tuần, streak và tỷ lệ tuân thủ.
- Biểu đồ tiến độ; xử lý ngày thiếu dữ liệu rõ ràng, không nội suy gây hiểu nhầm.

**Demo:** người dùng thấy tiến độ mục tiêu và xu hướng từ dữ liệu đã ghi nhận.

### Phase 4 — Tư vấn và gamification (Sprint 7–8)

#### Sprint 7 — Tư vấn dựa trên rule

- Engine tính BMI/BMR/TDEE có version.
- Rule gợi ý mức calo/mức vận động và nội dung liên quan.
- Guardrail: phạm vi input, cảnh báo, từ chối câu hỏi bệnh lý và hướng dẫn tìm chuyên gia.
- Lưu lịch sử input/output/rule version và disclaimer.
- Kiểm thử các boundary case đã được chuyên gia nghiệp vụ duyệt.

**Demo:** hệ thống đưa ra gợi ý giải thích được, truy vết được và từ chối yêu cầu ngoài phạm vi.

#### Sprint 8 — Điểm, thành tích và xếp hạng

- Point event bất biến, rule chống cộng trùng và giới hạn theo ngày.
- Achievement, streak và bảng xếp hạng tuần/tháng.
- Cho phép opt-out khỏi bảng xếp hạng và dùng nickname.
- Không thưởng hành vi giảm cân cực đoan hoặc nhập thiếu dinh dưỡng.

**Demo:** người dùng nhận điểm/thành tích từ hành vi lành mạnh và xem xếp hạng an toàn.

### Phase 5 — Hardening và phát hành MVP (Sprint 9)

- Hoàn thiện luồng end-to-end và sửa lỗi ưu tiên cao.
- Security review: auth, IDOR, mass assignment, upload, rate limit và dữ liệu nhạy cảm.
- Kiểm thử tải cho login, danh mục, dashboard và tracking.
- Backup/restore PostgreSQL, migration rollback plan và seed production.
- Metrics, log correlation ID, alert cơ bản và runbook xử lý sự cố.
- Chính sách riêng tư, điều khoản sử dụng, disclaimer và luồng xóa tài khoản/dữ liệu.
- UAT, checklist release và deployment production/staging.

**Đầu ra:** phiên bản MVP có thể triển khai, quan sát, sao lưu và vận hành.

## 6. Thứ tự ưu tiên backlog

### Must have trước khi phát hành

- Authentication, authorization và hồ sơ.
- Danh mục bài tập/nhóm cơ, thực phẩm/calo, công thức và bài viết.
- Chế độ mẫu và chế độ cá nhân.
- Tracking tập luyện, ăn uống và số đo.
- Mục tiêu và tiến độ.
- Tư vấn rule-based có guardrail và disclaimer.
- Admin CRUD, audit cơ bản, backup và monitoring.

### Should have nếu còn năng lực

- Refresh token theo thiết bị và quản lý phiên đăng nhập.
- Achievement/streak và bảng xếp hạng opt-in.
- Import danh mục bằng CSV có validation.
- Yêu thích bài tập, món ăn và bài viết.
- Email verification/quên mật khẩu.

### Chưa làm trong MVP

- Chat AI tự do hoặc tư vấn bệnh lý.
- Kết nối wearable/HealthKit/Google Health Connect.
- Social feed, chat giữa người dùng hoặc marketplace huấn luyện viên.
- Realtime notification, microservice và message queue.
- Thanh toán/subscription.

## 7. Definition of Done

Một story chỉ hoàn thành khi:

- acceptance criteria được đáp ứng và có test phù hợp;
- API contract/OpenAPI và migration được cập nhật;
- validation, authorization, error handling và logging đã được kiểm tra;
- không lộ thông tin nhạy cảm trong response hoặc log;
- code review hoàn tất, CI xanh và không còn lỗi severity cao;
- thay đổi có thể demo trên môi trường staging;
- nội dung liên quan sức khỏe đã qua checklist về phạm vi tư vấn và disclaimer.

## 8. Tiêu chí MVP sẵn sàng phát hành

- Người dùng mới có thể đăng ký, thiết lập hồ sơ, chọn chế độ, ghi nhận ít nhất một buổi tập và một bữa ăn, sau đó xem tiến độ.
- Admin có thể quản lý đầy đủ các danh mục mà không cần sửa trực tiếp database.
- Dữ liệu dinh dưỡng hiển thị đơn vị, khẩu phần và nguồn/nhãn ước tính.
- Tư vấn luôn giải thích dữ liệu đầu vào, công thức/rule áp dụng và giới hạn chuyên môn.
- Các luồng auth và dữ liệu riêng tư vượt qua security checklist; backup/restore được diễn tập.
- P95 latency cho API đọc phổ biến đạt mục tiêu nội bộ đã thống nhất; error rate và saturation có dashboard.
- Không còn lỗi blocker/critical; lỗi high còn lại phải có quyết định chấp nhận rủi ro rõ ràng.

## 9. Rủi ro và quyết định cần chốt sớm

1. **Nguồn dữ liệu dinh dưỡng:** cần xác định nguồn hợp pháp, giấy phép sử dụng và quy trình cập nhật.
2. **Công thức tư vấn:** cần người có chuyên môn duyệt rule, ngưỡng an toàn và câu chữ disclaimer trước Sprint 7.
3. **Nền tảng client:** web responsive, mobile native hay cả hai ảnh hưởng lớn đến timeline; roadmap này chỉ cố định phần backend và luồng nghiệp vụ.
4. **Media storage:** chọn nhà cung cấp object storage trước Sprint 3.
5. **Xếp hạng:** phải opt-in và tránh cơ chế tạo động lực cho hành vi giảm cân/tập luyện quá mức.
6. **Quyền riêng tư:** chốt khu vực triển khai, yêu cầu pháp lý và thời hạn lưu dữ liệu trước production.

## 10. Mốc bàn giao dự kiến

- Hết Sprint 2: foundation + auth + danh mục lõi.
- Hết Sprint 4: người dùng chọn được chế độ và xem nội dung ăn uống/chế biến.
- Hết Sprint 6: vòng lặp chính “lập kế hoạch → tracking → xem tiến độ” hoạt động.
- Hết Sprint 8: tư vấn cơ bản, điểm và xếp hạng hoàn thiện.
- Hết Sprint 9: MVP đủ điều kiện phát hành có kiểm soát.

Tổng thời gian tham khảo: **19 tuần** gồm Sprint 0 một tuần và 9 sprint hai tuần. Timeline cần điều chỉnh theo số lượng client, năng lực đội ngũ và mức độ sẵn sàng của dữ liệu nội dung.
