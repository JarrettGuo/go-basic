package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type MiddlewareBuilder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc: fn,
	}
}

func (b *MiddlewareBuilder) AllowReqBody() *MiddlewareBuilder {
	b.allowReqBody = true
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		url := context.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: context.Request.Method,
			Url:    url,
		}
		if b.allowReqBody && context.Request.Body != nil {
			body, _ := context.GetRawData()
			// body 是io流，读取后需要重新设置
			reader := io.NopCloser(bytes.NewBuffer(body))
			context.Request.Body = reader
			if len(body) > 1024 {
				body = body[:1024]
			}
			al.ReqBody = string(body)
		}

		if b.allowReqBody {
			context.Writer = responseWriter{
				al:             al,
				ResponseWriter: context.Writer,
			}
		}

		defer func() {
			al.Duration = time.Since(start).String()
			b.loggerFunc(context, al)
		}()
		// 执行到业务逻辑
		context.Next()
	}
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
func (w responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.al.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

type AccessLog struct {
	// HTTP 方法
	Method string
	// 请求路径
	Url string
	// 请求体
	ReqBody string
	// 响应体
	RespBody string
	Duration string
	Status   int
}
