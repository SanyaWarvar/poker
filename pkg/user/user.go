package user

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/google/uuid"
)

type User struct {
	Id            uuid.UUID
	Username      string `json:"username" binding:"required" db:"username"`
	Email         string `json:"email" binding:"required" db:"email"`
	Password      string `json:"password" binding:"required" db:"password_hash"`
	ProfilePic    string `db:"profile_picture"`
	ProfilePicUrl string `json:"profile_picture_url"`
	Balance       int    `json:"balance" db:"balance"`
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

func (u *User) GenerateUrl(host string) {
	u.ProfilePicUrl = fmt.Sprintf("%s/profiles/%s", host, u.Username)
}

func (u *User) SetDeafultPic() error {
	file, err := os.OpenFile("user_data/profile_pictures/default_pic.jpg", os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	fileString := base64.RawStdEncoding.EncodeToString(fileBytes)
	u.ProfilePic = fileString
	return nil
}
