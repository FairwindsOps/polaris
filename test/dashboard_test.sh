#!/bin/bash
set -euo pipefail

go run main.go dashboard --port 3000 --audit-path ./pkg/config/examples &
sleep 30
curl -f http://localhost:3000 > /dev/null
curl -f http://localhost:3000/health > /dev/null
curl -f http://localhost:3000/favicon.ico > /dev/null
curl -f http://localhost:3000/static/css/main.css > /dev/null
curl -f http://localhost:3000/results.json > /dev/null
curl -f http://localhost:3000/details/security > /dev/null
