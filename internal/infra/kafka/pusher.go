package kafka

type Pusher interface {
	PushToUser(uid string, msg map[string]interface{})
}
