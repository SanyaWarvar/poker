package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/SanyaWarvar/poker/pkg/auth"
	emailsmtp "github.com/SanyaWarvar/poker/pkg/email_smtp"
	"github.com/SanyaWarvar/poker/pkg/game"
	"github.com/SanyaWarvar/poker/pkg/handlers"
	"github.com/SanyaWarvar/poker/pkg/notifications"
	"github.com/SanyaWarvar/poker/pkg/server"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	err = generateStatics(db)
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

	repos := handlers.NewRepository(db, cacheDb, emailCfg, jwtCfg)
	services := handlers.NewService(repos)
	lt := game.NewLobbyTracker(services.HoldemService)
	o := game.NewWsObserver()
	b := game.NewBalanceObserver(services.UserService)
	engine := game.NewHoldemEngine(
		services.HoldemService,
		o,
		b,
		lt,
	)
	h := handlers.NewHandler(services, engine)
	srv := server.NewServer(h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	go engine.Lt.GameMonitor(game.DefaultTTS)
	srv.Run(port)
}

type StaticFile struct {
	Filename     string `db:"file_path"`
	FileAsString string `db:"file_data"`
	File         []byte
}

func generateStatics(db *sqlx.DB) error {
	pic, err := os.Open("./user_data/profile_pictures/default_pic.jpg")
	if err != nil {
		return err
	}
	defaultPic, err := io.ReadAll(pic)
	if err != nil {
		return err
	}
	defaultBase64 := base64.RawStdEncoding.EncodeToString(defaultPic)
	var files []StaticFile

	query := `
		SELECT file_data, file_path FROM files
	`
	err = db.Select(&files, query)
	if err != nil {
		return err
	}
	fmt.Printf("Необходимо создать %d файлов\n", len(files))
	for ind, item := range files {
		if item.FileAsString == defaultBase64 {
			continue
		}
		files[ind].File, err = base64.RawStdEncoding.DecodeString(item.FileAsString)
		if err != nil {
			continue
		}

		os.WriteFile(files[ind].Filename, files[ind].File, 0755)
	}
	return nil
}

func generateNotifications(repo notifications.INotificationRepository, userId uuid.UUID) {
	notifyData := os.Getenv("NOTIFY_DATA")
	if notifyData == "" {
		fmt.Println("NOTIFY_DATA не установлена")
		return
	}

	var messages []string
	err := json.Unmarshal([]byte(notifyData), &messages)
	if err != nil {
		fmt.Printf("Ошибка парсинга NOTIFY_DATA: %v\n", err)
		return
	}
	for {

		n, err := repo.GetNotifyCount(userId)
		if err != nil {
			log.Warnf("notify gen: %s", n)
			continue
		}

		if n < 5 {
			for ind := range messages {
				repo.CreateNotification(notifications.Notification{
					Id:         uuid.New(),
					UserId:     userId,
					Payload:    messages[ind],
					LastSendAt: time.Now().Add(-1 * 30 * time.Second),
					Readed:     false,
				})
				time.Sleep(time.Second * 5)
			}
		}

	}
}
