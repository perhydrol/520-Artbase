package image

type ListRange struct {
	Offset int `form:"offset" binding:"gte=0"`
	Limit  int `form:"limit" binding:"required,gte=1,lte=100"`
}
