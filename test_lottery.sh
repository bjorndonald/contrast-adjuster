#!/bin/bash

# Test script for lottery scraping functionality
# Make sure the server is running on port 8080

echo "Testing Lottery Scraping API"
echo "=============================="

# Test 1: Mega Millions
echo -e "\n1. Testing Mega Millions for 08/19/2025"
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "08/19/2025",
    "lottery_type": "megamillions"
  }' | jq '.'

# Test 2: Powerball
echo -e "\n2. Testing Powerball for 08/19/2025"
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "08/19/2025",
    "lottery_type": "powerball"
  }' | jq '.'

# Test 3: Invalid date format
echo -e "\n3. Testing invalid date format"
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-08-19",
    "lottery_type": "megamillions"
  }' | jq '.'

# Test 4: Invalid lottery type
echo -e "\n4. Testing invalid lottery type"
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "08/19/2025",
    "lottery_type": "invalid"
  }' | jq '.'

echo -e "\n=============================="
echo "Testing completed!"
