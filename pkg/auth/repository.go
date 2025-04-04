package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type JwtManagerCfg struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	SigningKey      string
	Method          jwt.SigningMethod
}

type IJwtManagerRepo interface {
	GenerateAccessToken(userId, refreshId uuid.UUID) (string, error)
	GenerateRefreshToken(userId uuid.UUID) (string, error)
	GeneratePairToken(userId uuid.UUID) (string, string, uuid.UUID, error)
	CompareTokens(hashedToken, token string) bool
	HashToken(refreshToken string) (string, error)
	SaveRefreshToken(hashedToken string, tokenId, userId uuid.UUID) error
	DeleteRefreshTokenById(tokenId uuid.UUID) error
	GetRefreshTokenById(tokenId uuid.UUID) (string, error)
	ParseToken(accessToken string, expCheck bool) (*AccessTokenClaims, error)
	CheckRefreshTokenExp(tokenId uuid.UUID) bool
	GetTokensTtl() (time.Duration, time.Duration)
}

type JwtManagerPostgres struct {
	db     *sqlx.DB
	config *JwtManagerCfg
}

func NewJwtManagerCfg(AccessTokenTTL, RefreshTokenTTL time.Duration, SigningKey string, Method jwt.SigningMethod) *JwtManagerCfg {
	return &JwtManagerCfg{AccessTokenTTL: AccessTokenTTL, RefreshTokenTTL: RefreshTokenTTL, SigningKey: SigningKey, Method: Method}
}

func NewJwtManagerPostgres(db *sqlx.DB, cfg *JwtManagerCfg) *JwtManagerPostgres {
	return &JwtManagerPostgres{db: db, config: cfg}
}

func (m *JwtManagerPostgres) GenerateAccessToken(userId, refreshId uuid.UUID) (string, error) {
	jwtClaims := AccessTokenClaims{
		UserId:    userId,
		RefreshId: refreshId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.config.AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(
		m.config.Method,
		jwtClaims,
	)

	return token.SignedString([]byte(m.config.SigningKey))
}

func (m *JwtManagerPostgres) GenerateRefreshToken(userId uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

func (m *JwtManagerPostgres) GeneratePairToken(userId uuid.UUID) (string, string, uuid.UUID, error) {
	refreshId := uuid.New()
	refreshToken, err := m.GenerateRefreshToken(userId)
	if err != nil {
		return "", "", refreshId, err
	}

	accessToken, err := m.GenerateAccessToken(userId, refreshId)
	if err != nil {
		return "", "", refreshId, err
	}

	return accessToken, refreshToken, refreshId, err
}

func (m *JwtManagerPostgres) SaveRefreshToken(hashedToken string, tokenId, userId uuid.UUID) error {
	query := fmt.Sprintf(`INSERT INTO tokens (id, user_id, token, exp_date) VALUES ($1, $2, $3, $4)`)
	expDate := time.Now().Add(m.config.RefreshTokenTTL)
	_, err := m.db.Exec(query, tokenId, userId, hashedToken, expDate)
	return err
}

func (m *JwtManagerPostgres) GetRefreshTokenById(tokenId uuid.UUID) (string, error) {
	var tokenHash string
	query := fmt.Sprintf(`SELECT token FROM tokens WHERE id=$1`)
	err := m.db.Get(&tokenHash, query, tokenId)
	return tokenHash, err
}

func (m *JwtManagerPostgres) CompareTokens(hashedToken, token string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token)) == nil
}

func (m *JwtManagerPostgres) HashToken(token string) (string, error) {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	return string(hashedToken), err
}

func (m *JwtManagerPostgres) ParseToken(accessToken string, expCheck bool) (*AccessTokenClaims, error) {
	var opts []jwt.ParserOption

	if !expCheck {
		opts = append(opts, jwt.WithoutClaimsValidation())
	}
	parsedToken, err := jwt.ParseWithClaims(accessToken, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(m.config.SigningKey), nil
	}, opts...)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JwtManagerPostgres) DeleteRefreshTokenById(tokenId uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM tokens WHERE id=$1`)
	_, err := m.db.Exec(query, tokenId)
	return err
}

func (m *JwtManagerPostgres) CheckRefreshTokenExp(tokenId uuid.UUID) bool {
	var expDate time.Time
	query := fmt.Sprintf(`SELECT exp_date FROM tokens WHERE id=$1`)
	err := m.db.Get(&expDate, query, tokenId)
	if err != nil {
		return false
	}

	return expDate.After(time.Now())
}

func (m *JwtManagerPostgres) GetTokensTtl() (time.Duration, time.Duration) {
	accessTokenTtl, refreshTokenTtl := m.config.AccessTokenTTL, m.config.RefreshTokenTTL
	return accessTokenTtl, refreshTokenTtl
}
