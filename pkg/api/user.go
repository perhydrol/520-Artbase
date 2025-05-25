package api

type LoginRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password"  valid:"required,min=6,max=32"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" valid:"required,min=6,max=32"`
	NewPassword string `json:"new_password" valid:"required,min=6,max=32"`
}

type CreateUserRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required,min=6,max=32"`
}

type GetUserInfoRequest UserInfo

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
