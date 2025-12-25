package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"dev_ctr_hello/pkg/auth"
	"dev_ctr_hello/pkg/storage"

	"github.com/gin-gonic/gin"
)

const (
	// Version 服务版本
	Version = "1.0.0"
)

func main() {
	// 从环境变量获取配置
	port := getEnv("PORT", "8080")
	authToken := getEnv("AUTH_TOKEN", "your-secret-token-change-me")
	storageDir := getEnv("STORAGE_DIR", "/app/storage")
	defaultTTLStr := getEnv("DEFAULT_TTL", "1h")

	// 解析默认过期时间
	defaultTTL, err := time.ParseDuration(defaultTTLStr)
	if err != nil {
		defaultTTL = 1 * time.Hour
	}

	// 初始化文件存储
	fileStorage, err := storage.NewStorage(storageDir, defaultTTL)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer fileStorage.Stop()

	// 创建 Gin 引擎
	r := gin.Default()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建认证中间件
	authMiddleware := auth.NewAuthMiddleware(authToken)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": Version,
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 上传文件（需要认证）
		api.POST("/upload", authMiddleware.RequireAuth(), handleUpload(fileStorage))

		// 下载文件（不需要认证）
		api.GET("/download/:id", handleDownload(fileStorage))

		// 获取文件元数据（不需要认证）
		api.GET("/file/:id/metadata", handleGetMetadata(fileStorage))

		// 删除文件（需要认证）
		api.DELETE("/file/:id", authMiddleware.RequireAuth(), handleDelete(fileStorage))
	}

	// 启动服务器
	r.Run(":" + port)
}

// handleUpload 处理文件上传
func handleUpload(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{
				"error": "file is required",
			})
			return
		}

		// 打开文件
		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{
				"error": "failed to open file",
			})
			return
		}
		defer src.Close()

		// 获取自定义过期时间（可选）
		var ttl time.Duration
		if ttlStr := c.PostForm("ttl"); ttlStr != "" {
			ttl, _ = time.ParseDuration(ttlStr)
		}

		// 保存文件
		metadata, err := s.SaveFromReader(src, file.Filename, file.Header.Get("Content-Type"), file.Size, ttl)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "failed to save file",
			})
			return
		}

		// 构建下载URL
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		downloadURL := fmt.Sprintf("%s://%s/api/v1/download/%s", scheme, c.Request.Host, metadata.ID)

		c.JSON(200, gin.H{
			"id":            metadata.ID,
			"name":          metadata.Name,
			"size":          metadata.Size,
			"content_type":  metadata.ContentType,
			"upload_time":   metadata.UploadTime,
			"expires_at":    metadata.ExpiresAt,
			"download_url":  downloadURL,
		})
	}
}

// handleDownload 处理文件下载
func handleDownload(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// 获取文件
		metadata, content, err := s.Get(id)
		if err != nil {
			if err == storage.ErrFileNotFound {
				c.JSON(404, gin.H{
					"error": "file not found",
				})
				return
			}
			if err == storage.ErrFileExpired {
				c.JSON(410, gin.H{
					"error": "file has expired",
				})
				return
			}
			c.JSON(500, gin.H{
				"error": "failed to get file",
			})
			return
		}

		// 设置响应头
		c.Header("Content-Type", metadata.ContentType)
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, metadata.Name))
		c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))
		c.Header("X-File-Name", metadata.Name)
		c.Header("X-Upload-Time", metadata.UploadTime.Format(time.RFC3339))
		c.Header("X-Expires-At", metadata.ExpiresAt.Format(time.RFC3339))

		// 返回文件内容
		c.Data(200, metadata.ContentType, content)
	}
}

// handleGetMetadata 处理获取文件元数据
func handleGetMetadata(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		metadata, err := s.GetMetadata(id)
		if err != nil {
			if err == storage.ErrFileNotFound {
				c.JSON(404, gin.H{
					"error": "file not found",
				})
				return
			}
			if err == storage.ErrFileExpired {
				c.JSON(410, gin.H{
					"error": "file has expired",
				})
				return
			}
			c.JSON(500, gin.H{
				"error": "failed to get metadata",
			})
			return
		}

		c.JSON(200, metadata)
	}
}

// handleDelete 处理文件删除
func handleDelete(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := s.Delete(id); err != nil {
			c.JSON(500, gin.H{
				"error": "failed to delete file",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "file deleted successfully",
		})
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
