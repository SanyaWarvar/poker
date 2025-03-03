package emailsmtp

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type IEmailSmtpService interface {
	CheckEmailConfirm(email string) (bool, error)
	ConfirmEmail(email, code string) error
	SendConfirmEmailMessage(email string) error
	SendMessage(email, messageText, title string) error
	GenerateConfirmCode() string
}

type EmailSmtpService struct {
	repo  IEmailSmtpRepo
	cache IEmailCacheRepo
}

func NewEmailSmtpService(repo IEmailSmtpRepo, cacheRepo IEmailCacheRepo) *EmailSmtpService {
	return &EmailSmtpService{
		repo:  repo,
		cache: cacheRepo,
	}
}

func (s *EmailSmtpService) SendMessage(email, messageText, title string) error {
	return s.repo.SendMessage(email, messageText, title)
}

func (s *EmailSmtpService) SendConfirmEmailMessage(email string) error {
	minTtl, _ := time.ParseDuration(os.Getenv("MIN_TTL"))
	maxTtl, _ := time.ParseDuration(os.Getenv("CODE_EXP"))

	_, ttl, err := s.cache.GetConfirmCode(email)

	if err == nil && minTtl < ttl {
		return errors.New(fmt.Sprintf("Сode has already been sent %s ago", maxTtl-ttl))
	}

	code := s.GenerateConfirmCode()
	s.cache.SaveConfirmCode(email, code)
	go func() {
		err = s.repo.SendConfirmEmailMessage(email, code)
		if err != nil {
			logrus.Errorf("error while sending confirm email message: %s", err.Error())
		}
	}()

	if err != nil && err.Error() == "redis: nil" {
		return nil
	}

	return err
}

func (s *EmailSmtpService) CheckEmailConfirm(email string) (bool, error) {
	return s.repo.CheckEmailConfirm(email)
}

func (s *EmailSmtpService) ConfirmEmail(email, code string) error {
	trueCode, _, err := s.cache.GetConfirmCode(email)
	if err != nil {
		return err
	}
	if trueCode != code {
		return errors.New("bad code")
	}
	return s.repo.ConfirmEmail(email)
}

func (s *EmailSmtpService) GetConfirmCode(email string) (string, time.Duration, error) {
	return s.cache.GetConfirmCode(email)
}

func (s *EmailSmtpService) GenerateConfirmCode() string {
	return s.repo.GenerateConfirmCode()
}
