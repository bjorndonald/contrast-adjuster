package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// getLotteryWinningNumbers retrieves winning numbers for a specific date and lottery type
// This function handles the main logic for fetching lottery data
func getLotteryWinningNumbers(date string, lotteryType string) (*LotteryResponse, error) {
	// Validate lottery type - currently supporting megamillions and powerball
	lotteryTypeLower := strings.ToLower(lotteryType)
	if lotteryTypeLower != "megamillions" && lotteryTypeLower != "powerball" {
		return &LotteryResponse{
			Success:     false,
			Date:        date,
			LotteryType: lotteryType,
			Error:       "Unsupported lottery type. Currently only 'megamillions' and 'powerball' are supported.",
		}, nil
	}

	// Parse the date to ensure it's in the correct format
	parsedDate, err := time.Parse("01/02/2006", date)
	if err != nil {
		return &LotteryResponse{
			Success:     false,
			Date:        date,
			LotteryType: lotteryType,
			Error:       fmt.Sprintf("Invalid date format. Please use MM/DD/YYYY format. Error: %v", err),
		}, nil
	}

	// Handle different lottery types
	if lotteryTypeLower == "megamillions" {
		return getMegaMillionsWinningNumbers(date, parsedDate)
	} else if lotteryTypeLower == "powerball" {
		return getPowerballWinningNumbers(date, parsedDate)
	}

	return &LotteryResponse{
		Success:     false,
		Date:        date,
		LotteryType: lotteryType,
		Error:       "Unknown lottery type",
	}, nil
}

// getMegaMillionsWinningNumbers retrieves winning numbers for Mega Millions
// This function handles the two-step API process for Mega Millions
func getMegaMillionsWinningNumbers(date string, parsedDate time.Time) (*LotteryResponse, error) {
	// Format date for the API call (MM/DD/YYYY)
	formattedDate := parsedDate.Format("01/02/2006")

	// Step 1: Get drawing data using the first API endpoint
	drawingData, err := getDrawingPagingData(formattedDate)
	if err != nil {
		return &LotteryResponse{
			Success:     false,
			Date:        date,
			LotteryType: "megamillions",
			Error:       fmt.Sprintf("Failed to get drawing data: %v", err),
		}, nil
	}

	// Check if we have any drawing data
	if len(drawingData.DrawingData) == 0 {
		return &LotteryResponse{
			Success:     false,
			Date:        date,
			LotteryType: "megamillions",
			Error:       "No drawing data found for the specified date",
		}, nil
	}

	// Get the first drawing item (should be the only one for a specific date)
	drawingItem := drawingData.DrawingData[0]

	// Step 2: Get detailed draw data using the second API endpoint
	detailedData, err := getDrawDataByTickWithMatrix(drawingItem.PlayDateTicks)
	if err != nil {
		// If detailed data fails, we can still return basic winning numbers
		winningNumbers := &WinningNumbers{
			PlayDate:    drawingItem.PlayDate,
			N1:          drawingItem.N1,
			N2:          drawingItem.N2,
			N3:          drawingItem.N3,
			N4:          drawingItem.N4,
			N5:          drawingItem.N5,
			MBall:       drawingItem.MBall,
			Megaplier:   drawingItem.Megaplier,
			UpdatedBy:   drawingItem.UpdatedBy,
			UpdatedTime: drawingItem.UpdatedTime,
		}

		return &LotteryResponse{
			Success:        true,
			Date:           date,
			LotteryType:    "megamillions",
			WinningNumbers: winningNumbers,
		}, nil
	}

	// Use detailed data if available
	winningNumbers := &WinningNumbers{
		PlayDate:    detailedData.Drawing.PlayDate,
		N1:          detailedData.Drawing.N1,
		N2:          detailedData.Drawing.N2,
		N3:          detailedData.Drawing.N3,
		N4:          detailedData.Drawing.N4,
		N5:          detailedData.Drawing.N5,
		MBall:       detailedData.Drawing.MBall,
		Megaplier:   detailedData.Drawing.Megaplier,
		UpdatedBy:   detailedData.Drawing.UpdatedBy,
		UpdatedTime: detailedData.Drawing.UpdatedTime,
	}

	return &LotteryResponse{
		Success:        true,
		Date:           date,
		LotteryType:    "megamillions",
		WinningNumbers: winningNumbers,
	}, nil
}

// getPowerballWinningNumbers retrieves winning numbers for Powerball by scraping the website
// This function scrapes the Powerball draw result page for a specific date
func getPowerballWinningNumbers(date string, parsedDate time.Time) (*LotteryResponse, error) {
	// Format date for Powerball URL (YYYY-MM-DD)
	formattedDate := parsedDate.Format("2006-01-02")

	// Construct the Powerball URL
	powerballURL := fmt.Sprintf("https://www.powerball.com/draw-result?gc=powerball&date=%s&oc=fl", formattedDate)

	// Scrape the Powerball page
	winningNumbers, err := scrapePowerballPage(powerballURL)
	if err != nil {
		return &LotteryResponse{
			Success:     false,
			Date:        date,
			LotteryType: "powerball",
			Error:       fmt.Sprintf("Failed to scrape Powerball data: %v", err),
		}, nil
	}

	return &LotteryResponse{
		Success:        true,
		Date:           date,
		LotteryType:    "powerball",
		WinningNumbers: winningNumbers,
	}, nil
}

// scrapePowerballPage scrapes the Powerball draw result page to extract winning numbers
// This function parses the HTML to find the winning numbers and Power Play multiplier
func scrapePowerballPage(url string) (*WinningNumbers, error) {
	// Create HTTP client with timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Lottery-API/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// Remove Accept-Encoding to prevent gzip compression
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Powerball website returned non-OK status: %d", resp.StatusCode)
	}

	// Read response body and handle compression
	var body []byte
	var err2 error

	// Check if response is gzipped
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		// Handle gzipped response
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		body, err2 = io.ReadAll(gzReader)
		if err2 != nil {
			return nil, fmt.Errorf("failed to read gzipped response body: %v", err2)
		}
	} else {
		// Handle uncompressed response
		body, err2 = io.ReadAll(resp.Body)
		if err2 != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err2)
		}
	}

	// Parse HTML to extract winning numbers
	winningNumbers, err := parsePowerballHTML(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Powerball HTML: %v", err)
	}

	return winningNumbers, nil
}

// parsePowerballHTML parses the HTML content to extract winning numbers
// This function uses regex patterns to find the winning numbers in the HTML
func parsePowerballHTML(htmlContent string) (*WinningNumbers, error) {
	// Extract white ball numbers (first 5 numbers)
	whiteBallPattern := regexp.MustCompile(`<div class="form-control col white-balls item-powerball">(\d+)</div>`)
	whiteBallMatches := whiteBallPattern.FindAllStringSubmatch(htmlContent, -1)

	if len(whiteBallMatches) < 5 {
		return nil, fmt.Errorf("could not find all 5 white ball numbers, found %d", len(whiteBallMatches))
	}

	// Extract Powerball number (6th number)
	powerballPattern := regexp.MustCompile(`<div class="form-control col powerball item-powerball">(\d+)</div>`)
	powerballMatches := powerballPattern.FindStringSubmatch(htmlContent)

	if len(powerballMatches) < 2 {
		return nil, fmt.Errorf("could not find Powerball number")
	}

	// Extract Power Play multiplier
	powerPlayPattern := regexp.MustCompile(`<span class="multiplier">(\d+)x</span>`)
	powerPlayMatches := powerPlayPattern.FindStringSubmatch(htmlContent)

	// Extract date from the page
	datePattern := regexp.MustCompile(`<h5 class="card-title mx-auto mb-3 lh-1 text-center  title-date">([^<]+)</h5>`)
	dateMatches := datePattern.FindStringSubmatch(htmlContent)

	// Parse the extracted numbers
	var n1, n2, n3, n4, n5 int
	var powerball int
	var powerPlay int = -1 // Default to -1 if not found

	// Parse white ball numbers
	if len(whiteBallMatches) >= 5 {
		fmt.Sscanf(whiteBallMatches[0][1], "%d", &n1)
		fmt.Sscanf(whiteBallMatches[1][1], "%d", &n2)
		fmt.Sscanf(whiteBallMatches[2][1], "%d", &n3)
		fmt.Sscanf(whiteBallMatches[3][1], "%d", &n4)
		fmt.Sscanf(whiteBallMatches[4][1], "%d", &n5)
	}

	// Parse Powerball number
	fmt.Sscanf(powerballMatches[1], "%d", &powerball)

	// Parse Power Play multiplier if found
	if len(powerPlayMatches) >= 2 {
		fmt.Sscanf(powerPlayMatches[1], "%d", &powerPlay)
	}

	// Parse date if found
	var playDate string
	if len(dateMatches) >= 2 {
		// Convert the date format (e.g., "Wed, Aug 27, 2025" to ISO format)
		parsedDate, err := time.Parse("Mon, Jan 02, 2006", dateMatches[1])
		if err == nil {
			playDate = parsedDate.Format("2006-01-02T00:00:00")
		} else {
			playDate = time.Now().Format("2006-01-02T00:00:00")
		}
	} else {
		playDate = time.Now().Format("2006-01-02T00:00:00")
	}

	// Create and return the winning numbers structure
	winningNumbers := &WinningNumbers{
		PlayDate:    playDate,
		N1:          n1,
		N2:          n2,
		N3:          n3,
		N4:          n4,
		N5:          n5,
		MBall:       powerball, // For Powerball, MBall represents the Powerball number
		Megaplier:   powerPlay, // For Powerball, Megaplier represents the Power Play multiplier
		UpdatedBy:   "POWERBALL_SCRAPER",
		UpdatedTime: time.Now().Format("2006-01-02T15:04:05"),
	}

	return winningNumbers, nil
}

// getDrawingPagingData calls the first API endpoint to get drawing data by date
// This endpoint returns basic drawing information including the PlayDateTicks needed for the second API call
func getDrawingPagingData(date string) (*DrawingData, error) {
	// API endpoint for getting drawing data
	url := "https://www.megamillions.com/cmspages/utilservice.asmx/GetDrawingPagingData"

	// Prepare the request body
	requestBody := map[string]interface{}{
		"endDate":    date,
		"pageNumber": 1,
		"pageSize":   20,
		"startDate":  date,
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Lottery-API/1.0)")

	// Make the HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	// Parse the outer response structure
	var outerResponse DrawingPagingResponse
	if err := json.Unmarshal(body, &outerResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outer response: %v", err)
	}

	// Parse the inner response structure (the actual drawing data)
	var drawingData DrawingData
	if err := json.Unmarshal([]byte(outerResponse.D), &drawingData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal drawing data: %v", err)
	}

	return &drawingData, nil
}

// getDrawDataByTickWithMatrix calls the second API endpoint to get detailed draw data
// This endpoint provides comprehensive information about a specific drawing using the PlayDateTicks
func getDrawDataByTickWithMatrix(playDateTicks int64) (*DetailedDrawData, error) {
	// API endpoint for getting detailed draw data
	url := "https://www.megamillions.com/cmspages/utilservice.asmx/GetDrawDataByTickWithMatrix"

	// Prepare the request body
	requestBody := map[string]interface{}{
		"PlayDateTicks": playDateTicks,
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Lottery-API/1.0)")

	// Make the HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	// Parse the outer response structure
	var outerResponse DrawDataResponse
	if err := json.Unmarshal(body, &outerResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outer response: %v", err)
	}

	// Parse the inner response structure (the actual detailed draw data)
	var detailedData DetailedDrawData
	if err := json.Unmarshal([]byte(outerResponse.D), &detailedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal detailed draw data: %v", err)
	}

	return &detailedData, nil
}
