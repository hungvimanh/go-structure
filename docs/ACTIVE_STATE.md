# Active State

Trạng thái làm việc hiện tại của `category-service`. Cập nhật file này mỗi khi có thay đổi quy tắc hoặc tiến độ đáng kể — không phải changelog chi tiết từng dòng code.

## Quy tắc làm việc

- **Không refactor lớn.** Sửa từng phần nhỏ theo yêu cầu cụ thể, để người dùng kiểm soát dần từng bước. Không tự ý mở rộng phạm vi ngoài những gì được yêu cầu.
- **Xác nhận trước khi làm bước tiếp theo.** Sau mỗi thay đổi, báo lại kết quả (build/vet/test) và chờ "ok" / yêu cầu kế tiếp thay vì tự chạy tiếp một loạt thay đổi.
- **Không thêm test cố định vào repo.** Khi cần xác minh hành vi, viết file `*_test.go` tạm, chạy `go test`, sau đó xóa — trừ khi được yêu cầu giữ lại test.
- Khi field name lặp lại giữa các lớp (model, validator, error), dùng hằng số định nghĩa 1 lần trong `model` (`Category_ID`, `Category_Code`, `Category_Name`, `Category_Description`), không hard-code string.

## Kiến trúc phân lớp lỗi

Dự án tách lỗi thành 3 loại độc lập, mỗi loại do đúng 1 lớp chịu trách nhiệm:

| Loại lỗi | Ai phát hiện | Ai không được check | HTTP status | Cơ chế |
|---|---|---|---|---|
| **Binding / syntax** | Handler | — | 421 | Request id không parse được (`uuid.Parse`), JSON body decode lỗi. Dùng chung helper `writeBindingError`. |
| **Business** | Service / Validator | Handler (không check syntax), Repository (không check nghiệp vụ) | 400 (field validation), 404 (not found) | `validation.Errors` (`internal/shared/validation`) — field-level, gom hết lỗi rồi mới trả, dịch qua i18n catalog. |
| **Technical / infra** | Repository | Service, Handler (không tự phán đoán loại lỗi) | 500 | `apperror` (`internal/shared/apperror`) — sentinel errors (`ErrBadRequest`, `ErrInternal`), message chung chung trả về client, raw error log server-side qua `log.Printf`, không lộ ra ngoài. |
| **Authentication** | Middleware (`internal/auth`) | Handler/Service/Repository (không tự verify token) | 401 | Chạy trước handler, ghi response trực tiếp qua `httpresponse.WriteError` khi thiếu/sai token — không đi qua `respondError`. Message trả về client luôn chung chung `"unauthorized"`, chi tiết lỗi log server-side qua `log.Printf`, không phân biệt "thiếu token" / "sai chữ ký" / "hết hạn" ra ngoài (tránh lộ thông tin). |

Nguyên tắc cốt lõi: **"không có row" không phải lỗi kỹ thuật** — repository trả `(nil, nil)` khi query không có kết quả; service tự diễn giải `nil` thành business error (`validator.NotFound` = `CATEGORY_NOT_FOUND`), không phải lỗi hạ tầng.

### Validation — gom lỗi, không fail-fast

`validator.Create` / `validator.Update` chạy **hết** `validateCode` + `validateName` + `validateDescription` trước khi trả lỗi, kể cả khi bước kiểm tra DB (`checker.ExistsByCode`) bị lỗi giữa chừng — lỗi field luôn được ưu tiên trả trước lỗi DB (`dbErr` chỉ trả về khi `errs == nil`). Trong `validateCode`, DB uniqueness check chỉ chạy sau khi input hợp lệ (không rỗng, không quá dài) — required/too-long dừng sớm vì response format là dictionary `field -> message` (tối đa 1 lỗi/field).

`Create` và `Update` dùng chung `validateCode`/`validateName`/`validateDescription`; phân biệt qua `excludeID` (`uuid.Nil` cho Create = không loại trừ gì, `id` thật cho Update = loại trừ chính nó khỏi check trùng).

### i18n

- Catalog JSON co-located theo module: `<module>_validator.i18n.json`, cấu trúc 3 lớp `error_code -> lang -> template`.
- Load 1 lần lúc khởi động qua `i18n.LoadCatalog(root)` (`filepath.WalkDir`, trùng code giữa các file là lỗi load-time).
- Locale lấy từ header `Accept-Language` chuẩn HTTP, parser tự viết (`internal/i18n/locale.go`), không dùng thư viện ngoài.
- Response validation error: `map[string]string` (field -> message đã dịch), không phải mảng object — do mỗi field tối đa 1 lỗi (early-return trong từng hàm validate field).
- Message có tham số dùng `%d`/`%s` + `fe.Params []any`, dịch qua `fmt.Sprintf`.

## Pagination

- `internal/shared/pagination`: package dùng chung để module khác tái dùng, khác convention của `i18n.ParseAcceptLanguage` (nhận string thô) — `pagination.Parse` nhận thẳng `*http.Request` vì việc đọc `skip`/`take` gắn với query string.
- `pagination.Parse(r *http.Request)`: parse `skip` và `take` **độc lập** (không fail-fast ở field đầu) rồi mới kiểm tra lỗi — chỉ khi cả hai không lỗi mới gán giá trị/default. Default `skip=0`, `take=10` (`DefaultSkip`, `DefaultTake`). `take` bị chặn trần `MaxTake=100` (âm thầm hạ xuống, không lỗi). Giá trị không parse được thành số hoặc âm (ở field nào cũng vậy) → `pagination.ErrInvalid`.
- Handler (`CategoryHandler.List`) gọi `pagination.Parse(r)`; lỗi parse trả **421** qua `writeBindingError` — coi ngang hàng với lỗi cú pháp khác (uuid.Parse, JSON decode), không phải business error.
- Luồng xuống: `handler` parse → truyền `pagination.Params` qua `service.List` → `repository.List` dùng làm `LIMIT $1 OFFSET $2` trong SQL.

## Authentication

- `internal/auth`: middleware chi-compatible verify access token (JWT RS256).
- **Route public vs protected**: KHÔNG dùng `router.Use(...)` ở root router nữa (chi bắt buộc mọi `Use` phải khai báo trước route trên cùng 1 router, nên không thể vừa có route public vừa protected trên root nếu áp `Use` global). Thay vào đó dùng `router.Group(func(r chi.Router) {...})` bọc riêng phần route cần token — `r.Use(auth.Middleware(...))` chỉ scope trong group đó, không lan ra route đăng ký ngoài group. Route public đăng ký thẳng trên `router`, route protected đăng ký trong `Group`.
- Ví dụ đang có trong `main.go`: `GET /categories/sample` (handler `CategoryHandler.Sample`, dữ liệu tĩnh, không qua service/repository) — route **public**, minh họa cách 1 endpoint trong cùng handler với các endpoint khác vẫn có thể public dù phần còn lại của handler đang bị bảo vệ. Route này chỉ để demo pattern, không phải API nghiệp vụ thật.
- Chi tự ưu tiên route tĩnh (`/categories/sample`) trước route có param (`/categories/{id}`) dù đăng ký ở router/group khác nhau — không bị conflict. Đã verify bằng test tạm (routing static-vs-param + middleware scope), xóa sau khi verify theo quy tắc.
- Chỉ **verify** — chưa có login/issue token/refresh token/authorization theo role hay permission. Middleware chỉ trả lời "token này có hợp lệ và đúng do hệ thống ký hay không", chưa quyết định request có được phép thực hiện hành động cụ thể hay không.
- Cơ chế verify: đọc header `Authorization: Bearer <token>` → `jwt.ParseWithClaims` với RSA public key, ép cứng signing method phải là RSA (chặn alg-confusion attack). Expiry (`exp`/`nbf`/`iat`) được thư viện `golang-jwt/v5` tự validate bên trong `ParseWithClaims`, không có code check riêng — hết hạn hay sai chữ ký đều gộp chung thành `ErrInvalidToken` → 401 `"unauthorized"` (không phân biệt lý do ra ngoài).
- `auth.Claims` (`internal/auth/claims.go`) là struct tùy biến nhúng `jwt.RegisteredClaims`, hiện có thêm `UserID`, `Email`, `FullName` (map theo JSON key `user_id`/`email`/`full_name`) — field giả định theo yêu cầu, **chưa xác nhận với payload token thật**. Đây là **điểm nối 1**: JWT có thêm field mới thì bổ sung field + json tag vào struct này trước.
- **`userContext` layer** (`internal/shared/usercontext`): sau khi verify, middleware map `*Claims` → `*usercontext.UserContext` qua hàm `auth.newUserContext` (**điểm nối 2** — Claims có field mới muốn lộ ra ngoài thì map thêm vào đây), rồi lưu vào `context` qua `usercontext.WithUser`. `UserContext` là struct nghiệp vụ độc lập với JWT/RSA — service/repository chỉ cần `usercontext.FromContext(ctx)`, không import `internal/auth`. Vì `context.Context` đã truyền xuyên suốt mọi layer sẵn (`r.Context()` → `service.X(ctx, ...)` → `repository.X(ctx, ...)`), lấy được `UserContext` ở bất kỳ layer nào mà không cần đổi signature hàm nào. `UserContext` cũng là chỗ mở để bổ sung dữ liệu khác về request sau này (không nhất thiết chỉ từ JWT).
- Đã xóa `auth.ClaimsFromContext`/`withClaims` (bản nháp cũ, chưa ai dùng) — thay thế hoàn toàn bởi `usercontext` ở trên.
- `internal/shared/httpresponse` (mới, tách từ `category/handler`): `WriteJSON`/`WriteError` dùng chung giữa `category/handler` và `auth` middleware — tránh lặp hàm ghi JSON response ở lớp thứ 2.
- Config: `JWT_PUBLIC_KEY` bắt buộc để middleware chạy được (fail-fast ở `main.go` qua `auth.LoadPublicKey`, không phải ở `config.Load()` vì lúc thêm biến giá trị còn để trống). `JWT_PRIVATE_KEY` đã có trong `Config`/`.env` nhưng **chưa được dùng ở đâu** — chỉ chuẩn bị sẵn cho tính năng issue token sau này.
- `.env`/`.env.example`: giá trị PEM đặt trên 1 dòng, xuống dòng bằng `\n`, bọc trong dấu ngoặc kép (godotenv tự chuyển `\n` thành newline thật lúc load — không cần xử lý escape thủ công trong code).

## Tiến độ

### Đã xong
- `internal/shared/validation`: kiểu `FieldError`/`Errors` dùng chung, generic, không phụ thuộc module cụ thể.
- `internal/shared/apperror`: sentinel errors kỹ thuật (`ErrBadRequest`, `ErrInternal`).
- `internal/i18n`: locale parser, catalog loader, translator.
- `internal/category/validator`: validate đầy đủ Code/Name/Description, check trùng code qua `CodeChecker` interface (không phụ thuộc trực tiếp `repository`), business error `CATEGORY_NOT_FOUND` cho case not-found.
- `internal/category/repository`: `Get` trả `(nil, nil)` khi không có row thay vì lỗi; `ExistsByCode` hỗ trợ loại trừ id (Update).
- `internal/category/service`: `Get`/`Update` phát hiện `category == nil` → business NotFound error.
- `internal/category/handler`: `respondError` phân loại lỗi (`validation.Errors` → 400/404 tùy code, còn lại qua `classifyError` → 400/500), binding error → 421 qua `writeBindingError`.
- `cmd/api/main.go`: DI thủ công (`repository.NewCategoryRepository` → `service.NewCategoryService` → `handler.NewCategoryHandler`), `chi.NewRouter()` với route group `/categories` (`GET /`, `POST /`, `GET /{id}`, `PUT /{id}`, `DELETE /{id}`), chạy qua `http.ListenAndServe`. `Count` chưa có route vì handler chưa có method tương ứng.
- `internal/auth`: middleware verify access token JWT RS256. Xem chi tiết ở mục Authentication phía trên.
- `internal/shared/usercontext`: layer `UserContext` tách khỏi JWT, truyền qua `context.Context` xuống mọi layer. Xem mục Authentication phía trên.

### Chưa làm (không tự ý làm nếu chưa được yêu cầu)
- `repository.Delete` dùng `Exec` không check số row bị ảnh hưởng — xóa id không tồn tại vẫn trả 204 âm thầm, chưa có check not-found như `Get`/`Update`.
- Các sentinel error khác (403/415/...) chưa định nghĩa trong `apperror`, sẽ thêm khi có nhu cầu thực tế (content-type check, v.v.). Riêng 401 đã có (xem mục Authentication).
- **Login / issue token / refresh token**: chưa làm, chỉ mới verify token do bên khác cấp.
- **Authorization theo policy (role/permission)**: middleware hiện tại mới trả lời "token hợp lệ hay không", chưa quyết định request có được phép thực hiện hành động cụ thể hay không (ví dụ: role nào được xóa category). Cần thiết kế riêng khi có yêu cầu.
- `userContext` layer đã có (xem mục Authentication) nhưng **chưa có handler/service/repository nào thực sự gọi `usercontext.FromContext`** — mới dừng ở middleware ghi vào context, chưa có nhu cầu nghiệp vụ đọc ra (ví dụ audit `created_by`). Sẽ dùng khi có yêu cầu cụ thể.
