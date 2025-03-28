package auth

import (
	"time"

	"github.com/google/uuid"
)

type IJwtManagerService interface {
	GeneratePairToken(userId uuid.UUID) (string, string, uuid.UUID, error)
	CompareTokens(refreshId uuid.UUID, token string) bool
	SaveRefreshToken(hashedToken string, userId, tokenId uuid.UUID) error
	DeleteRefreshTokenById(tokenId uuid.UUID) error
	GetRefreshTokenById(tokenId uuid.UUID) (string, error)
	ParseToken(accessToken string, expCheck bool) (*AccessTokenClaims, error)
	CheckRefreshTokenExp(tokenId uuid.UUID) bool
	GetTokensTtl() (time.Duration, time.Duration)
}

type JwtManagerService struct {
	repo IJwtManagerRepo
}

func NewJwtManagerService(repo IJwtManagerRepo) *JwtManagerService {
	return &JwtManagerService{repo: repo}
}

func (s *JwtManagerService) ParseToken(accessToken string, expCheck bool) (*AccessTokenClaims, error) {
	return s.repo.ParseToken(accessToken, expCheck)
}

func (s *JwtManagerService) GeneratePairToken(userId uuid.UUID) (string, string, uuid.UUID, error) {
	var err3 error
	accessToken, refreshToken, refreshId, err := s.repo.GeneratePairToken(userId)
	if err == nil {
		refreshHash, err2 := s.repo.HashToken(refreshToken)
		if err2 != nil {
			return accessToken, refreshToken, refreshId, err2
		}
		err3 = s.repo.SaveRefreshToken(refreshHash, refreshId, userId)
	}
	return accessToken, refreshToken, refreshId, err3
}
func (s *JwtManagerService) CompareTokens(refreshId uuid.UUID, token string) bool {
	hashedToken, err := s.repo.GetRefreshTokenById(refreshId)
	if err != nil {
		return false
	}
	return s.repo.CompareTokens(hashedToken, token)
}

func (s *JwtManagerService) SaveRefreshToken(hashedToken string, userId, tokenId uuid.UUID) error {
	return s.repo.SaveRefreshToken(hashedToken, userId, tokenId)
}
func (s *JwtManagerService) GetRefreshTokenById(tokenId uuid.UUID) (string, error) {
	return s.repo.GetRefreshTokenById(tokenId)
}

func (s *JwtManagerService) DeleteRefreshTokenById(tokenId uuid.UUID) error {
	return s.repo.DeleteRefreshTokenById(tokenId)
}

func (s *JwtManagerService) CheckRefreshTokenExp(tokenId uuid.UUID) bool {
	return s.repo.CheckRefreshTokenExp(tokenId)
}

func (s *JwtManagerService) GetTokensTtl() (time.Duration, time.Duration) {
	return s.repo.GetTokensTtl()
}
