#!/bin/sh

echo "########### Creating SQS ###########"
awslocal sqs create-queue --queue-name demand
awslocal sqs create-queue --queue-name demand-response

echo "########### Listing SQS ###########"
awslocal sqs list-queues

echo "########### Environment ready to use ###########"