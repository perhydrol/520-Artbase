package api

type HasUserUUID interface {
	GetUserUUID() string
}

type CreateImageRequest struct {
	UserUUID string   `json:"owneruuid" valid:"required,uuidv4"`
	IsPublic bool     `json:"is_public"`
	Tags     []string `json:"tags"`
}

func (req *CreateImageRequest) GetUserUUID() string {
	return req.UserUUID
}

type CreateImageResponse ImageInfo

type GetImageInfoResponse ImageInfo

type ImageInfo struct {
	ImageUUID string   `json:"imageuuid"`
	Token     string   `json:"token"`
	UserUUID  string   `json:"owneruuid"`
	IsPublic  bool     `json:"is_public"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type UpdateImageTagsRequest struct {
	Tags []string `json:"tags"`
}

type ListImageResponse struct {
	Count     int         `json:"count"`
	ImageList []ImageInfo `json:"image_list"`
}

type DeleteImageRequest struct {
	ImageUUID string `json:"image_uuid" valid:"required,uuidv4"`
}
