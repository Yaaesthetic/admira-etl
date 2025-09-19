package transformer

import (
    "fmt"
    "regexp"
    "strings"
    "time"
    
    "admira-etl/internal/models"
)

type Transformer struct {
    emailRegex *regexp.Regexp
}

func New() *Transformer {
    return &Transformer{
        emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    }
}

func (t *Transformer) NormalizeAdsRecords(records []models.AdsRecord) []models.NormalizedAdsRecord {
    var normalized []models.NormalizedAdsRecord
    
    for i, record := range records {
        quality := models.RecordQuality{
            RecordID:    fmt.Sprintf("ads_%d", i),
            IsValid:     true,
            FieldErrors: make(map[string]models.FieldQuality),
            ErrorCount:  0,
        }
        
        normalizedRecord := models.NormalizedAdsRecord{
            Date:        t.validateAndParseDate(record.Date, "date", &quality),
            CampaignID:  t.validateCampaignID(record.CampaignID, "campaign_id", &quality),
            Channel:     t.validateChannel(record.Channel, "channel", &quality),
            Clicks:      t.validateClicks(record.Clicks, "clicks", &quality),
            Impressions: t.validateImpressions(record.Impressions, "impressions", &quality),
            Cost:        t.validateCost(record.Cost, "cost", &quality),
            UTMCampaign: t.validateUTMCampaign(record.UTMCampaign, "utm_campaign", &quality),
            UTMSource:   t.validateUTMSource(record.UTMSource, "utm_source", &quality),
            UTMMedium:   t.validateUTMMedium(record.UTMMedium, "utm_medium", &quality),
            Quality:     quality,
        }
        
        normalizedRecord.UTMKey = t.generateUTMKey(
            normalizedRecord.UTMCampaign,
            normalizedRecord.UTMSource,
            normalizedRecord.UTMMedium,
        )
        
        // Final record validation
        normalizedRecord.Quality.IsValid = normalizedRecord.Quality.ErrorCount == 0
        
        normalized = append(normalized, normalizedRecord)
    }
    
    return t.deduplicateAdsRecords(normalized)
}

func (t *Transformer) NormalizeCRMRecords(records []models.CRMRecord) []models.NormalizedCRMRecord {
    var normalized []models.NormalizedCRMRecord
    
    for i, record := range records {
        quality := models.RecordQuality{
            RecordID:    fmt.Sprintf("crm_%d", i),
            IsValid:     true,
            FieldErrors: make(map[string]models.FieldQuality),
            ErrorCount:  0,
        }
        
        normalizedRecord := models.NormalizedCRMRecord{
            OpportunityID: t.validateOpportunityID(record.OpportunityID, "opportunity_id", &quality),
            ContactEmail:  t.validateEmail(record.ContactEmail, "contact_email", &quality),
            Stage:         t.validateStage(record.Stage, "stage", &quality),
            Amount:        t.validateAmount(record.Amount, "amount", &quality),
            CreatedAt:     t.validateAndParseDateTime(record.CreatedAt, "created_at", &quality),
            UTMCampaign:   t.validateUTMCampaign(record.UTMCampaign, "utm_campaign", &quality),
            UTMSource:     t.validateUTMSource(record.UTMSource, "utm_source", &quality),
            UTMMedium:     t.validateUTMMedium(record.UTMMedium, "utm_medium", &quality),
            Quality:       quality,
        }
        
        normalizedRecord.UTMKey = t.generateUTMKey(
            normalizedRecord.UTMCampaign,
            normalizedRecord.UTMSource,
            normalizedRecord.UTMMedium,
        )
        
        // Final record validation
        normalizedRecord.Quality.IsValid = normalizedRecord.Quality.ErrorCount == 0
        
        normalized = append(normalized, normalizedRecord)
    }
    
    return t.deduplicateCRMRecords(normalized)
}

// ADS Field Validators
func (t *Transformer) validateAndParseDate(dateStr string, fieldName string, quality *models.RecordQuality) time.Time {
    if strings.TrimSpace(dateStr) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Date field is empty",
            OriginalValue: dateStr,
        }
        quality.ErrorCount++
        return time.Time{}
    }
    
    // Handle different date formats
    formats := []string{
        "2006-01-02",
        "2006/01/02",
    }
    
    for _, format := range formats {
        if date, err := time.Parse(format, dateStr); err == nil {
            quality.FieldErrors[fieldName] = models.FieldQuality{
                IsValid:       true,
                Description:   "Valid date",
                OriginalValue: dateStr,
            }
            return date
        }
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       false,
        Description:   "Invalid date format - Expected YYYY-MM-DD or YYYY/MM/DD",
        OriginalValue: dateStr,
    }
    quality.ErrorCount++
    return time.Time{}
}

func (t *Transformer) validateCampaignID(id string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(id) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Campaign ID is empty, using 'unknown'",
            OriginalValue: id,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid campaign ID",
        OriginalValue: id,
    }
    return id
}

func (t *Transformer) validateChannel(channel string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(channel) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Channel is empty",
            OriginalValue: channel,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    validChannels := []string{"google_ads", "facebook_ads", "tiktok_ads", "linkedin_ads", "twitter_ads"}
    for _, validChannel := range validChannels {
        if channel == validChannel {
            quality.FieldErrors[fieldName] = models.FieldQuality{
                IsValid:       true,
                Description:   "Valid channel",
                OriginalValue: channel,
            }
            return channel
        }
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       false,
        Description:   fmt.Sprintf("Unknown channel type: %s", channel),
        OriginalValue: channel,
    }
    quality.ErrorCount++
    return channel
}

func (t *Transformer) validateClicks(clicks int, fieldName string, quality *models.RecordQuality) int {
    if clicks < 0 {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Invalid - Clicks cannot be negative, setting to 0",
            OriginalValue: clicks,
        }
        quality.ErrorCount++
        return 0
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid clicks count",
        OriginalValue: clicks,
    }
    return clicks
}

func (t *Transformer) validateImpressions(impressions int, fieldName string, quality *models.RecordQuality) int {
    if impressions < 0 {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Invalid - Impressions cannot be negative, setting to 0",
            OriginalValue: impressions,
        }
        quality.ErrorCount++
        return 0
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid impressions count",
        OriginalValue: impressions,
    }
    return impressions
}

func (t *Transformer) validateCost(cost float64, fieldName string, quality *models.RecordQuality) float64 {
    if cost < 0 {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Invalid - Cost cannot be negative, setting to 0",
            OriginalValue: cost,
        }
        quality.ErrorCount++
        return 0
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid cost amount",
        OriginalValue: cost,
    }
    return cost
}

// CRM Field Validators
func (t *Transformer) validateOpportunityID(id string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(id) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Opportunity ID is empty",
            OriginalValue: id,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid opportunity ID",
        OriginalValue: id,
    }
    return id
}

func (t *Transformer) validateEmail(email string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(email) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Email is empty",
            OriginalValue: email,
        }
        quality.ErrorCount++
        return email
    }
    
    if !t.emailRegex.MatchString(email) {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Invalid email format",
            OriginalValue: email,
        }
        quality.ErrorCount++
        return email
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid email",
        OriginalValue: email,
    }
    return email
}

func (t *Transformer) validateStage(stage string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(stage) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - Stage is empty",
            OriginalValue: stage,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    validStages := []string{"lead", "opportunity", "closed_won", "closed_lost"}
    for _, validStage := range validStages {
        if stage == validStage {
            quality.FieldErrors[fieldName] = models.FieldQuality{
                IsValid:       true,
                Description:   "Valid stage",
                OriginalValue: stage,
            }
            return stage
        }
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       false,
        Description:   fmt.Sprintf("Unknown stage: %s", stage),
        OriginalValue: stage,
    }
    quality.ErrorCount++
    return stage
}

func (t *Transformer) validateAmount(amount float64, fieldName string, quality *models.RecordQuality) float64 {
    if amount < 0 {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Invalid - Amount cannot be negative, setting to 0",
            OriginalValue: amount,
        }
        quality.ErrorCount++
        return 0
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid amount",
        OriginalValue: amount,
    }
    return amount
}

func (t *Transformer) validateAndParseDateTime(dateTimeStr string, fieldName string, quality *models.RecordQuality) time.Time {
    if strings.TrimSpace(dateTimeStr) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - DateTime field is empty",
            OriginalValue: dateTimeStr,
        }
        quality.ErrorCount++
        return time.Time{}
    }
    
    // Handle different datetime formats
    formats := []string{
        "2006-01-02T15:04:05Z",
        "2006-01-02T15:04:05.000Z",
        "2006-01-02 15:04:05",
        "2006/01/02 15:04:05",
    }
    
    for _, format := range formats {
        if dateTime, err := time.Parse(format, dateTimeStr); err == nil {
            quality.FieldErrors[fieldName] = models.FieldQuality{
                IsValid:       true,
                Description:   "Valid datetime",
                OriginalValue: dateTimeStr,
            }
            return dateTime
        }
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       false,
        Description:   "Invalid datetime format - Expected ISO format or YYYY-MM-DD HH:MM:SS",
        OriginalValue: dateTimeStr,
    }
    quality.ErrorCount++
    return time.Time{}
}

// UTM Validators
func (t *Transformer) validateUTMCampaign(campaign string, fieldName string, quality *models.RecordQuality) string {
    if strings.TrimSpace(campaign) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - UTM Campaign is empty, using 'unknown'",
            OriginalValue: campaign,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid UTM campaign",
        OriginalValue: campaign,
    }
    return campaign
}

func (t *Transformer) validateUTMSource(source *string, fieldName string, quality *models.RecordQuality) string {
    if source == nil || strings.TrimSpace(*source) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - UTM Source is null or empty, using 'unknown'",
            OriginalValue: source,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid UTM source",
        OriginalValue: *source,
    }
    return strings.TrimSpace(*source)
}

func (t *Transformer) validateUTMMedium(medium *string, fieldName string, quality *models.RecordQuality) string {
    if medium == nil || strings.TrimSpace(*medium) == "" {
        quality.FieldErrors[fieldName] = models.FieldQuality{
            IsValid:       false,
            Description:   "Missing - UTM Medium is null or empty, using 'unknown'",
            OriginalValue: medium,
        }
        quality.ErrorCount++
        return "unknown"
    }
    
    quality.FieldErrors[fieldName] = models.FieldQuality{
        IsValid:       true,
        Description:   "Valid UTM medium",
        OriginalValue: *medium,
    }
    return strings.TrimSpace(*medium)
}

func (t *Transformer) generateUTMKey(campaign, source, medium string) string {
    if strings.TrimSpace(campaign) == "" {
        campaign = "unknown"
    }
    return fmt.Sprintf("%s|%s|%s", campaign, source, medium)
}

func (t *Transformer) deduplicateAdsRecords(records []models.NormalizedAdsRecord) []models.NormalizedAdsRecord {
    seen := make(map[string]int) // map to track index of first occurrence
    var unique []models.NormalizedAdsRecord
    
    for i, record := range records {
        key := fmt.Sprintf("%s|%s|%s", 
            record.Date.Format("2006-01-02"), 
            record.CampaignID, 
            record.Channel)
        
        if existingIndex, exists := seen[key]; !exists {
            seen[key] = i
            unique = append(unique, record)
        } else {
            // Mark the duplicate with quality issue
            record.Quality.FieldErrors["duplicate"] = models.FieldQuality{
                IsValid:       false,
                Description:   fmt.Sprintf("Duplicate record found (original at index %d)", existingIndex),
                OriginalValue: key,
            }
            record.Quality.ErrorCount++
            record.Quality.IsValid = false
        }
    }
    
    return unique
}

func (t *Transformer) deduplicateCRMRecords(records []models.NormalizedCRMRecord) []models.NormalizedCRMRecord {
    seen := make(map[string]int)
    var unique []models.NormalizedCRMRecord
    
    for i, record := range records {
        if existingIndex, exists := seen[record.OpportunityID]; !exists {
            seen[record.OpportunityID] = i
            unique = append(unique, record)
        } else {
            // Mark the duplicate with quality issue
            record.Quality.FieldErrors["duplicate"] = models.FieldQuality{
                IsValid:       false,
                Description:   fmt.Sprintf("Duplicate opportunity ID found (original at index %d)", existingIndex),
                OriginalValue: record.OpportunityID,
            }
            record.Quality.ErrorCount++
            record.Quality.IsValid = false
        }
    }
    
    return unique
}

// Generate Quality Report
func (t *Transformer) GenerateQualityReport(adsRecords []models.NormalizedAdsRecord, crmRecords []models.NormalizedCRMRecord) models.DataQualityReport {
    var adsQuality []models.RecordQuality
    var crmQuality []models.RecordQuality
    
    validAds := 0
    for _, record := range adsRecords {
        adsQuality = append(adsQuality, record.Quality)
        if record.Quality.IsValid {
            validAds++
        }
    }
    
    validCRM := 0
    for _, record := range crmRecords {
        crmQuality = append(crmQuality, record.Quality)
        if record.Quality.IsValid {
            validCRM++
        }
    }
    
    adsScore := 0.0
    if len(adsRecords) > 0 {
        adsScore = float64(validAds) / float64(len(adsRecords)) * 100
    }
    
    crmScore := 0.0
    if len(crmRecords) > 0 {
        crmScore = float64(validCRM) / float64(len(crmRecords)) * 100
    }
    
    overallScore := 0.0
    totalRecords := len(adsRecords) + len(crmRecords)
    if totalRecords > 0 {
        overallScore = float64(validAds+validCRM) / float64(totalRecords) * 100
    }
    
    // Identify common issues
    commonIssues := t.identifyCommonIssues(adsRecords, crmRecords)
    
    return models.DataQualityReport{
        Summary: models.QualitySummary{
            TotalAdsRecords:     len(adsRecords),
            ValidAdsRecords:     validAds,
            AdsQualityScore:     adsScore,
            TotalCRMRecords:     len(crmRecords),
            ValidCRMRecords:     validCRM,
            CRMQualityScore:     crmScore,
            OverallQualityScore: overallScore,
            CommonIssues:        commonIssues,
        },
        AdsReport: adsQuality,
        CRMReport: crmQuality,
        Timestamp: time.Now().Format(time.RFC3339),
    }
}

func (t *Transformer) identifyCommonIssues(adsRecords []models.NormalizedAdsRecord, crmRecords []models.NormalizedCRMRecord) []string {
    issueCount := make(map[string]int)
    
    for _, record := range adsRecords {
        for _, fieldError := range record.Quality.FieldErrors {
            if !fieldError.IsValid {
                issueCount[fieldError.Description]++
            }
        }
    }
    
    for _, record := range crmRecords {
        for _, fieldError := range record.Quality.FieldErrors {
            if !fieldError.IsValid {
                issueCount[fieldError.Description]++
            }
        }
    }
    
    var commonIssues []string
    for issue, count := range issueCount {
        if count > 1 { // Only include issues that appear more than once
            commonIssues = append(commonIssues, fmt.Sprintf("%s (occurs %d times)", issue, count))
        }
    }
    
    return commonIssues
}
