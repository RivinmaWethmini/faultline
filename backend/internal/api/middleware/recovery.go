package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/faultline/faultline/internal/api/response"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				stack := debug.Stack()
				slog.Error("panic recovered",
					"error", rvr,
					"stack", string(stack),
					"path", r.URL.Path,
				)

				response.InternalError(w, "Internal server error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
