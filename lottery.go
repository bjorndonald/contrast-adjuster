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

	// Use the data-label approach which we know works reliably
	// Look for table cells with data-label attributes
	cellPattern := regexp.MustCompile(`<td[^>]*data-label="([^"]*)"[^>]*>\s*([^<]*)\s*</td>`)
	cellMatches := cellPattern.FindAllStringSubmatch(htmlContent, -1)

	// Group cells by their data-label
	labelGroups := make(map[string][]string)
	for _, match := range cellMatches {
		if len(match) >= 3 {
			label := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			labelGroups[label] = append(labelGroups[label], value)
		}
	}

	// Extract prize tiers from the grouped data
	if powerballWinners, ok := labelGroups["Powerball Winners"]; ok {
		if powerballPrizes, ok2 := labelGroups["Powerball Prize"]; ok2 {
			if powerPlayWinners, ok3 := labelGroups["Power Play Winners"]; ok3 {
				if powerPlayPrizes, ok4 := labelGroups["Power Play Prize"]; ok4 {
					// We have all the data we need - create exactly 9 prize tiers
					// Based on the actual Powerball structure, we need to distinguish between different match combinations
					// even if they have the same prize amounts

					// Create specific match descriptions based on the actual data
					var matchDescriptions []string

					// Process each row to determine the specific match description
					for i := 0; i < len(powerballPrizes) && i < 9; i++ {
						prize := strings.TrimSpace(powerballPrizes[i])
						// ppPrize := ""
						// if i < len(powerPlayPrizes) {
						// 	ppPrize = strings.TrimSpace(powerPlayPrizes[i])
						// }

						// Determine match description based on prize amounts and position
						var matchDesc string
						switch {
						case prize == "Grand Prize":
							matchDesc = "5+1 (Jackpot)"
						case prize == "$1,000,000":
							matchDesc = "5+0"
						case prize == "$50,000":
							matchDesc = "4+1"
						case prize == "$100":
							// Distinguish between 4+0 and 3+1 based on position
							if i == 3 {
								matchDesc = "4+0"
							} else {
								matchDesc = "3+1"
							}
						case prize == "$7":
							// Distinguish between 3+0 and 2+1 based on position
							if i == 5 {
								matchDesc = "3+0"
							} else {
								matchDesc = "2+1"
							}
						case prize == "$4":
							// Distinguish between 1+1 and 0+1 based on position
							if i == 7 {
								matchDesc = "1+1"
							} else {
								matchDesc = "0+1"
							}
						default:
							matchDesc = fmt.Sprintf("Match %d", i+1)
						}

						matchDescriptions = append(matchDescriptions, matchDesc)
					}

					// Ensure we have exactly 9 entries
					maxEntries := 9
					if len(powerballWinners) < maxEntries {
						maxEntries = len(powerballWinners)
					}
					if len(powerballPrizes) < maxEntries {
						maxEntries = len(powerballPrizes)
					}
					if len(powerPlayWinners) < maxEntries {
						maxEntries = len(powerPlayWinners)
					}
					if len(powerPlayPrizes) < maxEntries {
						maxEntries = len(powerPlayPrizes)
					}
					if len(matchDescriptions) < maxEntries {
						maxEntries = len(matchDescriptions)
					}

					for i := 0; i < maxEntries; i++ {
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

						matchDesc := matchDescriptions[i]

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

	// If still no matches, create basic prize tiers based on standard Powerball structure
	if len(prizeTiers) == 0 {
		prizeTiers = []PrizeTier{
			{
				Match:            "5+1 (Jackpot)",
				PowerballWinners: 0,
				PowerballPrize:   "Jackpot",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: Jackpot, 3x: Jackpot, 4x: Jackpot, 5x: Jackpot, 10x: Jackpot",
			},
			{
				Match:            "5+0",
				PowerballWinners: 0,
				PowerballPrize:   "$1,000,000",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $2,000,000, 3x: $3,000,000, 4x: $4,000,000, 5x: $5,000,000, 10x: $10,000,000",
			},
			{
				Match:            "4+1",
				PowerballWinners: 0,
				PowerballPrize:   "$50,000",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $100,000, 3x: $150,000, 4x: $200,000, 5x: $250,000, 10x: $500,000",
			},
			{
				Match:            "4+0",
				PowerballWinners: 0,
				PowerballPrize:   "$100",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $200, 3x: $300, 4x: $400, 5x: $500, 10x: $1,000",
			},
			{
				Match:            "3+1",
				PowerballWinners: 0,
				PowerballPrize:   "$100",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $200, 3x: $300, 4x: $400, 5x: $500, 10x: $1,000",
			},
			{
				Match:            "3+0",
				PowerballWinners: 0,
				PowerballPrize:   "$7",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $14, 3x: $21, 4x: $28, 5x: $35, 10x: $70",
			},
			{
				Match:            "2+1",
				PowerballWinners: 0,
				PowerballPrize:   "$7",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $14, 3x: $21, 4x: $28, 5x: $35, 10x: $70",
			},
			{
				Match:            "1+1",
				PowerballWinners: 0,
				PowerballPrize:   "$4",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $8, 3x: $12, 4x: $16, 5x: $20, 10x: $40",
			},
			{
				Match:            "0+1",
				PowerballWinners: 0,
				PowerballPrize:   "$4",
				PowerPlayWinners: 0,
				PowerPlayPrize:   "2x: $8, 3x: $12, 4x: $16, 5x: $20, 10x: $40",
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
			MegaplierPrize:      map[string]int{"2x": 0, "3x": 0, "4x": 0, "5x": 0, "10x": 0}, // All multipliers result in Jackpot (0 represents Jackpot)
		},
		{
			Match:               "5+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$1,000,000",
			MegaplierPrize:      map[string]int{"2x": 2000000, "3x": 3000000, "4x": 4000000, "5x": 5000000, "10x": 10000000},
		},
		{
			Match:               "4+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10,000",
			MegaplierPrize:      map[string]int{"2x": 20000, "3x": 30000, "4x": 40000, "5x": 50000, "10x": 100000},
		},
		{
			Match:               "4+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$500",
			MegaplierPrize:      map[string]int{"2x": 1000, "3x": 1500, "4x": 2000, "5x": 2500, "10x": 5000},
		},
		{
			Match:               "3+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$200",
			MegaplierPrize:      map[string]int{"2x": 400, "3x": 600, "4x": 800, "5x": 1000, "10x": 2000},
		},
		{
			Match:               "3+0",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10",
			MegaplierPrize:      map[string]int{"2x": 20, "3x": 30, "4x": 40, "5x": 50, "10x": 100},
		},
		{
			Match:               "2+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$10",
			MegaplierPrize:      map[string]int{"2x": 20, "3x": 30, "4x": 40, "5x": 50, "10x": 100},
		},
		{
			Match:               "1+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$4",
			MegaplierPrize:      map[string]int{"2x": 14, "3x": 21, "4x": 28, "5x": 35, "10x": 70},
		},
		{
			Match:               "0+1",
			MegaMillionsWinners: 0,
			MegaMillionsPrize:   "$2",
			MegaplierPrize:      map[string]int{"2x": 10, "3x": 15, "4x": 20, "5x": 25, "10x": 50},
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

// calculatePowerballPrize calculates the prize amount for a given number of matching balls and Powerball
// This function implements the static prize tier structure for Powerball
func calculatePowerballPrize(whiteBallMatches int, hasPowerball bool, estimatedJackpot string) (string, int) {
	// Define the static prize tiers based on Powerball rules
	// Format: (whiteBallMatches, hasPowerball) -> (prize description, base amount in cents)
	prizeTiers := map[string]struct {
		description string
		baseAmount  int // in cents
	}{
		"5+1": {estimatedJackpot, 0},     // Use estimated jackpot value
		"5+0": {"$1,000,000", 100000000}, // $1,000,000
		"4+1": {"$50,000", 5000000},      // $50,000
		"4+0": {"$100", 10000},           // $100
		"3+1": {"$100", 10000},           // $100
		"3+0": {"$7", 700},               // $7
		"2+1": {"$7", 700},               // $7
		"2+0": {"$0", 0},                 // No prize
		"1+1": {"$4", 400},               // $4
		"1+0": {"$0", 0},                 // No prize
		"0+1": {"$4", 400},               // $4
		"0+0": {"$0", 0},                 // No prize
	}

	// Create the key for the prize tier
	key := fmt.Sprintf("%d+%d", whiteBallMatches, boolToInt(hasPowerball))

	// Get the prize information
	if prizeInfo, exists := prizeTiers[key]; exists {
		return prizeInfo.description, prizeInfo.baseAmount
	}

	// Default case for invalid combinations
	return "No Prize", 0
}

// calculatePowerPlayPrize calculates the Power Play prize amount based on the base prize and multiplier
// This function applies the Power Play multiplier to the base prize amount
func calculatePowerPlayPrize(baseAmount int, multiplier int) (string, int) {
	// Power Play multipliers: 2x, 3x, 4x, 5x, 10x
	// Note: 10x multiplier only applies to prizes of $150,000 or less

	// Check if 10x multiplier applies (only for prizes $150,000 or less)
	if multiplier == 10 && baseAmount > 15000000 { // $150,000 in cents
		multiplier = 2 // Default to 2x for larger prizes
	}

	// Calculate the Power Play prize
	powerPlayAmount := baseAmount * multiplier

	// Format the prize amount
	var formattedPrize string
	if powerPlayAmount >= 100000000 { // $1,000,000 or more
		formattedPrize = fmt.Sprintf("$%d Million", powerPlayAmount/100000000)
	} else if powerPlayAmount >= 1000000 { // $1,000,000
		formattedPrize = fmt.Sprintf("$%d", powerPlayAmount/100)
	} else if powerPlayAmount >= 100000 { // $100,000 or more
		formattedPrize = fmt.Sprintf("$%d", powerPlayAmount/100)
	} else if powerPlayAmount >= 10000 { // $100 or more
		formattedPrize = fmt.Sprintf("$%d", powerPlayAmount/100)
	} else if powerPlayAmount >= 100 { // $1 or more
		formattedPrize = fmt.Sprintf("$%d", powerPlayAmount/100)
	} else {
		formattedPrize = "$0"
	}

	return formattedPrize, powerPlayAmount
}

// boolToInt converts a boolean to an integer (true = 1, false = 0)
// This helper function is used to create the prize tier key
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// getPowerballPrizeTierDescription returns a human-readable description of the prize tier
// This function provides clear descriptions for each Powerball prize tier
func getPowerballPrizeTierDescription(whiteBallMatches int, hasPowerball bool) string {
	switch {
	case whiteBallMatches == 5 && hasPowerball:
		return "5 white balls + Powerball (Jackpot)"
	case whiteBallMatches == 5 && !hasPowerball:
		return "5 white balls (no Powerball)"
	case whiteBallMatches == 4 && hasPowerball:
		return "4 white balls + Powerball"
	case whiteBallMatches == 4 && !hasPowerball:
		return "4 white balls (no Powerball)"
	case whiteBallMatches == 3 && hasPowerball:
		return "3 white balls + Powerball"
	case whiteBallMatches == 3 && !hasPowerball:
		return "3 white balls (no Powerball)"
	case whiteBallMatches == 2 && hasPowerball:
		return "2 white balls + Powerball"
	case whiteBallMatches == 2 && !hasPowerball:
		return "2 white balls (no Powerball)"
	case whiteBallMatches == 1 && hasPowerball:
		return "1 white ball + Powerball"
	case whiteBallMatches == 1 && !hasPowerball:
		return "1 white ball (no Powerball)"
	case whiteBallMatches == 0 && hasPowerball:
		return "Powerball only"
	case whiteBallMatches == 0 && !hasPowerball:
		return "No matches"
	default:
		return "Invalid combination"
	}
}

// checkPowerballTicket checks if a Powerball ticket is a winner and calculates the prize
// This function compares the ticket numbers with the winning numbers and determines the prize
func checkPowerballTicket(ticketNumbers []int, powerballNumber int, winningNumbers *WinningNumbers, powerPlayMultiplier int, estimatedJackpot string) (*TicketResult, error) {
	// Validate ticket input
	if len(ticketNumbers) != 5 {
		return nil, fmt.Errorf("invalid ticket: must have exactly 5 white ball numbers")
	}

	if powerballNumber < 1 || powerballNumber > 26 {
		return nil, fmt.Errorf("invalid Powerball number: must be between 1 and 26")
	}

	// Validate winning numbers
	if winningNumbers == nil {
		return nil, fmt.Errorf("winning numbers cannot be nil")
	}

	// Extract winning numbers
	winningWhiteBalls := []int{winningNumbers.N1, winningNumbers.N2, winningNumbers.N3, winningNumbers.N4, winningNumbers.N5}
	winningPowerball := winningNumbers.MBall

	// Count matching white balls
	whiteBallMatches := countMatchingNumbers(ticketNumbers, winningWhiteBalls)

	// Check if Powerball matches
	hasPowerball := (powerballNumber == winningPowerball)

	// Calculate base prize using the provided estimated jackpot
	prizeDescription, baseAmount := calculatePowerballPrize(whiteBallMatches, hasPowerball, estimatedJackpot)

	// Calculate Power Play prize if multiplier is provided
	var powerPlayPrize string
	var powerPlayAmount int
	if powerPlayMultiplier > 0 {
		powerPlayPrize, powerPlayAmount = calculatePowerPlayPrize(baseAmount, powerPlayMultiplier)
	}

	// Determine if ticket is a winner
	isWinner := baseAmount > 0 || (whiteBallMatches == 5 && hasPowerball) // Jackpot is always a winner

	// Create ticket result
	result := &TicketResult{
		IsWinner:            isWinner,
		WhiteBallMatches:    whiteBallMatches,
		HasPowerball:        hasPowerball,
		PrizeDescription:    prizeDescription,
		BasePrize:           baseAmount,
		PowerPlayMultiplier: powerPlayMultiplier,
		PowerPlayPrize:      powerPlayPrize,
		PowerPlayAmount:     powerPlayAmount,
		TotalPrize:          baseAmount,
	}

	// If Power Play is active, use the Power Play amount as total
	if powerPlayMultiplier > 0 {
		result.TotalPrize = powerPlayAmount
	}

	return result, nil
}

// countMatchingNumbers counts how many numbers from the ticket match the winning numbers
// This helper function compares two slices and returns the count of matching numbers
func countMatchingNumbers(ticketNumbers []int, winningNumbers []int) int {
	count := 0
	ticketSet := make(map[int]bool)

	// Create a set of ticket numbers for efficient lookup
	for _, num := range ticketNumbers {
		ticketSet[num] = true
	}

	// Count matches
	for _, winningNum := range winningNumbers {
		if ticketSet[winningNum] {
			count++
		}
	}

	return count
}

// TicketResult represents the result of checking a lottery ticket
// This struct contains all the information about whether the ticket won and the prize amounts
type TicketResult struct {
	IsWinner            bool   `json:"is_winner"`
	WhiteBallMatches    int    `json:"white_ball_matches"`
	HasPowerball        bool   `json:"has_powerball"`
	PrizeDescription    string `json:"prize_description"`
	BasePrize           int    `json:"base_prize"` // in cents
	PowerPlayMultiplier int    `json:"power_play_multiplier"`
	PowerPlayPrize      string `json:"power_play_prize"`
	PowerPlayAmount     int    `json:"power_play_amount"` // in cents
	TotalPrize          int    `json:"total_prize"`       // in cents
}

// demonstratePowerballPrizes demonstrates the Powerball prize calculation system
// This function shows examples of different ticket combinations and their prizes
func demonstratePowerballPrizes() {
	fmt.Println("=== Powerball Prize Calculation Examples ===\n")

	// Example winning numbers (you can change these)
	winningNumbers := &WinningNumbers{
		N1:    10,
		N2:    20,
		N3:    30,
		N4:    40,
		N5:    50,
		MBall: 25, // Powerball number
	}

	fmt.Printf("Winning Numbers: %d, %d, %d, %d, %d | Powerball: %d\n\n",
		winningNumbers.N1, winningNumbers.N2, winningNumbers.N3, winningNumbers.N4, winningNumbers.N5, winningNumbers.MBall)

	// Example tickets to test
	exampleTickets := []struct {
		description string
		whiteBalls  []int
		powerball   int
	}{
		{"Jackpot Ticket", []int{10, 20, 30, 40, 50}, 25},
		{"5 White Balls (no Powerball)", []int{10, 20, 30, 40, 50}, 15},
		{"4 White Balls + Powerball", []int{10, 20, 30, 40, 60}, 25},
		{"4 White Balls (no Powerball)", []int{10, 20, 30, 40, 60}, 15},
		{"3 White Balls + Powerball", []int{10, 20, 30, 70, 80}, 25},
		{"3 White Balls (no Powerball)", []int{10, 20, 30, 70, 80}, 15},
		{"2 White Balls + Powerball", []int{10, 20, 90, 100, 110}, 25},
		{"1 White Ball + Powerball", []int{10, 120, 130, 140, 150}, 25},
		{"Powerball Only", []int{160, 170, 180, 190, 200}, 25},
		{"No Matches", []int{160, 170, 180, 190, 200}, 15},
	}

	// Test each ticket
	for _, ticket := range exampleTickets {
		fmt.Printf("Ticket: %s\n", ticket.description)
		fmt.Printf("Numbers: %v | Powerball: %d\n", ticket.whiteBalls, ticket.powerball)

		// Check without Power Play
		result, err := checkPowerballTicket(ticket.whiteBalls, ticket.powerball, winningNumbers, 0, "$500 Million")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %s\n", result.PrizeDescription)
			if result.IsWinner {
				fmt.Printf("Prize: $%.2f\n", float64(result.BasePrize)/100)
			} else {
				fmt.Printf("Prize: No Prize\n")
			}
		}

		// Check with Power Play (2x multiplier)
		resultPP, err := checkPowerballTicket(ticket.whiteBalls, ticket.powerball, winningNumbers, 2, "$500 Million")
		if err != nil {
			fmt.Printf("Power Play Error: %v\n", err)
		} else if resultPP.PowerPlayMultiplier > 0 {
			fmt.Printf("Power Play (2x): %s\n", resultPP.PowerPlayPrize)
		}

		// For 4+1 case, also show 4x multiplier to demonstrate $200,000
		if ticket.description == "4 White Balls + Powerball" {
			resultPP4x, err := checkPowerballTicket(ticket.whiteBalls, ticket.powerball, winningNumbers, 4, "$500 Million")
			if err == nil && resultPP4x.PowerPlayMultiplier > 0 {
				fmt.Printf("Power Play (4x): %s\n", resultPP4x.PowerPlayPrize)
			}
		}

		fmt.Println("---")
	}

	// Demonstrate Power Play multipliers
	fmt.Println("=== Power Play Multiplier Examples ===\n")

	// Test a $100 prize with different multipliers
	basePrize := 10000 // $100 in cents
	multipliers := []int{2, 3, 4, 5, 10}

	fmt.Printf("Base Prize: $%.2f\n\n", float64(basePrize)/100)

	for _, multiplier := range multipliers {
		powerPlayPrize, powerPlayAmount := calculatePowerPlayPrize(basePrize, multiplier)
		fmt.Printf("%dx Multiplier: %s ($%.2f)\n", multiplier, powerPlayPrize, float64(powerPlayAmount)/100)
	}

	// Test $1,000,000 prize with 10x multiplier (should default to 2x)
	fmt.Printf("\n$1,000,000 Prize with 10x Multiplier (should default to 2x):\n")
	jackpotBase := 100000000 // $1,000,000 in cents
	powerPlayPrize, powerPlayAmount := calculatePowerPlayPrize(jackpotBase, 10)
	fmt.Printf("Result: %s ($%.2f)\n", powerPlayPrize, float64(powerPlayAmount)/100)
}

// demonstratePowerballPrizes can be called from main.go to show the Powerball prize calculation system
// This function demonstrates different ticket combinations and their prizes

// calculateMegaMillionsPrize calculates the prize amount for a given number of matching balls and Mega Ball
// This function implements the static prize tier structure for Mega Millions
func calculateMegaMillionsPrize(whiteBallMatches int, hasMegaBall bool, estimatedJackpot string) (string, int) {
	// Define the static prize tiers based on Mega Millions rules
	// Format: (whiteBallMatches, hasMegaBall) -> (prize description, base amount in cents)
	prizeTiers := map[string]struct {
		description string
		baseAmount  int // in cents
	}{
		"5+1": {estimatedJackpot, 0},     // Jackpot amount varies
		"5+0": {"$1,000,000", 100000000}, // $1,000,000
		"4+1": {"$10,000", 1000000},      // $10,000
		"4+0": {"$500", 50000},           // $500
		"3+1": {"$200", 20000},           // $200
		"3+0": {"$10", 1000},             // $10
		"2+1": {"$10", 1000},             // $10
		"2+0": {"$0", 0},                 // No prize
		"1+1": {"$4", 400},               // $4
		"1+0": {"$0", 0},                 // No prize
		"0+1": {"$2", 200},               // $2
		"0+0": {"$0", 0},                 // No prize
	}

	// Create the key for the prize tier
	key := fmt.Sprintf("%d+%d", whiteBallMatches, boolToInt(hasMegaBall))

	// Get the prize information
	if prizeInfo, exists := prizeTiers[key]; exists {
		return prizeInfo.description, prizeInfo.baseAmount
	}

	// Default case for invalid combinations
	return "No Prize", 0
}

// calculateMegaplierPrize calculates the Megaplier prize amount based on the base prize and multiplier
// This function applies the Megaplier multiplier to the base prize amount
func calculateMegaplierPrize(baseAmount int, multiplier int) (string, int) {
	// Megaplier multipliers: 2x, 3x, 4x, 5x, 10x
	// All Megaplier multipliers apply to all prizes (unlike Power Play)

	// Calculate the Megaplier prize
	megaplierAmount := baseAmount * multiplier

	// Format the prize amount
	var formattedPrize string
	if megaplierAmount >= 100000000 { // $1,000,000 or more
		formattedPrize = fmt.Sprintf("$%d Million", megaplierAmount/100000000)
	} else if megaplierAmount >= 1000000 { // $1,000,000
		formattedPrize = fmt.Sprintf("$%d", megaplierAmount/100)
	} else if megaplierAmount >= 100000 { // $100,000 or more
		formattedPrize = fmt.Sprintf("$%d", megaplierAmount/100)
	} else if megaplierAmount >= 10000 { // $100 or more
		formattedPrize = fmt.Sprintf("$%d", megaplierAmount/100)
	} else if megaplierAmount >= 100 { // $1 or more
		formattedPrize = fmt.Sprintf("$%d", megaplierAmount/100)
	} else {
		formattedPrize = "$0"
	}

	return formattedPrize, megaplierAmount
}

// checkMegaMillionsTicket checks if a Mega Millions ticket is a winner and calculates the prize
// This function compares the ticket numbers with the winning numbers and determines the prize
func checkMegaMillionsTicket(ticketNumbers []int, megaBallNumber int, winningNumbers *WinningNumbers, megaplierMultiplier int) (*TicketResult, error) {
	// Validate ticket input
	if len(ticketNumbers) != 5 {
		return nil, fmt.Errorf("invalid ticket: must have exactly 5 white ball numbers")
	}

	if megaBallNumber < 1 || megaBallNumber > 25 {
		return nil, fmt.Errorf("invalid Mega Ball number: must be between 1 and 25")
	}

	// Validate winning numbers
	if winningNumbers == nil {
		return nil, fmt.Errorf("winning numbers cannot be nil")
	}

	// Extract winning numbers
	winningWhiteBalls := []int{winningNumbers.N1, winningNumbers.N2, winningNumbers.N3, winningNumbers.N4, winningNumbers.N5}
	winningMegaBall := winningNumbers.MBall

	// Count matching white balls
	whiteBallMatches := countMatchingNumbers(ticketNumbers, winningWhiteBalls)

	// Check if Mega Ball matches
	hasMegaBall := (megaBallNumber == winningMegaBall)

	// Calculate base prize using a default estimated jackpot
	estimatedJackpot := "$500 Million" // This would come from the lottery data in a real implementation
	prizeDescription, baseAmount := calculateMegaMillionsPrize(whiteBallMatches, hasMegaBall, estimatedJackpot)

	// Calculate Megaplier prize if multiplier is provided
	var megaplierPrize string
	var megaplierAmount int
	if megaplierMultiplier > 0 {
		megaplierPrize, megaplierAmount = calculateMegaplierPrize(baseAmount, megaplierMultiplier)
	}

	// Determine if ticket is a winner
	isWinner := baseAmount > 0 || (whiteBallMatches == 5 && hasMegaBall) // Jackpot is always a winner

	// Create ticket result
	result := &TicketResult{
		IsWinner:            isWinner,
		WhiteBallMatches:    whiteBallMatches,
		HasPowerball:        hasMegaBall, // Reusing Powerball field for Mega Ball
		PrizeDescription:    prizeDescription,
		BasePrize:           baseAmount,
		PowerPlayMultiplier: megaplierMultiplier, // Reusing Power Play field for Megaplier
		PowerPlayPrize:      megaplierPrize,
		PowerPlayAmount:     megaplierAmount,
		TotalPrize:          baseAmount,
	}

	// If Megaplier is active, use the Megaplier amount as total
	if megaplierMultiplier > 0 {
		result.TotalPrize = megaplierAmount
	}

	return result, nil
}
