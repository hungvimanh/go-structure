# Full-Parity Generic Repository Implementation Plan

## Context

Dự án đang dùng PostgreSQL qua `pgx/v5`, nhưng `internal/category/repository/category_repository.go` phải lặp lại SQL, placeholder, scan, soft-delete và filter cho từng action. Mẫu C# trong `docs/plans/major template/` cung cấp một contract phong phú hơn — typed filters, search, sort, projection, pagination, CRUD và replace-reference — nhưng implementation phụ thuộc sâu vào EF Core/Thinktecture, expression tree, reflection vào private API và còn thiếu transaction ở các mutation nhiều bước.

Mục tiêu là tự xây một data-access framework trong `internal/shared` với **full feature parity theo capability nhưng corrected semantics**: giữ ý tưởng tốt của template, không sao chép lỗi/quirk, không dùng GORM/sqlc hay query-builder ngoài, và không biến service/handler thành consumer của một mini-ORM. `CategoryRepository` là pilot migration; contract bên trên repository được giữ nguyên.

**Khả thi:** Có. Tuy nhiên không thể loại bỏ hoàn toàn metadata đặc thù từng entity mà vẫn an toàn. Mỗi module vẫn phải khai báo table/column, projection/scan, write values, searchable/sortable fields và predicate nghiệp vụ; shared framework sẽ loại bỏ phần SQL cơ học và chuẩn hóa semantics.

**Source of truth:** yêu cầu hiện tại; `docs/plans/major template/*.cs`; `internal/category/repository/category_repository.go`; `internal/shared/filter/filter.go`; `internal/shared/pagination/pagination.go`; migration `internal/database/migration/migrations/000001_create_category.up.sql`.

**Chosen approach:** framework PostgreSQL/pgx dùng explicit descriptors + typed query AST/compiler + generic repository engine. Không dùng reflection theo struct tag, không nhận raw identifiers từ request, không mô phỏng LINQ, và không thêm code generation.

**Implementation boundary:** Các task dưới đây là implementation work. Phần verification cuối tài liệu là handoff bắt buộc của Plan Mode, không phải implementation task và sẽ được thực hiện ở stage riêng.

## Impact Summary

- `CategoryRepository`: GitNexus risk **LOW**; direct consumers là `internal/category/service/category_service.go` và `internal/category/module.go`; downstream có `internal/category/handler/category_handler.go` và `cmd/api/main.go`.
- `buildFilter`: risk **LOW**; direct callers chỉ là concrete `Count` và `List`, ảnh hưởng hai filter execution flows trong repository.
- Concrete `Count`, `List`, `Create` hiện không có direct upstream caller được graph nhận diện; vì interface mới là compatibility boundary thực tế, plan giữ nguyên interface này.
- Không có cảnh báo HIGH/CRITICAL.

## Requirements

| ID | Required behavior or constraint | Source |
|---|---|---|
| R1 | Xây shared repository/query framework hoàn toàn nội bộ, không dùng ORM, sqlc, codegen hoặc query-builder ngoài. | User request |
| R2 | Target PostgreSQL + `pgx/v5`; mọi API nhận `context.Context` và chạy được với pool hoặc transaction. | Current code + corrected parity |
| R3 | Hỗ trợ typed filters đầy đủ: string reverse/combine/not operators; scalar equality; ordered range; `IN`/`NOT IN` cho int/int64/decimal-compatible values/time/UUID/bool-compatible values. | C# filter contract |
| R4 | Hỗ trợ nested `AND`/`OR`, search-field whitelist, sort whitelist + direction, stable tie-breaker, projection/select whitelist và offset pagination. | C# query contract + corrected semantics |
| R5 | Mọi value phải parameterized; table/column/order/projection identifiers chỉ đến từ static descriptors và được quote an toàn; request không được truyền raw SQL fragments. | Security/correctness constraint |
| R6 | String operators dùng case-insensitive semantics theo template; active operators trên cùng field kết hợp `AND`, combine operator tạo nhóm `OR` nội bộ. | Selected behavior |
| R7 | Corrected null/empty semantics: nil filter là no-op; empty/whitespace scalar/search là no-op; `IN []` là false; `NOT IN []` là no-op; string predicates không match SQL NULL. | Corrected parity |
| R8 | Full generic read surface gồm `Count`, `Exists`, `List`, `Get`; write surface gồm `Create`, `Update`, soft/hard `Delete`, affected-row result và explicit not-found handling. | Repository parity |
| R9 | Entity mapping phải explicit: column lists, ID, soft-delete, insert/update value extractors và projection-specific scanner; không reflection theo property/struct tag. | Go design constraint |
| R10 | Có transaction abstraction và repository rebinding để aggregate mutation dùng cùng `pgx.Tx`; relation replace-all phải atomic và phân biệt absent / clear / replace. | Corrected C# mutation semantics |
| R11 | PostgreSQL set filtering dùng deduplicated array parameters với `ANY`/`ALL` thay cho Thinktecture temp tables; không tạo threshold/temp-table path khi chưa có profiling. | Corrected PostgreSQL design |
| R12 | Migrate Category làm pilot nhưng giữ nguyên `CategoryRepository`, service và handler contracts; `Get` missing vẫn trả `nil, nil`; write wrappers giữ technical error flow hiện tại. | Current compatibility boundary |
| R13 | SQL sinh ra cho Category phải dùng đúng snake_case schema trong migration (`category`, `created_at`, `deleted_at`, ...), thay vì CamelCase identifiers hiện đang lệch schema. | Migration vs current repository |
| R14 | Không thêm upsert ngầm; nếu cần sau này phải là operation riêng với conflict key rõ ràng. | C# contract has no upsert |
| R15 | Không đưa validation/i18n/reflection behavior của `DataEntity.cs` vào persistence layer. | Layer boundary |

## Scope

### In scope

- Full filter/query capability của template, với corrected semantics.
- Safe PostgreSQL compiler, projection, sort, search, pagination và set filtering.
- Generic pgx repository engine, transaction abstraction và relation replace-all utility.
- Category metadata/mapper/filter adapter và migration khỏi handwritten action SQL.
- Documentation về extension contract cho module mới.

### Out of scope

- Bug-for-bug compatibility với C# (`OrFilter` bị bỏ qua, invalid sort reflection, boolean success không kiểm tra row, sync-over-async, private EF reflection).
- GORM/sqlc/ent, third-party query builder, struct-tag ORM mapping hoặc generated code.
- Multi-dialect support; SQL Server hints, `READ UNCOMMITTED`, `FORCE ORDER` và Thinktecture temp tables.
- Upsert, optimistic locking/version columns, cursor pagination, auto-join/navigation loading.
- Thay đổi HTTP route/body/status contract hoặc schema migration.
- Validation/i18n/resource loading từ `DataEntity.cs`.

## Current and Target Behavior

### Current behavior

- `CategoryRepository` có bảy action rõ ràng nhưng mỗi method tự sở hữu SQL và scan.
- `buildFilter`, `appendStringFilter`, `appendIntFilter`, placeholder numbering và LIKE escaping nằm riêng trong Category repository.
- Filter hiện chỉ có `StringFilter` và `IntFilter`; search là OR giữa code/name; sort cố định `CreatedAt DESC`.
- Repository giữ trực tiếp `*pgxpool.Pool`, chưa có shared `DBTX`/transaction rebinding.
- Migration tạo snake_case columns, trong khi repository dùng unquoted CamelCase names; PostgreSQL sẽ fold thành tên không khớp (`CreatedAt` → `createdat`, không phải `created_at`).

### Target behavior

- Shared query AST compile deterministic PostgreSQL + ordered args; không module nào tự quản placeholder.
- Shared generic repository sinh read/write SQL từ explicit entity spec và projection spec.
- Module repository chỉ map domain filter/query vào typed expressions và giữ query đặc thù ở module khi generic surface không phù hợp.
- Category interface/service/handler không đổi; concrete repository delegate sang shared engine.
- Future aggregate repository có thể chạy parent + link-table replacement trong một transaction.

### Business rules and edge cases

- `R6`: `Eq`/`Neq`, contains/start/end, reverse/not/combine đều case-insensitive. Reverse nghĩa là filter value chứa/bắt đầu/kết thúc bằng column value; combine là normal OR reverse.
- `R7`: nhiều operator cùng active trên một filter được AND; nil/empty/whitespace scalar không phát sinh predicate; empty search không lọc.
- `R7`: `IN []` compile thành constant false; `NOT IN []` không thêm predicate. Set values được deduplicate trước bind.
- `R4`: search term có giá trị nhưng không chọn search field sẽ dùng toàn bộ configured searchable fields, không trả zero rows như template.
- `R4`: sort luôn có deterministic ID tie-breaker; invalid sort/projection key trả typed query error trước khi chạm DB.
- `R8`: generic update/delete báo affected row; Category adapter giữ API `error` hiện có và không tự đổi HTTP/business semantics trong pilot.
- `R10`: relation patch absent = không đổi; present + empty = clear; present + values = replace bằng tập deduplicated, tất cả trong cùng transaction.
- `R12`: Category soft-delete invariant áp dụng cho read/update/delete; `ExistsByCode` vẫn hỗ trợ exclude ID.

## Implementation Design

### Components and responsibilities

1. **`internal/shared/filter` — transport-neutral typed values**
   - Giữ `StringFilter`/`IntFilter` và JSON names hiện có.
   - Bổ sung corrected reverse/not/combine string fields.
   - Bổ sung reusable scalar/ordered/set filter shapes cho int64, time, UUID và module-defined decimal-compatible value.
   - Chỉ chứa data contract; không chứa SQL strings hay pgx dependency.

2. **`internal/shared/query` — typed AST và PostgreSQL compiler**
   - Opaque `Table`, `Column`, `ProjectionKey`, `SortKey`; chỉ descriptor/module code tạo được identifiers.
   - Expression nodes: constants, compare, set, string operations, null checks, `And`, `Or`, `Not`.
   - Central placeholder allocator và argument list.
   - Search, sort, stable tie-breaker, pagination và projection compilation.
   - Validation limits: tối đa depth 4, 64 predicates và 1,000 set values cho một request; vượt giới hạn trả typed query error.
   - PostgreSQL array strategy dùng `= ANY($n)` / `<> ALL($n)` hoặc case-normalized equivalent.

3. **`internal/shared/repository` — execution/runtime**
   - `DBTX` khớp `pgxpool.Pool` và `pgx.Tx`: `Exec`, `Query`, `QueryRow`.
   - `RowScanner` chung cho `pgx.Row`/`pgx.Rows`.
   - `Transactor.WithinTx(ctx, fn)` quản lý begin/commit/rollback và giữ context errors.
   - `EntitySpec[T, ID]`: table, key, columns, soft-delete policy, default order, insert/update extractors.
   - `Projection[TResult]`: selected columns + matching scan callback; dynamic selections không được tách rời scan layout.
   - Generic engine: `Count`, `Exists`, `List`, `Get`, `Create`, `Update`, `Delete`; `WithDB(tx)` tạo bound copy cho transaction.
   - `RelationSpec` + `RelationPatch` cho link-table replace-all, deduplicate và atomic execution.

4. **Category adapter**
   - Static descriptor dùng exact identifiers từ migration.
   - Full entity projection/scanner cho tám Category columns.
   - Filter adapter chuyển `model.CategoryFilter` thành query AST; current search maps tới code/name.
   - Current pagination maps sang shared query options; default sort `created_at DESC, id DESC`.
   - Concrete methods delegate sang engine; `ExistsByCode` dùng shared `Exists` với code + excluded ID.

### Interfaces and data flow

```text
handler/service contracts (unchanged)
  → CategoryRepository adapter
    → Category descriptor + filter/query mapping
      → shared repository engine
        → shared query PostgreSQL compiler
          → DBTX (*pgxpool.Pool or pgx.Tx)
```

Representative contracts (final naming may follow surrounding Go idiom):

```go
type DBTX interface {
    Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
    Query(context.Context, string, ...any) (pgx.Rows, error)
    QueryRow(context.Context, string, ...any) pgx.Row
}

type RowScanner interface { Scan(...any) error }

type EntitySpec[T any, ID any] struct { /* static metadata + mappers */ }
type Projection[T any] struct { /* columns + scanner */ }
type Repository[T any, ID any] struct { /* DBTX + EntitySpec */ }

type RelationPatch[ID any] struct {
    Present bool
    IDs     []ID
}
```

The query API must not expose `Raw(string)` to request/model code. Any unavoidable SQL expression such as `CURRENT_TIMESTAMP` is a closed internal enum/builtin, not caller text.

### String compilation

Use PostgreSQL text functions rather than wildcard interpolation where possible:

- equality: normalized `LOWER(column) = LOWER($n)`;
- contains: `STRPOS(LOWER(column), LOWER($n)) > 0`;
- reverse contains: `STRPOS(LOWER($n), LOWER(column)) > 0`;
- starts/ends: closed compiler forms using normalized operands;
- combine: explicit OR group of normal + reverse;
- not variants: require non-null column and negate the positive expression.

This preserves literal substring semantics without treating `%`/`_` in user values as wildcard syntax.

### State, migration, and compatibility

- Không tạo migration mới; descriptors phải bám migration hiện có.
- `CategoryRepository` interface và Fx provider shape được giữ nếu có thể; shared engine nằm hoàn toàn dưới concrete adapter.
- Existing JSON fields `eq`, `neq`, `contains`, `startsWith`, `endsWith`, `in`, `notIn` tiếp tục hợp lệ; advanced fields chỉ mở rộng contract.
- Việc chuyển Category string scalar filters sang case-insensitive là intentional compatibility change đã chọn; tài liệu phải ghi rõ.
- Không map generic not-found trực tiếp lên service/HTTP trong pilot; adapter duy trì `Get` missing → `nil, nil`.

## Change Map

| Path / area | Action | Responsibility | Requirements |
|---|---|---|---|
| `internal/shared/filter/filter.go` | Modify | Mở rộng typed filter contracts, giữ fields hiện có | R3, R6, R7 |
| `internal/shared/query/identifier.go` | Create | Opaque/sanitized table-column-sort identifiers | R5 |
| `internal/shared/query/expression.go` | Create | Predicate AST, AND/OR/NOT, comparison/string/set nodes | R3, R4, R6, R7 |
| `internal/shared/query/compiler.go` | Create | PostgreSQL SQL/placeholder/args compiler | R2, R5, R11 |
| `internal/shared/query/options.go` | Create | Search, sort, projection, pagination, complexity limits | R4, R7 |
| `internal/shared/query/errors.go` | Create | Invalid query/sort/projection/limit errors | R4, R5 |
| `internal/shared/repository/db.go` | Create | `DBTX`, `RowScanner`, command result abstractions | R2, R8 |
| `internal/shared/repository/transactor.go` | Create | pgx transaction lifecycle and rebinding | R10 |
| `internal/shared/repository/spec.go` | Create | Entity, projection, soft-delete and mapper specs | R8, R9 |
| `internal/shared/repository/repository.go` | Create | Generic Count/Exists/List/Get/Create/Update/Delete | R8, R9, R14 |
| `internal/shared/repository/relation.go` | Create | Atomic relation replace-all utility | R10 |
| `internal/category/repository/category_spec.go` | Create | Exact schema metadata, projection, scanner, write mappings | R9, R13 |
| `internal/category/repository/category_query.go` | Create | Category filter/search/order adapter | R3, R4, R6, R7, R12 |
| `internal/category/repository/category_repository.go` | Modify | Preserve interface; replace handwritten action SQL/helpers with shared engine delegation | R1, R8, R12, R13 |
| `docs/ACTIVE_STATE.md` | Modify | Record generic repository boundary, corrected string semantics and Category pilot status | R12, R13, R15 |

## Task Dependencies

`Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6 → Task 7`

## Implementation Tasks

### Task 1: Expand the typed filter contract

**Requirements:** R3, R6, R7

**Depends on:** None

**Outcome:** Shared filter DTOs represent the complete corrected feature set without SQL dependencies.

**Files / areas:**
- Modify: `internal/shared/filter/filter.go`

**Implementation steps:**
- [ ] Retain current concrete fields and JSON names used by Category.
- [ ] Add not/reverse/combine string operators matching the intended C# semantics, not the miswired implementation.
- [ ] Add reusable scalar/ordered/set shapes for int64, time, UUID, bool-compatible and decimal-compatible values while keeping current `IntFilter` source compatibility.
- [ ] Define canonical presence/empty semantics in package comments: scalar empty/whitespace no-op, `IN []` false, `NOT IN []` no-op.

**Business rules and edge cases:** Multiple active operators on one filter are ANDed; combine fields own an internal OR group; string operations are case-insensitive.

**Interfaces produced or changed:** Extended `filter.StringFilter`; reusable generic/concrete filter value types.

**Completion state:** Modules can express the full filter feature set without importing query/repository packages.

### Task 2: Build the typed query AST and PostgreSQL compiler

**Requirements:** R2, R3, R4, R5, R6, R7, R11

**Depends on:** Task 1

**Outcome:** A deterministic, parameterized PostgreSQL compiler owns all dynamic SQL construction.

**Files / areas:**
- Create: `internal/shared/query/identifier.go`
- Create: `internal/shared/query/expression.go`
- Create: `internal/shared/query/compiler.go`
- Create: `internal/shared/query/options.go`
- Create: `internal/shared/query/errors.go`

**Implementation steps:**
- [ ] Introduce opaque descriptor-owned identifiers and sanitize/quote them with pgx-compatible PostgreSQL rules.
- [ ] Implement immutable/append-only expression nodes for comparisons, string operators, sets, null checks and nested boolean groups.
- [ ] Compile placeholders and args centrally; values never enter SQL text.
- [ ] Implement corrected case-insensitive normal/reverse/combine string semantics using literal text functions.
- [ ] Compile deduplicated sets through PostgreSQL array parameters; preserve empty-set truth rules.
- [ ] Add search-field, sort-key/direction, stable tie-breaker, projection and pagination compilation.
- [ ] Reject unknown keys and queries exceeding depth/predicate/set limits before execution.

**Business rules and edge cases:** Search with no selected fields defaults to all descriptor search fields; sort always appends deterministic ID order; no public raw SQL escape hatch.

**Interfaces produced or changed:** Typed `Expression`, query `Options`, projection/sort/search registries and PostgreSQL compiler result `(sql, args)`.

**Completion state:** A module can safely describe a full dynamic query without assembling SQL or placeholders.

### Task 3: Add pgx execution and transaction primitives

**Requirements:** R2, R8, R10

**Depends on:** Task 2

**Outcome:** Shared persistence code runs identically on a pool or transaction and can guarantee atomic aggregate mutations.

**Files / areas:**
- Create: `internal/shared/repository/db.go`
- Create: `internal/shared/repository/transactor.go`

**Implementation steps:**
- [ ] Define the minimal `DBTX` and `RowScanner` interfaces around pgx method signatures.
- [ ] Implement a pool-backed transactor with begin, callback, rollback-on-error and commit handling.
- [ ] Preserve `context.Canceled`/`context.DeadlineExceeded`; wrap infrastructure errors without losing causes.
- [ ] Provide repository rebinding/clone semantics for a transaction-scoped `DBTX`.

**Business rules and edge cases:** Rollback is best-effort after callback failure; commit failure is returned; no nested transaction promise is introduced.

**Interfaces produced or changed:** `DBTX`, `RowScanner`, `Transactor`, transaction callback contract.

**Completion state:** Later repository tasks do not depend directly on `*pgxpool.Pool`.

### Task 4: Implement entity specs, projections and the generic repository engine

**Requirements:** R5, R8, R9, R12, R14

**Depends on:** Task 3

**Outcome:** CRUD/read operations are generated from explicit metadata while domain adapters control compatibility.

**Files / areas:**
- Create: `internal/shared/repository/spec.go`
- Create: `internal/shared/repository/repository.go`

**Implementation steps:**
- [ ] Define `EntitySpec` with table, key, selected/write columns, soft-delete policy, default order and value extractors.
- [ ] Define projection objects that bind a column sequence to its exact scanner callback.
- [ ] Implement Count/Exists/List/Get and row iteration with close/`rows.Err()` handling.
- [ ] Implement Create/Update/Delete, generated placeholders, affected-row results and soft/hard delete policy.
- [ ] Keep not-found as an engine result that adapters may map to `nil`, boolean or typed domain error.
- [ ] Validate descriptor consistency without reflection or struct tags.

**Business rules and edge cases:** Update/delete on a soft-deleted row do not affect it; no automatic upsert; dynamic projection may only use registered column/scanner layouts.

**Interfaces produced or changed:** `EntitySpec[T, ID]`, `Projection[T]`, generic repository and command result types.

**Completion state:** Entity repositories can delegate mechanical SQL while retaining module-owned contracts and custom queries.

### Task 5: Add atomic relation replacement support

**Requirements:** R10, R14

**Depends on:** Task 4

**Outcome:** The corrected equivalent of `SaveReference` is reusable and transaction-safe.

**Files / areas:**
- Create: `internal/shared/repository/relation.go`

**Implementation steps:**
- [ ] Define static junction-table metadata and a typed `RelationPatch` with explicit `Present` state.
- [ ] Deduplicate child IDs and implement clear/replace operations through the caller-provided transaction-bound `DBTX`.
- [ ] Require parent mutation and relation replacement to execute inside `Transactor.WithinTx`; do not begin hidden independent transactions.
- [ ] Return affected/error information without boolean success masking DB failures.

**Business rules and edge cases:** Absent patch leaves links unchanged; present empty clears; present non-empty replaces the full set atomically; this utility does not validate referenced domain entities.

**Interfaces produced or changed:** `RelationSpec`, `RelationPatch`, relation replacer API.

**Completion state:** Future aggregate modules can reproduce Major reference behavior without partial commits.

### Task 6: Migrate Category as the pilot adapter

**Requirements:** R3, R4, R6, R7, R8, R9, R12, R13

**Depends on:** Task 5

**Outcome:** Category keeps its current upper-layer contract while no longer defining SQL per action.

**Files / areas:**
- Create: `internal/category/repository/category_spec.go`
- Create: `internal/category/repository/category_query.go`
- Modify: `internal/category/repository/category_repository.go`

**Implementation steps:**
- [ ] Register exact table/column names from the migration and a full Category projection/scanner.
- [ ] Map Category insert/update values and soft-delete metadata into `EntitySpec`.
- [ ] Translate current code/name/status filters and code/name search to the query AST.
- [ ] Apply selected case-insensitive string semantics and stable `created_at DESC, id DESC` ordering.
- [ ] Delegate Count/List/Get/Create/Update/Delete to the generic engine.
- [ ] Implement `ExistsByCode` with shared Exists plus excluded-ID predicate.
- [ ] Remove local `buildFilter`, append helpers and LIKE escaping after all callers move.
- [ ] Preserve `CategoryRepository` method signatures and `Get` missing → `nil, nil`; do not leak shared query types to service/handler.

**Business rules and edge cases:** Category always excludes `deleted_at IS NOT NULL`; current pagination limits remain owned by `internal/shared/pagination`; affected-row information is not promoted to HTTP/business behavior in this pilot.

**Interfaces produced or changed:** Concrete Category adapter internals only; public `CategoryRepository` remains stable.

**Completion state:** Category demonstrates the complete shared architecture while upper layers remain unchanged.

### Task 7: Document the extension boundary

**Requirements:** R1, R11, R12, R15

**Depends on:** Task 6

**Outcome:** Future modules can adopt the framework without recreating ORM behavior or leaking persistence concerns upward.

**Files / areas:**
- Modify: `docs/ACTIVE_STATE.md`
- Add package comments alongside `internal/shared/query` and `internal/shared/repository` public contracts.

**Implementation steps:**
- [ ] Document which metadata each module must own: schema identifiers, filter mapping, projections/scanners, write values and custom predicates.
- [ ] Document when to use generic operations versus a module-specific SQL query.
- [ ] Record corrected string/empty-set/search/sort/relation semantics and Category compatibility boundary.
- [ ] Record non-goals: ORM relationships, auto joins, reflection mapping, upsert and multi-dialect abstraction.

**Business rules and edge cases:** Documentation must not claim all SQL disappears; custom aggregate/analytics queries remain module-owned.

**Interfaces produced or changed:** Stable package usage contract and active-state metadata.

**Completion state:** The project has an explicit adoption guide and no ambiguity about framework responsibilities.

## Requirements Coverage

| Requirement | Implemented by | Notes |
|---|---|---|
| R1 | Tasks 2, 4, 7 | Fully internal framework |
| R2 | Tasks 2, 3 | PostgreSQL + pgx pool/tx |
| R3 | Tasks 1, 2, 6 | Complete typed filter capabilities |
| R4 | Tasks 2, 6 | OR/search/sort/projection/pagination |
| R5 | Tasks 2, 4 | Static identifiers and parameterized values |
| R6 | Tasks 1, 2, 6 | Case-insensitive corrected string behavior |
| R7 | Tasks 1, 2, 6 | Corrected nil/empty/set behavior |
| R8 | Tasks 3, 4, 6 | Generic CRUD/read surface |
| R9 | Tasks 4, 6 | Explicit metadata and scanners |
| R10 | Tasks 3, 5 | Atomic transaction/relation support |
| R11 | Task 2 | PostgreSQL arrays, no temp-table clone |
| R12 | Tasks 4, 6, 7 | Category upper-layer compatibility |
| R13 | Task 6 | Exact migration identifiers |
| R14 | Tasks 4, 5 | No implicit upsert |
| R15 | Task 7 | Persistence boundary documented |

## Resolved Decisions

- Scope: full repository/query feature parity, not only a Category MVP.
- Semantics: corrected behavior, not bug-for-bug C# compatibility.
- String filtering: case-insensitive semantics from the .NET template, including Category after migration.
- Target: PostgreSQL + pgx/v5; no premature multi-dialect API.
- Mapping: explicit descriptor/scanner callbacks; no reflection or struct-tag ORM.
- Large sets: PostgreSQL array parameters with deduplication; no Thinktecture-style temp-table path without measured need.
- Projection: registered columns and matching scanners; no reflection-driven arbitrary fields or silent zero-value projection.
- Transactions: explicit `Transactor` and `DBTX` rebinding; no hidden transaction per helper.
- Compatibility: preserve Category repository/service/handler method contracts; correct current schema identifier mismatch.
- Upsert: excluded because the reference template does not implement it.

## Preconditions

- PostgreSQL schema created by `000001_create_category.up.sql` remains authoritative.
- `github.com/jackc/pgx/v5` remains the database driver.
- Before implementation edits, rerun GitNexus impact for each concrete symbol being changed, especially `NewCategoryRepository`, `Count`, `List`, `Get`, `ExistsByCode`, `Create`, `Update`, `Delete`, `buildFilter` and helper functions.

## Verification Handoff (separate stage, not implementation tasks)

After implementation, a later verification/unit-test stage should:

- run formatting, `go test ./...`, `go vet ./...` and `go build ./cmd/api`;
- exercise query compiler matrices for all string normal/not/reverse/combine operators, nested groups, empty sets, placeholder ordering and invalid identifiers/limits;
- exercise pool- and transaction-bound repository paths, rollback/commit behavior and relation absent/clear/replace semantics;
- exercise Category Count/List/Get/Exists/Create/Update/Delete against the snake_case migration and verify case-insensitive filter behavior plus stable pagination;
- inspect generated SQL/args without executing request-derived raw identifiers;
- run `gitnexus_detect_changes(scope: "all")` before any requested commit and confirm only shared persistence/query, Category repository adapter and documentation flows changed.
