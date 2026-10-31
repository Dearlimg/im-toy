package dto

type CreateGroupReq struct {
	Name       string   `json:"name" binding:"required,min=2,max=50"`
	MemberUIDs []string `json:"member_uids"`
}

type GroupResp struct {
	GroupID   string `json:"group_id"`
	Name      string `json:"name"`
	OwnerUID  string `json:"owner_uid"`
	Avatar    string `json:"avatar"`
	MaxMember int    `json:"max_member"`
}

type GroupMemberResp struct {
	IsOnline bool   `json:"is_online"`
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     int    `json:"role"`
}

type GroupMessageResp struct {
	MsgID        string `json:"msg_id"`
	GroupID      string `json:"group_id"`
	FromUID      string `json:"from_uid"`
	FromUsername string `json:"from_username"`
	FromNickname string `json:"from_nickname"`
	Content      string `json:"content"`
	SendTime     int64  `json:"send_time"`
	ReadCount    int64  `json:"read_count,omitempty"`
	TotalCount   int    `json:"total_count,omitempty"`
}
