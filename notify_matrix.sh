#!/bin/bash

# This is a sample alertscript for zabbix. This script is called by Zabbix which in turn calls the zabbix-matrix-bot using curl

TARGET="$1"
SUBJECT="$2"
MESSAGE="$3"

PAYLOAD="$( jq -nc --arg send_to "$TARGET" --arg subject "$SUBJECT" --arg message "$MESSAGE" '{"send_to": $send_to, "subject": $subject, "message": $message}' )"

curl -s "http://localhost:8080" -d"$PAYLOAD"
