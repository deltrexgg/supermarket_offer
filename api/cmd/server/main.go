package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/config"
	"github.com/deltrexgg/supermarket_offer/api/internal/db"
	"github.com/deltrexgg/supermarket_offer/api/internal/httpapi"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/render"
)

func main() {
	cfg := config.Load()
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := gdb.AutoMigrate(models.All()...); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	rdb, err := cache.Connect(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	rn, err := render.New(cfg.ChromePath)
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}
	defer rn.Close()

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: httpapi.NewRouter(httpapi.Deps{
			DB: gdb, Redis: rdb, Renderer: rn, CORSOrigins: cfg.CORSOrigins,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
