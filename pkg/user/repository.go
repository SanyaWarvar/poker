package user

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type IUserRepo interface {
	CreateUser(user User) error
	GetUserByUP(username, hashedPassword string) (User, error)
	GetUserByEP(email, hashedPassword string) (User, error)
	GetUserById(userId uuid.UUID) (User, error)
	GetUserByU(username string) (User, error)
	GetUserByE(email string) (User, error)
	HashPassword(password string) (string, error)
	ComparePassword(password, hashedPassword string) bool
	UpdateProfilePic(userId uuid.UUID, path string) error
}

type UserPostgres struct {
	db *sqlx.DB
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

func (r *UserPostgres) CreateUser(user User) error {
	id := uuid.NewString()
	query := fmt.Sprint(`INSERT INTO users (id, username, email, password_hash, profile_picture) VALUES ($1, $2, $3, $4, $5)`)
	_, err := r.db.Exec(query, id, user.Username, user.Email, user.Password, user.ProfilePic)
	return err
}

func (r *UserPostgres) GetUserByUP(username, hashedPassword string) (User, error) {
	var user User
	query := fmt.Sprint(`SELECT * FROM users WHERE username = $1 AND password_hash = $2`)
	err := r.db.Get(&user, query, username, hashedPassword)
	return user, err
}

func (r *UserPostgres) GetUserByU(username string) (User, error) {
	var user User
	query := fmt.Sprint(`SELECT * FROM users WHERE username = $1`)
	err := r.db.Get(&user, query, username)
	return user, err
}

func (r *UserPostgres) GetUserByE(email string) (User, error) {
	var user User
	query := fmt.Sprintf(`SELECT * FROM users WHERE email = $1`)
	err := r.db.Get(&user, query, email)
	return user, err
}

func (r *UserPostgres) GetUserByEP(email, hashedPassword string) (User, error) {
	var user User
	query := fmt.Sprintf(`SELECT * FROM users WHERE email = $1 AND password_hash = $2`)
	err := r.db.Get(&user, query, email, hashedPassword)
	return user, err
}

func (r *UserPostgres) GetUserById(userId uuid.UUID) (User, error) {
	var user User
	query := fmt.Sprintf(`SELECT *, username, email, password_hash FROM users WHERE id = $1`)
	err := r.db.Get(&user, query, userId)
	return user, err
}

func (m *UserPostgres) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (r *UserPostgres) ComparePassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (r *UserPostgres) UpdateProfilePic(userId uuid.UUID, path string) error {

	query := fmt.Sprintf(
		`
		UPDATE users
		SET profile_picture = $1
		WHERE user_id = $2
		`,
	)

	_, err := r.db.Exec(query, path, userId)
	return err
}
