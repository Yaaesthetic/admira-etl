package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/sirupsen/logrus"
    "admira-etl/internal/config"
    "admira-etl/internal/models"
)

type HTTPClient struct {
    client        *http.Client
    retryAttempts int
    logger        *logrus.Logger
}

func NewHTTPClient(cfg *config.Config, logger *logrus.Logger) *HTTPClient {
    return &HTTPClient{
        client: &http.Client{
            Timeout: cfg.HTTPTimeout,
        },
        retryAttempts: cfg.RetryAttempts,
        logger:        logger,
    }
}

func (c *HTTPClient) FetchAdsData(url string) (*models.AdsResponse, error) {
    var adsResponse models.AdsResponse
    
    err := c.retryRequest(url, &adsResponse)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch ads data: %w", err)
    }
    
    c.logger.WithField("records", len(adsResponse.External.Ads.Performance)).Info("Fetched ads data")
    return &adsResponse, nil
}

func (c *HTTPClient) FetchCRMData(url string) (*models.CRMResponse, error) {
    var crmResponse models.CRMResponse
    
    err := c.retryRequest(url, &crmResponse)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch CRM data: %w", err)
    }
    
    c.logger.WithField("records", len(crmResponse.External.CRM.Opportunities)).Info("Fetched CRM data")
    return &crmResponse, nil
}

func (c *HTTPClient) PostExportData(url string, data interface{}, signature string) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal export data: %w", err)
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create export request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Signature", signature)
    
    return c.retryPostRequest(req)
}

func (c *HTTPClient) retryRequest(url string, target interface{}) error {
    var lastErr error
    
    for attempt := 0; attempt < c.retryAttempts; attempt++ {
        if attempt > 0 {
            backoffTime := time.Duration(attempt*attempt) * time.Second
            c.logger.WithFields(logrus.Fields{
                "attempt": attempt + 1,
                "backoff": backoffTime,
                "url":     url,
            }).Warn("Retrying request after backoff")
            time.Sleep(backoffTime)
        }
        
        resp, err := c.client.Get(url)
        if err != nil {
            lastErr = err
            continue
        }
        
        if resp.StatusCode >= 500 {
            resp.Body.Close()
            lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
            continue
        }
        
        if resp.StatusCode >= 400 {
            resp.Body.Close()
            return fmt.Errorf("client error: %d", resp.StatusCode)
        }
        
        body, err := io.ReadAll(resp.Body)
        resp.Body.Close()
        
        if err != nil {
            lastErr = err
            continue
        }
        
        if err := json.Unmarshal(body, target); err != nil {
            lastErr = err
            continue
        }
        
        c.logger.WithFields(logrus.Fields{
            "attempt":     attempt + 1,
            "status_code": resp.StatusCode,
            "url":         url,
        }).Info("Request successful")
        
        return nil
    }
    
    return fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

func (c *HTTPClient) retryPostRequest(req *http.Request) error {
    var lastErr error
    
    for attempt := 0; attempt < c.retryAttempts; attempt++ {
        if attempt > 0 {
            backoffTime := time.Duration(attempt*attempt) * time.Second
            time.Sleep(backoffTime)
        }
        
        resp, err := c.client.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        
        resp.Body.Close()
        
        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            return nil
        }
        
        if resp.StatusCode >= 400 && resp.StatusCode < 500 {
            return fmt.Errorf("client error: %d", resp.StatusCode)
        }
        
        lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
    }
    
    return fmt.Errorf("export failed after retries: %w", lastErr)
}
