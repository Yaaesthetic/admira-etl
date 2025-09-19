package models

import (
    "time"
)

// Data Quality Tracking Structures
type FieldQuality struct {
    IsValid     bool   `json:"is_valid"`
    Description string `json:"description"`
    OriginalValue interface{} `json:"original_value,omitempty"`
}

type RecordQuality struct {
    RecordID    string                    `json:"record_id"`
    IsValid     bool                      `json:"is_valid"`
    FieldErrors map[string]FieldQuality   `json:"field_errors"`
    ErrorCount  int                       `json:"error_count"`
}

// External API Response Structures
type AdsResponse struct {
    External struct {
        Ads struct {
            Performance []AdsRecord `json:"performance"`
        } `json:"ads"`
    } `json:"external"`
}

type CRMResponse struct {
    External struct {
        CRM struct {
            Opportunities []CRMRecord `json:"opportunities"`
        } `json:"crm"`
    } `json:"external"`
}

// Raw data records
type AdsRecord struct {
    Date         string  `json:"date"`
    CampaignID   string  `json:"campaign_id"`
    Channel      string  `json:"channel"`
    Clicks       int     `json:"clicks"`
    Impressions  int     `json:"impressions"`
    Cost         float64 `json:"cost"`
    UTMCampaign  string  `json:"utm_campaign"`
    UTMSource    *string `json:"utm_source"`
    UTMMedium    *string `json:"utm_medium"`
}

type CRMRecord struct {
    OpportunityID string  `json:"opportunity_id"`
    ContactEmail  string  `json:"contact_email"`
    Stage         string  `json:"stage"`
    Amount        float64 `json:"amount"`
    CreatedAt     string  `json:"created_at"`
    UTMCampaign   string  `json:"utm_campaign"`
    UTMSource     *string `json:"utm_source"`
    UTMMedium     *string `json:"utm_medium"`
}

// Normalized internal structures with Quality Tracking
type NormalizedAdsRecord struct {
    Date         time.Time
    CampaignID   string
    Channel      string
    Clicks       int
    Impressions  int
    Cost         float64
    UTMCampaign  string
    UTMSource    string
    UTMMedium    string
    UTMKey       string
    
    // Data Quality Tracking
    Quality      RecordQuality `json:"quality"`
}

type NormalizedCRMRecord struct {
    OpportunityID string
    ContactEmail  string
    Stage         string
    Amount        float64
    CreatedAt     time.Time
    UTMCampaign   string
    UTMSource     string
    UTMMedium     string
    UTMKey        string
    
    // Data Quality Tracking
    Quality       RecordQuality `json:"quality"`
}

// Business metrics
type ChannelMetrics struct {
    Channel       string  `json:"channel"`
    Date          string  `json:"date"`
    Clicks        int     `json:"clicks"`
    Impressions   int     `json:"impressions"`
    Cost          float64 `json:"cost"`
    Leads         int     `json:"leads"`
    Opportunities int     `json:"opportunities"`
    ClosedWon     int     `json:"closed_won"`
    Revenue       float64 `json:"revenue"`
    CPC           float64 `json:"cpc"`
    CPA           float64 `json:"cpa"`
    CVRLeadToOpp  float64 `json:"cvr_lead_to_opp"`
    CVROppToWon   float64 `json:"cvr_opp_to_won"`
    ROAS          float64 `json:"roas"`
    
    // Data Quality Summary
    QualityScore  float64 `json:"quality_score"`  // Percentage of valid records
    TotalRecords  int     `json:"total_records"`
    ValidRecords  int     `json:"valid_records"`
}

type FunnelMetrics struct {
    UTMCampaign   string  `json:"utm_campaign"`
    UTMSource     string  `json:"utm_source"`
    UTMMedium     string  `json:"utm_medium"`
    Clicks        int     `json:"clicks"`
    Impressions   int     `json:"impressions"`
    Cost          float64 `json:"cost"`
    Leads         int     `json:"leads"`
    Opportunities int     `json:"opportunities"`
    ClosedWon     int     `json:"closed_won"`
    Revenue       float64 `json:"revenue"`
    CPC           float64 `json:"cpc"`
    CPA           float64 `json:"cpa"`
    CVRLeadToOpp  float64 `json:"cvr_lead_to_opp"`
    CVROppToWon   float64 `json:"cvr_opp_to_won"`
    ROAS          float64 `json:"roas"`
    
    // Data Quality Summary
    QualityScore  float64 `json:"quality_score"`
    TotalRecords  int     `json:"total_records"`
    ValidRecords  int     `json:"valid_records"`
}

// Data Quality Report Structures
type DataQualityReport struct {
    Summary    QualitySummary    `json:"summary"`
    AdsReport  []RecordQuality   `json:"ads_quality"`
    CRMReport  []RecordQuality   `json:"crm_quality"`
    Timestamp  string            `json:"timestamp"`
}

type QualitySummary struct {
    TotalAdsRecords    int     `json:"total_ads_records"`
    ValidAdsRecords    int     `json:"valid_ads_records"`
    AdsQualityScore    float64 `json:"ads_quality_score"`
    TotalCRMRecords    int     `json:"total_crm_records"`
    ValidCRMRecords    int     `json:"valid_crm_records"`
    CRMQualityScore    float64 `json:"crm_quality_score"`
    OverallQualityScore float64 `json:"overall_quality_score"`
    CommonIssues       []string `json:"common_issues"`
}

// API response structures
type MetricsResponse struct {
    Data       interface{} `json:"data"`
    Total      int         `json:"total"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    HasMore    bool        `json:"has_more"`
}

type IngestResponse struct {
    Status        string `json:"status"`
    AdsRecords    int    `json:"ads_records"`
    CRMRecords    int    `json:"crm_records"`
    ProcessedAt   string `json:"processed_at"`
    Message       string `json:"message"`
    
    // Data Quality Summary
    QualitySummary QualitySummary `json:"quality_summary"`
}

type ExportRecord struct {
    Date          string  `json:"date"`
    Channel       string  `json:"channel"`
    CampaignID    string  `json:"campaign_id"`
    Clicks        int     `json:"clicks"`
    Impressions   int     `json:"impressions"`
    Cost          float64 `json:"cost"`
    Leads         int     `json:"leads"`
    Opportunities int     `json:"opportunities"`
    ClosedWon     int     `json:"closed_won"`
    Revenue       float64 `json:"revenue"`
    CPC           float64 `json:"cpc"`
    CPA           float64 `json:"cpa"`
    CVRLeadToOpp  float64 `json:"cvr_lead_to_opp"`
    CVROppToWon   float64 `json:"cvr_opp_to_won"`
    ROAS          float64 `json:"roas"`
}
