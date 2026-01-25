CREATE SCHEMA IF NOT EXISTS crypto;

CREATE TABLE IF NOT EXISTS crypto.crypto_daily_metrics (
    id VARCHAR(50),
    symbol VARCHAR(10),
    current_price DECIMAL(18, 8),
    market_cap BIGINT,
    high_24h DECIMAL(18, 8),
    low_24h DECIMAL(18, 8),
    is_volatile BOOLEAN,
    ingestion_date DATE,
    PRIMARY KEY (id, ingestion_date)
);
