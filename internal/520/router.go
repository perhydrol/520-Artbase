package demo520

import (
	"demo520/internal/520/controller/image"
	"demo520/internal/520/controller/user"
	"demo520/internal/520/store"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

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
		auth.POST("/change-password/:email", uc.ChangePassword)
		auth.POST("/register", uc.Create)
		auth.POST("/login", uc.Create)
	}
	userg := g.Group("/users")
	{
		userg.GET(":email", uc.Get)
		userg.PATCH(":email", uc.Update)
	}
	imageg := g.Group("/images")
	{
		imageg.POST("", imagec.Create)
		imageg.GET("", imagec.GetPublicList)
		imageg.GET("/user/:user_uuid/public", imagec.GetUserPublicList)
		imageg.GET("/user/:user_uuid/images", imagec.GetUserImagesList)
		imageg.GET(":iamge_uuid", imagec.Get)
		imageg.PATCH(":image_uuid", imagec.UpdateImageTags)
	}
	return nil
}
