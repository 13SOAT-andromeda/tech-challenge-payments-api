package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(webhook *WebhookHandler, health *HealthHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/webhooks/mercadopago", gin.WrapF(webhook.Handle))
	r.GET("/health", gin.WrapF(health.Handle))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
