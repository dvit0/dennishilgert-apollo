#!/bin/bash

# Prompt for user input
echo "Please enter the MinIO address (e.g., 10.71.29.50:9000):"
read MINIO_ADDRESS
echo "Please enter the Kafka broker address (e.g., 10.71.29.50:9092):"
read KAFKA_BROKER_ADDRESS

echo "Setting up Kafka notifications for MinIO ..."

# Running the Docker command with user inputs
docker container run -it --entrypoint /bin/sh minio/mc -c "
  /usr/bin/mc alias set minio http://${MINIO_ADDRESS} root 7dbcf34172f78a6719f410a64bdb94fc;
  /usr/bin/mc mb minio/apollo-kernels;
  /usr/bin/mc mb minio/apollo-functions;
  /usr/bin/mc mb minio/apollo-dependencies;
  /usr/bin/mc admin config set minio notify_kafka:apollo-fn-upload brokers='${KAFKA_BROKER_ADDRESS}' topic='apollo_function_code_uploaded' queue_dir='' queue_limit='0' enable='on';
  /usr/bin/mc admin service restart minio;
  /usr/bin/mc event add minio/apollo-functions arn:minio:sqs::apollo-fn-upload:kafka --event put;
  exit 0;"