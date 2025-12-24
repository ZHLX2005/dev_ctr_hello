package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Service is 测试deploy分支 版本号发生更新 1224",
		})
	})

	// 根路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Gin DevContainer!",
			"version": "1.0.0",
		})
	})

	// Ping 端点
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// Hello API
		api.GET("/hello/:name", func(c *gin.Context) {
			name := c.Param("name")
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello " + name + "!",
			})
		})

		// 用户 API 组
		users := api.Group("/users")
		{
			users.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"users": []string{"Alice", "Bob", "Charlie"},
				})
			})

			users.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(http.StatusOK, gin.H{
					"user_id": id,
					"name":    "User " + id,
				})
			})

			users.POST("", func(c *gin.Context) {
				var json struct {
					Name string `json:"name" binding:"required"`
				}
				if err := c.ShouldBindJSON(&json); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, gin.H{
					"message": "User created successfully",
					"name":    json.Name,
				})
			})
		}
	}

	// 启动服务器
	r.Run(":8080")
}
