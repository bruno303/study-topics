#!/bin/bash
DOMAIN="subdomain.duckdns.org"
TOKEN="<token>"
TXT_RECORD="_acme-challenge.$DOMAIN"
TXT_VALUE="$CERTBOT_VALIDATION"

# Update DuckDNS with the TXT record
curl -s "https://www.duckdns.org/update?domains=$DOMAIN&token=$TOKEN&txt=$TXT_VALUE"
sleep 30  # Wait for DNS propagation
