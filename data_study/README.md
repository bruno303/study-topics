# Data Study - Cryptocurrency ETL Pipeline

A data engineering study project implementing a complete ETL (Extract, Transform, Load) pipeline for cryptocurrency market data using Apache Airflow, PySpark, PostgreSQL, and Metabase.

## Overview

This project demonstrates a modern data engineering stack by:
- **Extracting** cryptocurrency market data from the CoinGecko API
- **Transforming** the data using PySpark to calculate volatility metrics
- **Loading** processed data into a PostgreSQL data warehouse
- **Visualizing** insights through Metabase dashboards

## Architecture

```
┌─────────────┐      ┌──────────────┐      ┌────────────┐      ┌──────────────┐
│  CoinGecko  │ ───> │   Airflow    │ ───> │   PySpark  │ ───> │  PostgreSQL  │
│     API     │      │ Orchestrator │      │ Transform  │      │  Warehouse   │
└─────────────┘      └──────────────┘      └────────────┘      └──────────────┘
                                                                        │
                                                                        ↓
                                                               ┌──────────────┐
                                                               │   Metabase   │
                                                               │  Dashboard   │
                                                               └──────────────┘
```

## Tech Stack

- **Apache Airflow** - Workflow orchestration
- **Apache Spark** (PySpark) - Data transformation
- **PostgreSQL** - Data warehouse
- **Metabase** - Business intelligence and visualization
- **Docker Compose** - Container orchestration
- **Python 3.10** - Programming language

## Project Structure

```
data_study/
├── docker-compose.yml          # Infrastructure definition
├── airflow/
│   └── Dockerfile              # Custom Airflow image with PySpark
├── dags/
│   └── crypto_pipeline.py      # Airflow DAG definition
├── data/
│   └── processed/              # Archived raw data files
├── pg_data/                    # PostgreSQL data volume
├── schemas/
│   └── crypto/
│       └── crypto_daily_metrics.sql  # Table schema
└── scripts/
    └── process_crypto.py       # PySpark transformation script
```

## Getting Started

### Prerequisites

- Docker and Docker Compose installed
- At least 4GB of available RAM
- Ports 5432, 8080, and 3000 available

### Installation

1. Clone the repository:
```bash
cd /home/bruno/dev/projects/study-topics/data_study
```

2. Download the PostgreSQL JDBC driver:
```bash
wget https://jdbc.postgresql.org/download/postgresql-42.6.0.jar -P scripts/
```

3. Start the infrastructure:
```bash
docker-compose up -d
```

4. Wait for services to initialize (1-2 minutes):
```bash
docker-compose logs -f airflow
```

5. Access the services:
   - **Airflow UI**: http://localhost:8080
   - **Metabase**: http://localhost:3000
   - **PostgreSQL**: localhost:5432

### Initial Setup

1. **Create the database schema**:
```bash
docker exec -it postgres psql -U user -d datalake -f /schemas/crypto/crypto_daily_metrics.sql
```

2. **Enable the Airflow DAG**:
   - Navigate to http://localhost:8080
   - Login with the Airflow standalone credentials (check logs)
   - Enable the `crypto_etl` DAG
   - Trigger a manual run

3. **Configure Metabase**:
   - Navigate to http://localhost:3000
   - Complete the initial setup
   - Connect to PostgreSQL:
     - Host: `postgres`
     - Port: `5432`
     - Database: `datalake`
     - Username: `user`
     - Password: `password`

## 📊 Pipeline Details

### DAG: `crypto_etl`

**Schedule**: Daily (`@daily`)

**Tasks**:
1. **extract_data**: Fetches top 50 cryptocurrencies by market cap from CoinGecko API
2. **transform_data**: Processes data with PySpark, calculating volatility metrics
3. **archive_data**: Moves raw files to the processed folder

### Data Schema

The pipeline creates a table `crypto.crypto_daily_metrics` with the following structure:

| Column          | Type           | Description                           |
|----------------|----------------|---------------------------------------|
| id             | VARCHAR(50)    | Cryptocurrency identifier             |
| symbol         | VARCHAR(10)    | Trading symbol (e.g., BTC, ETH)      |
| current_price  | DECIMAL(18,8)  | Current price in USD                 |
| market_cap     | BIGINT         | Total market capitalization          |
| high_24h       | DECIMAL(18,8)  | Highest price in last 24 hours       |
| low_24h        | DECIMAL(18,8)  | Lowest price in last 24 hours        |
| is_volatile    | BOOLEAN        | True if 24h price range > 5%         |
| ingestion_date | DATE           | Date of data ingestion               |

**Primary Key**: (id, ingestion_date)

### Volatility Calculation

A cryptocurrency is marked as volatile (`is_volatile = true`) when:

$$
\frac{high_{24h} - low_{24h}}{low_{24h}} > 0.05
$$

This indicates a price fluctuation greater than 5% in the last 24 hours.

## Sample Queries

### Top 10 Volatile Cryptocurrencies Today
```sql
SELECT symbol, current_price, high_24h, low_24h
FROM crypto.crypto_daily_metrics
WHERE ingestion_date = CURRENT_DATE
  AND is_volatile = true
ORDER BY market_cap DESC
LIMIT 10;
```

### Market Cap Trends Over Time
```sql
SELECT ingestion_date, symbol, market_cap
FROM crypto.crypto_daily_metrics
WHERE symbol IN ('btc', 'eth', 'bnb')
ORDER BY ingestion_date DESC, market_cap DESC;
```

## Troubleshooting

### PySpark JDBC driver missing
The PostgreSQL JDBC driver should be included. If missing, download it:
```bash
wget https://jdbc.postgresql.org/download/postgresql-42.6.0.jar -P scripts/
```

## Learning Objectives

This project covers:
- ETL pipeline design and implementation
- Workflow orchestration with Airflow
- Distributed data processing with Spark
- Data warehousing concepts
- BI tool integration
