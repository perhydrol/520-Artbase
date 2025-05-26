package api

type LoginRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password"  valid:"required,min=6,max=32"`
	SeedTime int64  `json:"seedTime" valid:"required,min=0"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" valid:"required,min=6,max=32"`
	NewPassword string `json:"new_password" valid:"required,min=6,max=32"`
}

type CreateUserRequest struct {
	Nickname string `json:"nickname" valid:"required,min=6,max=32"`
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required,min=6,max=32"`
}

type GetUserInfoResponse UserInfo

type UserInfo struct {
	Email    string `json:"email"`
	UserUUID string `json:"user_uuid"`
	Nickname string `json:"nickname"`
	CreateAt string `json:"create_at"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Nickname string `json:"nickname" valid:"required,min=6,max=32"`
}
