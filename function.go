package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/net/html"
)

// changeContrast processes the image
func changeContrast(img image.Image, contrast float64) (*image.RGBA, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	contrastFactor := float64(contrast)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := newImg.At(x, y).RGBA()

			rNorm := (float64(r>>8)/255.0-0.5)*contrastFactor + 0.5
			gNorm := (float64(g>>8)/255.0-0.5)*contrastFactor + 0.5
			bNorm := (float64(b>>8)/255.0-0.5)*contrastFactor + 0.5

			rFinal := uint8(math.Max(0, math.Min(1.0, rNorm)) * 255.0)
			gFinal := uint8(math.Max(0, math.Min(1.0, gNorm)) * 255.0)
			bFinal := uint8(math.Max(0, math.Min(1.0, bNorm)) * 255.0)

			newImg.Set(x, y, color.RGBA{rFinal, gFinal, bFinal, uint8(a >> 8)})
		}
	}
	return newImg, nil
}

// processImage takes a base64 string and contrast factor, returns a new base64 string
func processImage(base64Str string, contrast float64) (string, error) {
	// Determine image type (e.g., "image/jpeg") from base64 string prefix
	parts := strings.SplitN(base64Str, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 image format")
	}
	mimeType, base64Data := parts[0], parts[1]

	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(bytes.NewReader(decodedData))
	if err != nil {
		return "", err
	}

	processedImg, err := changeContrast(img, contrast)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if strings.Contains(mimeType, "jpeg") {
		err = jpeg.Encode(&buf, processedImg, nil)
	} else if strings.Contains(mimeType, "png") {
		err = png.Encode(&buf, processedImg)
	} else {
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	if err != nil {
		return "", err
	}

	return mimeType + "," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// scrapeLotteryNumbers scrapes winning numbers from lottery websites for a given date
// This function takes a date string in format '08/19/2025' and returns lottery results
func scrapeLotteryNumbers(date string, lotteryType string) (*LotteryResponse, error) {
	// Create the response structure
	response := &LotteryResponse{
		Date:           date,
		Lottery:        lotteryType,
		Success:        false,
		WinningNumbers: []int{},
	}

	// Validate date format
	if !isValidDateFormat(date) {
		response.Message = "Invalid date format. Please use MM/DD/YYYY format."
		return response, nil
	}

	// Route to appropriate scraper based on lottery type
	switch strings.ToLower(lotteryType) {
	case "megamillions":
		return scrapeMegaMillions(date, lotteryType)
	case "powerball":
		return scrapePowerball(date, lotteryType)
	default:
		response.Message = "Unsupported lottery type. Currently supports: megamillions, powerball"
		return response, nil
	}
}

// scrapeMegaMillions scrapes winning numbers from the Mega Millions website for a given date
// This function handles the specific structure of the Mega Millions website
func scrapeMegaMillions(date string, lotteryType string) (*LotteryResponse, error) {
	// Create the response structure
	response := &LotteryResponse{
		Date:           date,
		Lottery:        lotteryType,
		Success:        false,
		WinningNumbers: []int{},
	}

	// Construct the URL for the specific date
	url := fmt.Sprintf("https://www.megamillions.com/Winning-Numbers/Previous-Drawings.aspx?pageNumber=1&pageSize=20&startDate=%s&endDate=%s", date, date)

	// Make HTTP request to the website
	resp, err := http.Get(url)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to fetch data: %v", err)
		return response, nil
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		response.Message = fmt.Sprintf("HTTP request failed with status: %d", resp.StatusCode)
		return response, nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to read response body: %v", err)
		return response, nil
	}

	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		response.Message = fmt.Sprintf("Failed to parse HTML: %v", err)
		return response, nil
	}

	// Extract winning numbers from the HTML
	numbers, megaBall, megaplier, err := extractWinningNumbers(doc, date)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to extract numbers: %v", err)
		return response, nil
	}

	// Check if we found any numbers
	if len(numbers) == 0 {
		response.Message = "No winning numbers found for the specified date"
		return response, nil
	}

	// Set the response data
	response.WinningNumbers = numbers
	response.MegaBall = megaBall
	response.Megaplier = megaplier
	response.Success = true
	response.Message = "Successfully retrieved winning numbers"

	return response, nil
}

// scrapePowerball scrapes winning numbers from the Powerball website for a given date
// This function handles the specific structure of the Powerball website
func scrapePowerball(date string, lotteryType string) (*LotteryResponse, error) {
	// Create the response structure
	response := &LotteryResponse{
		Date:           date,
		Lottery:        lotteryType,
		Success:        false,
		WinningNumbers: []int{},
	}

	// Convert date from MM/DD/YYYY to YYYY-MM-DD format for Powerball URL
	powerballDate, err := convertDateForPowerball(date)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to convert date format: %v", err)
		return response, nil
	}

	// Construct the URL for the specific date
	url := fmt.Sprintf("https://www.powerball.com/draw-result?gc=powerball&date=%s", powerballDate)

	// Make HTTP request to the website
	resp, err := http.Get(url)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to fetch data: %v", err)
		return response, nil
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		response.Message = fmt.Sprintf("HTTP request failed with status: %d", resp.StatusCode)
		return response, nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to read response body: %v", err)
		return response, nil
	}

	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		response.Message = fmt.Sprintf("Failed to parse HTML: %v", err)
		return response, nil
	}

	// Extract winning numbers from the Powerball HTML
	numbers, powerball, powerPlay, err := extractPowerballNumbers(doc, date)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to extract numbers: %v", err)
		return response, nil
	}

	// Check if we found any numbers
	if len(numbers) == 0 {
		response.Message = "No winning numbers found for the specified date"
		return response, nil
	}

	// Set the response data
	response.WinningNumbers = numbers
	response.MegaBall = powerball  // Reusing MegaBall field for Powerball number
	response.Megaplier = powerPlay // Reusing Megaplier field for Power Play multiplier
	response.Success = true
	response.Message = "Successfully retrieved winning numbers"

	return response, nil
}

// convertDateForPowerball converts date from MM/DD/YYYY to YYYY-MM-DD format
// This function is needed because Powerball uses a different date format in their URL
func convertDateForPowerball(date string) (string, error) {
	// Parse the input date (MM/DD/YYYY)
	parsedDate, err := time.Parse("01/02/2006", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	// Format to YYYY-MM-DD
	return parsedDate.Format("2006-01-02"), nil
}

// extractPowerballNumbers parses the Powerball HTML document to extract winning numbers
// This function is specifically designed for the Powerball website structure
func extractPowerballNumbers(doc *html.Node, targetDate string) ([]int, *int, *int, error) {
	var numbers []int
	var powerball *int
	var powerPlay *int

	// Function to traverse HTML nodes recursively
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Look for specific elements that contain lottery numbers
			if n.Data == "div" || n.Data == "span" {
				// Check if this element contains the winning numbers section
				if containsPowerballWinningNumbers(n) {
					// Extract numbers from this section
					sectionNumbers, pb, pp := extractPowerballSection(n)
					if len(sectionNumbers) > 0 {
						numbers = sectionNumbers
						powerball = pb
						powerPlay = pp
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return numbers, powerball, powerPlay, nil
}

// containsPowerballWinningNumbers checks if an HTML element contains Powerball winning numbers
// This function looks for specific text patterns that indicate the winning numbers section
func containsPowerballWinningNumbers(n *html.Node) bool {
	var contains bool
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.ToLower(strings.TrimSpace(node.Data))
			// Look for indicators of winning numbers section
			if strings.Contains(text, "winning numbers") || strings.Contains(text, "draw results") {
				contains = true
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return contains
}

// extractPowerballSection extracts winning numbers from a Powerball HTML section
// This function parses the specific structure of Powerball result sections
func extractPowerballSection(section *html.Node) ([]int, *int, *int) {
	var numbers []int
	var powerball *int
	var powerPlay *int

	// Regular expressions to match lottery numbers
	numberRegex := regexp.MustCompile(`\b(\d{1,2})\b`)
	powerPlayRegex := regexp.MustCompile(`(\d+)x`)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)

			// Look for Power Play multiplier (e.g., "4x")
			if powerPlay == nil {
				ppMatches := powerPlayRegex.FindStringSubmatch(text)
				if len(ppMatches) > 1 {
					if pp, err := strconv.Atoi(ppMatches[1]); err == nil {
						powerPlay = &pp
					}
				}
			}

			// Look for regular numbers
			matches := numberRegex.FindAllString(text, -1)
			for _, match := range matches {
				if num, err := strconv.Atoi(match); err == nil {
					// Assume first 5 numbers are regular winning numbers
					if len(numbers) < 5 {
						numbers = append(numbers, num)
					} else if powerball == nil {
						// Next number might be the Powerball
						powerball = &num
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(section)
	return numbers, powerball, powerPlay
}

// extractWinningNumbers parses the HTML document to extract winning numbers, mega ball, and megaplier
// This function navigates through the HTML structure to find the lottery results table
func extractWinningNumbers(doc *html.Node, targetDate string) ([]int, *int, *int, error) {
	var numbers []int
	var megaBall *int
	var megaplier *int

	// Function to traverse HTML nodes recursively
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			// Look for table rows that might contain lottery data
			if containsDate(n, targetDate) {
				// Extract numbers from this row
				rowNumbers, mb, mp := extractNumbersFromRow(n)
				if len(rowNumbers) > 0 {
					numbers = rowNumbers
					megaBall = mb
					megaplier = mp
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return numbers, megaBall, megaplier, nil
}

// containsDate checks if a table row contains the target date
// This function searches for date patterns within the HTML content
func containsDate(n *html.Node, targetDate string) bool {
	var contains bool
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			if strings.Contains(node.Data, targetDate) {
				contains = true
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return contains
}

// extractNumbersFromRow extracts winning numbers, mega ball, and megaplier from a table row
// This function parses the specific structure of lottery result rows
func extractNumbersFromRow(row *html.Node) ([]int, *int, *int) {
	var numbers []int
	var megaBall *int
	var megaplier *int

	// Regular expressions to match lottery numbers
	numberRegex := regexp.MustCompile(`\b(\d{1,2})\b`)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			// Look for numbers in the text content
			matches := numberRegex.FindAllString(n.Data, -1)
			for _, match := range matches {
				if num, err := strconv.Atoi(match); err == nil {
					// Assume first 5 numbers are regular winning numbers
					if len(numbers) < 5 {
						numbers = append(numbers, num)
					} else if megaBall == nil {
						// Next number might be the mega ball
						megaBall = &num
					} else if megaplier == nil {
						// Next number might be the megaplier
						megaplier = &num
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(row)
	return numbers, megaBall, megaplier
}

// isValidDateFormat validates that the date string is in the correct MM/DD/YYYY format
// This function ensures the input date follows the expected pattern
func isValidDateFormat(date string) bool {
	// Check if the date matches MM/DD/YYYY pattern
	pattern := regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`)
	if !pattern.MatchString(date) {
		return false
	}

	// Try to parse the date to ensure it's valid
	_, err := time.Parse("01/02/2006", date)
	return err == nil
}
