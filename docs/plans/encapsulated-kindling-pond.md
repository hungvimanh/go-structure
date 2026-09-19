# Context

API hiện dùng Chi cho router, route group và path parameter, trong khi error pipeline, auth và timeout được viết theo `net/http`. Người dùng đã chọn **Echo toàn diện**: dùng handlers/middleware Echo-native và `c.JSON`, thay vì bọc các handler `net/http`. Mục tiêu là thay Chi bằng Echo v5 mà vẫn giữ URI, status/JSON của lỗi ứng dụng, auth scope, timeout 504 và lifecycle server/Fx.

Project này dùng để xây khung structure backend Go, chưa giải quyết một bài toán nghiệp vụ cụ thể. Vì vậy migration phải đồng thời đặt ra boundary rõ ràng cho các module tương lai: Echo chỉ thuộc HTTP transport/composition layer, không được rò xuống service, repository, model hay validator.

Các HTTP policy hiện có tiếp tục được giữ nguyên trong migration:

- Lỗi parse/binding input (UUID, JSON, query/pagination) trả **421**.
- Resource hợp lệ về cú pháp nhưng không tồn tại, ví dụ `GET` một `id` không có dữ liệu, tiếp tục là business validation error trả **400**.
- **404 chỉ dành cho route/API không tồn tại**, không dùng để biểu diễn resource nghiệp vụ không tồn tại. Trong deployment bình thường, gateway là tầng chịu trách nhiệm trả 404. Echo default 404 chỉ là fallback phòng thủ khi backend bị gọi trực tiếp.

GitNexus impact cho `provideRouter`, `registerHTTPServer`, `registerRoutes`, `CategoryHandler.Get`, `Update`, `Delete` đều LOW: 0 direct caller/process/module được index (kết quả `Delete` là partial nhưng cũng LOW). Chạy lại impact ngay trước khi edit theo quy tắc project.

## Kế hoạch

1. **Thay dependency framework và làm sạch module graph**
   - Trong `go.mod`/`go.sum`, thay `github.com/go-chi/chi/v5` bằng Echo v5 stable `github.com/labstack/echo/v5 v5.3.1`.
   - Chạy `go mod tidy` và `go mod verify` sau khi sửa import.
   - Xác nhận Chi không còn là direct/transitive dependency hoặc còn import/reference trong source.
   - Xác nhận Echo và `go.uber.org/fx` đều là direct dependency; hiện Fx đang bị đánh dấu `// indirect` dù app import trực tiếp.

2. **Chuyển error pipeline sang Echo-native và giữ một điểm quyết định application error** — `internal/shared/httpresponse/httpresponse.go`
   - Bỏ `HandlerFunc`, `ErrorResponder.Wrap`, `ErrorResponder.Recoverer`, `WriteJSON`, `WriteError` và `WriteBindingError`: các API này chỉ phục vụ adapter `net/http` hiện tại.
   - Đổi `ErrorResponder.RespondError` nhận `*echo.Context` và trả response qua `c.JSON`, nhưng giữ nguyên thứ tự phân loại lỗi, locale, body của validation/binding/technical errors, logging và chính sách `DEV_MODE`.
   - Giữ nguyên `BindingError` và policy parse/binding input trả 421.
   - Bổ sung một loại/sentinel auth error có giữ cause gốc. Auth error luôn map thành `401 {"error":"unauthorized"}`, được log server-side và **không bao giờ lộ cause qua `DEV_MODE`**.
   - Thêm method theo đúng chữ ký `echo.HTTPErrorHandler`. Method phải kiểm tra `c.Response().Committed` trước khi ghi response để không double-write.
   - Với router/framework errors, dùng `echo.StatusCode(err)` thay vì chỉ dựa vào type assertion `*echo.HTTPError`, vì Echo v5 không đảm bảo mọi routing error đều có concrete type đó.
   - Routing 404/405 không đi qua application error classifier và không được biến thành 500. Delegate chúng cho `echo.DefaultHTTPErrorHandler(false)`. Đây chỉ là fallback khi gọi trực tiếp backend; gateway vẫn là nơi sở hữu contract 404 trong deployment bình thường.
   - Không map business/resource not-found thành 404. `validation.Errors` chứa `CATEGORY_NOT_FOUND` tiếp tục trả 400 như contract hiện tại.
   - Mọi error từ handler, auth middleware và panic recovery đều quay về `HTTPErrorHandler`; không còn middleware/handler tự gọi responder riêng.
   - `context.DeadlineExceeded` tiếp tục map 504.

3. **Chuyển custom middleware sang Echo-native và thống nhất invariant trả lỗi**
   - `internal/shared/reqtimeout/reqtimeout.go`: đổi `Middleware` sang `echo.MiddlewareFunc`; tạo request context deadline, gọi `c.SetRequest(c.Request().WithContext(ctx))`, rồi `return next(c)`. Không dùng `middleware.ContextTimeout`, vì default của Echo map deadline thành 503 thay vì contract 504 hiện có.
   - Ghi rõ đây là cooperative cancellation: middleware đặt deadline và các layer phía dưới phải tôn trọng `context.Context`; nó không cưỡng chế dừng code bỏ qua context.
   - `internal/auth/middleware.go`: đổi `Middleware` sang `echo.MiddlewareFunc`; giữ nguyên token verification và `usercontext`. Khi thiếu/sai token, trả auth error có cause về chain thay vì tự gọi `c.JSON`; `HTTPErrorHandler` chịu trách nhiệm trả exact `401 {"error":"unauthorized"}`.
   - Sau khi verify, thay request bằng `c.SetRequest(...)` rồi `return next(c)`.
   - Giữ middleware order ở composition root: Echo `middleware.Recover()` ngoài cùng, custom request-timeout sau đó. Panic trả về từ Recover sẽ đi qua custom `HTTPErrorHandler`; không dùng Echo JWT middleware để giữ exact token verification, 401 body và user-context mapping.
   - Invariant cho middleware tương lai: ưu tiên `return error`, không tự phân loại/log/ghi error response. Chỉ middleware thực sự sở hữu một success response mới được ghi trực tiếp.

4. **Chuyển composition root sang Echo và khóa framework boundary** — `cmd/api/app.go`
   - Đổi `provideRouter` từ `*chi.Mux` sang `*echo.Echo`, khởi tạo `echo.New()`, gán `HTTPErrorHandler` của router bằng method trên `ErrorResponder`, rồi đăng ký `middleware.Recover()` và custom timeout middleware.
   - Đổi `provideAuthMiddleware` và `registerHTTPServer` sang các type Echo. Giữ nguyên `http.Server`, `net.Listen`, Fx lifecycle, timeout cấu hình và graceful shutdown vì `*echo.Echo` vẫn là `http.Handler`.
   - Echo chỉ được import tại HTTP transport/composition boundary: `cmd/api`, handler, route module, auth HTTP middleware, request-timeout middleware và shared HTTP response package.
   - Không được thêm import Echo vào service, repository, model, validator, database, config, i18n hoặc các shared business primitives.

5. **Chuyển Category handlers và route registration**
   - `internal/category/handler/category_handler.go`: đổi cả bảy methods (`Sample`, `List`, `Search`, `Get`, `Create`, `Update`, `Delete`) sang `echo.HandlerFunc` (`func(c *echo.Context) error`). Dùng `c.Param("id")` cho Get/Update/Delete và `return c.JSON(...)` cho success response; Delete dùng `return c.NoContent(204)`.
   - Giữ `c.Request()` cho pagination, strict JSON decoder hiện có (`DisallowUnknownFields` và body trailing-data check) và service contexts. Không đổi sang `c.Bind`, vì binder của Echo làm thay đổi policy bind hiện có.
   - Handler invariant: chỉ bind transport input, gọi service và ghi success response; khi lỗi chỉ `return err`. Handler không gọi `ErrorResponder`, không tự log lỗi service/repository và không tự phân loại application error.
   - `internal/category/module.go`: thay Chi router bằng `*echo.Echo`; đăng ký sample route public trực tiếp, tạo Echo group có auth middleware cho category routes, và đăng ký methods trực tiếp không qua `ErrorResponder.Wrap`. Đổi `/{id}` thành `/:id`; giữ toàn bộ HTTP methods, paths và public/protected scope.
   - Resource `id` không tồn tại vẫn chảy từ service thành `validation.Errors` và trả 400; không dùng Echo/router 404 cho case này.

6. **Cập nhật tài liệu** — `docs/ACTIVE_STATE.md`
   - Thay mô tả HandlerFunc/Wrap bằng Echo handler trả error + `HTTPErrorHandler` tập trung.
   - Thay tài liệu custom `net/http` recoverer/adapter bằng Echo Recover, Echo-native custom timeout/auth, và lý do không dùng `ContextTimeout`.
   - Ghi rõ auth middleware trả error về centralized pipeline, nhưng client vẫn luôn nhận generic 401 và không bị ảnh hưởng bởi `DEV_MODE`.
   - Ghi rõ framework boundary: Echo chỉ tồn tại ở HTTP transport/composition layer.
   - Ghi rõ status policy vẫn giữ nguyên: binding/parse 421, resource not-found 400, route/API not-found 404 do gateway sở hữu; Echo 404 chỉ là direct-backend fallback.
   - Cập nhật module wiring, auth group, route syntax `:id`, static route precedence và tiến độ migration.
   - Không sửa các file có thay đổi không liên quan: `AGENTS.md`, `CLAUDE.md`.

7. **Xác minh**
   - Định dạng các Go file đã đổi; xác nhận không còn import/reference Chi hoặc APIs HTTP adapter đã xoá.
   - Chạy `go mod tidy`, `go mod verify`, `go test ./...`, `go vet ./...`, `go build ./cmd/api` và giữ `TestAppOptions_ValidateApp` xanh.
   - Xác nhận Echo và Fx là direct dependencies trong `go.mod`.
   - Xác nhận không có import Echo ngoài HTTP transport/composition boundary đã liệt kê.
   - Theo quy ước không giữ test cố định, tạo `httptest` tạm rồi xoá sau khi chạy:
     - public sample trả 200;
     - protected route không token trả exact `{"error":"unauthorized"}`/401 qua centralized `HTTPErrorHandler`;
     - `DEV_MODE=1` không làm lộ auth cause;
     - static `/categories/sample` thắng `:id`;
     - path parameter parse sai trả 421;
     - `id` hợp lệ nhưng resource không tồn tại trả business validation 400, không phải 404;
     - strict JSON binding và success body giữ nguyên;
     - routing 404/405 không bị phân loại thành application 500 và được delegate cho Echo default handler; không coi fallback body này là application contract vì gateway sở hữu 404;
     - error sau khi response đã committed không gây double-write;
     - recovery trả 500 theo ErrorResponder rồi server phục vụ request tiếp;
     - timeout vẫn đưa deadline đến service path và trả 504.
   - Không chạy binary thật khi thiếu Postgres/JWT; `httptest` kiểm chứng HTTP layer độc lập.
   - Trước commit (nếu được yêu cầu), chạy `gitnexus_detect_changes(scope: "all")` để xác nhận scope chỉ gồm router, shared HTTP pipeline/middleware, category HTTP layer, dependencies và documentation.

## Giới hạn

- Không đổi service, repository, model, validation, database access hay HTTP server lifecycle.
- Không đổi status policy hiện tại: parse/binding tiếp tục 421; resource nghiệp vụ không tồn tại tiếp tục 400.
- 404 chỉ biểu diễn route/API không tồn tại. Gateway sở hữu contract 404 trong deployment bình thường; Echo default 404/405 chỉ là fallback phòng thủ khi backend bị gọi trực tiếp.
- Không mở rộng migration sang chuẩn hóa JSON envelope, đổi status semantics, health endpoint, structured logging, interface placement hay config/database hardening.
- Sau migration và báo kết quả verify, dừng chờ chỉ dẫn tiếp theo theo quy tắc làm việc trong `docs/ACTIVE_STATE.md`.
