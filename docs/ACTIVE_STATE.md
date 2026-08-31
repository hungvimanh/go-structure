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
| **Binding / syntax** | Handler | — | 421 | Request id không parse được (`uuid.Parse`), JSON body decode lỗi. Handler `return httpresponse.NewBindingError(message, cause)`, `ErrorResponder.RespondError` nhận diện và ghi 421 qua `WriteBindingError`. |
| **Business** | Service / Validator | Handler (không check syntax), Repository (không check nghiệp vụ) | 400 (mọi `validation.Errors`, kể cả not-found) | `validation.Errors` (`internal/shared/validation`) — field-level, gom hết lỗi rồi mới trả, dịch qua i18n catalog. |
| **Technical / infra** | Repository | Service, Handler (không tự phán đoán loại lỗi) | 400 (`ErrBadRequest`) / 500 (còn lại) | `apperror` (`internal/shared/apperror`) — sentinel errors (`ErrBadRequest`, `ErrInternal`), message chung chung trả về client, raw error log server-side qua `log.Printf`, không lộ ra ngoài **trừ khi `DEV_MODE=1`** (xem bên dưới). |
| **Authentication** | Middleware (`internal/auth`) | Handler/Service/Repository (không tự verify token) | 401 | Chạy trước handler, ghi response trực tiếp qua `httpresponse.WriteError` khi thiếu/sai token — không đi qua `ErrorResponder.RespondError`. Message trả về client luôn chung chung `"unauthorized"`, chi tiết lỗi log server-side qua `log.Printf`, không phân biệt "thiếu token" / "sai chữ ký" / "hết hạn" ra ngoài (tránh lộ thông tin, **không** bị ảnh hưởng bởi `DEV_MODE` — 401 luôn chung chung dù dev hay production). |

### Error handling middleware — HandlerFunc + Wrap

Handler không tự gọi `RespondError`/ghi response lỗi nữa — lý do: `http.Handler`/`http.HandlerFunc` chuẩn không có kênh trả lỗi (chữ ký `func(w, r)` không return gì), nên cần đổi chữ ký + 1 adapter để mô phỏng middleware xử lý lỗi tập trung.

- `httpresponse.HandlerFunc` = `func(w http.ResponseWriter, r *http.Request) error`. Method handler (`CategoryHandler.Get`, `.List`, ...) theo chữ ký này: chỉ mô tả cổng vào/ra của API — bind request, gọi service, `WriteJSON` khi thành công, **return error thay vì tự ghi response lỗi**. `CategoryHandler` không còn giữ `*httpresponse.ErrorResponder` nữa (không cần, xem bên dưới).
- `httpresponse.NewBindingError(message, cause)` trả về `*BindingError` (struct implement `error`, giữ `Cause` là error gốc như `uuid.Parse`/`json.Decode` để không mất thông tin) — handler dùng cái này cho lỗi binding/syntax thay vì gọi `WriteBindingError` trực tiếp.
- `(*ErrorResponder).Wrap(h HandlerFunc) http.HandlerFunc`: adapter duy nhất biến `HandlerFunc` thành `http.HandlerFunc` để đăng ký với chi — gọi `h(w, r)`, nếu có error thì gọi `RespondError` 1 lần. Đây là **điểm nối** để đăng ký route: `main.go` gọi `errorResponder.Wrap(categoryHandler.Get)` thay vì đăng ký thẳng `categoryHandler.Get`.
- `RespondError` giờ phân loại theo thứ tự: `validation.Errors` (400, dịch i18n) → `*BindingError` (421, qua `WriteBindingError`) → còn lại qua `classifyError` (400/500, có áp `DEV_MODE`).
- Log lỗi technical (500) tách thành method riêng `(*ErrorResponder).logError(err)` (hiện chỉ `log.Printf`) — đây là **điểm duy nhất** cần sửa khi sau này muốn lưu log vào DB hoặc bắn lên queue, không phải đụng vào `RespondError` hay bất kỳ handler nào.
- Kết quả: `category_handler.go` không còn import `errors`/`log`/`apperror`/`validation`/`i18n` gì cả, cũng không giữ field `responder` — chỉ còn phụ thuộc `service` + `httpresponse` (để `WriteJSON`/`NewBindingError`) + `pagination`/`model`. Module tương lai làm y hệt: viết handler theo `httpresponse.HandlerFunc`, đăng ký route qua `errorResponder.Wrap(...)`.

### DEV_MODE

- Env var `DEV_MODE` (`"0"`/`"1"`, mặc định `"0"` khi không set) — parse trong `config.Load()` thành `Config.DevMode bool` (`getEnv("DEV_MODE", "0") == "1"`), không fail-fast nếu thiếu (khác `DB_PASSWORD`).
- Chỉ ảnh hưởng nhánh lỗi **technical** trong `ErrorResponder.RespondError` (`internal/shared/httpresponse`) — tức là mọi thứ **không phải** `validation.Errors`: khi `DevMode=true`, message trả về client là `err.Error()` đầy đủ thay vì message chung chung (`"bad request"`/`"internal server error"`) từ `classifyError`. Status code không đổi.
- **Không** ảnh hưởng: `validation.Errors` (luôn dịch qua i18n catalog, không liên quan che/không che), lỗi binding/syntax (`httpresponse.WriteBindingError`, message vốn đã cụ thể theo call site), lỗi authentication (401 luôn `"unauthorized"` — cố ý không cho `DEV_MODE` lộ chi tiết verify token dù ở dev).
- `ErrorResponder` nhận `devMode` qua constructor (`httpresponse.NewErrorResponder(catalog, cfg.DevMode)`, gọi 1 lần trong `main.go`) — không đọc env trực tiếp trong `httpresponse`, giữ package này không phụ thuộc `config`.
- **Không được bật `DEV_MODE=1` ở production** — lộ chi tiết lỗi hạ tầng (ví dụ raw SQL error) ra ngoài cho client.

Nguyên tắc cốt lõi: **"không có row" không phải lỗi kỹ thuật** — repository trả `(nil, nil)` khi query không có kết quả; service tự diễn giải `nil` thành business error (`validator.NotFound` = `CATEGORY_NOT_FOUND`), không phải lỗi hạ tầng. Về HTTP status, `NotFound` **không** còn được ưu tiên thành 404 — nó vẫn chỉ là 1 code trong `validation.Errors` như các lỗi field khác, trả 400 kèm message đã dịch (xem lý do ở mục "Đã xong" bên dưới).

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
- Handler (`CategoryHandler.List`) gọi `pagination.Parse(r)`; lỗi parse `return httpresponse.NewBindingError("invalid pagination params", err)` → **421** — coi ngang hàng với lỗi cú pháp khác (uuid.Parse, JSON decode), không phải business error.
- Luồng xuống: `handler` parse → truyền `pagination.Params` qua `service.List` → `repository.List` dùng làm `LIMIT $1 OFFSET $2` trong SQL.

## Server timeout + graceful shutdown

- `cmd/api/main.go` đổi từ `http.ListenAndServe` trần sang `&http.Server{...}` tường minh với 4 timeout: `ReadHeaderTimeout=5s` (chặn slowloris — client gửi header nhỏ giọt để giữ connection), `ReadTimeout=15s` (đọc hết header+body), `IdleTimeout=90s` (keep-alive connection rảnh), và `WriteTimeout=cfg.RequestTimeout+5s` — cố ý **suy ra** từ `cfg.RequestTimeout` (không phải hằng số độc lập) vì `WriteTimeout` phải lớn hơn thời gian tối đa 1 request được phép chạy (`reqtimeout`), nếu không `http.Server` sẽ tự cắt connection trước khi `ErrorResponder` kịp ghi xong response 504. 3 timeout còn lại hardcode như hằng số (giống cách `posgres.go` hardcode `MaxConns`/`MinConns`) — không cần chỉnh theo môi trường như `REQUEST_TIMEOUT_SECONDS`.
- Graceful shutdown: `server.ListenAndServe()` chạy trong goroutine riêng (kênh `serverErrs` nhận lỗi start-up thật, không nhận `http.ErrServerClosed`). `main()` chờ ở `select` giữa kênh đó và `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` — nhận tín hiệu dừng (Ctrl+C hoặc SIGTERM từ container orchestrator lúc deploy/scale down) thì gọi `server.Shutdown(shutdownCtx)` với timeout = `cfg.RequestTimeout+5s` (cùng lý do với `WriteTimeout` — request đang chạy có thể mất tối đa `cfg.RequestTimeout` để tự kết thúc qua `reqtimeout`). `Shutdown` tự chờ mọi request đang xử lý hoàn tất (hoặc hết `shutdownCtx` thì force-close) trước khi return — `db.Close()` (`defer` sẵn có) nhờ vậy chỉ chạy sau khi mọi request đã xong, không cắt ngang.
- Đã verify bằng test tạm (2 case: `Shutdown` chờ request đang chạy hoàn tất mới return + trả đúng response; connection mới bị từ chối ngay khi `Shutdown` bắt đầu) — xóa sau khi verify theo quy tắc. Không verify được qua chạy binary thật vì môi trường dev không có Postgres/`JWT_PUBLIC_KEY` sẵn sàng.
- Lưu ý Windows: `syscall.SIGTERM` compile được trên Windows (Go định nghĩa hằng số này cho mọi platform) nhưng Windows không thực sự gửi SIGTERM như POSIX — chỉ `os.Interrupt` (Ctrl+C) hoạt động khi chạy local. Code vẫn đúng khi deploy lên Linux (Docker/K8s gửi SIGTERM thật khi dừng container).

## Panic recovery

- `(*httpresponse.ErrorResponder).Recoverer`: middleware chi-compatible (`func(http.Handler) http.Handler`), method trên `ErrorResponder` (giống `Wrap`) thay vì package riêng — vì cần gọi thẳng `RespondError` để panic đi qua đúng 1 con đường xử lý lỗi duy nhất (không tạo nhánh log/response thứ 2 song song với `RespondError`).
- Cơ chế: `defer recover()` bọc `next.ServeHTTP` — panic bắt được thì đổi thành `error` kèm stack trace (`fmt.Errorf("panic: %v\n%s", rec, debug.Stack())`), truyền vào `RespondError` y hệt mọi lỗi technical khác → luôn **500**, luôn log qua `logError`, chỉ lộ stack trace ra client khi `DEV_MODE=1` (không đặc cách so với lỗi technical khác).
- Đăng ký ở `main.go`: `router.Use(errorResponder.Recoverer)` — **middleware ngoài cùng**, đặt trước `reqtimeout.Middleware` — để bắt được panic từ bất kỳ middleware/handler nào phía trong, không riêng gì handler nghiệp vụ. Không dùng `chi/middleware.Recoverer` (built-in của chi) vì nó tự ghi response riêng, không đi qua `ErrorResponder` → phá vỡ nguyên tắc "1 điểm duy nhất quyết định status/message" của kiến trúc hiện tại.
- Mục đích: trước đây 1 panic (nil pointer, index out of range, nil map write...) trong request handler sẽ crash toàn bộ server (mọi request khác đang chạy song song trong goroutine khác cũng bị kéo theo) — nay chỉ request gây panic đó nhận 500, server tiếp tục phục vụ các request khác bình thường.

## Request timeout

- `internal/shared/reqtimeout`: middleware `Middleware(timeout time.Duration) func(http.Handler) http.Handler` — bọc `r.Context()` bằng `context.WithTimeout`, **không tự ghi response** khi hết giờ (khác `http.TimeoutHandler`/`chi/middleware.Timeout`) — để giữ đúng kiến trúc: mọi response lỗi phải đi qua `ErrorResponder.RespondError` (xem mục "Error handling middleware"), không có 2 nơi tự quyết định status code.
- Đăng ký ở `main.go`: `router.Use(reqtimeout.Middleware(cfg.RequestTimeout))` ngay sau `chi.NewRouter()`, trước mọi route (public lẫn protected) — đây là middleware hạ tầng, không phải auth, nên áp dụng toàn cục, không đặt trong `Group`.
- Cơ chế: context bị cancel → `pgxpool` (nhận `ctx` ở mọi method) tự gửi cancel request xuống Postgres và trả `context.DeadlineExceeded` — lỗi này chảy ngược qua repository → service (`fmt.Errorf("...: %w", err)`) → handler → `ErrorResponder.RespondError`. `classifyError` (`internal/shared/httpresponse`) nhận diện `errors.Is(err, context.DeadlineExceeded)` → **504** (tách riêng khỏi 500 "internal server error" vì đây không phải lỗi server, là timeout có chủ đích). Log qua `logError` áp dụng cho cả 500 và 504.
- Mục đích: 1 query bị treo (deadlock, DB chậm) sẽ tự hủy sau N giây thay vì giữ connection trong `pgxpool` (giới hạn `MaxConns=10`, xem `posgres.go`) vô thời hạn — tránh cạn pool kéo domino các request khác khi traffic cao.
- Config: `REQUEST_TIMEOUT_SECONDS` (mặc định `10` khi không set hoặc parse lỗi, không fail-fast — khác `DB_PASSWORD`), parse trong `config.Load()` qua `getEnvDurationSeconds` thành `Config.RequestTimeout time.Duration`.
- **Không** ảnh hưởng: `validation.Errors` (400) và `*BindingError` (421) — 2 loại này được `RespondError` nhận diện trước `classifyError`, không đi qua nhánh timeout.

## Module wiring pattern

- Vấn đề: `main.go` khởi tạo trực tiếp từng repository/service/handler + đăng ký route của từng module ngay trong `main()` — module càng nhiều thì `main.go` càng dài, và nhiều người cùng thêm module mới cùng lúc dễ conflict ngay tại chỗ khởi tạo/đăng ký route (dễ resolve sai, mất code).
- Giải pháp: mỗi module tự expose 1 file wiring (`internal/<module>/module.go`, package cùng tên module) với 2 hàm:
  - `New(deps...) *Module`: gói toàn bộ chuỗi khởi tạo `repository -> service -> handler` của module đó, trả về struct `Module` giữ handler bên trong (không export field).
  - `(*Module) RegisterRoutes(router chi.Router, errorResponder *httpresponse.ErrorResponder, ...deps khác)`: đăng ký toàn bộ route (public lẫn protected) của module đó vào router truyền vào.
- `main.go` chỉ còn là bootstrap chung (config/db/i18n/jwt/server timeout/graceful shutdown) + với mỗi module đúng 2 dòng: `xModule := x.New(db)` và `xModule.RegisterRoutes(router, errorResponder, ...)` — không còn biết chi tiết khởi tạo bên trong từng module.
- Đã áp dụng cho `category`: xem `internal/category/module.go`. Đây là mẫu cho module tiếp theo — không phải refactor bắt buộc phải làm lại toàn bộ, chỉ áp dụng khi thêm module mới hoặc khi được yêu cầu.
- Tradeoff đã biết: vẫn còn 1 điểm chèn chung trong `main.go` (chỗ gọi `New`/`RegisterRoutes`), nên chưa triệt tiêu 100% khả năng conflict khi 2 module được thêm cùng lúc — nhưng diff ở đó chỉ còn 1-2 dòng/module thay vì cả khối khởi tạo + định nghĩa route, nên nếu có conflict cũng dễ resolve đúng hơn nhiều.

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
- `internal/shared/httpresponse` (mới, tách từ `category/handler`): `WriteJSON`/`WriteError` dùng chung giữa `category/handler` và `auth` middleware — tránh lặp hàm ghi JSON response ở lớp thứ 2. `ErrorResponder` (struct, tạo 1 lần trong `main.go` qua `NewErrorResponder(catalog, devMode)`) cũng ở đây. Xem chi tiết cơ chế wrap-as-middleware ở mục "Error handling middleware" bên dưới.
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
- `internal/category/handler`: mọi method chuyển sang chữ ký `httpresponse.HandlerFunc` (`return error` thay vì tự ghi response lỗi) — xem mục "Error handling middleware" phía trên. Handler không còn import gì liên quan xử lý lỗi (`errors`/`log`/`apperror`/`validation`/`i18n`) và không giữ field `responder` nữa — chỉ mô tả bind request → gọi service → `WriteJSON`. Đồng thời bỏ luôn nhánh đặc cách `NotFound → 404`: `Get` với id không tồn tại giờ trả **400** kèm lỗi validate (`CATEGORY_NOT_FOUND`) như mọi lỗi field khác, thay vì 404 — quyết định chủ động, không phải do quên xử lý.
- `cmd/api/main.go`: bootstrap chung (config/db/i18n/jwt) + gọi `category.New(db)` / `categoryModule.RegisterRoutes(router, errorResponder, auth.Middleware(jwtPublicKey))` — DI thủ công (`repository.NewCategoryRepository` → `service.NewCategoryService` → `handler.NewCategoryHandler(service)`) và route (`GET /`, `POST /`, `GET /{id}`, `PUT /{id}`, `DELETE /{id}` dưới `/categories`, qua `errorResponder.Wrap(...)`) đã chuyển vào `internal/category/module.go` — xem mục "Module wiring pattern" phía trên. Chạy qua `&http.Server{}` với timeout + graceful shutdown (xem mục "Server timeout + graceful shutdown"). `Count` chưa có route vì handler chưa có method tương ứng — quyết định chủ động không gộp vào `List` (xem mục "Tiến độ" bên dưới), nếu cần expose thì thêm route riêng khi có nhu cầu thực tế.
- `internal/auth`: middleware verify access token JWT RS256. Xem chi tiết ở mục Authentication phía trên.
- `internal/shared/usercontext`: layer `UserContext` tách khỏi JWT, truyền qua `context.Context` xuống mọi layer. Xem mục Authentication phía trên.
- `DEV_MODE`: env var bật/tắt lộ chi tiết lỗi technical trong response. Xem mục "DEV_MODE" phía trên.
- `internal/shared/reqtimeout`: middleware cancel context sau `REQUEST_TIMEOUT_SECONDS` giây, `classifyError` map `context.DeadlineExceeded` → 504. Xem mục "Request timeout" phía trên.
- `(*httpresponse.ErrorResponder).Recoverer`: middleware bắt panic, đổi thành lỗi 500 qua `RespondError` thay vì crash server. Xem mục "Panic recovery" phía trên.
- `cmd/api/main.go`: `http.Server` với timeout tường minh (`ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout`) + graceful shutdown qua `signal.NotifyContext`/`server.Shutdown`. Xem mục "Server timeout + graceful shutdown" phía trên.

### Chưa làm (không tự ý làm nếu chưa được yêu cầu)
- `repository.Delete` dùng `Exec` không check số row bị ảnh hưởng — xóa id không tồn tại vẫn trả 204 âm thầm, chưa có check not-found như `Get`/`Update`.
- Các sentinel error khác (403/415/...) chưa định nghĩa trong `apperror`, sẽ thêm khi có nhu cầu thực tế (content-type check, v.v.). Riêng 401 đã có (xem mục Authentication).
- **Login / issue token / refresh token**: chưa làm, chỉ mới verify token do bên khác cấp.
- **Authorization theo policy (role/permission)**: middleware hiện tại mới trả lời "token hợp lệ hay không", chưa quyết định request có được phép thực hiện hành động cụ thể hay không (ví dụ: role nào được xóa category). Cần thiết kế riêng khi có yêu cầu.
- Production readiness (server timeout, graceful shutdown, panic recovery, request timeout) đã làm xong theo các mục phía trên — không còn mục nào tồn đọng từ khảo sát ban đầu.
- **Quyết định chủ động: không gộp `List` + `Count` thành 1 endpoint.** Từng cân nhắc chạy song song qua `errgroup` cho response dạng `{items, total}`, nhưng bị từ chối — client tự gọi `List` hoặc `Count` riêng theo nhu cầu lấy dữ liệu, không áp đặt hình dạng response cố định gộp cả hai. `List`/`Count` tiếp tục là 2 method độc lập trong `CategoryService`/`CategoryRepository` như hiện tại.
- `userContext` layer đã có (xem mục Authentication) nhưng **chưa có handler/service/repository nào thực sự gọi `usercontext.FromContext`** — mới dừng ở middleware ghi vào context, chưa có nhu cầu nghiệp vụ đọc ra (ví dụ audit `created_by`). Sẽ dùng khi có yêu cầu cụ thể.
