from pyspark.sql import SparkSession
from pyspark.sql.functions import col, current_date, when

# Initialize Spark
spark = SparkSession.builder \
    .appName("Crypto ETL") \
    .config("spark.jars", "/opt/airflow/scripts/postgresql-42.6.0.jar") \
    .getOrCreate()

# Read the raw JSON we just downloaded
df = spark.read.json("/opt/airflow/data/raw_crypto_*.json")

# Transformation: Let's create a "Volatility Alert" flag
# If high_24h is 5% higher than low_24h, mark as volatile
df_transformed = df.withColumn(
    "is_volatile", 
    when((col("high_24h") - col("low_24h")) / col("low_24h") > 0.05, True).otherwise(False)
).withColumn("ingestion_date", current_date())

# Select only clean columns for our warehouse
final_df = df_transformed.select(
    "id", "symbol", "current_price", "market_cap", "high_24h", "low_24h", "is_volatile", "ingestion_date"
)

# Write to Postgres (JDBC)
final_df.write \
    .format("jdbc") \
    .option("url", "jdbc:postgresql://postgres:5432/datalake") \
    .option("driver", "org.postgresql.Driver") \
    .option("dbtable", "crypto.crypto_daily_metrics") \
    .option("user", "user") \
    .option("password", "password") \
    .mode("append") \
    .save()

spark.stop()
