package user

import (
	"errors"

	"github.com/google/uuid"
)

type IUserService interface {
	CreateUser(user User) error
	GetUserByUP(user User) (User, error)
	GetUserByEP(email, password string) (User, error)
	HashPassword(password string) (string, error)
	GetUserById(userId uuid.UUID) (User, error)
	UpdateProfilePic(userId uuid.UUID, path string) error
}

type UserService struct {
	repo IUserRepo
}

func NewUserService(repo IUserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(user User) error {
	var err error
	user.Password, err = s.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Id, err = uuid.NewUUID()
	if err != nil {
		return err
	}
	err = user.SetDeafultPic()
	if err != nil {
		return err
	}
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUserByUP(user User) (User, error) {
	targetUser, err := s.repo.GetUserByU(user.Username)
	if err != nil {
		return user, err
	}

	if s.repo.ComparePassword(user.Password, targetUser.Password) {
		return targetUser, err
	}
	return user, errors.New("incorrect password")
}

func (s *UserService) GetUserByEP(email, password string) (User, error) {
	var user User
	targetUser, err := s.repo.GetUserByE(email)
	if err != nil {
		return user, err
	}

	if s.repo.ComparePassword(password, targetUser.Password) {
		return targetUser, err
	}
	return user, errors.New("incorrect password")
}

func (s *UserService) HashPassword(password string) (string, error) {
	return s.repo.HashPassword(password)
}

func (s *UserService) UpdateProfilePic(userId uuid.UUID, path string) error {
	return s.repo.UpdateProfilePic(userId, path)
}

func (s *UserService) GetUserById(userId uuid.UUID) (User, error) {
	return s.repo.GetUserById(userId)
}
