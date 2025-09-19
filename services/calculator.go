package metrics

import (
    "math"
    "time"
    
    "admira-etl/internal/models"
)

type Calculator struct{}

func NewCalculator() *Calculator {
    return &Calculator{}
}

func (c *Calculator) CalculateChannelMetrics(adsRecords []models.NormalizedAdsRecord, crmRecords []models.NormalizedCRMRecord, channel string) []models.ChannelMetrics {
    // Group ads records by date and channel
    adsGrouped := make(map[string][]models.NormalizedAdsRecord)
    for _, record := range adsRecords {
        if channel == "" || record.Channel == channel {
            key := record.Date.Format("2006-01-02") + "|" + record.Channel
            adsGrouped[key] = append(adsGrouped[key], record)
        }
    }
    
    var results []models.ChannelMetrics
    
    for key, adsGroup := range adsGrouped {
        if len(adsGroup) == 0 {
            continue
        }
        
        date := adsGroup[0].Date.Format("2006-01-02")
        channelName := adsGroup[0].Channel
        
        // Aggregate ads metrics
        totalClicks := 0
        totalImpressions := 0
        totalCost := 0.0
        utmKeys := make(map[string]bool)
        
        for _, record := range adsGroup {
            totalClicks += record.Clicks
            totalImpressions += record.Impressions
            totalCost += record.Cost
            utmKeys[record.UTMKey] = true
        }
        
        // Find matching CRM records
        leads := 0
        opportunities := 0
        closedWon := 0
        revenue := 0.0
        
        for _, crmRecord := range crmRecords {
            recordDate := crmRecord.CreatedAt.Format("2006-01-02")
            if recordDate == date && utmKeys[crmRecord.UTMKey] {
                switch crmRecord.Stage {
                case "lead":
                    leads++
                case "opportunity":
                    opportunities++
                case "closed_won":
                    closedWon++
                    revenue += crmRecord.Amount
                case "closed_lost":
                    // Count as opportunity that didn't convert
                    opportunities++
                }
            }
        }
        
        // Calculate business metrics
        metrics := models.ChannelMetrics{
            Channel:       channelName,
            Date:          date,
            Clicks:        totalClicks,
            Impressions:   totalImpressions,
            Cost:          totalCost,
            Leads:         leads,
            Opportunities: opportunities + closedWon, // Total opportunities including won
            ClosedWon:     closedWon,
            Revenue:       revenue,
            CPC:           c.safeDivide(totalCost, float64(totalClicks)),
            CPA:           c.safeDivide(totalCost, float64(leads)),
            CVRLeadToOpp:  c.safeDivide(float64(opportunities+closedWon), float64(leads)),
            CVROppToWon:   c.safeDivide(float64(closedWon), float64(opportunities+closedWon)),
            ROAS:          c.safeDivide(revenue, totalCost),
        }
        
        results = append(results, metrics)
    }
    
    return results
}

func (c *Calculator) CalculateFunnelMetrics(adsRecords []models.NormalizedAdsRecord, crmRecords []models.NormalizedCRMRecord, utmCampaign string) []models.FunnelMetrics {
    // Group by UTM parameters
    utmGroups := make(map[string][]models.NormalizedAdsRecord)
    
    for _, record := range adsRecords {
        if utmCampaign == "" || record.UTMCampaign == utmCampaign {
            key := record.UTMKey
            utmGroups[key] = append(utmGroups[key], record)
        }
    }
    
    var results []models.FunnelMetrics
    
    for utmKey, adsGroup := range utmGroups {
        if len(adsGroup) == 0 {
            continue
        }
        
        // Aggregate ads metrics
        totalClicks := 0
        totalImpressions := 0
        totalCost := 0.0
        
        campaign := adsGroup[0].UTMCampaign
        source := adsGroup[0].UTMSource
        medium := adsGroup[0].UTMMedium
        
        for _, record := range adsGroup {
            totalClicks += record.Clicks
            totalImpressions += record.Impressions
            totalCost += record.Cost
        }
        
        // Find matching CRM records
        leads := 0
        opportunities := 0
        closedWon := 0
        revenue := 0.0
        
        for _, crmRecord := range crmRecords {
            if crmRecord.UTMKey == utmKey {
                switch crmRecord.Stage {
                case "lead":
                    leads++
                case "opportunity":
                    opportunities++
                case "closed_won":
                    closedWon++
                    revenue += crmRecord.Amount
                case "closed_lost":
                    opportunities++
                }
            }
        }
        
        metrics := models.FunnelMetrics{
            UTMCampaign:   campaign,
            UTMSource:     source,
            UTMMedium:     medium,
            Clicks:        totalClicks,
            Impressions:   totalImpressions,
            Cost:          totalCost,
            Leads:         leads,
            Opportunities: opportunities + closedWon,
            ClosedWon:     closedWon,
            Revenue:       revenue,
            CPC:           c.safeDivide(totalCost, float64(totalClicks)),
            CPA:           c.safeDivide(totalCost, float64(leads)),
            CVRLeadToOpp:  c.safeDivide(float64(opportunities+closedWon), float64(leads)),
            CVROppToWon:   c.safeDivide(float64(closedWon), float64(opportunities+closedWon)),
            ROAS:          c.safeDivide(revenue, totalCost),
        }
        
        results = append(results, metrics)
    }
    
    return results
}

func (c *Calculator) safeDivide(numerator, denominator float64) float64 {
    if denominator == 0 {
        return 0
    }
    result := numerator / denominator
    if math.IsNaN(result) || math.IsInf(result, 0) {
        return 0
    }
    return math.Round(result*1000) / 1000 // Round to 3 decimal places
}
