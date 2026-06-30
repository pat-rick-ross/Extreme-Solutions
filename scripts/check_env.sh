#!/bin/bash

echo "Checking environment variables..."

# Required variables
REQUIRED_VARS=(
    "DB_PASSWORD"
    "JWT_SECRET"
    "MPESA_CONSUMER_KEY"
    "MPESA_CONSUMER_SECRET"
    "MPESA_PASSKEY"
)

MISSING=0
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ] && [ -z "$(grep "^$var=" .env 2>/dev/null | cut -d= -f2)" ]; then
        echo "❌ $var is not set"
        MISSING=1
    else
        echo "✅ $var is set"
    fi
done

if [ $MISSING -eq 1 ]; then
    echo "⚠️  Some required variables are missing. Please check your .env file."
    exit 1
else
    echo "✅ All required environment variables are set!"
fi
