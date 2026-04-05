package di

import (
	"im-chat/internal/application/service"
	"im-chat/internal/application/service/impl"
	domain_repo "im-chat/internal/domain/repository"
	"im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/handler"
)

type Container struct {
	UserRepo        domain_repo.UserRepository
	FriendRepo      domain_repo.FriendRepository
	GroupRepo       domain_repo.GroupRepository
	MsgRepo         domain_repo.PrivateMessageRepository
	ReadRepo        domain_repo.UserRoomReadRepository
	MessageReadRepo domain_repo.MessageReadRepository

	UserService   service.UserService
	FriendService service.FriendService
	GroupService  service.GroupService

	UserHandler       *handler.UserHandler
	FriendHandler     *handler.FriendHandler
	GroupHandler      *handler.GroupHandler
	PrivateMsgHandler *handler.PrivateMessageHandler
	MessageHandler    *handler.MessageHandler
}

func NewContainer() *Container {
	c := &Container{}

	c.UserRepo = repository.NewUserRepository()
	c.FriendRepo = repository.NewFriendRepository()
	c.GroupRepo = repository.NewGroupRepository()
	c.MsgRepo = repository.NewPrivateMessageRepository()
	c.ReadRepo = repository.NewUserRoomReadRepository()
	c.MessageReadRepo = repository.NewMessageReadRepository()

	c.UserService = impl.NewUserServiceImpl(c.UserRepo)
	c.FriendService = impl.NewFriendServiceImpl(c.FriendRepo, c.UserRepo)
	c.GroupService = impl.NewGroupServiceImpl(c.GroupRepo, c.MessageReadRepo, c.UserRepo)

	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.FriendHandler = handler.NewFriendHandler(c.FriendService)
	c.GroupHandler = handler.NewGroupHandler(c.GroupService)
	c.PrivateMsgHandler = handler.NewPrivateMessageHandler(c.MsgRepo, c.ReadRepo)
	c.MessageHandler = handler.NewMessageHandler(c.ReadRepo)

	return c
}
