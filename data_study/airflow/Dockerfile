FROM apache/airflow:2.7.1-python3.10

USER root

RUN apt-get update && \
    apt-get install -y openjdk-11-jdk-headless && \
    apt-get clean

ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64

USER airflow

RUN pip install pyspark==3.4.1
