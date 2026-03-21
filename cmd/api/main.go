package main

import (
	_ "github.com/georgifotev1/nuvelaone-api/docs"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/georgifotev1/nuvelaone-api/config"
	"github.com/georgifotev1/nuvelaone-api/pkg/database"
	"github.com/georgifotev1/nuvelaone-api/pkg/logger"
	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/georgifotev1/nuvelaone-api/pkg/ratelimiter"
	"github.com/georgifotev1/nuvelaone-api/pkg/redis"
)

// @title           NuvelaOne API
// @version         1.0.0
// @description     REST API for NuvelaOne application
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.NewZapLogger(cfg.Env == "development")
	defer log.Sync()

	db, err := database.Connect(cfg.DB.URL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	log.Info("database connection established")

	redisClient := redis.NewRedisClient(cfg.Redis.Addr, cfg.Redis.PW, cfg.Redis.DB)
	defer redisClient.Close()
	log.Info("redis connection established")

	var rateLimiter ratelimiter.Limiter
	if cfg.RateLimiter.Enabled {
		rateLimiter = ratelimiter.NewFixedWindowLimiter(
			cfg.RateLimiter.RequestsPerTimeFrame,
			cfg.RateLimiter.TimeFrame,
		)
		log.Info("rate limiter enabled")
	}

	var mail mailer.Mailer
	if cfg.Resend.APIKey != "" && cfg.Resend.FromEmail != "" {
		mail = mailer.NewResendClient(cfg.Resend.APIKey, cfg.Resend.FromEmail)
		log.Info("email mailer initialized")
	}

	app := &application{
		config:      cfg,
		db:          db,
		redis:       redisClient,
		rateLimiter: rateLimiter,
		mailer:      mail,
		logger:      log,
	}

	if err := app.run(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
