#!/bin/sh

send_message() {
    aws --endpoint-url="http://localhost:4566" sqs send-message --queue-url="http://localhost:4566/000000000000/demand" --message-body '
{
    "name": "Name",
    "age": 20
}
' >> /dev/null
    echo "Message sent"
}

echo "Press [CTRL+C] to exit this loop..."

while :
do
    send_message
    sleep 3
done