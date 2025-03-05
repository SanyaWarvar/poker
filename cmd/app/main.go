package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SanyaWarvar/poker/pkg/auth"
	emailsmtp "github.com/SanyaWarvar/poker/pkg/email_smtp"
	"github.com/SanyaWarvar/poker/pkg/server"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := godotenv.Load(".env"); err != nil {
		logrus.Fatalf("Error while load dotenv: %s", err.Error())
	}
	

	db, err := server.NewPostgresDB(server.PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("Error while create connection to db: %s", err.Error())
	}

	dbNum, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		logrus.Fatalf("Error while create connection to db: %s", err.Error())
	}
	redisOptions := redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("CACHE_PASSWORD"),
		DB:       dbNum,
	}
	codeExp, err := time.ParseDuration(os.Getenv("CODE_EXP"))
	if err != nil {
		logrus.Fatalf("Error while create connection to db: %s", err.Error())
	}
	cacheDb, err := server.NewRedisDb(&redisOptions)
	if err != nil {
		logrus.Fatalf("Error while create connection to cache: %s", err.Error())
	}
	codeLenght, err := strconv.Atoi(os.Getenv("CODE_LENGHT"))
	if err != nil {
		logrus.Fatalf("Error while create connection to cache: %s", err.Error())
	}
	emailCfg := emailsmtp.NewEmailCfg(
		os.Getenv("OWNER_EMAIL"),
		os.Getenv("OWNER_PASSWORD"),
		os.Getenv("SMTP_ADDR"),
		codeLenght,
		codeExp,
	)

	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESSTOKENTTL"))
	if err != nil {
		logrus.Fatalf("Errof while parse accessTokenTTL: %s", err.Error())
	}
	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESHTOKENTTL"))
	if err != nil {
		logrus.Fatalf("Errof while parse refreshTokenTTL: %s", err.Error())
	}
	jwtCfg := auth.NewJwtManagerCfg(accessTokenTTL, refreshTokenTTL, os.Getenv("SIGNINGKEY"), jwt.SigningMethodHS256)

	repos := server.NewRepository(db, cacheDb, emailCfg, jwtCfg)

	services := server.NewService(repos)
	srv := server.NewServer(services)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	srv.Run(port)
}
