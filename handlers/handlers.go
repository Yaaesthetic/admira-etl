package handlers

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    
    "admira-etl/internal/config"
    "admira-etl/internal/client"
    "admira-etl/internal/transformer"
    "admira-etl/internal/storage"
    "admira-etl/internal/metrics"
    "admira-etl/internal/export"
    "admira-etl/internal/models"
)

type Handler struct {
    config      *config.Config
    httpClient  *client.HTTPClient
    transformer *transformer.Transformer
    store       *storage.MemoryStore
    calculator  *metrics.Calculator
    exporter    *export.Exporter
    logger      *logrus.Logger
}

func New(cfg *config.Config, httpClient *client.HTTPClient, transformer *transformer.Transformer, 
         store *storage.MemoryStore, calculator *metrics.Calculator, exporter *export.Exporter, 
         logger *logrus.Logger) *Handler {
    return &Handler{
        config:      cfg,
        httpClient:  httpClient,
        transformer: transformer,
        store:       store,
        calculator:  calculator,
        exporter:    exporter,
        logger:      logger,
    }
}

func (h *Handler) HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "ok",
        "timestamp": time.Now().Format(time.RFC3339),
        "service":   "admira-etl",
    })
}

func (h *Handler) ReadinessCheck(c *gin.Context) {
    if h.store.HasData() {
        c.JSON(http.StatusOK, gin.H{
            "status":        "ready",
            "has_data":      true,
            "last_ingest":   h.store.GetLastIngestTime().Format(time.RFC3339),
        })
    } else {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status":   "not ready",
            "has_data": false,
            "message":  "No data ingested yet",
        })
    }
}

func (h *Handler) IngestData(c *gin.Context) {
    startTime := time.Now()
    
    since := c.Query("since")
    var sinceTime time.Time
    if since != "" {
        if t, err := time.Parse("2006-01-02", since); err == nil {
            sinceTime = t
            h.logger.WithField("since", sinceTime).Info("Filtering data since date")
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
            return
        }
    }
    
    h.logger.Info("Starting data ingestion")
    
    // Fetch ads data
    adsResponse, err := h.httpClient.FetchAdsData(h.config.AdsAPIURL)
    if err != nil {
        h.logger.WithError(err).Error("Failed to fetch ads data")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ads data"})
        return
    }
    
    // Fetch CRM data
    crmResponse, err := h.httpClient.FetchCRMData(h.config.CRMAPIURL)
    if err != nil {
        h.logger.WithError(err).Error("Failed to fetch CRM data")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch CRM data"})
        return
    }
    
    // Transform and filter data with quality validation
    normalizedAds := h.transformer.NormalizeAdsRecords(adsResponse.External.Ads.Performance)
    normalizedCRM := h.transformer.NormalizeCRMRecords(crmResponse.External.CRM.Opportunities)
    
    // Apply since filter if specified
    if !sinceTime.IsZero() {
        filteredAds := []models.NormalizedAdsRecord{}
        for _, record := range normalizedAds {
            if record.Date.Equal(sinceTime) || record.Date.After(sinceTime) {
                filteredAds = append(filteredAds, record)
            }
        }
        normalizedAds = filteredAds
        
        filteredCRM := []models.NormalizedCRMRecord{}
        for _, record := range normalizedCRM {
            recordDate := time.Date(record.CreatedAt.Year(), record.CreatedAt.Month(), record.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)
            if recordDate.Equal(sinceTime) || recordDate.After(sinceTime) {
                filteredCRM = append(filteredCRM, record)
            }
        }
        normalizedCRM = filteredCRM
    }
    
    // Generate quality report
    qualityReport := h.transformer.GenerateQualityReport(normalizedAds, normalizedCRM)
    
    // Store data
    h.store.StoreAdsRecords(normalizedAds)
    h.store.StoreCRMRecords(normalizedCRM)
    
    duration := time.Since(startTime)
    h.logger.WithFields(logrus.Fields{
        "ads_records":    len(normalizedAds),
        "crm_records":    len(normalizedCRM),
        "duration_ms":    duration.Milliseconds(),
        "quality_score":  qualityReport.Summary.OverallQualityScore,
        "valid_ads":      qualityReport.Summary.ValidAdsRecords,
        "valid_crm":      qualityReport.Summary.ValidCRMRecords,
    }).Info("Data ingestion completed with quality validation")
    
    // Log quality issues if any
    if len(qualityReport.Summary.CommonIssues) > 0 {
        h.logger.WithField("common_issues", qualityReport.Summary.CommonIssues).Warn("Data quality issues detected")
    }
    
    c.JSON(http.StatusOK, models.IngestResponse{
        Status:         "success",
        AdsRecords:     len(normalizedAds),
        CRMRecords:     len(normalizedCRM),
        ProcessedAt:    time.Now().Format(time.RFC3339),
        Message:        "Data ingested and processed with quality validation",
        QualitySummary: qualityReport.Summary,
    })
}

func (h *Handler) GetDataQualityReport(c *gin.Context) {
    adsRecords := h.store.GetAdsRecords()
    crmRecords := h.store.GetCRMRecords()
    
    if len(adsRecords) == 0 && len(crmRecords) == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "No data available for quality analysis. Please run ingestion first.",
        })
        return
    }
    
    qualityReport := h.transformer.GenerateQualityReport(adsRecords, crmRecords)
    
    c.JSON(http.StatusOK, qualityReport)
}

func (h *Handler) GetChannelMetrics(c *gin.Context) {
    from := c.Query("from")
    to := c.Query("to")
    channel := c.Query("channel")
    limitStr := c.DefaultQuery("limit", "10")
    offsetStr := c.DefaultQuery("offset", "0")
    
    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)
    
    // Parse dates
    var fromTime, toTime time.Time
    var err error
    
    if from != "" {
        fromTime, err = time.Parse("2006-01-02", from)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format, use YYYY-MM-DD"})
            return
        }
    }
    
    if to != "" {
        toTime, err = time.Parse("2006-01-02", to)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format, use YYYY-MM-DD"})
            return
        }
    }
    
    // Get filtered data
    var adsRecords []models.NormalizedAdsRecord
    var crmRecords []models.NormalizedCRMRecord
    
    if !fromTime.IsZero() && !toTime.IsZero() {
        adsRecords = h.store.GetAdsRecordsByDateRange(fromTime, toTime)
        crmRecords = h.store.GetCRMRecordsByDateRange(fromTime, toTime)
    } else {
        adsRecords = h.store.GetAdsRecords()
        crmRecords = h.store.GetCRMRecords()
    }
    
    // Calculate metrics with quality scores
    metrics := h.calculator.CalculateChannelMetricsWithQuality(adsRecords, crmRecords, channel)
    
    // Apply pagination
    total := len(metrics)
    start := offset
    end := offset + limit
    
    if start > total {
        start = total
    }
    if end > total {
        end = total
    }
    
    paginatedMetrics := metrics[start:end]
    
    response := models.MetricsResponse{
        Data:    paginatedMetrics,
        Total:   total,
        Page:    offset/limit + 1,
        Limit:   limit,
        HasMore: end < total,
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *Handler) GetFunnelMetrics(c *gin.Context) {
    from := c.Query("from")
    to := c.Query("to")
    utmCampaign := c.Query("utm_campaign")
    limitStr := c.DefaultQuery("limit", "10")
    offsetStr := c.DefaultQuery("offset", "0")
    
    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)
    
    // Parse dates
    var fromTime, toTime time.Time
    var err error
    
    if from != "" {
        fromTime, err = time.Parse("2006-01-02", from)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format, use YYYY-MM-DD"})
            return
        }
    }
    
    if to != "" {
        toTime, err = time.Parse("2006-01-02", to)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format, use YYYY-MM-DD"})
            return
        }
    }
    
    // Get filtered data
    var adsRecords []models.NormalizedAdsRecord
    var crmRecords []models.NormalizedCRMRecord
    
    if !fromTime.IsZero() && !toTime.IsZero() {
        adsRecords = h.store.GetAdsRecordsByDateRange(fromTime, toTime)
        crmRecords = h.store.GetCRMRecordsByDateRange(fromTime, toTime)
    } else {
        adsRecords = h.store.GetAdsRecords()
        crmRecords = h.store.GetCRMRecords()
    }
    
    // Calculate metrics with quality scores
    metrics := h.calculator.CalculateFunnelMetricsWithQuality(adsRecords, crmRecords, utmCampaign)
    
    // Apply pagination
    total := len(metrics)
    start := offset
    end := offset + limit
    
    if start > total {
        start = total
    }
    if end > total {
        end = total
    }
    
    paginatedMetrics := metrics[start:end]
    
    response := models.MetricsResponse{
        Data:    paginatedMetrics,
        Total:   total,
        Page:    offset/limit + 1,
        Limit:   limit,
        HasMore: end < total,
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *Handler) ExportData(c *gin.Context) {
    dateStr := c.Query("date")
    if dateStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter is required (YYYY-MM-DD)"})
        return
    }
    
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
        return
    }
    
    // Get data for the specific date
    adsRecords := h.store.GetAdsRecordsByDateRange(date, date)
    crmRecords := h.store.GetCRMRecordsByDateRange(date, date)
    
    if len(adsRecords) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "No data found for the specified date"})
        return
    }
    
    // Calculate metrics for export
    channelMetrics := h.calculator.CalculateChannelMetricsWithQuality(adsRecords, crmRecords, "")
    exportRecords := h.exporter.ConvertChannelMetricsToExport(channelMetrics)
    
    // Export to sink if URL is configured
    if h.config.SinkURL != "" {
        if err := h.exporter.ExportDailyData(h.config.SinkURL, exportRecords); err != nil {
            h.logger.WithError(err).Error("Failed to export to sink")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
            return
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":         "success",
        "date":           dateStr,
        "records_count":  len(exportRecords),
        "exported_at":    time.Now().Format(time.RFC3339),
        "sink_url":       h.config.SinkURL,
        "data":           exportRecords,
    })
}
