#!/bin/bash

# Test script for the Lottery Winning Numbers API
# Make sure the server is running on localhost:8080

echo "Testing Lottery Winning Numbers API"
echo "=================================="
echo

# Test 1: Valid Mega Millions request for 08/19/2025
echo "Test 1: Valid Mega Millions request for 08/19/2025"
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/19/2025", "lottery_type": "megamillions"}' \
  | jq '.'
echo
echo

# Test 2: Valid Powerball request for 08/27/2025
echo "Test 2: Valid Powerball request for 08/27/2025"
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/27/2025", "lottery_type": "powerball"}' \
  | jq '.'
echo
echo

# Test 3: Valid Mega Millions request for 08/22/2025
echo "Test 3: Valid Mega Millions request for 08/22/2025"
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/22/2025", "lottery_type": "megamillions"}' \
  | jq '.'
echo
echo

# Test 4: Unsupported lottery type
echo "Test 4: Unsupported lottery type"
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/19/2025", "lottery_type": "lotto"}' \
  | jq '.'
echo
echo

# Test 5: Invalid date format
echo "Test 5: Invalid date format"
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "invalid-date", "lottery_type": "megamillions"}' \
  | jq '.'
echo
echo

# Test 6: Health check
echo "Test 6: Health check"
curl -X GET http://localhost:8080/health | jq '.'
echo
echo

echo "Testing completed!"
