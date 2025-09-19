# Admira ETL Service

A production-ready **Extract, Transform, Load (ETL)** system built in Go that processes marketing campaign data from advertising platforms and CRM systems, with comprehensive data quality validation and business metrics calculation.

## Overview

This ETL service integrates advertising spend data with customer relationship management data to calculate key marketing performance indicators including CPC, CPA, conversion rates, and ROAS. The system includes advanced data quality tracking that validates every field and provides detailed error reporting.

### Key Features

- **Robust Data Ingestion**: Fetches data from external APIs with retry logic and exponential backoff
- **Advanced Data Quality**: Field-level validation with detailed error descriptions and quality scores
- **Business Metrics**: Calculates CPC, CPA, CVR, ROAS and other marketing KPIs
- **REST API**: Clean endpoints for data ingestion, metrics queries, and quality reporting
- **Export Capabilities**: Secure data export with HMAC authentication
- **Docker-Ready**: Full containerization with Docker Compose support

## Architecture

```
┌─────────────────────────────────────────────┐
│             HTTP API Layer (Gin)            │
├─────────────────────────────────────────────┤
│  Business Logic (Handlers, Metrics, Export) │
├─────────────────────────────────────────────┤
│  Data Processing (ETL, Quality Validation)  │
├─────────────────────────────────────────────┤
│  Foundation (Config, Models, Logging)       │
└─────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- **Docker Desktop** (only requirement - no Go installation needed)

### Installation & Setup

1. **Clone or create the project directory**:
```bash
mkdir admira-etl && cd admira-etl
```

2. **Create all the Go source files** (main.go, internal/ structure as provided)

3. **Generate dependencies** (critical step):
```bash
docker run --rm -v ${PWD}:/app -w /app golang:1.21 sh -c "go clean -modcache && go mod tidy"
```

4. **Set up environment**:
```bash
cp .env.example .env
# Edit .env if needed for custom API URLs
```

5. **Build and run**:
```bash
docker compose up --build
```

The service will start on **http://localhost:8080**

## Project Structure

```
admira-etl/
├── main.go                           # Application entry point
├── go.mod & go.sum                   # Go modules & dependencies
├── docker-compose.yml               # Docker orchestration
├── .env.example                      # Environment template
├── Makefile                          # Build automation
└── internal/                         # Private application code
    ├── config/                       # Configuration management
    ├── models/                       # Data structures & types
    ├── client/                       # HTTP client (retry logic)
    ├── transformer/                  # ETL & data quality validation
    ├── storage/                      # In-memory data storage
    ├── handlers/                     # HTTP request handlers
    ├── metrics/                      # Business metrics calculation
    └── export/                       # Data export functionality
```

## API Endpoints

### Health & Status
```bash
GET  /healthz                 # Health check
GET  /readyz                  # Readiness check (has data)
```

### Data Ingestion
```bash
POST /ingest/run              # Trigger ETL pipeline
POST /ingest/run?since=2025-08-01  # Filter data from specific date
```

### Metrics & Analytics
```bash
GET /metrics/channel          # Channel performance metrics
GET /metrics/funnel           # Campaign funnel analysis
```

**Query Parameters**:
- `from` & `to`: Date range (YYYY-MM-DD)
- `channel`: Filter by advertising channel
- `utm_campaign`: Filter by campaign name
- `limit` & `offset`: Pagination

### Data Quality
```bash
GET /quality/report           # Comprehensive data quality analysis
```

### Export
```bash
POST /export/run?date=2025-08-01  # Export daily consolidated data
```

## Testing the System

### 1. Health Check
```bash
curl http://localhost:8080/healthz
```

### 2. Data Ingestion with Quality Validation
```bash
curl -X POST http://localhost:8080/ingest/run
```

### 3. View Data Quality Report
```bash
curl http://localhost:8080/quality/report
```

### 4. Query Metrics
```bash
curl "http://localhost:8080/metrics/channel?from=2025-08-01&to=2025-08-10&channel=google_ads"
```

## Data Quality Features

The system validates every field during ETL processing:

- **Missing Fields**: Automatically detected and flagged as "Missing"
- **Invalid Formats**: Date, email, and numeric validation with error descriptions
- **Duplicates**: Detected and prevented during ingestion
- **Quality Scores**: Calculated at record and dataset levels
- **Detailed Reports**: Field-by-field validation results

Example quality response:
```json
{
  "summary": {
    "overall_quality_score": 87.5,
    "total_ads_records": 16,
    "valid_ads_records": 14,
    "common_issues": ["Missing - UTM Source is null or empty"]
  }
}
```

## Development Commands

```bash
# View logs
docker compose logs -f

# Rebuild after code changes
docker compose up --build

# Stop services
docker compose down

# Clean Docker resources
docker system prune -f

# Generate fresh dependencies
docker run --rm -v ${PWD}:/app -w /app golang:1.21 sh -c "go clean -modcache && go mod tidy"
```

## Configuration

Environment variables (see `.env.example`):

```env
ADS_API_URL=https://mocki.io/v1/9dcc2981-2bc8-465a-bce3-47767e1278e6
CRM_API_URL=https://mocki.io/v1/6a064f10-829d-432c-9f0d-24d5b8cb71c7
SINK_URL=https://httpbin.org/post
SINK_SECRET=admira_secret_example
PORT=8080
LOG_LEVEL=info
HTTP_TIMEOUT=30s
RETRY_ATTEMPTS=3
```

## Production Considerations

- **Monitoring**: JSON structured logging with correlation IDs
- **Reliability**: Exponential backoff retry logic for API calls
- **Security**: HMAC-signed exports for data integrity
- **Performance**: In-memory storage with thread-safe operations
- **Quality**: Comprehensive field validation and error tracking

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure `docker compose up --build` works
6. Submit a pull request

## System Requirements

- **Runtime**: Docker Desktop
- **Memory**: 512MB minimum
- **Network**: Internet access for external API calls

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## External Dependencies

- **Gin Framework**: HTTP web framework
- **Logrus**: Structured logging
- **Testify**: Testing utilities
- **GoDotEnv**: Environment variable management

---

**Built with Go 1.21 | Containerized with Docker | Production Ready**