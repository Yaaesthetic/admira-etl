package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    
    "admira-etl/internal/config"
    "admira-etl/internal/client"
    "admira-etl/internal/storage"
    "admira-etl/internal/transformer"
    "admira-etl/internal/handlers"
    "admira-etl/internal/metrics"
    "admira-etl/internal/export"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Setup logger
    logger := logrus.New()
    level, err := logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        level = logrus.InfoLevel
    }
    logger.SetLevel(level)
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    logger.Info("Starting Admira ETL Service with Data Quality Tracking")
    
    // Initialize components
    httpClient := client.NewHTTPClient(cfg, logger)
    transformer := transformer.New()
    store := storage.NewMemoryStore()
    calculator := metrics.NewCalculator()
    exporter := export.NewExporter(cfg.SinkSecret, httpClient, logger)
    
    // Initialize handlers
    handler := handlers.New(cfg, httpClient, transformer, store, calculator, exporter, logger)
    
    // Setup Gin router
    if cfg.LogLevel != "debug" {
        gin.SetMode(gin.ReleaseMode)
    }
    router := gin.New()
    router.Use(gin.Logger(), gin.Recovery())
    
    // Health endpoints
    router.GET("/healthz", handler.HealthCheck)
    router.GET("/readyz", handler.ReadinessCheck)
    
    // Ingestion endpoint
    router.POST("/ingest/run", handler.IngestData)
    
    // Data quality endpoint
    router.GET("/quality/report", handler.GetDataQualityReport)
    
    // Metrics endpoints
    router.GET("/metrics/channel", handler.GetChannelMetrics)
    router.GET("/metrics/funnel", handler.GetFunnelMetrics)
    
    // Export endpoint
    router.POST("/export/run", handler.ExportData)
    
    // Start server
    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: router,
    }
    
    go func() {
        logger.WithField("port", cfg.Port).Info("Server started with data quality tracking")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.WithError(err).Fatal("Failed to start server")
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        logger.WithError(err).Fatal("Server forced to shutdown")
    }
    
    logger.Info("Server exited")
}
