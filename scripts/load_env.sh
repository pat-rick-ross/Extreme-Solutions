#!/bin/bash
# Load environment variables from .env file, ignoring comments and empty lines

if [ -f .env ]; then
    # Read .env file line by line
    while IFS= read -r line || [ -n "$line" ]; do
        # Skip comments and empty lines
        if [[ ! -z "$line" ]] && [[ ! "$line" =~ ^[[:space:]]*# ]]; then
            # Export the variable
            export "$line"
        fi
    done < .env
    echo "✅ Environment variables loaded from .env"
else
    echo "❌ .env file not found"
    exit 1
fi
