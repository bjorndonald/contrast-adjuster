# Contrast Adjuster with Lottery Scraping

This Go application provides two main functionalities:
1. **Image Contrast Adjustment** - Adjusts the contrast of images
2. **Lottery Number Scraping** - Scrapes winning lottery numbers from official websites

## Features

### Lottery Scraping
- Supports **Mega Millions** and **Powerball** lottery types
- Scrapes winning numbers based on specific dates
- Extracts main numbers, bonus balls, and multipliers
- Handles both lottery websites with different HTML structures

## API Endpoints

### 1. Contrast Adjustment
- **POST** `/adjust-contrast`
- Adjusts image contrast based on provided factor

### 2. Lottery Numbers
- **POST** `/lottery-numbers`
- Scrapes winning lottery numbers for a specific date and lottery type

## Lottery Scraping Usage

### Request Format
```json
{
  "date": "08/19/2025",
  "lottery_type": "megamillions"
}
```

### Supported Lottery Types
- `megamillions` - Scrapes from Mega Millions website
- `powerball` - Scrapes from Powerball website

### Response Format
```json
{
  "date": "08/19/2025",
  "lottery_type": "megamillions",
  "numbers": ["10", "19", "24", "49", "68"],
  "mega_ball": "10",
  "megaplier": "N/A"
}
```

For Powerball:
```json
{
  "date": "08/19/2025",
  "lottery_type": "powerball",
  "numbers": ["9", "12", "22", "41", "61"],
  "power_ball": "25",
  "power_play": "4x"
}
```

## URL Construction

The application automatically constructs the appropriate URLs:

- **Mega Millions**: `https://www.megamillions.com/Winning-Numbers/Previous-Drawings.aspx?pageNumber=1&pageSize=20&startDate={date}&endDate={date}`
- **Powerball**: `https://www.powerball.com/draw-result?gc=powerball&date={date}`

## Date Format

Dates must be provided in the format: `MM/DD/YYYY` (e.g., `08/19/2025`)

## Installation

1. Ensure Go 1.24.6+ is installed
2. Clone the repository
3. Run `go mod tidy` to download dependencies
4. Run `go run .` to start the server

## Dependencies

- `github.com/gin-gonic/gin` - Web framework
- `github.com/PuerkitoBio/goquery` - HTML parsing
- `golang.org/x/image` - Image processing
- `golang.org/x/net` - Network utilities

## Error Handling

The application provides detailed error messages for:
- Invalid date formats
- Unsupported lottery types
- Network failures
- HTML parsing errors
- Missing lottery data for specified dates

## CORS Support

The application includes CORS middleware to support cross-origin requests from web applications.
