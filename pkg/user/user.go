package user

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	targetName := u.Id.String()
	var foundFile string

	err := filepath.Walk("./user_data/profile_pictures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileNameWithoutExt := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if fileNameWithoutExt == targetName {
				foundFile = filepath.Ext(info.Name())
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("ошибка при обходе директории: %w", err)
	}

	if foundFile == "" {
		u.ProfilePicUrl = fmt.Sprintf("%s/profiles/default_pic.jpg", host)
		return nil
	}
	u.ProfilePicUrl = fmt.Sprintf("%s/profiles/%s%s", host, u.Id.String(), foundFile)
	return nil
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
