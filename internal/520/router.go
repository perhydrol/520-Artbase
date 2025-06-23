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

	api_v1 := g.Group("/v1")

	auth := api_v1.Group("/auth")
	{
		auth.POST("/register", uc.Create)
		auth.POST("/login", uc.Login)
		auth.POST("/change-password/:email", uc.ChangePassword)
	}

	users := api_v1.Group("/users").Use(middleware.JWTAuth())
	{
		users.GET("/:email", uc.Get)
		users.PATCH("/:email", uc.Update)
	}

	images := api_v1.Group("/images")
	{
		// 公开接口
		images.GET("", imagec.GetPublicList)
		images.GET("/user/:user_uuid/public", imagec.GetUserPublicList)

		// 私有接口：用子路由组加 JWT
		privateImages := images.Group("").Use(middleware.JWTAuth())
		{
			privateImages.POST("", imagec.Create)
			privateImages.GET("/:image_uuid", imagec.Get)
			privateImages.DELETE("/:image_uuid", imagec.DeleteImage)
			privateImages.GET("/user/:user_uuid/images", imagec.GetUserImagesList)
			privateImages.PATCH("/:image_uuid", imagec.UpdateImageTags)
			privateImages.GET("/file/:imageUUIDFileName", imagec.GetImageFile)
		}
	}
	return nil
}
