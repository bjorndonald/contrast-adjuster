#!/bin/bash

# Test script for the Lottery Prize Amounts API
# Make sure the server is running on localhost:8080

echo "Testing Lottery Prize Amounts API"
echo "================================="
echo

# Test 1: Powerball prize amounts for 08/27/2025
echo "Test 1: Powerball prize amounts for 08/27/2025"
curl -X POST http://localhost:8080/lottery-prize-amounts \
  -H "Content-Type: application/json" \
  -d '{"date": "08/27/2025", "lottery_type": "powerball"}' \
  | jq '.'
echo
echo

# Test 2: Mega Millions prize amounts for 08/19/2025
echo "Test 2: Mega Millions prize amounts for 08/19/2025"
curl -X POST http://localhost:8080/lottery-prize-amounts \
  -H "Content-Type: application/json" \
  -d '{"date": "08/19/2025", "lottery_type": "megamillions"}' \
  | jq '.'
echo
echo

# Test 3: Powerball prize amounts for 08/20/2025 (as requested)
echo "Test 3: Powerball prize amounts for 08/20/2025"
curl -X POST http://localhost:8080/lottery-prize-amounts \
  -H "Content-Type: application/json" \
  -d '{"date": "08/20/2025", "lottery_type": "powerball"}' \
  | jq '.'
echo
echo

# Test 4: Unsupported lottery type
echo "Test 4: Unsupported lottery type"
curl -X POST http://localhost:8080/lottery-prize-amounts \
  -H "Content-Type: application/json" \
  -d '{"date": "08/27/2025", "lottery_type": "lotto"}' \
  | jq '.'
echo
echo

# Test 5: Invalid date format
echo "Test 5: Invalid date format"
curl -X POST http://localhost:8080/lottery-prize-amounts \
  -H "Content-Type: application/json" \
  -d '{"date": "invalid-date", "lottery_type": "powerball"}' \
  | jq '.'
echo
echo

echo "Testing completed!"
