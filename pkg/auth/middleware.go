package auth

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// SignatureHeader 签名 HTTP 头名称
	SignatureHeader = "X-Signature"
	// TimestampHeader 时间戳 HTTP 头名称
	TimestampHeader = "X-Timestamp"
	// SignatureTolerance 签名时间容差（秒）
	// 防止重放攻击
	SignatureTolerance = 300 // 5 分钟
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	validator *SignatureValidator
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(validator *SignatureValidator) *AuthMiddleware {
	return &AuthMiddleware{
		validator: validator,
	}
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取签名
		signature := c.GetHeader(SignatureHeader)
		if signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "signature header is required",
			})
			c.Abort()
			return
		}

		// 获取时间戳
		timestamp := c.GetHeader(TimestampHeader)
		if timestamp == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "timestamp header is required",
			})
			c.Abort()
			return
		}

		// 验证时间戳（防止重放攻击）
		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid timestamp format",
			})
			c.Abort()
			return
		}

		now := time.Now()
		diff := now.Sub(ts).Seconds()

		if diff > SignatureTolerance || diff < -SignatureTolerance {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "timestamp is too old or in the future",
			})
			c.Abort()
			return
		}

		// 构建待签名字符串: method + path + timestamp
		// 不包含 body 和 query，保持请求头简洁
		method := c.Request.Method
		path := c.Request.URL.Path

		signData := method + ":" + path + ":" + timestamp

		// 验证签名
		if err := m.validator.VerifyString(signData, signature); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid signature",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateSignData 生成待签名字符串（供客户端使用）
func GenerateSignData(method, path, query, timestamp, body string) string {
	if query != "" {
		return method + path + "?" + query + "|" + timestamp + "|" + body
	}
	return method + path + "|" + timestamp + "|" + body
}

// ParseResponseHeaders 解析响应头获取签名信息
func ParseResponseHeaders(headers http.Header) (signature, timestamp string) {
	signature = headers.Get(SignatureHeader)
	timestamp = headers.Get(TimestampHeader)
	return
}

// SignRequest 对 HTTP 请求进行签名（供客户端使用）
func SignRequest(method, path, query, body string, privateKey *rsa.PrivateKey) (signature, timestamp string, err error) {
	timestamp = time.Now().Format(time.RFC3339)
	signData := GenerateSignData(method, path, query, timestamp, body)

	signature, err = SignWithKey([]byte(signData), privateKey)
	if err != nil {
		return "", "", err
	}

	return signature, timestamp, nil
}

// SignRequestForm 对表单请求进行签名（用于文件上传）
func SignRequestForm(method, path, query string, formData map[string]string, privateKey *rsa.PrivateKey) (signature, timestamp string, err error) {
	// 将表单数据按 key 排序后拼接
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}

	// 简单排序（实际项目中可以使用更稳定的排序）
	body := ""
	for _, k := range keys {
		if body != "" {
			body += "&"
		}
		body += k + "=" + formData[k]
	}

	return SignRequest(method, path, query, body, privateKey)
}
