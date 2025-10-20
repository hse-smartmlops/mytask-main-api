package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type traceResponseRecorder struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (r *traceResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *traceResponseRecorder) WriteHeader(code int) {
	r.status = code
}

func TraceMeta() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			recorder := &traceResponseRecorder{ResponseWriter: res.Writer, status: http.StatusOK}
			res.Writer = recorder

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			res.Writer = recorder.ResponseWriter

			status := recorder.status
			if res.Status != 0 {
				status = res.Status
			}
			body := recorder.body.Bytes()
			contentType := res.Header().Get(echo.HeaderContentType)

			if status == http.StatusNoContent || len(body) == 0 || !strings.Contains(contentType, echo.MIMEApplicationJSON) {
				if !res.Committed {
					res.WriteHeader(status)
				}
				if len(body) > 0 {
					_, _ = res.Writer.Write(body)
				}
				return nil
			}

			if traceID := traceIDFromContext(c); traceID != "" {
				var payload map[string]interface{}
				if err := json.Unmarshal(body, &payload); err == nil {
					meta, _ := payload["meta"].(map[string]interface{})
					if meta == nil {
						meta = make(map[string]interface{})
					}
					meta["trace_id"] = traceID
					payload["meta"] = meta
					if updated, err := json.Marshal(payload); err == nil {
						body = updated
						res.Header().Set(echo.HeaderContentLength, strconv.Itoa(len(body)))
					}
				}
			}

			if !res.Committed {
				res.WriteHeader(status)
			}
			if len(body) > 0 {
				_, _ = res.Writer.Write(body)
			}
			return nil
		}
	}
}

func traceIDFromContext(c echo.Context) string {
	if traceID := c.Response().Header().Get(echo.HeaderXRequestID); traceID != "" {
		return traceID
	}
	return c.Request().Header.Get(echo.HeaderXRequestID)
}
