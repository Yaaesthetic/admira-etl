package storage

import (
    "sync"
    "time"
    
    "admira-etl/internal/models"
)

type MemoryStore struct {
    mu         sync.RWMutex
    adsRecords []models.NormalizedAdsRecord
    crmRecords []models.NormalizedCRMRecord
    lastIngest time.Time
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        adsRecords: make([]models.NormalizedAdsRecord, 0),
        crmRecords: make([]models.NormalizedCRMRecord, 0),
    }
}

func (s *MemoryStore) StoreAdsRecords(records []models.NormalizedAdsRecord) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.adsRecords = records
    s.lastIngest = time.Now()
}

func (s *MemoryStore) StoreCRMRecords(records []models.NormalizedCRMRecord) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.crmRecords = records
}

func (s *MemoryStore) GetAdsRecords() []models.NormalizedAdsRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    records := make([]models.NormalizedAdsRecord, len(s.adsRecords))
    copy(records, s.adsRecords)
    return records
}

func (s *MemoryStore) GetCRMRecords() []models.NormalizedCRMRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    records := make([]models.NormalizedCRMRecord, len(s.crmRecords))
    copy(records, s.crmRecords)
    return records
}

func (s *MemoryStore) GetAdsRecordsByDateRange(from, to time.Time) []models.NormalizedAdsRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    var filtered []models.NormalizedAdsRecord
    for _, record := range s.adsRecords {
        if (record.Date.Equal(from) || record.Date.After(from)) && 
           (record.Date.Equal(to) || record.Date.Before(to)) {
            filtered = append(filtered, record)
        }
    }
    return filtered
}

func (s *MemoryStore) GetCRMRecordsByDateRange(from, to time.Time) []models.NormalizedCRMRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    var filtered []models.NormalizedCRMRecord
    for _, record := range s.crmRecords {
        recordDate := time.Date(record.CreatedAt.Year(), record.CreatedAt.Month(), record.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)
        if (recordDate.Equal(from) || recordDate.After(from)) && 
           (recordDate.Equal(to) || recordDate.Before(to)) {
            filtered = append(filtered, record)
        }
    }
    return filtered
}

func (s *MemoryStore) GetLastIngestTime() time.Time {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.lastIngest
}

func (s *MemoryStore) HasData() bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return len(s.adsRecords) > 0 && len(s.crmRecords) > 0
}
