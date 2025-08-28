package main

// Request payload structure
type ContrastRequest struct {
	ImageData      string  `json:"image_data" binding:"required"`
	ContrastFactor float64 `json:"contrast_factor" binding:"required"`
}

// Response payload structure
type ContrastResponse struct {
	ProcessedImage string `json:"processed_image"`
}

// LotteryRequest represents the request payload for lottery number scraping
type LotteryRequest struct {
	Date     string `json:"date" binding:"required"`      // Date in format '08/19/2025'
	Lottery  string `json:"lottery" binding:"required"`   // Type of lottery (e.g., "megamillions")
}

// LotteryResponse represents the response payload for lottery number scraping
type LotteryResponse struct {
	Date         string   `json:"date"`                    // The date requested
	Lottery      string   `json:"lottery"`                 // The lottery type
	WinningNumbers []int  `json:"winning_numbers"`         // Array of winning numbers
	MegaBall     *int     `json:"mega_ball,omitempty"`     // Mega ball number (if applicable)
	Megaplier   *int     `json:"megaplier,omitempty"`     // Megaplier value (if applicable)
	Success     bool     `json:"success"`                  // Whether the scraping was successful
	Message     string   `json:"message"`                  // Success or error message
}
