package broker

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/sestack/smq/docs"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/middleware"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"go.uber.org/zap"
)

func InitHTTPMoniter(b *Broker) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	swagger := router.Group("/api/v1/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	user_view := &userView{Cluster: b.cluster}
	login_api := router.Group("/api/v1/login")
	{
		login_api.POST("/", user_view.Login)
	}

	changePwd_api := router.Group("/api/v1/change_password").Use(middleware.JWTAuth())
	{
		changePwd_api.POST("/", user_view.ChangePassword)
	}

	user_api := router.Group("/api/v1/user").Use(middleware.JWTAuth())
	{
		user_api.GET("/", user_view.List)
		user_api.GET("/:name/", user_view.Retrieve)
		user_api.POST("/", user_view.Create)
		user_api.PUT("/:name/", user_view.Update)
		user_api.DELETE("/:name/", user_view.Delete)
	}

	node_view := &nodeView{Cluster: b.cluster}
	middleware.JWTAuth()
	node_api := router.Group("/api/v1/node").Use(middleware.JWTAuth())
	{
		node_api.GET("/", node_view.List)
		node_api.GET("/:id/", node_view.Retrieve)
	}

	client_view := &clientView{Cluster: b.cluster}
	client_api := router.Group("/api/v1/client").Use(middleware.JWTAuth())
	{
		client_api.GET("/", client_view.List)
		client_api.GET("/:id/", client_view.Retrieve)
	}

	subscription_view := &subscriptionView{Cluster: b.cluster}
	subscription_api := router.Group("/api/v1/subscription").Use(middleware.JWTAuth())
	{
		subscription_api.GET("/", subscription_view.List)
		subscription_api.GET("/:id/", subscription_view.Retrieve)
	}

	bridge_view := &bridgeView{Cluster: b.cluster}
	bridge_api := router.Group("/api/v1/bridge").Use(middleware.JWTAuth())
	{
		bridge_api.GET("/", bridge_view.List)
		bridge_api.GET("/:topic/", bridge_view.Retrieve)
		bridge_api.POST("/", bridge_view.Create)
		bridge_api.PUT("/:topic/", bridge_view.Update)
		bridge_api.DELETE("/:topic/", bridge_view.Delete)
	}

	acl_view := &aclView{Cluster: b.cluster}
	acl_api := router.Group("/api/v1/acl").Use(middleware.JWTAuth())
	{
		acl_api.GET("/", acl_view.List)
		acl_api.GET("/:id/", acl_view.Retrieve)
		acl_api.POST("/", acl_view.Create)
		acl_api.PUT("/:id/", acl_view.Update)
		acl_api.DELETE("/:id/", acl_view.Delete)
	}

	global.LOGGER.Info("start http server on",zap.String("hp",fmt.Sprintf("%s:%s", b.config.Http.Host, b.config.Http.Port)))
	router.Run(fmt.Sprintf("%s:%s", b.config.Http.Host, b.config.Http.Port))
}
