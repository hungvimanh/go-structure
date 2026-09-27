package repository

import (
	"context"
	"fmt"
	"strings"

	"category-service/internal/shared/query"

	"github.com/jackc/pgx/v5"
)

// RelationSpec defines static metadata for a junction table.
type RelationSpec struct {
	Table        query.Table
	ParentColumn query.Column
	ChildColumn  query.Column
}

// RelationPatch distinguishes an absent field from an explicit clear request.
type RelationPatch[ID comparable] struct {
	Present bool
	IDs     []ID
}

// RelationResult records the delete and insert effects of one replacement.
type RelationResult struct {
	Deleted CommandResult
	Inserted CommandResult
}

// ReplaceRelation clears and optionally replaces a parent relation set through
// the caller-supplied pgx transaction. Call it inside Transactor.WithinTx
// together with the related aggregate mutation; its signature prevents
// pool-backed autocommit execution.
func ReplaceRelation[ParentID any, ChildID comparable](ctx context.Context, transaction pgx.Tx, spec RelationSpec, parentID ParentID, patch RelationPatch[ChildID]) (RelationResult, error) {
	if !patch.Present {
		return RelationResult{}, nil
	}
	if transaction == nil || spec.Table.SQL() == "" || spec.ParentColumn.SQL() == "" || spec.ChildColumn.SQL() == "" {
		return RelationResult{}, fmt.Errorf("%w: relation metadata is invalid", errInvalidSpec)
	}

	deleteStatement := "DELETE FROM " + spec.Table.SQL() + " WHERE " + spec.ParentColumn.SQL() + " = $1"
	deleteTag, err := transaction.Exec(ctx, deleteStatement, parentID)
	if err != nil {
		return RelationResult{}, err
	}
	result := RelationResult{Deleted: commandResult(deleteTag)}

	ids := deduplicate(patch.IDs)
	if len(ids) == 0 {
		return result, nil
	}

	values := make([]string, len(ids))
	arguments := make([]any, 0, len(ids)*2)
	for index, id := range ids {
		first := index*2 + 1
		values[index] = "($" + integerString(first) + ", $" + integerString(first+1) + ")"
		arguments = append(arguments, parentID, id)
	}
	insertStatement := "INSERT INTO " + spec.Table.SQL() + " (" + spec.ParentColumn.SQL() + ", " + spec.ChildColumn.SQL() + ") VALUES " + strings.Join(values, ", ")
	insertTag, err := transaction.Exec(ctx, insertStatement, arguments...)
	if err != nil {
		return result, err
	}
	result.Inserted = commandResult(insertTag)
	return result, nil
}

func deduplicate[T comparable](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[T]struct{}, len(values))
	result := make([]T, 0, len(values))
	for _, value := range values {
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
