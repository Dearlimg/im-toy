package response

const (
	CodeSuccess                 = 0
	CodeFailed                  = -1
	CodeInvalidParams           = 40001
	CodeUnauthorized            = 40101
	CodeForbidden               = 40301
	CodeNotFound                = 40401
	CodeInternalError           = 50001
	CodeUserAlreadyExist        = 10001
	CodeUserNotFound            = 10002
	CodeWrongPassword           = 10003
	CodePasswordHashFailed      = 10004
	CodeUserRegisterFailed      = 10005
	CodeTokenGenerateFailed     = 10006
	CodeUserDisabled            = 10007
	CodeCannotAddSelfAsFriend   = 20001
	CodeFriendAlreadyExist      = 20002
	CodeFriendRequestSent       = 20003
	CodeFriendNotFound          = 20004
	CodeFriendAddFailed         = 20005
	CodeFriendRequestNotFound   = 20006
	CodeFriendRequestInvalid    = 20007
	CodeFriendAcceptFailed      = 20008
	CodeFriendCreateFailed      = 20009
	CodeFriendDeleteFailed      = 20010
	CodeFriendListGetFailed     = 20011
	CodeFriendRequestsGetFailed = 20012
	CodeUserSearchFailed        = 20013
	CodeGroupCreateFailed       = 30001
	CodeGroupAddOwnerFailed     = 30002
	CodeGroupNotFound           = 30003
	CodeGroupMemberGetFailed    = 30004
	CodeGroupUserInfoGetFailed  = 30005
	CodeGroupUnauthorized       = 30006
	CodeGroupMessageGetFailed   = 30007
)

var codeMsg = map[int]string{
	CodeSuccess:                 "success",
	CodeFailed:                  "failed",
	CodeInvalidParams:           "invalid params",
	CodeUnauthorized:            "unauthorized",
	CodeForbidden:               "forbidden",
	CodeNotFound:                "not found",
	CodeInternalError:           "internal error",
	CodeUserAlreadyExist:        "user already exist",
	CodeUserNotFound:            "user not found",
	CodeWrongPassword:           "wrong password",
	CodePasswordHashFailed:      "password hash failed",
	CodeUserRegisterFailed:      "user register failed",
	CodeTokenGenerateFailed:     "token generate failed",
	CodeUserDisabled:            "user disabled",
	CodeCannotAddSelfAsFriend:   "cannot add self as friend",
	CodeFriendAlreadyExist:      "friend already exist",
	CodeFriendRequestSent:       "friend request already sent",
	CodeFriendNotFound:          "friend not found",
	CodeFriendAddFailed:         "add friend failed",
	CodeFriendRequestNotFound:   "friend request not found",
	CodeFriendRequestInvalid:    "friend request status invalid",
	CodeFriendAcceptFailed:      "accept friend failed",
	CodeFriendCreateFailed:      "create friend relation failed",
	CodeFriendDeleteFailed:      "delete friend failed",
	CodeFriendListGetFailed:     "get friend list failed",
	CodeFriendRequestsGetFailed: "get friend requests failed",
	CodeUserSearchFailed:        "search user failed",
	CodeGroupCreateFailed:       "create group failed",
	CodeGroupAddOwnerFailed:     "add group owner failed",
	CodeGroupNotFound:           "group not found",
	CodeGroupMemberGetFailed:    "get group members failed",
	CodeGroupUserInfoGetFailed:  "get group user info failed",
	CodeGroupUnauthorized:       "no permission for group",
	CodeGroupMessageGetFailed:   "get group messages failed",
}

func GetMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return codeMsg[CodeFailed]
}
