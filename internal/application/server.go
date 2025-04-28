package application

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprom "github.com/zsais/go-gin-prometheus"
	"project_reference/infrastructure/database"
	"project_reference/infrastructure/rabbit"
	"project_reference/internal/controller"
)

func setupServer(ctx context.Context, db *database.DB, mq *rabbit.RabbitMQ) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(controller.LoggerMiddleware())

	prom := ginprom.NewPrometheus("gin")
	prom.Use(r)

	r.GET("/api/ping", controller.Ping)
	r.GET("/api/docs/spec", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})
	r.GET("/api/docs/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/docs/spec")),
	)

	return r, nil
}
