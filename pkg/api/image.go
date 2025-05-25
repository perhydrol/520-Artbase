package api

type CreateImageRequest struct {
	OwnerUUID string   `json:"owneruuid"`
	IsPublic  bool     `json:"is_public"`
	Tags      []string `json:"tags"`
}

type GetImageInfoRequest ImageInfo

type ImageInfo struct {
	ImageUUID string   `json:"image_uuid"`
	Token     string   `json:"token"`
	OwnerUUID string   `json:"owneruuid"`
	IsPublic  bool     `json:"is_public"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
