package export

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/sirupsen/logrus"
    "admira-etl/internal/client"
    "admira-etl/internal/models"
)

type Exporter struct {
    secret     string
    httpClient *client.HTTPClient
    logger     *logrus.Logger
}

func NewExporter(secret string, httpClient *client.HTTPClient, logger *logrus.Logger) *Exporter {
    return &Exporter{
        secret:     secret,
        httpClient: httpClient,
        logger:     logger,
    }
}

func (e *Exporter) ExportDailyData(sinkURL string, records []models.ExportRecord) error {
    if len(records) == 0 {
        return fmt.Errorf("no records to export")
    }
    
    for _, record := range records {
        // Create HMAC signature
        signature, err := e.createSignature(record)
        if err != nil {
            e.logger.WithError(err).Error("Failed to create signature")
            return fmt.Errorf("failed to create signature: %w", err)
        }
        
        // Send to sink
        if err := e.httpClient.PostExportData(sinkURL, record, signature); err != nil {
            e.logger.WithError(err).WithField("record", record).Error("Failed to export record")
            return fmt.Errorf("failed to export record: %w", err)
        }
        
        e.logger.WithFields(logrus.Fields{
            "date":       record.Date,
            "channel":    record.Channel,
            "campaign_id": record.CampaignID,
        }).Info("Successfully exported record")
    }
    
    return nil
}

func (e *Exporter) ConvertChannelMetricsToExport(metrics []models.ChannelMetrics) []models.ExportRecord {
    var records []models.ExportRecord
    
    for _, metric := range metrics {
        record := models.ExportRecord{
            Date:          metric.Date,
            Channel:       metric.Channel,
            CampaignID:    "aggregated", // Since channel metrics are aggregated
            Clicks:        metric.Clicks,
            Impressions:   metric.Impressions,
            Cost:          metric.Cost,
            Leads:         metric.Leads,
            Opportunities: metric.Opportunities,
            ClosedWon:     metric.ClosedWon,
            Revenue:       metric.Revenue,
            CPC:           metric.CPC,
            CPA:           metric.CPA,
            CVRLeadToOpp:  metric.CVRLeadToOpp,
            CVROppToWon:   metric.CVROppToWon,
            ROAS:          metric.ROAS,
        }
        records = append(records, record)
    }
    
    return records
}

func (e *Exporter) createSignature(data interface{}) (string, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return "", err
    }
    
    h := hmac.New(sha256.New, []byte(e.secret))
    h.Write(jsonData)
    signature := hex.EncodeToString(h.Sum(nil))
    
    return "sha256=" + signature, nil
}
