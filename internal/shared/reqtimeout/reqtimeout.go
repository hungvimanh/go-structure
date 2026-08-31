// Package reqtimeout gan context timeout cho moi request truoc khi vao handler.
// Muc dich: mot query DB bi treo se tu huy sau N giay thay vi giu connection
// trong pgxpool (gioi han so luong) vo thoi han, tranh can kiet pool khi traffic cao.
package reqtimeout

import (
	"context"
	"net/http"
	"time"
)

// Middleware khong tu ghi response khi timeout — no chi cancel context, con
// lai de handler/repository tra ve context.DeadlineExceeded nhu 1 error binh
// thuong, roi ErrorResponder.RespondError phan loai thanh 504 (xem classifyError).
func Middleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
