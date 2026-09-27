# Active State

Trạng thái làm việc hiện tại của `category-service`. Cập nhật file này mỗi khi có thay đổi quy tắc hoặc tiến độ đáng kể — không phải changelog chi tiết từng dòng code.

## Quy tắc làm việc

- **Không refactor lớn.** Sửa từng phần nhỏ theo yêu cầu cụ thể, để người dùng kiểm soát dần từng bước. Không tự ý mở rộng phạm vi ngoài những gì được yêu cầu.
- **Xác nhận trước khi làm bước tiếp theo.** Sau mỗi thay đổi, báo lại kết quả (build/vet/test) và chờ "ok" / yêu cầu kế tiếp thay vì tự chạy tiếp một loạt thay đổi.
- **Không thêm test cố định vào repo.** Khi cần xác minh hành vi, viết file `*_test.go` tạm, chạy `go test`, sau đó xóa — trừ khi được yêu cầu giữ lại test.
- Khi field name lặp lại giữa các lớp (model, validator, error), dùng hằng số định nghĩa 1 lần trong `model` (`Category_ID`, `Category_Code`, `Category_Name`, `Category_Description`), không hard-code string.

## Framework boundary

Dự án dùng Echo v5 tại HTTP transport/composition layer. Echo chỉ được import tại:

- `cmd/api`;
- HTTP handler và route module;
- `internal/auth/middleware.go`;
- `internal/shared/reqtimeout`;
- `internal/shared/httpresponse`.

Không đưa Echo xuống service, repository, model, validator, database, config, i18n hoặc shared business primitives. `*echo.Echo` vẫn là `http.Handler`, nên lifecycle `http.Server`, Fx và graceful shutdown không phụ thuộc framework router.

## Kiến trúc phân lớp lỗi

Dự án tách lỗi thành 4 loại độc lập, mỗi loại do đúng 1 lớp chịu trách nhiệm:

| Loại lỗi | Ai phát hiện | Ai không được check | HTTP status | Cơ chế |
|---|---|---|---|---|
| **Binding / syntax** | Handler | — | 421 | Handler trả `httpresponse.NewBindingError(message, cause)`. `ErrorResponder.RespondError` ghi JSON qua `c.JSON`. |
| **Business** | Service / Validator | Handler (không check syntax), Repository (không check nghiệp vụ) | 400 | `validation.Errors` — field-level, gom hết lỗi rồi mới trả, dịch qua i18n catalog. `CATEGORY_NOT_FOUND` cũng là 400. |
| **Authentication** | Middleware (`internal/auth`) | Handler/Service/Repository | 401 | Middleware trả `*httpresponse.AuthenticationError`, giữ cause để log server-side; centralized handler luôn trả `{"error":"unauthorized"}`. |
| **Technical / infra** | Repository | Service, Handler | 400 (`ErrBadRequest`) / 500 (còn lại) / 504 (deadline) | `apperror` sentinel errors; raw error chỉ log server-side, trừ technical error khi `DEV_MODE=1`. |

### Error handling — Echo `HTTPErrorHandler`

`(*httpresponse.ErrorResponder).HTTPErrorHandler` là điểm vào duy nhất cho mọi error trả về từ Echo handler, custom middleware và Echo Recover. Nó kiểm tra `c.Response().Committed` trước khi ghi để tránh double-write.

- Handler có chữ ký `func(c *echo.Context) error`: chỉ bind transport input, gọi service và ghi success response qua `c.JSON`/`c.NoContent`. Khi lỗi, handler chỉ `return err`; không tự gọi responder, log hay phân loại error.
- `RespondError` phân loại theo thứ tự `validation.Errors` (400) → `*BindingError` (421) → `*AuthenticationError` (401 generic) → technical error. `context.DeadlineExceeded` vẫn map 504.
- `AuthenticationError` không chịu ảnh hưởng của `DEV_MODE`: cause chỉ được log server-side, không được lộ qua response.
- Lỗi technical 500/504 tiếp tục log tại `(*ErrorResponder).logError` — đây là một điểm thay đổi duy nhất nếu sau này thay log storage/queue.
- Router/framework error 404 và 405 không đi qua application classifier: handler delegate sang `echo.DefaultHTTPErrorHandler(false)`. Đây chỉ là fallback khi backend bị gọi trực tiếp; trong deployment bình thường gateway sở hữu contract 404.
- Resource nghiệp vụ không tồn tại không dùng Echo 404. Service trả `validation.Errors` với `CATEGORY_NOT_FOUND`, nên client nhận 400 theo contract.

### DEV_MODE

- Env var `DEV_MODE` (`"0"`/`"1"`, mặc định `"0"`) được parse trong `config.Load()` thành `Config.DevMode bool`.
- Chỉ ảnh hưởng technical error trong `ErrorResponder.RespondError`: khi `DevMode=true`, message là `err.Error()` thay vì message chung chung. Status không đổi.
- Không ảnh hưởng `validation.Errors`, binding/syntax errors hoặc authentication. 401 luôn là `{"error":"unauthorized"}`.
- Không bật `DEV_MODE=1` ở production vì technical error có thể lộ chi tiết hạ tầng.

## Pagination

- `internal/shared/pagination` dùng chung cho các module. `pagination.Parse(r *http.Request)` đọc `skip`/`take` độc lập, default `skip=0`, `take=10`, và giới hạn `take` tại `MaxTake=100`.
- Giá trị không parse được hoặc âm trả `pagination.ErrInvalid`; `CategoryHandler.List` bọc thành `NewBindingError`, nên trả 421.
- Luồng xuống: handler parse → `service.List` → `repository.List` dùng `LIMIT`/`OFFSET`.

## Server lifecycle

`cmd/api` vẫn chạy `http.Server` với `ReadHeaderTimeout=5s`, `ReadTimeout=15s`, `IdleTimeout=90s` và `WriteTimeout=cfg.RequestTimeout+5s`. Fx hook vẫn dùng `net.Listen`, `server.Serve` và `server.Shutdown`; database chỉ đóng sau khi request đang chạy được drain hoặc shutdown timeout hết hạn.

## Panic recovery và request timeout

- `middleware.Recover()` của Echo là middleware ngoài cùng, trước custom timeout. Panic được Echo trả về centralized `HTTPErrorHandler`; error technical vẫn log và map 500, server tiếp tục phục vụ request sau đó.
- `reqtimeout.Middleware` là `echo.MiddlewareFunc`: đặt deadline trên request context bằng `c.SetRequest(c.Request().WithContext(ctx))`, rồi `return next(c)`.
- Timeout là **cooperative cancellation**: các layer dưới phải tôn trọng `context.Context`; middleware không cưỡng chế dừng code bỏ qua context.
- Không dùng `middleware.ContextTimeout`, vì default status 503 của nó phá vỡ contract 504 hiện có.

## Module wiring và routing

- Category module expose Fx `Module`, cung cấp repository → service → handler và gọi `registerRoutes` qua `fx.Invoke`.
- `registerRoutes` nhận `*echo.Echo` cùng `echo.MiddlewareFunc` auth. `GET /categories/sample` đăng ký trực tiếp và public.
- Các route category còn lại nằm trong `router.Group("/categories", authMiddleware)`: `GET /`, `POST /search`, `POST /`, `GET /:id`, `PUT /:id`, `DELETE /:id`.
- Echo ưu tiên static route `/categories/sample` trước parameter route `/:id`; scope auth không lan vào sample route.

## Authentication

- Middleware verify JWT RS256 từ `Authorization: Bearer <token>`, ép signing method RSA và map `Claims` thành `usercontext.UserContext` trong request context.
- Thiếu token, token hết hạn hoặc sai chữ ký đều trả `AuthenticationError` về centralized pipeline. Client không phân biệt nguyên nhân và luôn nhận generic 401; server log cause.
- Chưa có login, issue/refresh token hoặc authorization theo role/permission.

## Shared persistence/query boundary

`internal/shared/query` và `internal/shared/repository` là framework PostgreSQL/pgx nội bộ. Module mới phải tự sở hữu descriptor tĩnh cho table/column, mapping filter sang `query.Expression`, registry search/sort/projection, scanner chính xác theo projection, extractor giá trị insert/update và predicate nghiệp vụ. Giá trị luôn được bind parameter; identifier, projection và sort key chỉ lấy từ descriptor, không lấy raw SQL từ request.

- String predicate dùng literal, case-insensitive semantics. Nhiều operator active trên cùng field được nối `AND`; `combine` là nhóm `OR` nội bộ; scalar string rỗng/whitespace là no-op. `IN []` là false, `NOT IN []` là no-op; set được deduplicate trước khi bind PostgreSQL array.
- Search rỗng là no-op. Khi có search term nhưng không chọn field, compiler dùng toàn bộ searchable field đã cấu hình. Sort phải có ID tie-breaker xác định; sort/projection key không có trong registry trả query error trước DB.
- `EntitySpec`/`Projection` giữ explicit column + scanner, không reflection hoặc struct tag. Generic engine phù hợp Count/Exists/List/Get/Create/Update/Delete; query aggregate/analytics không vừa contract vẫn do module viết riêng bằng SQL descriptor-owned.
- Relation replace-all nhận `RelationPatch`: absent giữ nguyên, present empty clear, present values replace sau deduplicate. Parent mutation và replacement phải cùng nằm trong `Transactor.WithinTx`; helper không tự mở transaction.
- Category là pilot: public repository/service/handler contract giữ nguyên, `Get` missing vẫn là `nil, nil`, schema dùng snake_case (`category`, `created_at`, `deleted_at`), và list mặc định `created_at DESC, id DESC`.

Không dùng ORM relationship, auto-join, reflection mapping, upsert ngầm hoặc multi-dialect abstraction. Không phải mọi SQL đều biến mất: chỉ SQL cơ học của CRUD/query động được shared framework quản lý.

## Tiến độ

### Đã xong

- `internal/shared/validation`, `internal/shared/apperror`, `internal/i18n` và category validator/repository/service tiếp tục giữ nguyên boundary nghiệp vụ.
- Category handler hỗ trợ List/Search/Get/Create/Update/Delete, filter typed hiện có và strict decoder của search request.
- HTTP transport đã migrate từ Chi sang Echo v5: composition root, error pipeline, auth middleware, request-timeout middleware, Category handlers và route module đều dùng Echo-native API.
- Module graph khai báo `github.com/labstack/echo/v5` và `go.uber.org/fx` là direct dependencies; Chi được loại khỏi graph sau `go mod tidy`.
- HTTP policy được giữ: bind/parse 421, business resource not-found 400, timeout 504, route/API 404 do gateway sở hữu với Echo fallback cho direct backend.
- `http.Server` timeout, graceful shutdown, JWT verification và user-context layer được giữ nguyên về boundary/lifecycle.

### Chưa làm (không tự ý làm nếu chưa được yêu cầu)

- `repository.Delete` chưa kiểm tra số row ảnh hưởng; xóa id không tồn tại vẫn có thể trả 204.
- Login / issue token / refresh token và authorization theo policy (role/permission).
- Các sentinel error khác như 403/415, JSON envelope chuẩn hóa, health endpoint, structured logging, config/database hardening.
- Xác minh độc lập sau migration: format, `go mod verify`, test, build, vet và HTTP cases của Echo. Không giữ test tạm trong repo khi chưa có yêu cầu.
