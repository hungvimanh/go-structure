package usercontext

import "context"

// UserContext la du lieu ve nguoi goi request, doc lap voi JWT/RSA.
// Service/repository chi phu thuoc vao struct nay, khong phu thuoc internal/auth.
// Sau nay can them du lieu gi khac ve request (khong chi tu JWT), bo sung field vao day.
type UserContext struct {
	UserID   string
	Email    string
	FullName string
}

type contextKey struct{}

var ctxKey = contextKey{}

func WithUser(ctx context.Context, u *UserContext) context.Context {
	return context.WithValue(ctx, ctxKey, u)
}

func FromContext(ctx context.Context) (*UserContext, bool) {
	u, ok := ctx.Value(ctxKey).(*UserContext)
	return u, ok
}
