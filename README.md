# Contrast Adjuster with Lottery API

This project provides two main functionalities:
1. **Image Contrast Adjustment**: Adjust the contrast of images using base64 encoding
2. **Lottery Winning Numbers API**: Retrieve winning lottery numbers for specific dates

## Features

### Lottery Winning Numbers API
- **Endpoint**: `POST /lottery-winning-numbers`
- **Supported Lotteries**: 
  - **Mega Millions**: Uses official API endpoints
  - **Powerball**: Uses web scraping from official website
- **Data Source**: 
  - Mega Millions: Official API endpoints
  - Powerball: Official Powerball website scraping
- **Date Format**: MM/DD/YYYY (e.g., "08/19/2025")

## API Endpoints

### 1. Lottery Winning Numbers
```
POST /lottery-winning-numbers
```

**Request Body:**
```json
{
  "date": "08/19/2025",
  "lottery_type": "megamillions"
}
```

**Response (Success - Mega Millions):**
```json
{
  "success": true,
  "date": "08/19/2025",
  "lottery_type": "megamillions",
  "winning_numbers": {
    "play_date": "2025-08-19T00:00:00",
    "n1": 10,
    "n2": 19,
    "n3": 24,
    "n4": 49,
    "n5": 68,
    "m_ball": 10,
    "megaplier": -1,
    "updated_by": "SERVICE",
    "updated_time": "2025-08-19T23:07:01"
  }
}
```

**Response (Success - Powerball):**
```json
{
  "success": true,
  "date": "08/27/2025",
  "lottery_type": "powerball",
  "winning_numbers": {
    "play_date": "2025-08-27T00:00:00",
    "n1": 9,
    "n2": 12,
    "n3": 22,
    "n4": 41,
    "n5": 61,
    "m_ball": 25,
    "megaplier": 4,
    "updated_by": "POWERBALL_SCRAPER",
    "updated_time": "2025-08-28T16:35:27"
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "date": "08/19/2025",
  "lottery_type": "lotto",
  "error": "Unsupported lottery type. Currently only 'megamillions' and 'powerball' are supported."
}
```

### 2. Health Check
```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "message": "Lottery API is running"
}
```

### 3. Image Contrast Adjustment (Existing)
```
POST /adjust-contrast
```

## How It Works

### Mega Millions
The Mega Millions API uses a two-step process to retrieve winning numbers:

1. **First API Call**: `GetDrawingPagingData`
   - Endpoint: `https://www.megamillions.com/cmspages/utilservice.asmx/GetDrawingPagingData`
   - Purpose: Get basic drawing information and `PlayDateTicks` for a specific date
   - Returns: Basic drawing data including the unique tick identifier

2. **Second API Call**: `GetDrawDataByTickWithMatrix`
   - Endpoint: `https://www.megamillions.com/cmspages/utilservice.asmx/GetDrawDataByTickWithMatrix`
   - Purpose: Get detailed drawing information using the tick identifier
   - Returns: Comprehensive drawing data including jackpot and prize tier information

### Powerball
The Powerball implementation uses web scraping to extract winning numbers:

1. **URL Construction**: Builds the Powerball draw result URL with the specified date
   - Format: `https://www.powerball.com/draw-result?gc=powerball&date=YYYY-MM-DD&oc=fl`

2. **HTML Scraping**: Parses the HTML response to extract:
   - White ball numbers (5 numbers from 1-69)
   - Powerball number (1 number from 1-26)
   - Power Play multiplier (if available)
   - Drawing date

3. **Data Extraction**: Uses regex patterns to find winning numbers in the HTML structure

## Installation and Setup

### Prerequisites
- Go 1.24.6 or higher
- Git

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd contrast-adjuster

# Install dependencies
go mod tidy

# Build the application
go build -o contrast-adjuster .

# Run the application
./contrast-adjuster
```

The server will start on port 8080.

## Usage Examples

### Using curl

**Get Mega Millions winning numbers for a specific date:**
```bash
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/19/2025", "lottery_type": "megamillions"}'
```

**Get Powerball winning numbers for a specific date:**
```bash
curl -X POST http://localhost:8080/lottery-winning-numbers \
  -H "Content-Type: application/json" \
  -d '{"date": "08/27/2025", "lottery_type": "powerball"}'
```

**Check server health:**
```bash
curl -X GET http://localhost:8080/health
```

### Using the test script
```bash
# Make the test script executable
chmod +x test_lottery.sh

# Run the tests (requires jq for JSON formatting)
./test_lottery.sh
```

## Error Handling

The API provides comprehensive error handling for:
- Invalid date formats
- Unsupported lottery types
- API communication failures (Mega Millions)
- Web scraping failures (Powerball)
- Missing or malformed request data

All errors return appropriate HTTP status codes and descriptive error messages.

## Supported Lottery Types

### Mega Millions
- **Numbers**: 5 white balls (1-70) + 1 Mega Ball (1-25)
- **Data Source**: Official API endpoints
- **Multiplier**: Megaplier (2x, 3x, 4x, 5x, 10x)

### Powerball
- **Numbers**: 5 white balls (1-69) + 1 Powerball (1-26)
- **Data Source**: Web scraping from official website
- **Multiplier**: Power Play (2x, 3x, 4x, 5x, 10x)

## Data Structure

### Winning Numbers
- **n1-n5**: White ball numbers (varies by lottery)
- **m_ball**: 
  - Mega Millions: Mega Ball number
  - Powerball: Powerball number
- **megaplier**: 
  - Mega Millions: Megaplier value
  - Powerball: Power Play multiplier value
- **play_date**: Drawing date
- **updated_by**: Service that updated the data
- **updated_time**: Last update timestamp

## Rate Limiting and Performance

- HTTP client timeout: 30 seconds
- No built-in rate limiting (respects lottery website limits)
- Efficient JSON parsing and HTML regex parsing
- Proper error handling for both API and web scraping methods

## Security Considerations

- Input validation for date formats
- Lottery type validation
- Proper HTTP status codes for different error scenarios
- User-Agent header for API identification
- Secure web scraping with proper headers

## Web Scraping Details

### Powerball Scraping
- **URL Pattern**: `https://www.powerball.com/draw-result?gc=powerball&date=YYYY-MM-DD&oc=fl`
- **HTML Elements**: 
  - White balls: `<div class="form-control col white-balls item-powerball">`
  - Powerball: `<div class="form-control col powerball item-powerball">`
  - Power Play: `<span class="multiplier">Nx</span>`
- **Date Format**: Converts from "Mon, Jan 02, 2006" to ISO format
- **Fallback**: Uses current timestamp if date parsing fails

## Future Enhancements

- Support for additional lottery types
- Historical data retrieval
- Caching for frequently requested dates
- Batch processing for multiple dates
- WebSocket support for real-time updates
- Enhanced error handling for web scraping edge cases

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions, please create an issue in the repository.
