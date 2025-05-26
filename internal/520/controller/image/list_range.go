package image

type ListRange struct {
	Offset int `form:"offset" binding:"required,gt=0"`
	Limit  int `form:"limit" binding:"required,gte=10,lte=30"`
}
