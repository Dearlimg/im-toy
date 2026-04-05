package impl

import (
	"im-chat/internal/apperrors"
	"im-chat/internal/application/dto"
	"im-chat/internal/application/service"
	"im-chat/internal/domain/entity"
	"im-chat/internal/domain/repository"
	"im-chat/internal/infra/idgen"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/interface/http/response"

	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	userRepo repository.UserRepository
}

func NewUserServiceImpl(userRepo repository.UserRepository) service.UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

func (s *UserServiceImpl) Register(req *dto.UserRegisterReq) (*dto.UserLoginResp, error) {
	existingUser, _ := s.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		return nil, apperrors.New(response.CodeUserAlreadyExist, response.GetMsg(response.CodeUserAlreadyExist))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.New(response.CodePasswordHashFailed, response.GetMsg(response.CodePasswordHashFailed))
	}

	uid := idgen.Gen.NextUID()
	user := &entity.User{
		UID:      uid,
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, apperrors.New(response.CodeUserRegisterFailed, response.GetMsg(response.CodeUserRegisterFailed))
	}

	token, err := jwt.GenerateToken(uid)
	if err != nil {
		return nil, apperrors.New(response.CodeTokenGenerateFailed, response.GetMsg(response.CodeTokenGenerateFailed))
	}

	return &dto.UserLoginResp{
		Token: token,
		User: dto.UserResp{
			UID:      uid,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	}, nil
}

func (s *UserServiceImpl) Login(req *dto.UserLoginReq) (*dto.UserLoginResp, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, apperrors.New(response.CodeUserNotFound, response.GetMsg(response.CodeUserNotFound))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.New(response.CodeWrongPassword, response.GetMsg(response.CodeWrongPassword))
	}

	if user.Status != 1 {
		return nil, apperrors.New(response.CodeUserDisabled, response.GetMsg(response.CodeUserDisabled))
	}

	token, err := jwt.GenerateToken(user.UID)
	if err != nil {
		return nil, apperrors.New(response.CodeTokenGenerateFailed, response.GetMsg(response.CodeTokenGenerateFailed))
	}

	return &dto.UserLoginResp{
		Token: token,
		User: dto.UserResp{
			UID:      user.UID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	}, nil
}
