package user

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	UpdateProfilePic(userId uuid.UUID, encodedPicture, filepath string) error
	UpdateUsername(userId uuid.UUID, username string) error
	GetUserByUsername(username string) (User, error)
	SaveProfilePic(userId uuid.UUID, picture []byte, filename string) error
	ChangeBalance(userId uuid.UUID, delta int) error
	GetPlayersByIdLIst(idList []uuid.UUID) ([]User, error)
	UpdateManyUserBalance(userId []uuid.UUID, newBalance []int) error
	IncGameCount(playerId uuid.UUID) error
	UpdateMaxBalance(playerId uuid.UUID) error
	GetStatsByU(username string) (PlayerStats, error)
}

type UserPostgres struct {
	db *sqlx.DB
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

func (r *UserPostgres) CreateUser(user User) error {
	id := uuid.NewString()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	query := fmt.Sprint(`INSERT INTO users (id, username, email, password_hash, profile_picture) VALUES ($1, $2, $3, $4, $5)`)
	_, err = tx.Exec(query, id, user.Username, user.Email, user.Password, user.ProfilePic)
	if err != nil {
		tx.Rollback()
		return err
	}
	query = fmt.Sprint(`INSERT INTO player_stats(user_id) values($1)`)
	_, err = tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
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

func (r *UserPostgres) UpdateProfilePic(userId uuid.UUID, encodedPicture, filepath string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	query1 := fmt.Sprintf(
		`
		UPDATE users
		SET profile_picture = $1
		WHERE id = $2
		`,
	)
	_, err = tx.Exec(query1, filepath, userId.String())
	if err != nil {
		tx.Rollback()
		return err
	}
	fmt.Println(userId)
	query2 := fmt.Sprintf(
		`
		DELETE FROM files WHERE file_path like $1
		`,
	)
	_, err = tx.Exec(query2, userId.String()+"%")
	if err != nil {
		tx.Rollback()
		return err
	}

	query3 := fmt.Sprintf(
		`
		INSERT INTO files VALUES
		($1, $2)
		`,
	)

	_, err = tx.Exec(query3, encodedPicture, filepath)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
func (r *UserPostgres) UpdateUsername(userId uuid.UUID, username string) error {
	query := fmt.Sprintf(
		`
		UPDATE users
		SET username = $1
		WHERE id = $2
		`,
	)

	_, err := r.db.Exec(query, username, userId)
	return err
}
func (r *UserPostgres) GetUserByUsername(username string) (User, error) {
	var output User
	query := fmt.Sprintf(
		`
		SELECT * 
		FROM users 
		WHERE username = $1
		`,
	)

	err := r.db.Get(&output, query, username)
	return output, err
}

func (r *UserPostgres) SaveProfilePic(userId uuid.UUID, picture []byte, filename string) error {
	targetName := userId.String()
	filepath.Walk("./user_data/profile_pictures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileNameWithoutExt := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if fileNameWithoutExt == targetName {
				os.Remove(path)
			}
		}
		return nil
	})
	return os.WriteFile("user_data/profile_pictures/"+filename, picture, 0644)
}

func (r *UserPostgres) ChangeBalance(userId uuid.UUID, delta int) error {
	query := fmt.Sprintf(
		`
		UPDATE users SET balance = balance + $1
		WHERE (($1 >= 0) OR ($1 < 0 AND $1 <= balance)) AND id = $2
		`,
	)
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: false})
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, delta, userId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func (r *UserPostgres) GetPlayersByIdLIst(idList []uuid.UUID) ([]User, error) {
	var output []User
	query := fmt.Sprintf(
		`
		SELECT * FROM users
		WHERE id = any($1)
		`,
	)
	err := r.db.Select(&output, query, pq.Array(idList))

	return output, err
}
func (r *UserPostgres) UpdateManyUserBalance(userId []uuid.UUID, newBalance []int) error {
	if len(userId) != len(newBalance) {
		return fmt.Errorf("invalid arrays lenght")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Preparex(`
        UPDATE users 
        SET balance = $1 
        WHERE id = $2
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, userId := range userId {
		_, err := stmt.Exec(newBalance[i], userId)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil

}

func (r *UserPostgres) IncGameCount(playerId uuid.UUID) error {
	query := `
	UPDATE player_stats
	SET games_played = games_played + 1
	WHERE user_id = $1
	`
	_, err := r.db.Exec(query, playerId)
	return err
}

func (r *UserPostgres) UpdateMaxBalance(playerId uuid.UUID) error {
	query := `
	UPDATE player_stats
	SET max_balance = max(
	(select balance from users where user_id = $1)
	, max_balance)
	WHERE user_id = $1
	`
	_, err := r.db.Exec(query, playerId)
	return err
}

func (r *UserPostgres) GetStatsByU(username string) (PlayerStats, error) {
	var output PlayerStats
	query := `
	select * from player_stats where user_id = (select id from users where username = $1)
	`
	err := r.db.Get(&output, query, username)
	return output, err
}
