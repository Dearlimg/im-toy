package dto

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=2,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname" binding:"required,min=2,max=20"`
}

type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResp struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type UserLoginResp struct {
	Token string   `json:"token"`
	User  UserResp `json:"user"`
}
