// Package reqtimeout gan context timeout cho moi request truoc khi vao handler.
// Muc dich: mot query DB bi treo se tu huy sau N giay thay vi giu connection
// trong pgxpool (gioi han so luong) vo thoi han, tranh can kiet pool khi traffic cao.
package reqtimeout

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
)

// Middleware khong tu ghi response khi timeout — no chi cancel context, con
// lai de handler/repository tra ve context.DeadlineExceeded nhu 1 error binh
// thuong, roi ErrorResponder.RespondError phan loai thanh 504 (xem classifyError).
func Middleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
