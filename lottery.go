package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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

// getLotteryPrizeAmounts retrieves prize amounts for a specific date and lottery type
// This function handles the main logic for fetching lottery prize data
func getLotteryPrizeAmounts(date string, lotteryType string) (*PrizeResponse, error) {
	// Validate lottery type - currently supporting megamillions and powerball
	lotteryTypeLower := strings.ToLower(lotteryType)
	if lotteryTypeLower != "megamillions" && lotteryTypeLower != "powerball" {
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: lotteryType,
			Error:       "Unsupported lottery type. Currently only 'megamillions' and 'powerball' are supported.",
		}, nil
	}

	// Parse the date to ensure it's in the correct format
	parsedDate, err := time.Parse("01/02/2006", date)
	if err != nil {
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: lotteryType,
			Error:       fmt.Sprintf("Invalid date format. Please use MM/DD/YYYY format. Error: %v", err),
		}, nil
	}

	// Handle different lottery types
	if lotteryTypeLower == "megamillions" {
		return getMegaMillionsPrizeAmounts(date, parsedDate)
	} else if lotteryTypeLower == "powerball" {
		return getPowerballPrizeAmounts(date, parsedDate)
	}

	return &PrizeResponse{
		Success:     false,
		Date:        date,
		LotteryType: lotteryType,
		Error:       "Unknown lottery type",
	}, nil
}

// getMegaMillionsPrizeAmounts retrieves prize amounts for Mega Millions using the API
// This function uses the existing API endpoints to get prize information
func getMegaMillionsPrizeAmounts(date string, parsedDate time.Time) (*PrizeResponse, error) {
	// Format date for the API call (MM/DD/YYYY)
	formattedDate := parsedDate.Format("01/02/2006")

	// Step 1: Get drawing data using the first API endpoint
	drawingData, err := getDrawingPagingData(formattedDate)
	if err != nil {
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: "megamillions",
			Error:       fmt.Sprintf("Failed to get drawing data: %v", err),
		}, nil
	}

	// Check if we have any drawing data
	if len(drawingData.DrawingData) == 0 {
		return &PrizeResponse{
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
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: "megamillions",
			Error:       fmt.Sprintf("Failed to get detailed draw data: %v", err),
		}, nil
	}

	// Parse the prize information from the detailed data
	prizeInfo, err := parseMegaMillionsPrizeData(detailedData, drawingItem.PlayDate)
	if err != nil {
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: "megamillions",
			Error:       fmt.Sprintf("Failed to parse prize data: %v", err),
		}, nil
	}

	return &PrizeResponse{
		Success:     true,
		Date:        date,
		LotteryType: "megamillions",
		PrizeInfo:   prizeInfo,
	}, nil
}

// getPowerballPrizeAmounts retrieves prize amounts for Powerball by scraping the website
// This function scrapes the Powerball draw result page for prize information
func getPowerballPrizeAmounts(date string, parsedDate time.Time) (*PrizeResponse, error) {
	// Format date for Powerball URL (YYYY-MM-DD)
	formattedDate := parsedDate.Format("2006-01-02")

	// Construct the Powerball URL
	powerballURL := fmt.Sprintf("https://www.powerball.com/draw-result?gc=powerball&date=%s&oc=fl", formattedDate)

	// Scrape the Powerball page for prize information
	prizeInfo, err := scrapePowerballPrizePage(powerballURL)
	if err != nil {
		return &PrizeResponse{
			Success:     false,
			Date:        date,
			LotteryType: "powerball",
			Error:       fmt.Sprintf("Failed to scrape Powerball prize data: %v", err),
		}, nil
	}

	return &PrizeResponse{
		Success:     true,
		Date:        date,
		LotteryType: "powerball",
		PrizeInfo:   prizeInfo,
	}, nil
}

// scrapePowerballPrizePage scrapes the Powerball draw result page to extract prize information
// This function parses the HTML to find jackpot amounts, cash values, and prize tier information
func scrapePowerballPrizePage(url string) (*PrizeInfo, error) {
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
	fmt.Printf("Scraping Powerball prize page: %s\n", url)
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

	// Save HTML content to file for debugging (optional)
	htmlContent := string(body)
	// Debug file creation removed for production use

	// Parse HTML to extract prize information
	prizeInfo, err := parsePowerballPrizeHTML(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Powerball prize HTML: %v", err)
	}

	return prizeInfo, nil
}

// parsePowerballPrizeHTML parses the HTML content to extract prize information
// This function uses regex patterns to find jackpot amounts, cash values, and prize tier data
func parsePowerballPrizeHTML(htmlContent string) (*PrizeInfo, error) {
	// Extract estimated jackpot - try multiple patterns to be more flexible
	var jackpotMatches []string
	jackpotPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<span class="prize-label">\s*Estimated Jackpot:\s*</span>\s*<span>([^<]+)</span>`),
		regexp.MustCompile(`Estimated Jackpot[^<]*<[^>]*>([^<]+)</[^>]*>`),
		regexp.MustCompile(`Jackpot[^<]*<[^>]*>([^<]+)</[^>]*>`),
		regexp.MustCompile(`\$[\d,]+(?:\.\d{2})?\s*[Mm]illion`),
	}

	for _, pattern := range jackpotPatterns {
		if matches := pattern.FindStringSubmatch(htmlContent); len(matches) > 0 {
			jackpotMatches = matches
			break
		}
	}

	// Extract cash value - try multiple patterns to be more flexible
	var cashValueMatches []string
	cashValuePatterns := []*regexp.Regexp{
		regexp.MustCompile(`<span class="prize-label">\s*Cash Value:\s*</span>\s*<span>([^<]+)</span>`),
		regexp.MustCompile(`Cash Value[^<]*<[^>]*>([^<]+)</[^>]*>`),
		regexp.MustCompile(`Cash[^<]*<[^>]*>([^<]+)</[^>]*>`),
	}

	for _, pattern := range cashValuePatterns {
		if matches := pattern.FindStringSubmatch(htmlContent); len(matches) > 0 {
			cashValueMatches = matches
			break
		}
	}

	// Extract date from the page
	datePattern := regexp.MustCompile(`<h5 class="card-title mx-auto mb-3 lh-1 text-center  title-date">([^<]+)</h5>`)
	dateMatches := datePattern.FindStringSubmatch(htmlContent)

	// Parse the extracted values
	var estimatedJackpot string
	var cashValue string

	if len(jackpotMatches) >= 2 {
		estimatedJackpot = strings.TrimSpace(jackpotMatches[1])
	}

	if len(cashValueMatches) >= 2 {
		cashValue = strings.TrimSpace(cashValueMatches[1])
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

	// Extract prize tier information from the table
	prizeTiers, err := extractPowerballPrizeTiers(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract prize tiers: %v", err)
	}

	// Create and return the prize information structure
	prizeInfo := &PrizeInfo{
		PlayDate:         playDate,
		EstimatedJackpot: estimatedJackpot,
		CashValue:        cashValue,
		PrizeTiers:       prizeTiers,
		UpdatedBy:        "POWERBALL_SCRAPER",
		UpdatedTime:      time.Now().Format("2006-01-02T15:04:05"),
	}

	return prizeInfo, nil
}

// extractPowerballPrizeTiers extracts prize tier information from the Powerball HTML table
// This function parses the prize table to get match descriptions, winners, and prize amounts
func extractPowerballPrizeTiers(htmlContent string) ([]PrizeTier, error) {
	var prizeTiers []PrizeTier

	// Based on the actual HTML structure, we need to extract rows from the table
	// Each row has 5 cells: Match, Powerball Winners, Powerball Prize, Power Play Winners, Power Play Prize

	// Pattern to match table rows with the data-label attributes and extract the match pattern from CSS classes
	rowPattern := regexp.MustCompile(`<tr>\s*<td[^>]*data-label="Match"[^>]*>.*?item-powerball\s+([^"\s]+)[^>]*>.*?</td>\s*<td[^>]*data-label="Powerball Winners"[^>]*>\s*(\d*)\s*</td>\s*<td[^>]*data-label="Powerball Prize"[^>]*>\s*([^<]+)\s*</td>\s*<td[^>]*data-label="Power Play Winners"[^>]*>\s*(\d*)\s*</td>\s*<td[^>]*data-label="Power Play Prize"[^>]*>\s*([^<]*)\s*</td>\s*</tr>`)
	matches := rowPattern.FindAllStringSubmatch(htmlContent, -1)

	// Found prize tier rows with regex pattern

	// If we found structured data with data-label attributes, use that
	if len(matches) > 0 {
		for _, match := range matches {
			if len(match) >= 6 {
				// Extract match pattern, winner counts and prizes
				matchPattern := strings.TrimSpace(match[1])
				powerballWinnersStr := strings.TrimSpace(match[2])
				powerballPrize := strings.TrimSpace(match[3])
				powerPlayWinnersStr := strings.TrimSpace(match[4])
				powerPlayPrize := strings.TrimSpace(match[5])

				// Parse winner counts (handle empty strings)
				powerballWinners := 0
				if powerballWinnersStr != "" {
					if w, err := strconv.Atoi(powerballWinnersStr); err == nil {
						powerballWinners = w
					}
				}

				powerPlayWinners := 0
				if powerPlayWinnersStr != "" {
					if w, err := strconv.Atoi(powerPlayWinnersStr); err == nil {
						powerPlayWinners = w
					}
				}

				// Determine match description based on the CSS class pattern
				matchDesc := determinePowerballMatchDescriptionFromPattern(matchPattern, powerballPrize)

				prizeTier := PrizeTier{
					Match:            matchDesc,
					PowerballWinners: powerballWinners,
					PowerballPrize:   powerballPrize,
					PowerPlayWinners: powerPlayWinners,
					PowerPlayPrize:   powerPlayPrize,
				}

				prizeTiers = append(prizeTiers, prizeTier)
			}
		}
	}

	// If we still don't have prize tiers, try a more flexible approach
	if len(prizeTiers) == 0 {

		// Look for table cells with data-label attributes more broadly
		cellPattern := regexp.MustCompile(`<td[^>]*data-label="([^"]*)"[^>]*>\s*([^<]*)\s*</td>`)
		cellMatches := cellPattern.FindAllStringSubmatch(htmlContent, -1)

		// Found cells with data-label attributes

		// Group cells by their data-label
		labelGroups := make(map[string][]string)
		for _, match := range cellMatches {
			if len(match) >= 3 {
				label := strings.TrimSpace(match[1])
				value := strings.TrimSpace(match[2])
				labelGroups[label] = append(labelGroups[label], value)
			}
		}

		// Group cells by their data-label

		// Try to extract prize tiers from the grouped data
		if powerballWinners, ok := labelGroups["Powerball Winners"]; ok {
			if powerballPrizes, ok2 := labelGroups["Powerball Prize"]; ok2 {
				if powerPlayWinners, ok3 := labelGroups["Power Play Winners"]; ok3 {
					if powerPlayPrizes, ok4 := labelGroups["Power Play Prize"]; ok4 {
						// We have all the data we need
						maxLen := len(powerballWinners)
						if len(powerballPrizes) < maxLen {
							maxLen = len(powerballPrizes)
						}
						if len(powerPlayWinners) < maxLen {
							maxLen = len(powerPlayWinners)
						}
						if len(powerPlayPrizes) < maxLen {
							maxLen = len(powerPlayPrizes)
						}

						// Creating prize tiers from grouped data

						for i := 0; i < maxLen; i++ {
							// Parse winner counts (handle empty strings)
							powerballWinnersCount := 0
							if i < len(powerballWinners) && powerballWinners[i] != "" {
								if w, err := strconv.Atoi(powerballWinners[i]); err == nil {
									powerballWinnersCount = w
								}
							}

							powerPlayWinnersCount := 0
							if i < len(powerPlayWinners) && powerPlayWinners[i] != "" {
								if w, err := strconv.Atoi(powerPlayWinners[i]); err == nil {
									powerPlayWinnersCount = w
								}
							}

							prize := ""
							if i < len(powerballPrizes) {
								prize = strings.TrimSpace(powerballPrizes[i])
							}

							ppPrize := ""
							if i < len(powerPlayPrizes) {
								ppPrize = strings.TrimSpace(powerPlayPrizes[i])
							}

							matchDesc := determinePowerballMatchDescription(prize)

							prizeTier := PrizeTier{
								Match:            matchDesc,
								PowerballWinners: powerballWinnersCount,
								PowerballPrize:   prize,
								PowerPlayWinners: powerPlayWinnersCount,
								PowerPlayPrize:   ppPrize,
							}

							prizeTiers = append(prizeTiers, prizeTier)
						}
					}
				}
			}
		}
	}

	// If still no matches, create basic prize tiers based on standard Powerball structure
	if len(prizeTiers) == 0 {
		prizeTiers = []PrizeTier{
			{
				Match:            "5+1 (Jackpot)",
				PowerballWinners: 0,
				PowerballPrize:   "Grand Prize",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "",
			},
			{
				Match:            "5+0",
				PowerballWinners: 0,
				PowerballPrize:   "$1,000,000",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$2,000,000",
			},
			{
				Match:            "4+1",
				PowerballWinners: 0,
				PowerballPrize:   "$50,000",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$200,000",
			},
			{
				Match:            "4+0",
				PowerballWinners: 0,
				PowerballPrize:   "$100",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$400",
			},
			{
				Match:            "3+1",
				PowerballWinners: 0,
				PowerballPrize:   "$100",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$400",
			},
			{
				Match:            "3+0",
				PowerballWinners: 0,
				PowerballPrize:   "$7",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$28",
			},
			{
				Match:            "2+1",
				PowerballWinners: 0,
				PowerballPrize:   "$7",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$28",
			},
			{
				Match:            "1+1",
				PowerballWinners: 0,
				PowerballPrize:   "$4",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$16",
			},
			{
				Match:            "0+1",
				PowerballWinners: 0,
				PowerballPrize:   "$4",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "$16",
			},
		}
	}

	return prizeTiers, nil
}

// determinePowerballMatchDescription determines the match description based on the prize amount
// This function maps prize amounts to match descriptions for Powerball
func determinePowerballMatchDescription(prize string) string {
	switch strings.TrimSpace(prize) {
	case "Grand Prize":
		return "5+1 (Jackpot)"
	case "$1,000,000":
		return "5+0"
	case "$50,000":
		return "4+1"
	case "$100":
		return "4+0 or 3+1"
	case "$7":
		return "3+0 or 2+1"
	case "$4":
		return "1+1 or 0+1"
	default:
		return "Unknown Match"
	}
}

// determinePowerballMatchDescriptionFromPattern determines the match description based on the CSS class pattern
// This function maps CSS class patterns to specific match descriptions for Powerball
func determinePowerballMatchDescriptionFromPattern(pattern, prize string) string {
	switch pattern {
	case "m5-pb":
		return "5+1 (Jackpot)"
	case "m5":
		return "5+0"
	case "m4-pb":
		return "4+1"
	case "m4":
		return "4+0"
	case "m3-pb":
		return "3+1"
	case "m3":
		return "3+0"
	case "m2-pb":
		return "2+1"
	case "m2":
		return "2+0"
	case "m1-pb":
		return "1+1"
	case "m1":
		return "1+0"
	case "m0-pb":
		return "0+1"
	case "m0":
		return "0+0"
	default:
		// Fallback to prize-based description if pattern is not recognized
		return determinePowerballMatchDescription(prize)
	}
}

// parseMegaMillionsPrizeData parses prize information from Mega Millions API response
// This function extracts prize tier information from the detailed draw data
func parseMegaMillionsPrizeData(detailedData *DetailedDrawData, playDate string) (*PrizeInfo, error) {

	// Extract the actual jackpot value from the API response
	var jackpotValue string
	if detailedData.Jackpot != nil {
		// Convert the jackpot value to string, handling different possible types
		switch v := detailedData.Jackpot.(type) {
		case string:
			jackpotValue = v
		case float64:
			if v >= 1000000 {
				jackpotValue = fmt.Sprintf("$%.0f Million", v/1000000)
			} else if v >= 1000 {
				jackpotValue = fmt.Sprintf("$%.0f Thousand", v/1000)
			} else {
				jackpotValue = fmt.Sprintf("$%.0f", v)
			}
		case int:
			if v >= 1000000 {
				jackpotValue = fmt.Sprintf("$%d Million", v/1000000)
			} else if v >= 1000 {
				jackpotValue = fmt.Sprintf("$%d Thousand", v/1000)
			} else {
				jackpotValue = fmt.Sprintf("$%d", v)
			}
		case map[string]interface{}:
			// Handle the case where Jackpot is a map with CurrentPrizePool
			if currentPrizePool, ok := v["CurrentPrizePool"]; ok {
				switch poolValue := currentPrizePool.(type) {
				case float64:
					if poolValue >= 1000000 {
						jackpotValue = fmt.Sprintf("$%.0f Million", poolValue/1000000)
					} else if poolValue >= 1000 {
						jackpotValue = fmt.Sprintf("$%.0f Thousand", poolValue/1000)
					} else {
						jackpotValue = fmt.Sprintf("$%.0f", poolValue)
					}
				case int:
					if poolValue >= 1000000 {
						jackpotValue = fmt.Sprintf("$%d Million", poolValue/1000000)
					} else if poolValue >= 1000 {
						jackpotValue = fmt.Sprintf("$%d Thousand", poolValue/1000)
					} else {
						jackpotValue = fmt.Sprintf("$%d", poolValue)
					}
				default:
					jackpotValue = fmt.Sprintf("%v", poolValue)
				}
			} else {
				jackpotValue = "Unknown"
			}
		default:
			jackpotValue = fmt.Sprintf("%v", v)
		}
	} else {
		jackpotValue = "Unknown"
	}

	prizeTiers := []PrizeTier{
		{
			Match:               "5+1 (Jackpot)",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   jackpotValue,
		},
		{
			Match:               "5+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$1,000,000",
		},
		{
			Match:               "4+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10,000",
		},
		{
			Match:               "4+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$500",
		},
		{
			Match:               "3+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$200",
		},
		{
			Match:               "3+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10",
		},
		{
			Match:               "2+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10",
		},
		{
			Match:               "1+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$4",
		},
		{
			Match:               "0+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$2",
		},
	}

	// Also extract cash value if available
	var cashValue string
	if detailedData.Jackpot != nil {
		if jackpotMap, ok := detailedData.Jackpot.(map[string]interface{}); ok {
			if currentCashValue, ok := jackpotMap["CurrentCashValue"]; ok {
				switch cashVal := currentCashValue.(type) {
				case float64:
					if cashVal >= 1000000 {
						cashValue = fmt.Sprintf("$%.1f Million", cashVal/1000000)
					} else if cashVal >= 1000 {
						cashValue = fmt.Sprintf("$%.0f Thousand", cashVal/1000)
					} else {
						cashValue = fmt.Sprintf("$%.0f", cashVal)
					}
				case int:
					if cashVal >= 1000000 {
						cashValue = fmt.Sprintf("$%.1f Million", float64(cashVal)/1000000)
					} else if cashVal >= 1000 {
						cashValue = fmt.Sprintf("$%.0f Thousand", cashVal/1000)
					} else {
						cashValue = fmt.Sprintf("$%d", cashVal)
					}
				default:
					cashValue = fmt.Sprintf("%v", cashVal)
				}
			}
		}
	}

	prizeInfo := &PrizeInfo{
		PlayDate:         playDate,
		EstimatedJackpot: jackpotValue,
		CashValue:        cashValue,
		PrizeTiers:       prizeTiers,
		UpdatedBy:        "MEGAMILLIONS_API",
		UpdatedTime:      time.Now().Format("2006-01-02T15:04:05"),
	}

	return prizeInfo, nil
}
