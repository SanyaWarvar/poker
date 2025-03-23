package user

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// User представляет собой модель пользователя.
// @Schema
type User struct {
	Id               uuid.UUID `json:"-"`
	Username         string    `json:"username" binding:"required" db:"username"`
	Email            string    `json:"email" binding:"required" db:"email"`
	Password         string    `json:"-" binding:"required" db:"password_hash"`
	ProfilePic       string    `json:"-" db:"profile_picture"`
	ProfilePicUrl    string    `json:"profile_picture_url"`
	Balance          int       `json:"balance" db:"balance"`
	IsEmailConfirmed bool      `json:"-" db:"confirmed_email"`
	PicExt           string    `json:"-" db:"pic_ext"`
}

const (
	UsernamePattern = "^[-a-zA-Z0-9_#$&*]+$"
	PasswordPattern = "^[-a-zA-Z0-9_#$&*]+$"
	UsernameMaxLen  = 32
	UsernameMinLen  = 4
	PasswordMaxLen  = 32
	PasswordMinLen  = 8
)

func (u *User) IsValid() bool {
	matched, err := regexp.Match(UsernamePattern, []byte(u.Username))
	usernameLen := len([]rune(u.Username))
	passwordLen := len([]rune(u.Password))
	if err != nil || !matched {
		return false
	}

	matched, err = regexp.Match(PasswordPattern, []byte(u.Password))
	if err != nil || !matched {
		return false
	}

	if (usernameLen <= UsernameMaxLen && usernameLen >= UsernameMinLen) &&
		(passwordLen <= PasswordMaxLen && passwordLen >= PasswordMinLen) {
		return true
	}
	return false
}

func CheckUsername(username string) bool {
	matched, err := regexp.Match(UsernamePattern, []byte(username))
	usernameLen := len([]rune(username))
	if err != nil || !matched {
		return false
	}

	if usernameLen <= UsernameMaxLen && usernameLen >= UsernameMinLen {
		return true
	}
	return false
}

func (u *User) GenerateUrl(host string) error {
	filename := strings.Split(u.ProfilePic, "/")
	u.ProfilePicUrl = fmt.Sprintf("%s/profiles/%s", host, filename[len(filename)-1])
	return nil
}
