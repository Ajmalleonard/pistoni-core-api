package server

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"pistoni/core/internal/config"
	authhandler "pistoni/core/internal/modules/auth/handler"
	authrepo "pistoni/core/internal/modules/auth/repo"
	authservice "pistoni/core/internal/modules/auth/service"
	carhandler "pistoni/core/internal/modules/cars/handler"
	carrepo "pistoni/core/internal/modules/cars/repo"
	carservice "pistoni/core/internal/modules/cars/service"
	chatgateway "pistoni/core/internal/modules/chat/gateway"
	chathandler "pistoni/core/internal/modules/chat/handler"
	chatqueue "pistoni/core/internal/modules/chat/queue"
	chatrepo "pistoni/core/internal/modules/chat/repo"
	chatservice "pistoni/core/internal/modules/chat/service"
	notificationshandler "pistoni/core/internal/modules/notifications/handler"
	notificationsrepo "pistoni/core/internal/modules/notifications/repo"
	notificationsservice "pistoni/core/internal/modules/notifications/service"
	paymentmethodshandler "pistoni/core/internal/modules/paymentmethods/handler"
	paymentmethodsrepo "pistoni/core/internal/modules/paymentmethods/repo"
	paymentmethodsservice "pistoni/core/internal/modules/paymentmethods/service"
	providershandler "pistoni/core/internal/modules/providers/handler"
	providersrepo "pistoni/core/internal/modules/providers/repo"
	providersservice "pistoni/core/internal/modules/providers/service"
	rentalshandler "pistoni/core/internal/modules/rentals/handler"
	rentalsrepo "pistoni/core/internal/modules/rentals/repo"
	rentalsservice "pistoni/core/internal/modules/rentals/service"
	usershandler "pistoni/core/internal/modules/users/handler"
	usersrepo "pistoni/core/internal/modules/users/repo"
	usersservice "pistoni/core/internal/modules/users/service"
	"pistoni/core/internal/platform/auth"
	"pistoni/core/internal/platform/guard"
	"pistoni/core/internal/platform/mailer"
	"pistoni/core/internal/platform/middleware"
	"pistoni/core/internal/platform/model"
	platformredis "pistoni/core/internal/platform/redis"
	"pistoni/core/internal/platform/storage"
)

type Server struct {
	engine *gin.Engine
	cfg    config.Config
}

func New(cfg config.Config) (*Server, error) {
	r := gin.New()
	rateLimiter := guard.NewIPRateLimiter()
	r.Use(gin.Logger(), gin.Recovery(), rateLimiter.Middleware(2400, 600, 1*time.Minute))

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://core.api.pistoni.io",
			"https://pistoni.io",
			"https://www.pistoni.io",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store := storage.NewMemoryStore()
	userRepo := usersrepo.NewMemoryRepository(store)
	seedAdmin(userRepo)

	userSvc := usersservice.New(userRepo)
	usersH := usershandler.New(userSvc)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	var emailer authservice.EmailSender
	if m, err := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFromName, cfg.SMTPFromEmail, cfg.AppName); err == nil {
		emailer = m
		log.Printf("[mailer] smtp enabled host=%s port=%s from=%s app=%s", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFromEmail, cfg.AppName)
	} else {
		log.Printf("[mailer] smtp disabled, fallback to noop sender: %v", err)
	}
	authSvc := authservice.New(userRepo, authrepo.NewMemoryPendingRepository(store), jwtManager, emailer)
	authH := authhandler.New(authSvc)
	authRequired := auth.AuthRequired(jwtManager, func(userID string) (any, error) { return authSvc.UserForAuth(userID) })
	adminOnly := auth.RequireRoles(string(model.UserRoleAdmin))

	carSvc := carservice.New(carrepo.NewMemoryRepository(store, userRepo))
	carsH := carhandler.New(carSvc)

	providersSvc := providersservice.New(providersrepo.NewMemoryRepository(store, userRepo))
	providersH := providershandler.New(providersSvc)
	paymentMethodsSvc := paymentmethodsservice.New(paymentmethodsrepo.NewMemoryRepository(store))
	paymentMethodsH := paymentmethodshandler.New(paymentMethodsSvc)
	rentalsSvc := rentalsservice.New(rentalsrepo.NewMemoryRepository(store))
	rentalsH := rentalshandler.New(rentalsSvc)
	notificationsSvc := notificationsservice.New(notificationsrepo.NewMemoryRepository(store))
	notificationsH := notificationshandler.New(notificationsSvc)
	chatSvc := chatservice.New(chatrepo.NewMemoryRepository(store, userRepo))
	chatH := chathandler.New(chatSvc)

	redisConn := platformredis.Connect(cfg)
	var signalPublisher chatqueue.QueuePublisher
	if redisConn.Enabled {
		signalPublisher = chatqueue.NewRedisStreamPublisher(redisConn.Client, cfg.RedisQueueStream)
		log.Printf("[redis] connected endpoint=%s stream=%s", redisConn.Endpoint, cfg.RedisQueueStream)
	} else {
		signalPublisher = chatqueue.NewNoopPublisher()
		log.Printf("[redis] degraded mode enabled: %s", redisConn.Reason)
	}

	chatSocket, err := chatgateway.New(chatSvc, jwtManager, cfg.SocketPath, signalPublisher)
	if err != nil {
		return nil, err
	}
	sensitiveQueue := middleware.AdmissionQueue(24, 200, 2*time.Second)
	retryBackoff := guard.NewRetryBackoffLimiter().Middleware(60*time.Second, 15*time.Minute)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "pistoni-core-go"}) })
	authH.Register(r.Group("/auth"), authRequired)
	carsH.Register(r.Group("/cars"))
	usersH.Register(r.Group("/users"))
	providersH.Register(r.Group("/providers"), authRequired, adminOnly, sensitiveQueue, retryBackoff)
	paymentMethodsH.Register(r.Group("/payment-methods"))
	rentalsH.Register(r.Group("/rentals"))
	notificationsH.Register(r.Group("/notifications"))
	chatH.Register(r.Group("/chat"))

	go func() {
		if err := chatSocket.Server().Serve(); err != nil {
			log.Printf("[socketio] serve error: %v", err)
		}
	}()
	r.Any(cfg.SocketPath+"/*any", gin.WrapH(chatSocket.Server()))
	log.Printf("[socketio] enabled path=%s namespace=/ shared_http_port=%s socket_port=%s", cfg.SocketPath, cfg.Port, cfg.SocketPort)

	return &Server{engine: r, cfg: cfg}, nil
}

func (s *Server) Run() error { return s.engine.Run(fmt.Sprintf(":%s", s.cfg.Port)) }

func seedAdmin(userRepo usersrepo.Repository) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	pw := string(hash)
	_, _ = userRepo.Create(model.User{Email: "admin@pistoni.local", Password: &pw, Name: "Admin", Role: model.UserRoleAdmin, AuthProvider: model.AuthProviderEmail, IsVerified: true})
}
