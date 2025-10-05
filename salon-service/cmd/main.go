package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    sharedAuth "github.com/EricsAntony/salon/salon-shared/auth"
    sharedLogger "github.com/EricsAntony/salon/salon-shared/logger"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog/log"

    "salon-service/internal/api"
    "salon-service/internal/config"
    "salon-service/internal/repository"
    "salon-service/internal/service"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to load config")
    }

    sharedCfg := cfg.ToShared()

    sharedLogger.Init(sharedCfg)
    log.Info().Str("service", sharedCfg.Log.ServiceName).Msg("starting service")

    pool, err := pgxpool.New(context.Background(), sharedCfg.DB.URL)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to connect to database")
    }
    defer pool.Close()

    store := repository.NewStore(pool)
    staffAuthRepo := repository.NewStaffAuthRepository(pool)
    jwtManager := sharedAuth.NewJWTManager(sharedCfg)

    svc := service.New(store, staffAuthRepo, sharedCfg, jwtManager)
    handler := api.NewHandler(sharedCfg, svc, store)

    router := handler.Routes()

    server := &http.Server{
        Addr:         ":" + sharedCfg.Server.Port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        log.Info().Str("addr", server.Addr).Msg("http server listening")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal().Err(err).Msg("http server error")
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Error().Err(err).Msg("graceful shutdown failed")
    }

    log.Info().Msg("server stopped")
}
