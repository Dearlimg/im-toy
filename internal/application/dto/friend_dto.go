package dto

type AddFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
	Remark    string `json:"remark"`
}

type AcceptFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
}

type DeleteFriendReq struct {
	FriendUID string `json:"friend_uid" binding:"required"`
}

type SearchUserReq struct {
	Keyword string `json:"keyword" binding:"required"`
}

type FriendResp struct {
	UID       string `json:"uid"`
	FriendUID string `json:"friend_uid"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Remark    string `json:"remark"`
	Status    int    `json:"status"`
	IsOnline  bool   `json:"is_online"`
	PubKey    string `json:"pubkey"`
}
