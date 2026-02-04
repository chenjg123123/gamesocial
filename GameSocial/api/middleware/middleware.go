// middleware 提供一组可复用的 HTTP 中间件：链式组合、崩溃恢复、访问日志与 CORS。
package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gamesocial/modules/auth"
)

// Middleware 是对 http.Handler 的装饰器：输入一个 handler，输出包裹后的 handler。
type Middleware func(http.Handler) http.Handler

// Chain 将多个中间件按传入顺序组合起来，返回最终的 http.Handler。
// 例如：Chain(h, A(), B()) 等价于 A(B(h))。
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

// Recover 捕获 handler 链路中的 panic，避免整个进程崩溃，并返回 500。
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logging 打印每个请求的 method/path 与耗时，便于排查性能与请求轨迹。
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

func InjectUserIDFromToken(tokenSecret string) Middleware {
	secret := []byte(strings.TrimSpace(tokenSecret))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(secret) != 0 && r != nil && r.Header.Get("X-User-Id") == "" {
				authz := strings.TrimSpace(r.Header.Get("Authorization"))
				if authz != "" {
					token := authz
					parts := strings.Fields(authz)
					if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
						token = parts[1]
					}
					uid, err := auth.ParseTokenV1(token, secret, time.Now())
					if err == nil && uid != 0 {
						r.Header.Set("X-User-Id", strconv.FormatUint(uid, 10))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS 为跨域请求设置响应头，并在 OPTIONS 预检请求时直接返回 204。
func CORS(allowOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(allowOrigin)
			if origin == "" {
				origin = "*"
			}

			reqOrigin := r.Header.Get("Origin")
			if origin == "*" && reqOrigin != "" {
				origin = reqOrigin
				w.Header().Add("Vary", "Origin")
			} else if origin != "*" && reqOrigin != "" && strings.Contains(origin, ",") {
				allowed := false
				for _, item := range strings.Split(origin, ",") {
					if strings.TrimSpace(item) == reqOrigin {
						allowed = true
						break
					}
				}
				if allowed {
					origin = reqOrigin
					w.Header().Add("Vary", "Origin")
				} else {
					origin = ""
				}
			}

			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				w.Header().Add("Vary", "Access-Control-Request-Headers")
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if origin != "" && origin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Del("Access-Control-Allow-Credentials")
			}

			// 预检请求无需进入业务 handler。
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
