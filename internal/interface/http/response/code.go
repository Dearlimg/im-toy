package response

const (
	CodeSuccess            = 0
	CodeFailed             = -1
	CodeInvalidParams      = 40001
	CodeUnauthorized       = 40101
	CodeForbidden          = 40301
	CodeNotFound           = 40401
	CodeInternalError      = 50001
	CodeUserAlreadyExist   = 10001
	CodeUserNotFound       = 10002
	CodeWrongPassword      = 10003
	CodeFriendAlreadyExist = 20001
	CodeFriendRequestSent  = 20002
	CodeFriendNotFound     = 20003
)

var codeMsg = map[int]string{
	CodeSuccess:            "success",
	CodeFailed:             "failed",
	CodeInvalidParams:      "invalid params",
	CodeUnauthorized:       "unauthorized",
	CodeForbidden:          "forbidden",
	CodeNotFound:           "not found",
	CodeInternalError:      "internal error",
	CodeUserAlreadyExist:   "user already exist",
	CodeUserNotFound:       "user not found",
	CodeWrongPassword:      "wrong password",
	CodeFriendAlreadyExist: "friend already exist",
	CodeFriendRequestSent:  "friend request already sent",
	CodeFriendNotFound:     "friend not found",
}

func GetMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return codeMsg[CodeFailed]
}
