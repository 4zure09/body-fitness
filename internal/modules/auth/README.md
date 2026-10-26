# Auth module

Sở hữu đăng ký, đăng nhập, refresh/revoke token và authorization policy. Password hasher và token signer được khai báo qua port; HTTP handler không truy cập database trực tiếp.

- Mật khẩu dùng Argon2id với salt ngẫu nhiên.
- Access token và refresh token dùng JWT HS256, có `iss`, `aud`, `sub`, `jti`, `iat`, `nbf`, `exp` và `token_use`.
- Refresh token chỉ được lưu dưới dạng SHA-256 fingerprint, được rotate một lần và có thể revoke.
- Tài khoản mới luôn nhận role `user`.
- Role `admin` được bootstrap bằng management command `go run ./cmd/admin --email ...`.

