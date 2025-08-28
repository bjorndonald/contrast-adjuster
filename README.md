# Contrast Adjuster with Lottery Number Scraper

This Go application provides two main functionalities:
1. **Image Contrast Adjustment** - Adjusts the contrast of images
2. **Lottery Number Scraping** - Scrapes winning lottery numbers from lottery websites

## Features

### Image Contrast Adjustment
- POST `/adjust-contrast` - Adjusts image contrast based on provided factor

### Lottery Number Scraping
- POST `/lottery-numbers` - Scrapes winning lottery numbers for a specific date

**Supported Lottery Types:**
- **Mega Millions** - Scrapes from [megamillions.com](https://www.megamillions.com)
- **Powerball** - Scrapes from [powerball.com](https://www.powerball.com)

## Lottery Endpoint Usage

### Request Format
```json
{
  "date": "08/19/2025",
  "lottery": "powerball"
}
```

**Parameters:**
- `date`: Date in MM/DD/YYYY format (e.g., "08/19/2025")
- `lottery`: Type of lottery - supports "megamillions" or "powerball"

### Response Format

#### Mega Millions Response
```json
{
  "date": "08/19/2025",
  "lottery": "megamillions",
  "winning_numbers": [12, 24, 35, 41, 58],
  "mega_ball": 15,
  "megaplier": 3,
  "success": true,
  "message": "Successfully retrieved winning numbers"
}
```

#### Powerball Response
```json
{
  "date": "08/19/2025",
  "lottery": "powerball",
  "winning_numbers": [9, 12, 22, 41, 61],
  "mega_ball": 25,
  "megaplier": 4,
  "success": true,
  "message": "Successfully retrieved winning numbers"
}
```

**Response Fields:**
- `date`: The requested date
- `lottery`: The lottery type
- `winning_numbers`: Array of 5 main winning numbers
- `mega_ball`: Mega Ball number (Mega Millions) or Powerball number (Powerball)
- `megaplier`: Megaplier value (Mega Millions) or Power Play multiplier (Powerball)
- `success`: Boolean indicating if scraping was successful
- `message`: Success or error message

### Example Usage

#### Mega Millions
```bash
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "08/19/2025",
    "lottery": "megamillions"
  }'
```

#### Powerball
```bash
curl -X POST http://localhost:8080/lottery-numbers \
  -H "Content-Type: application/json" \
  -d '{
    "date": "08/19/2025",
    "lottery": "powerball"
  }'
```

## Dependencies

The application requires the following Go packages:
- `github.com/gin-gonic/gin` - Web framework
- `golang.org/x/image` - Image processing
- `golang.org/x/net/html` - HTML parsing for web scraping

## Running the Application

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Run the server:
   ```bash
   go run .
   ```

3. The server will start on port 8080

## Notes

- **Mega Millions**: Scrapes from the Previous Drawings page using MM/DD/YYYY date format
- **Powerball**: Scrapes from the draw-result page using YYYY-MM-DD date format (automatically converted)
- Date format must be exactly MM/DD/YYYY (e.g., "08/19/2025") for both lottery types
- The scraper attempts to extract winning numbers, special ball numbers, and multipliers from the HTML content
- Results may vary depending on the website structure and availability of data
- Powerball scraping looks for Power Play multipliers (e.g., "4x") in the results
