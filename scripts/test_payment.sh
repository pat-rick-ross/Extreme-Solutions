#!/bin/bash

# Configuration
URL="http://localhost:8080/api/v1/payments/initiate"
TOKEN="YOUR_BEARER_TOKEN"

echo "--- Testing Daraja Gateway ---"
curl -i -X POST $URL \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": "00000000-0000-0000-0000-000000000000",
    "amount": 100.00,
    "phone": "2547XXXXXXXX",
    "gateway": "daraja"
  }'

echo -e "\n\n--- Testing Paystack Gateway ---"
curl -i -X POST $URL \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": "00000000-0000-0000-0000-000000000000",
    "amount": 100.00,
    "phone": "2547XXXXXXXX",
    "gateway": "paystack"
  }'