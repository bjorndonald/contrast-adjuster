package main

// Request payload structure for contrast adjustment
type ContrastRequest struct {
	ImageData      string  `json:"image_data" binding:"required"`
	ContrastFactor float64 `json:"contrast_factor" binding:"required"`
}

// Response payload structure for contrast adjustment
type ContrastResponse struct {
	ProcessedImage string `json:"processed_image"`
}

// Request payload structure for lottery scraping
type LotteryRequest struct {
	Date        string `json:"date" binding:"required"`        // Date in format '08/19/2025'
	LotteryType string `json:"lottery_type" binding:"required"` // 'megamillions' or 'powerball'
}

// Response payload structure for lottery scraping
type LotteryResponse struct {
	Date        string   `json:"date"`
	LotteryType string   `json:"lottery_type"`
	Numbers     []string `json:"numbers"`
	PowerBall   string   `json:"power_ball,omitempty"`   // For Powerball
	MegaBall    string   `json:"mega_ball,omitempty"`    // For Mega Millions
	PowerPlay   string   `json:"power_play,omitempty"`   // For Powerball
	Megaplier   string   `json:"megaplier,omitempty"`    // For Mega Millions
	Error       string   `json:"error,omitempty"`
}
