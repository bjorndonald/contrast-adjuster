package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// scrapeLotteryNumbers scrapes winning lottery numbers based on date and lottery type
func scrapeLotteryNumbers(date, lotteryType string) (*LotteryResponse, error) {
	// Parse the date to ensure it's in the correct format
	parsedDate, err := time.Parse("01/02/2006", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format. Expected format: MM/DD/YYYY, got: %s", date)
	}

	// Format date for URL construction
	formattedDate := parsedDate.Format("01/02/2006")

	// Construct the appropriate URL based on lottery type
	var url string
	switch strings.ToLower(lotteryType) {
	case "megamillions":
		url = fmt.Sprintf("https://www.megamillions.com/Winning-Numbers/Previous-Drawings.aspx?pageNumber=1&pageSize=20&startDate=%s&endDate=%s", formattedDate, formattedDate)
	case "powerball":
		url = fmt.Sprintf("https://www.powerball.com/draw-result?gc=powerball&date=%s", formattedDate)
	default:
		return nil, fmt.Errorf("unsupported lottery type: %s. Supported types: megamillions, powerball", lotteryType)
	}

	// Make HTTP request to the lottery website
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lottery data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Extract winning numbers based on lottery type
	var response *LotteryResponse
	switch strings.ToLower(lotteryType) {
	case "megamillions":
		response, err = extractMegaMillionsNumbers(doc, formattedDate)
	case "powerball":
		response, err = extractPowerballNumbers(doc, formattedDate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract lottery numbers: %v", err)
	}

	response.Date = formattedDate
	response.LotteryType = lotteryType
	return response, nil
}

// extractMegaMillionsNumbers extracts winning numbers from Mega Millions HTML
func extractMegaMillionsNumbers(doc *goquery.Document, date string) (*LotteryResponse, error) {
	response := &LotteryResponse{}

	// Look for the drawing item that matches the date
	// The date is displayed in the h5.drawItemDate element
	drawItem := doc.Find("a.prevDrawItem").FilterFunction(func(i int, s *goquery.Selection) bool {
		dateText := s.Find("h5.drawItemDate").Text()
		// Normalize the date format by removing leading zeros
		normalizedDate := strings.TrimLeft(date, "0")
		normalizedDateText := strings.TrimLeft(strings.TrimSpace(dateText), "0")
		return normalizedDateText == normalizedDate
	})

	if drawItem.Length() == 0 {
		return nil, fmt.Errorf("no drawing found for date: %s", date)
	}

	// Extract the five main numbers (pastNum1 through pastNum5)
	var numbers []string
	for i := 1; i <= 5; i++ {
		selector := fmt.Sprintf("li.ball.pastNum%d", i)
		number := drawItem.Find(selector).Text()
		numbers = append(numbers, strings.TrimSpace(number))
	}

	// Extract the Mega Ball (yellow ball with pastNumMB class)
	megaBall := drawItem.Find("li.ball.yellowBall.pastNumMB").Text()
	megaBall = strings.TrimSpace(megaBall)

	// Extract the Megaplier
	megaplier := drawItem.Find("span.megaplier.pastNumMP").Text()
	megaplier = strings.TrimSpace(megaplier)

	response.Numbers = numbers
	response.MegaBall = megaBall
	response.Megaplier = megaplier

	return response, nil
}

// extractPowerballNumbers extracts winning numbers from Powerball HTML
func extractPowerballNumbers(doc *goquery.Document, date string) (*LotteryResponse, error) {
	response := &LotteryResponse{}

	// Look for the white balls (first 5 numbers)
	var numbers []string
	doc.Find("div.white-balls.item-powerball").Each(func(i int, s *goquery.Selection) {
		if i < 5 { // Only take the first 5 white balls
			numbers = append(numbers, strings.TrimSpace(s.Text()))
		}
	})

	// Extract the Power Ball (red ball)
	powerBall := doc.Find("div.powerball.item-powerball").Text()
	powerBall = strings.TrimSpace(powerBall)

	// Extract the Power Play multiplier
	powerPlay := doc.Find("span.multiplier").Text()
	powerPlay = strings.TrimSpace(powerPlay)

	response.Numbers = numbers
	response.PowerBall = powerBall
	response.PowerPlay = powerPlay

	return response, nil
}

// validateLotteryRequest validates the incoming lottery request
func validateLotteryRequest(req *LotteryRequest) error {
	if req.Date == "" {
		return fmt.Errorf("date is required")
	}

	if req.LotteryType == "" {
		return fmt.Errorf("lottery_type is required")
	}

	// Validate lottery type
	validTypes := map[string]bool{
		"megamillions": true,
		"powerball":    true,
	}

	if !validTypes[strings.ToLower(req.LotteryType)] {
		return fmt.Errorf("invalid lottery_type: %s. Supported types: megamillions, powerball", req.LotteryType)
	}

	// Validate date format
	_, err := time.Parse("01/02/2006", req.Date)
	if err != nil {
		return fmt.Errorf("invalid date format. Expected format: MM/DD/YYYY, got: %s", req.Date)
	}

	return nil
}
