package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IUserService interface {
	CreateUser(user User) error
	GetUserByUP(user User) (User, error)
	GetUserByEP(email, password string) (User, error)
	HashPassword(password string) (string, error)
	GetUserById(userId uuid.UUID) (User, error)
	GetUserByUsername(username string) (User, error)
	UpdateProfilePic(userId uuid.UUID, picture []byte, filename string) error
	UpdateUsername(userId uuid.UUID, username string) error // будем обновлять именно эту инфу.
	GetDaily(userId uuid.UUID) (DailyReward, error)
	ChangeBalance(userId uuid.UUID, delta int) error //TODO
}

type UserService struct {
	repo  IUserRepo
	cache IUserCacheRepo
}

func NewUserService(repo IUserRepo, cache IUserCacheRepo) *UserService {
	return &UserService{repo: repo, cache: cache}
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

func (s *UserService) UpdateProfilePic(userId uuid.UUID, picture []byte, filename string) error {
	encodedPicture := base64.RawStdEncoding.EncodeToString(picture)
	err := s.repo.SaveProfilePic(userId, picture, filename)
	if err != nil {
		return nil
	}
	return s.repo.UpdateProfilePic(userId, encodedPicture)
}

func (s *UserService) GetUserById(userId uuid.UUID) (User, error) {
	return s.repo.GetUserById(userId)
}

func (s *UserService) UpdateUsername(userId uuid.UUID, username string) error {
	return s.repo.UpdateUsername(userId, username)
}

func (s *UserService) GetUserByUsername(username string) (User, error) {
	return s.repo.GetUserByUsername(username)
}

func (s *UserService) GetDaily(userId uuid.UUID) (DailyReward, error) {
	var output DailyReward
	lastTime, err := s.cache.GetLastDailyReward(userId)
	if err != nil && !errors.Is(err, redis.Nil) {
		return output, err
	}
	fmt.Println(lastTime, time.Now().After(lastTime.Add(time.Hour*24)))
	if !time.Now().After(lastTime.Add(time.Hour * 24)) {
		return output, errors.New(
			fmt.Sprintf("next possible daily reward will available at %s", lastTime.Add(time.Hour*24).Format(time.UnixDate)),
		)
	}
	output = SpinWheel()
	err = s.cache.SaveLastDailyReward(userId)
	if err != nil {
		return output, err
	}
	err = s.ChangeBalance(userId, output.Amount)
	if err != nil {
		return output, err
	}
	return output, nil
}

func (s *UserService) ChangeBalance(userId uuid.UUID, delta int) error {
	return s.repo.ChangeBalance(userId, delta)
}
