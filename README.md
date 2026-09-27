# Go Backend Starter — Category Service

Khung backend Go theo hướng module hóa, dùng Echo v5, Uber Fx và PostgreSQL/pgx. Repository hiện có một module mẫu `category`, một query AST an toàn, generic repository, migration runner, JWT authentication, validation và i18n.

> **Trạng thái:** phù hợp để làm nền cho giai đoạn phát triển, prototype hoặc dự án nội bộ. Repository **chưa sẵn sàng để triển khai production** nếu chưa bổ sung các hạng mục trong [Production readiness](#production-readiness).

## Điểm chính

- Composition root và lifecycle bằng `go.uber.org/fx`.
- HTTP API bằng Echo v5 và `net/http.Server` với timeout, panic recovery, request timeout và graceful shutdown.
- PostgreSQL bằng `pgx/v5` và `pgxpool`.
- Generic repository hỗ trợ count, exists, list, get, create, update và soft/hard delete.
- Query AST đóng, dùng parameterized SQL và giới hạn độ phức tạp của filter.
- Filter có kiểu dữ liệu, search nhiều trường, sort ổn định và pagination.
- JWT bearer authentication dùng RSA public key.
- Validation theo field, gom nhiều lỗi và dịch message theo `Accept-Language`.
- Forward-only migration được embed vào binary, chạy trong transaction và có PostgreSQL advisory lock.
- Module mẫu `category` theo luồng handler → service → repository.

## Công nghệ

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ | Go 1.26.3 |
| HTTP | Echo v5 |
| Dependency injection | Uber Fx |
| Database | PostgreSQL, pgx v5 |
| Authentication | JWT RS256 |
| ID | UUID v7 |
| Configuration | Environment variables, `godotenv` cho local |

## Kiến trúc

```text
HTTP request
    │
    ▼
Echo router / middleware
    │
    ▼
Handler ──► Service ──► Repository adapter
                              │
                              ▼
                    Generic repository
                              │
                              ▼
                     Query AST compiler
                              │
                              ▼
                         PostgreSQL
```

Mỗi business module tự đăng ký dependency và route bằng `fx.Module`. `cmd/api` chỉ giữ composition root và lifecycle hạ tầng.

```text
.
├── cmd/
│   ├── api/                         # API entrypoint, Fx wiring, HTTP lifecycle
│   └── migrate/                     # Migration command
├── internal/
│   ├── auth/                        # JWT claims, verify và middleware
│   ├── category/                    # Module mẫu theo từng layer
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── service/
│   │   └── validator/
│   ├── config/                      # Environment configuration
│   ├── database/                    # pgxpool và migration runner
│   ├── i18n/                        # Catalog và translator
│   └── shared/
│       ├── apperror/
│       ├── filter/
│       ├── httpresponse/
│       ├── pagination/
│       ├── query/                   # Query AST và SQL compiler
│       ├── repository/              # Generic repository và transaction helpers
│       ├── reqtimeout/
│       ├── usercontext/
│       └── validation/
└── docs/                            # Thiết kế, tiêu chuẩn và kế hoạch kỹ thuật
```

## Yêu cầu

- Go theo phiên bản trong `go.mod`.
- PostgreSQL.
- RSA public key ở định dạng PEM để xác thực access token.
- Chạy lệnh từ repository root. Catalog i18n hiện được đọc từ đường dẫn filesystem tương đối.

## Khởi chạy local

### 1. Tạo file cấu hình

Linux/macOS/Git Bash:

```bash
cp .env.example .env
```

PowerShell:

```powershell
Copy-Item .env.example .env
```

Cập nhật tối thiểu các biến sau:

```dotenv
DB_PASSWORD=your-postgres-password
JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n"
```

`JWT_PUBLIC_KEY` là **nội dung PEM**, không phải đường dẫn tới file. Xem chú thích trong `.env.example` để biết cách biểu diễn newline bằng `\n`.

### 2. Tải dependency

```bash
go mod download
```

### 3. Chạy migration

```bash
go run ./cmd/migrate
```

API không tự chạy migration khi startup. Migration là một bước triển khai riêng biệt.

### 4. Chạy API

```bash
go run ./cmd/api
```

Mặc định API lắng nghe tại:

```text
http://localhost:5000
```

Kiểm tra public endpoint:

```bash
curl http://localhost:5000/categories/sample
```

## Cấu hình

| Biến | Bắt buộc | Mặc định | Mô tả |
|---|---:|---|---|
| `APP_PORT` | Không | `5000` | Port của HTTP server. |
| `DB_HOST` | Không | `localhost` | PostgreSQL host. |
| `DB_PORT` | Không | `5432` | PostgreSQL port. |
| `DB_USER` | Không | `postgres` | PostgreSQL user. |
| `DB_PASSWORD` | **Có** | — | PostgreSQL password; app fail-fast nếu để trống. |
| `DB_NAME` | Không | `postgres` | Database name. |
| `DB_SSLMODE` | Không | `disable` | pgx SSL mode. Production phải đặt giá trị phù hợp như `require` hoặc `verify-full`. |
| `I18N_DIR` | Không | `internal` | Thư mục gốc được quét để tìm `*.i18n.json`. |
| `JWT_PUBLIC_KEY` | **Có để API startup** | rỗng | Nội dung RSA public key PEM dùng để verify token. |
| `JWT_PRIVATE_KEY` | Chưa dùng | rỗng | Được giữ cho chức năng phát hành token trong tương lai. |
| `REQUEST_TIMEOUT_SECONDS` | Không | `10` | Timeout xử lý mỗi request. |
| `DEV_MODE` | Không | `0` | `1` cho phép trả technical error ra response; không bật ở production. |

## Authentication

Ngoại trừ endpoint sample, các route `category` yêu cầu header:

```http
Authorization: Bearer <access-token>
```

Token hiện được verify bằng RSA public key. Custom claims được hỗ trợ:

```json
{
  "user_id": "user-id",
  "email": "user@example.com",
  "full_name": "Example User"
}
```

Repository hiện chỉ **xác thực** token. Role-based access control (RBAC), attribute-based access control (ABAC), ownership check, token issuance, refresh và revocation chưa được triển khai.

## API mẫu

> Do cách route hiện tại được đăng ký, collection route dùng dấu `/` cuối: `/categories/`.

| Method | Path | Auth | Mô tả |
|---|---|---:|---|
| `GET` | `/categories/sample` | Không | Trả một category mẫu, không truy cập database. |
| `GET` | `/categories/?skip=0&take=10&search=cat` | Có | Danh sách và tìm kiếm đơn giản. |
| `POST` | `/categories/search` | Có | Danh sách với typed filter trong JSON body. |
| `POST` | `/categories/` | Có | Tạo category. |
| `GET` | `/categories/:id` | Có | Lấy category theo UUID. |
| `PUT` | `/categories/:id` | Có | Cập nhật category. |
| `DELETE` | `/categories/:id` | Có | Soft-delete category. |

### Tạo category

```bash
curl -X POST http://localhost:5000/categories/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "BOOK",
    "name": "Books",
    "description": "Book category",
    "status": 1
  }'
```

### Typed filter

```bash
curl -X POST http://localhost:5000/categories/search \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "skip": 0,
    "take": 20,
    "filter": {
      "code": { "startsWith": "BO" },
      "status": { "eq": 1 }
    }
  }'
```

Các toán tử chính:

- String: `eq`, `neq`, `contains`, `notContains`, `startsWith`, `notStartsWith`, `endsWith`, `notEndsWith`, `in`, `notIn` và các biến thể reverse/combine.
- Number: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `notIn`.
- Pagination: `skip` mặc định `0`, `take` mặc định `10`, tối đa `100`.

Query compiler giới hạn mặc định độ sâu, số predicate và tổng số phần tử trong set filter trước khi chạy SQL.

## Error response và i18n

Validation error được trả theo dạng:

```json
{
  "errors": {
    "code": "...",
    "name": "..."
  }
}
```

Ngôn ngữ được chọn từ `Accept-Language`. Technical error được che bằng message chung khi `DEV_MODE=0`.

## Migration

Migration nằm tại:

```text
internal/database/migration/migrations
```

Quy ước:

1. Tạo file forward-only theo số thứ tự, ví dụ `000002_add_category_slug.up.sql`.
2. Không sửa migration đã được áp dụng ở môi trường dùng chung.
3. Chạy `go run ./cmd/migrate` như một bước deployment riêng.
4. Dùng migration mới để sửa schema hoặc phục hồi từ backup; project không cung cấp automatic down migration.

Migration runner:

- Embed SQL vào migration binary.
- Tạo bảng `schema_migrations` nếu cần.
- Sắp xếp migration theo tên file.
- Dùng PostgreSQL advisory transaction lock để tránh chạy đồng thời.
- Commit schema change và version record trong cùng transaction.

Xem thêm: [`docs/MIGRATIONS.md`](docs/MIGRATIONS.md).

## Test

Chạy toàn bộ test hiện có:

```bash
go test ./...
```

Hiện repository chỉ có test kiểm tra Fx dependency graph tại `cmd/api/main_test.go`. Chưa có test suite cho query compiler, repository, migration, auth, handler, validation hoặc PostgreSQL integration. Đây là một blocker trước production.

## Thêm business module mới

Một module mới nên giữ cùng boundary với `category`:

1. Định nghĩa model và request DTO.
2. Khai báo static repository specification: table, columns, projection, sort và soft-delete.
3. Viết repository adapter trên generic repository.
4. Viết service chứa business rule và transaction boundary.
5. Viết validator độc lập với HTTP framework.
6. Viết Echo handler và route registration.
7. Gom dependency bằng `fx.Module`.
8. Thêm module vào `appOptions` trong `cmd/api/main.go`.
9. Thêm migration, unit test và integration test.

Không truyền raw SQL fragment, column name hoặc sort key từ request vào query. Hãy ánh xạ input qua registry tĩnh của module.

## Production readiness

### Đánh giá hiện tại

| Khu vực | Trạng thái | Ghi chú |
|---|---|---|
| Core architecture | Tốt | Boundary rõ; query AST, generic repository, lifecycle và migration có thiết kế tốt. |
| Data integrity | Chưa đạt | `category.code` chưa có unique constraint; check trùng ở application có race condition. |
| Security | Chưa đạt | JWT chưa bắt buộc issuer/audience/expiration; chưa có body limit, rate limit và authorization. |
| Observability | Chưa có | Chỉ dùng stdlib `log`; chưa có request ID, access log, metrics hoặc tracing. |
| Health checks | Chưa có | Chưa có liveness/readiness endpoint. |
| Testing | Chưa đạt | Mới có một Fx graph test. |
| Delivery | Chưa có | Chưa có Dockerfile, Compose, CI/CD, linter config hoặc deployment manifest. |
| API documentation | Một phần | README mô tả API mẫu; chưa có OpenAPI specification. |

### Blocker cần xử lý trước production

1. Thêm unique constraint phù hợp cho `category.code`; map lỗi constraint về domain error và dùng transaction cho các write workflow nhiều bước.
2. Embed i18n catalog hoặc đóng gói catalog bằng đường dẫn tuyệt đối, ổn định trong image.
3. Thêm `/healthz` và `/readyz`; readiness phải kiểm tra dependency cần thiết như PostgreSQL.
4. Thêm structured logging, request ID, access log, metrics và tracing.
5. Siết JWT bằng valid algorithm, expiration bắt buộc, issuer, audience và chiến lược key rotation.
6. Thêm body size limit, rate limit và authorization policy.
7. Thay `log.Fatalf` trong server goroutine bằng cơ chế trả lỗi để Fx thực hiện shutdown lifecycle.
8. Viết unit/integration test cho query compiler, repository, validation, auth, HTTP contract và migration.
9. Thêm Dockerfile multi-stage, CI pipeline, lint, vulnerability scan và artifact build có thể tái lập.
10. Cấu hình database pool theo môi trường và bắt buộc TLS ở production.

### API behavior cần chốt sớm

Một số hành vi hiện tại nên được xem là chưa ổn định:

- Binding error đang dùng HTTP `421 Misdirected Request` thay vì `400 Bad Request`.
- Category không tồn tại đang đi qua validation pipeline và trả `400` thay vì `404`.
- Delete ID không tồn tại vẫn có thể trả `204`.
- List rỗng có thể serialize thành `null` thay vì `[]` và chưa có `total`.
- Collection route hiện yêu cầu dấu `/` cuối.

Nên chuẩn hóa các contract này trước khi có consumer thật để tránh breaking change về sau.

## Tài liệu liên quan

- [`docs/MIGRATIONS.md`](docs/MIGRATIONS.md): cách vận hành migration.
- [`docs/SECURITY_STANDARDS.md`](docs/SECURITY_STANDARDS.md): tiêu chuẩn security của project.
- [`docs/PERFORMANCE_STANDARDS.md`](docs/PERFORMANCE_STANDARDS.md): tiêu chuẩn performance.
- [`docs/DESIGN_STANDARDS.md`](docs/DESIGN_STANDARDS.md): tiêu chuẩn thiết kế.
- [`docs/ACTIVE_STATE.md`](docs/ACTIVE_STATE.md): trạng thái công việc kỹ thuật hiện tại.
