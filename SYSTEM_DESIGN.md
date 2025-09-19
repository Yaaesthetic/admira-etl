# SYSTEM_DESIGN.md

## Admira ETL Service - System Design Document

### Document Information
- **System Name**: Admira ETL Service
- **Version**: 1.0
- **Date**: September 19, 2025
- **Architecture**: Microservice-based ETL Pipeline
- **Technology Stack**: Go 1.21, Docker, Gin Framework

***

## Executive Summary

The Admira ETL Service is a production-ready data integration system designed to process marketing campaign data from advertising platforms and CRM systems. The system transforms raw data into actionable business metrics while maintaining comprehensive data quality validation and providing real-time monitoring capabilities.

### Key Design Goals
- **Data Integrity**: Comprehensive field-level validation with detailed error reporting
- **Scalability**: Containerized architecture supporting horizontal scaling
- **Reliability**: Exponential backoff retry mechanisms and graceful error handling
- **Observability**: Structured logging and health monitoring endpoints
- **Security**: HMAC-signed data exports and secure API communications

***

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    External Data Sources                     │
│  ┌─────────────┐                    ┌─────────────┐        │
│  │  Ads API    │                    │   CRM API   │        │
│  │ (Campaign   │                    │(Opportunities│        │
│  │  Data)      │                    │  & Leads)   │        │
│  └─────────────┘                    └─────────────┘        │
└─────────────────┬─────────────────────────┬─────────────────┘
                  │                         │
                  ▼                         ▼
┌─────────────────────────────────────────────────────────────┐
│                 Admira ETL Service                          │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │               HTTP API Layer (Gin)                      │ │
│ │    /healthz  /ingest/run  /metrics/*  /quality/report  │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │            Business Logic Layer                         │ │
│ │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │ │
│ │  │Handlers │  │Metrics  │  │Export   │  │Quality  │   │ │
│ │  │         │  │Engine   │  │Service  │  │Validator│   │ │
│ │  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │           Data Processing Layer                         │ │
│ │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │ │
│ │  │HTTP     │  │ETL      │  │In-Memory│  │Config   │   │ │
│ │  │Client   │  │Transform│  │Storage  │  │Manager  │   │ │
│ │  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Export Destinations                      │
│              ┌─────────────────────────────┐               │
│              │         Sink API            │               │
│              │     (HMAC Secured)          │               │
│              └─────────────────────────────┘               │
└─────────────────────────────────────────────────────────────┘
```

### Component Architecture

The system follows a **layered architecture pattern** with clear separation of concerns:

1. **HTTP API Layer**: REST endpoints using Gin framework
2. **Business Logic Layer**: Core business operations and metrics calculation
3. **Data Processing Layer**: ETL operations, validation, and storage
4. **Foundation Layer**: Configuration, models, and logging utilities

---

## Detailed Component Design

### 1. Data Ingestion & ETL Pipeline

#### Idempotency & Reprocessing

**Design Decision**: Implement idempotency through unique record identification and deduplication logic.

**Implementation**:
- **ADS Records**: Composite key of `date|campaign_id|channel`
- **CRM Records**: Primary key on `opportunity_id`
- **Deduplication Strategy**: First occurrence wins, subsequent duplicates are marked with quality issues

**Benefits**:
- Safe re-running of ETL jobs without data corruption
- Audit trail of duplicate detection attempts
- Graceful handling of upstream data inconsistencies

#### Data Quality Validation Framework

**Field-Level Validation**:
```go
type FieldQuality struct {
    IsValid       bool        `json:"is_valid"`
    Description   string      `json:"description"`
    OriginalValue interface{} `json:"original_value"`
}
```

**Validation Rules**:
- **Missing Fields**: Null/empty values → "Missing" status with fallback values
- **Format Validation**: Date parsing, email regex, numeric range checks
- **Business Rules**: Valid channel types, stage progressions, non-negative costs
- **Cross-Field Validation**: Clicks ≤ Impressions, Amount consistency by stage

### 2. Storage Strategy

#### Partitioning & Retention

**Current Implementation**: In-memory storage with thread-safe operations

**Design Rationale**:
- **Development Simplicity**: Rapid prototyping without external dependencies
- **Performance**: Sub-millisecond query response times
- **Stateless Operation**: Each container restart provides clean state

**Production Evolution Path**:
```
Phase 1 (Current): In-Memory Storage
    ↓
Phase 2: Redis Cluster (Cache Layer)
    ↓
Phase 3: PostgreSQL (Persistent Storage)
    ↓
Phase 4: Partitioned Data Warehouse (BigQuery/Snowflake)
```

**Partitioning Strategy** (Future):
- **Time-based**: Daily/weekly partitions for ads and CRM data
- **Campaign-based**: Separate partitions for high-volume campaigns

### 3. Concurrency & Throughput

#### Goroutines & Worker Pools

**Current Design**:
- **HTTP Server**: Gin's built-in goroutine per request model
- **API Calls**: Sequential processing with retry logic

**Scaling Considerations**:
```go
// Future worker pool implementation
type WorkerPool struct {
    workers   int
    jobs      chan ETLJob
    results   chan ETLResult
    wg        sync.WaitGroup
}

// Parallel processing design
func (w *WorkerPool) ProcessBatch(records []RawRecord) {
    // Distribute records across workers
    // Collect results with error aggregation
}
```

**Performance Targets**:
- **Throughput**: 10,000 records/second per worker
- **Latency**: P95 < 500ms for metrics queries
- **Concurrency**: Support 100 concurrent API requests
using JMeter for non functional test

### 4. Data Quality & Fallbacks

#### Missing UTM Handling

**Problem**: UTM parameters are critical for attribution but frequently missing or inconsistent.

**Solution Strategy**:
1. **Standardization**: Convert null/empty → "unknown" 
2. **Fuzzy Matching**: Campaign name similarity for missing UTM campaigns
3. **Fallback Attribution**: Channel-based grouping when UTMs unavailable
4. **Quality Scoring**: Penalize records with missing attribution data

**UTM Key Generation**:
```go
func generateUTMKey(campaign, source, medium string) string {
    // Normalize empty values
    if strings.TrimSpace(campaign) == "" { campaign = "unknown" }
    if source == "" { source = "unknown" }
    if medium == "" { medium = "unknown" }
    
    return fmt.Sprintf("%s|%s|%s", campaign, source, medium)
}
```

### 5. Observability Strategy

#### Structured Logging

**Log Format**: JSON with consistent field naming
```json
{
  "level": "info",
  "timestamp": "2025-09-19T14:30:00Z",
  "correlation_id": "req_abc123",
  "component": "transformer",
  "operation": "normalize_ads",
  "records_processed": 150,
  "quality_score": 94.2,
  "duration_ms": 45
}
```

**Key Metrics**:
- **Ingestion Metrics**: Records processed, error rates, processing time
- **Quality Metrics**: Validation failure rates by field, overall quality scores
- **API Metrics**: Request latency, success rates, concurrent connections
- **Business Metrics**: CPC/CPA trends, attribution coverage, revenue accuracy

#### Health Monitoring

**Readiness vs. Liveness**:
- **`/healthz`**: Basic service availability (always responds if running)
- **`/readyz`**: Business readiness (has processed data, external APIs reachable)

## Evolution in Admira's Ecosystem

### Integration Patterns

#### Current State
- **Standalone Service**: Self-contained ETL with mock data sources
- **Direct API**: Synchronous request/response pattern
- **In-Memory**: Ephemeral data storage

#### Short-term Evolution (3-6 months)
```
ETL Service → Message Queue → Data Lake
     ↓              ↓           ↓
   Kafka        Streaming      S3/GCS
  Producer      Analytics     Parquet
```

#### Long-term Architecture (1-2 years)
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Source    │ →  │  Streaming  │ →  │    Data     │
│  Connectors │    │    ETL      │    │    Lake     │
└─────────────┘    └─────────────┘    └─────────────┘
       ↓                  ↓                  ↓
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Schema    │    │   Quality   │    │  Analytics  │
│  Registry   │    │  Monitoring │    │   Engine    │
└─────────────┘    └─────────────┘    └─────────────┘
```

### API Contract Evolution

**Current API**: Synchronous REST endpoints
**Future API**: Event-driven with backwards compatibility

```go
// V1 API (Current)
POST /ingest/run → {status, records_processed}

```

***

## Performance & Scalability Analysis

### Bottleneck Identification

**Current Bottlenecks**:
1. **Memory Constraints**: All data stored in RAM
2. **Single-threaded Processing**: Sequential ETL operations  
3. **Blocking API Calls**: Synchronous external API requests

### Capacity Planning

**Current Capacity** (Single Instance):
- **Memory**: ~100MB for 10K records
- **CPU**: Single-core utilization during processing
- **Network**: 50 req/sec sustained throughput

***

## Disaster Recovery & Business Continuity

### Failure Scenarios

**Service Failure**:
- **Detection**: Health check failures
- **Recovery**: Container restart with clean state
- **Data Loss**: Acceptable for development phase

**External API Failure**:
- **Detection**: HTTP error codes or timeouts
- **Recovery**: Exponential backoff retry with circuit breaker
- **Fallback**: Graceful degradation with cached data

**Data Corruption**:
- **Detection**: Quality validation failures above threshold
- **Recovery**: Re-ingestion from source systems
- **Prevention**: Comprehensive field-level validation

### Backup Strategy

**Current**: No persistent data backup needed (stateless service)

***

## Conclusion

The Admira ETL Service implements a robust, scalable architecture for marketing data integration with comprehensive quality validation. The current design prioritizes development velocity and operational simplicity while providing clear evolution paths for production scalability.

**Key Architectural Decisions**:
1. **Layered Architecture**: Clear separation of concerns for maintainability
2. **Quality-First Design**: Field-level validation with detailed error reporting  
3. **Container-Native**: Docker-first approach for deployment flexibility
4. **API-Driven**: REST endpoints for easy integration and testing

---

**Document Version**: 1.0  
**Last Updated**: September 19, 2025  
**Review Schedule**: Quarterly or upon major architectural changes