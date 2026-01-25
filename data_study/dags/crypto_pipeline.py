import json
import requests
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.bash import BashOperator
from datetime import datetime

def fetch_crypto_data():
    # Fetch top 50 coins
    url = "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=50&page=1&sparkline=false"
    response = requests.get(url)
    data = response.json()
    
    # Save raw data with timestamp (Idempotency!)
    date_str = datetime.now().strftime("%Y-%m-%d")
    with open(f'/opt/airflow/data/raw_crypto_{date_str}.json', 'w') as f:
        json.dump(data, f)

with DAG('crypto_etl', start_date=datetime(2026, 1, 24), schedule_interval='@daily') as dag:
    extract_task = PythonOperator(
        task_id='extract_data',
        python_callable=fetch_crypto_data
    )
    transform_task = BashOperator(
        task_id='transform_data',
        bash_command="""
        python /opt/airflow/scripts/process_crypto.py
        """
    )
    cleanup_task = BashOperator(
        task_id='archive_data',
        bash_command="mkdir -p /opt/airflow/data/processed && mv /opt/airflow/data/*.json /opt/airflow/data/processed/"
    )

    extract_task >> transform_task >> cleanup_task
