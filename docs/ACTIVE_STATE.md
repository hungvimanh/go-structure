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

## Tiến độ

### Đã xong
- `internal/shared/validation`: kiểu `FieldError`/`Errors` dùng chung, generic, không phụ thuộc module cụ thể.
- `internal/shared/apperror`: sentinel errors kỹ thuật (`ErrBadRequest`, `ErrInternal`).
- `internal/i18n`: locale parser, catalog loader, translator.
- `internal/category/validator`: validate đầy đủ Code/Name/Description, check trùng code qua `CodeChecker` interface (không phụ thuộc trực tiếp `repository`), business error `CATEGORY_NOT_FOUND` cho case not-found.
- `internal/category/repository`: `Get` trả `(nil, nil)` khi không có row thay vì lỗi; `ExistsByCode` hỗ trợ loại trừ id (Update).
- `internal/category/service`: `Get`/`Update` phát hiện `category == nil` → business NotFound error.
- `internal/category/handler`: `respondError` phân loại lỗi (`validation.Errors` → 400/404 tùy code, còn lại qua `classifyError` → 400/500), binding error → 421 qua `writeBindingError`.

### Chưa làm (không tự ý làm nếu chưa được yêu cầu)
- `cmd/api/main.go` chưa wire router/handler/service/repository đầy đủ (chưa có `chi.NewRouter()`, chưa đăng ký route, chưa gọi `NewCategoryRepository`/`NewCategoryService`/`NewCategoryHandler`).
- `repository.Delete` dùng `Exec` không check số row bị ảnh hưởng — xóa id không tồn tại vẫn trả 204 âm thầm, chưa có check not-found như `Get`/`Update`.
- Các sentinel error khác (401/403/415/...) chưa định nghĩa trong `apperror`, sẽ thêm khi có nhu cầu thực tế (auth, content-type check, v.v.).
