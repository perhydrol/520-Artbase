package demo520

import (
	"demo520/internal/520/controller/image"
	"demo520/internal/520/controller/user"
	"demo520/internal/520/store"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func installRouters(g *gin.Engine) error {
	// 注册 404 Handler.
	g.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})

	// 注册 /healthz handler.
	g.GET("/healthz", func(c *gin.Context) {
		log.C(c).Infow("Healthz function called")

		core.WriteResponse(c, nil, map[string]string{"status": "ok"})
	})

	uc := user.NewUserController(store.S)
	imagec := image.NewUserController(store.S)

	auth := g.Group("/auth")
	{
		auth.POST("/register", uc.Create)
		auth.POST("/login", uc.Login)
		auth.POST("/change-password/:email", uc.ChangePassword)
	}

	users := g.Group("/users").Use(middleware.JWTAuth())
	{
		users.GET("/:email", uc.Get)
		users.PATCH("/:email", uc.Update)
	}

	images := g.Group("/images")
	{
		// 公开接口
		images.GET("", imagec.GetPublicList)
		images.GET("/user/:user_uuid/public", imagec.GetUserPublicList)

		// 私有接口：用子路由组加 JWT
		privateImages := images.Group("").Use(middleware.JWTAuth())
		{
			privateImages.POST("", imagec.Create)
			privateImages.GET("/:image_uuid", imagec.Get)
			privateImages.GET("/user/:user_uuid/images", imagec.GetUserImagesList)
			privateImages.PATCH("/:image_uuid", imagec.UpdateImageTags)
		}
	}
	return nil
}
