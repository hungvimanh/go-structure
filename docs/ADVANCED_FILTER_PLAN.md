# Advanced Category List Filtering

## Context

Thay thiết kế query-string `field[operator]` bằng hai mức API:

- `GET /categories/`: pagination + một chuỗi `search` đơn giản.
- `POST /categories/search`: pagination + advanced filter bằng JSON typed.

`POST /categories/` hiện là Create nên không thể dùng lại cho List. Route `/categories/search` được
đặt riêng, vẫn nằm trong auth group. `Count` tiếp tục là method nội bộ, không có HTTP route. Response
List vẫn là JSON array để giữ tương thích.

Filter không có business validation riêng. Handler chỉ strict-decode JSON vào DTO:

- Decode được đúng type → nhận object và chuyển xuống service/repository.
- Malformed JSON, sai datatype, unknown field/operator hoặc có nhiều JSON value → binding error `421`.
- Field/operator absent hoặc `null` → nil, repository bỏ qua.
- Filter object rỗng hoặc operation không có giá trị → không thêm predicate.

Pagination vẫn giữ validation hiện tại vì đó là contract riêng: default `skip=0`, `take=10`, reject số
âm/sai kiểu và cap `take=100`; lỗi cũng là binding error 421.

## Đánh giá thiết kế

Thiết kế typed filter phù hợp hơn `map[string]any` hoặc condition list phẳng:

- JSON decoder kiểm tra datatype tự nhiên, không cần parser key hay validator filter.
- Mỗi entity khai báo đúng field được phép lọc; client không thể gửi tên cột SQL.
- Primitive theo datatype dùng lại được giữa module.
- Mỗi repository có `buildFilter` riêng, explicit mapping field model → cột SQL và dùng pgx placeholder.
- List và Count dùng cùng `CategoryFilter`, nên WHERE luôn nhất quán.

MVP chỉ mở `Code`, `Name`, `Status`. Không mirror `ID`, `Description`, timestamps hay `DeletedAt` chỉ
vì chúng tồn tại trong entity; mỗi field mới phải được mở chủ động khi đã xác định use case/operator.
`DeletedAt IS NULL` luôn là predicate bắt buộc do server quản lý.

## API contract

### GET `/categories/`

Ví dụ:

```http
GET /categories/?skip=0&take=10&search=home
```

- Chỉ nhận `skip`, `take`, `search`; query key khác trả 421 `invalid category list params`.
- `search` được `TrimSpace`; thiếu/rỗng thì bỏ qua.
- Search là literal case-insensitive trên Code OR Name:
  `(Code ILIKE $N ESCAPE '\\' OR Name ILIKE $N ESCAPE '\\')`.
- `%`, `_`, `\\` do client gửi được escape trước khi server thêm `%...%`.
- Kết quả vẫn là `200 []Category`, không thêm envelope/total.

### POST `/categories/search`

```json
{
  "skip": 0,
  "take": 25,
  "filter": {
    "code": { "contains": "CAT" },
    "name": { "startsWith": "Home" },
    "status": { "in": [0, 1] }
  }
}
```

- Body phẳng `{skip, take, filter}` để đồng nhất với GET và tránh DTO pagination lồng không cần thiết.
- `skip`, `take`, `filter` đều optional.
- Handler dùng `json.Decoder.DisallowUnknownFields()` và kiểm tra EOF sau value đầu tiên.
- Sai JSON/type/field/operator → `httpresponse.NewBindingError("invalid category search request", err)`
  → 421.
- Không gọi `validation.Errors`, không thêm i18n code, không có `ValidateFilter`.

### Presence semantics

Không cần custom `Optional.UnmarshalJSON`; pointer/slice đã đủ:

- Field/operator omitted hoặc `null` → pointer nil → bỏ qua.
- Filter field là `{}` → không có operation → bỏ qua.
- Scalar được gửi, kể cả `0` hoặc `""`, có pointer khác nil → là giá trị filter thật.
- `in/notIn` omitted, `null`, hoặc `[]` → length 0 → bỏ qua.
- Mảng có phần tử thì dùng nguyên typed values; không trim/rewrite và không semantic-validate.
- Nhiều operator/field có giá trị được kết hợp bằng AND, kể cả khi chúng mâu thuẫn và trả 0 rows.

## Model và shared datatype filters

Tạo `internal/shared/filter/filter.go`:

```go
type StringFilter struct {
    Eq         *string  `json:"eq"`
    Neq        *string  `json:"neq"`
    Contains   *string  `json:"contains"`
    StartsWith *string  `json:"startsWith"`
    EndsWith   *string  `json:"endsWith"`
    In         []string `json:"in"`
    NotIn      []string `json:"notIn"`
}

type IntFilter struct {
    Eq    *int  `json:"eq"`
    Neq   *int  `json:"neq"`
    Gt    *int  `json:"gt"`
    Gte   *int  `json:"gte"`
    Lt    *int  `json:"lt"`
    Lte   *int  `json:"lte"`
    In    []int `json:"in"`
    NotIn []int `json:"notIn"`
}
```

Tạo `internal/category/model/category_filter.go`:

```go
type CategoryFilter struct {
    Code   *filter.StringFilter `json:"code"`
    Name   *filter.StringFilter `json:"name"`
    Status *filter.IntFilter    `json:"status"`

    Search *string `json:"-"`
}

type CategoryListRequest struct {
    Skip   *int           `json:"skip"`
    Take   *int           `json:"take"`
    Filter CategoryFilter `json:"filter"`
}
```

`Search` là criteria nội bộ cho GET, không decode được từ POST. Nó nằm trong cùng object để signatures
List/Count chỉ nhận một filter object. Public advanced fields vẫn là tên entity và chỉ whitelist ba
field MVP.

Operator semantics:

| Datatype | Operators |
|---|---|
| string | `eq`, `neq`, `contains`, `startsWith`, `endsWith`, `in`, `notIn` |
| int | `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `notIn` |

`contains`/`startsWith`/`endsWith` dùng `ILIKE` case-insensitive. `eq`/`neq`/`in`/`notIn` string dùng
PostgreSQL comparison mặc định case-sensitive, nhất quán với `ExistsByCode(Code = $1)` hiện tại.

## Luồng implementation tương lai

1. **Shared/model**
   - Thêm typed filters và Category DTO như trên; thêm constant `Category_Status` thay literal.
   - Thêm `pagination.FromValues(skip, take *int)` và cho `pagination.Parse` tái sử dụng cùng normalizer.

2. **Handler/routes**
   - GET List whitelist query keys, parse pagination/search và tạo `CategoryFilter{Search: &value}`.
   - Thêm Search handler strict-decode `CategoryListRequest` rồi gọi cùng service List.
   - Thêm protected `r.Post("/search", errorResponder.Wrap(h.Search))`; giữ Create POST `/` nguyên vẹn.

3. **Service/contracts**

   ```go
   List(ctx context.Context, params pagination.Params, f model.CategoryFilter) ([]*model.Category, error)
   Count(ctx context.Context, f model.CategoryFilter) (int64, error)
   ```

   Service chỉ forward, không validate filter và không phụ thuộc package repository ở Handler.

4. **Repository**
   - Thêm private `buildFilter(f model.CategoryFilter, startArg int)` trả clause, args, next placeholder.
   - Builder kiểm tra từng field/operation explicit; pointer nil hoặc slice empty thì skip.
   - Tên cột `Code`, `Name`, `Status` chỉ xuất hiện trong repository. Không reflection/map từ JSON key.
   - Mọi data dùng pgx args. `in` dùng `= ANY($N)`; `notIn` dùng `NOT (column = ANY($N))`.
   - Search tạo một grouped OR; các predicate còn lại nối AND.
   - List và Count cùng gọi builder sau `WHERE DeletedAt IS NULL`; List append LIMIT/OFFSET sau args filter.

Không thêm migration/index trong MVP vì repository không có schema/migration. `ILIKE '%...%'` có thể
sequential scan; chỉ cân nhắc `pg_trgm` sau khi đo `EXPLAIN ANALYZE` trên database tương đương production.

## Critical files

- `docs/ADVANCED_FILTER_PLAN.md`
- `internal/shared/filter/filter.go`
- `internal/shared/pagination/pagination.go`
- `internal/category/model/category_filter.go`
- `internal/category/handler/category_handler.go`
- `internal/category/module.go`
- `internal/category/service/category_service.go`
- `internal/category/repository/category_repository.go`

Không sửa category validator hoặc i18n catalog cho filter.

## Verification khi được duyệt implementation

1. Test tạm rồi xóa: strict JSON decode/421, pointer zero values, nil/empty skip, LIKE escape, SQL
   placeholder/args, GET unknown query, POST route auth, List/Count cùng predicate.
2. Chạy `gofmt`, `go test ./...`, `go vet ./...`, `go build ./...`.
3. Smoke test GET pagination/search, POST advanced search, và xác nhận Create vẫn là POST `/categories/`.
4. Chạy `gitnexus_detect_changes()` trước commit và không trộn các thay đổi Fx đang có sẵn.

## Phạm vi hiện tại

Chỉ tạo/cập nhật tài liệu kế hoạch. Chưa implement source code, route, test hay migration.
