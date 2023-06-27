#!/bin/sh

echo "########### Creating SQS ###########"
awslocal sqs create-queue --queue-name sleuth-test
awslocal sqs create-queue --queue-name sleuth-test-response

echo "########### Listing SQS ###########"
awslocal sqs list-queues

echo "########### Environment ready to use ###########"