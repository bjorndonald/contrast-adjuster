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

// Request payload structure for lottery winning numbers
type LotteryRequest struct {
	Date        string `json:"date" binding:"required"`         // Date in MM/DD/YYYY format
	LotteryType string `json:"lottery_type" binding:"required"` // Type of lottery (e.g., "megamillions")
}

// Response payload structure for lottery winning numbers
type LotteryResponse struct {
	Success        bool            `json:"success"`
	Date           string          `json:"date"`
	LotteryType    string          `json:"lottery_type"`
	WinningNumbers *WinningNumbers `json:"winning_numbers,omitempty"`
	Error          string          `json:"error,omitempty"`
}

// Structure to hold winning numbers data
type WinningNumbers struct {
	PlayDate    string `json:"play_date"`
	N1          int    `json:"n1"`        // First white ball number
	N2          int    `json:"n2"`        // Second white ball number
	N3          int    `json:"n3"`        // Third white ball number
	N4          int    `json:"n4"`        // Fourth white ball number
	N5          int    `json:"n5"`        // Fifth white ball number
	MBall       int    `json:"m_ball"`    // Mega Ball number
	Megaplier   int    `json:"megaplier"` // Megaplier value
	UpdatedBy   string `json:"updated_by"`
	UpdatedTime string `json:"updated_time"`
}

// Structure for the first API response (GetDrawingPagingData)
type DrawingPagingResponse struct {
	D string `json:"d"` // Contains the actual data as a JSON string
}

// Structure for the drawing data within the first API response
type DrawingData struct {
	DrawingData  []DrawingItem `json:"DrawingData"`
	TotalResults int           `json:"TotalResults"`
}

// Structure for individual drawing items
type DrawingItem struct {
	PlayDate      string `json:"PlayDate"`
	N1            int    `json:"N1"`
	N2            int    `json:"N2"`
	N3            int    `json:"N3"`
	N4            int    `json:"N4"`
	N5            int    `json:"N5"`
	MBall         int    `json:"MBall"`
	Megaplier     int    `json:"Megaplier"`
	UpdatedBy     string `json:"UpdatedBy"`
	UpdatedTime   string `json:"UpdatedTime"`
	PlayDateTicks int64  `json:"PlayDateTicks"`
}

// Structure for the second API response (GetDrawDataByTickWithMatrix)
type DrawDataResponse struct {
	D string `json:"d"` // Contains the actual data as a JSON string
}

// Structure for the detailed draw data
type DetailedDrawData struct {
	Drawing    DrawingItem `json:"Drawing"`
	Jackpot    interface{} `json:"Jackpot"`    // Using interface{} for flexibility
	PrizeTiers interface{} `json:"PrizeTiers"` // Using interface{} for flexibility
}

// Request payload structure for prize amounts
type PrizeRequest struct {
	Date        string `json:"date" binding:"required"`         // Date in MM/DD/YYYY format
	LotteryType string `json:"lottery_type" binding:"required"` // Type of lottery (e.g., "megamillions", "powerball")
}

// Response payload structure for prize amounts
type PrizeResponse struct {
	Success     bool       `json:"success"`
	Date        string     `json:"date"`
	LotteryType string     `json:"lottery_type"`
	PrizeInfo   *PrizeInfo `json:"prize_info,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// Structure to hold prize information
type PrizeInfo struct {
	PlayDate         string      `json:"play_date"`
	EstimatedJackpot string      `json:"estimated_jackpot,omitempty"`
	CashValue        string      `json:"cash_value,omitempty"`
	PrizeTiers       []PrizeTier `json:"prize_tiers"`
	UpdatedBy        string      `json:"updated_by"`
	UpdatedTime      string      `json:"updated_time"`
}

// Structure for individual prize tiers
type PrizeTier struct {
	Match               string         `json:"match"`                          // Match description (e.g., "5+1", "5+0", "4+1")
	PowerballWinners    int            `json:"powerball_winners,omitempty"`    // Number of winners without Power Play
	PowerballPrize      string         `json:"powerball_prize,omitempty"`      // Prize amount without Power Play
	PowerPlayWinners    int            `json:"power_play_winners,omitempty"`   // Number of winners with Power Play
	PowerPlayPrize      string         `json:"power_play_prize,omitempty"`     // Prize amount with Power Play
	MegaMillionsWinners int            `json:"megamillions_winners,omitempty"` // Number of Mega Millions winners
	MegaMillionsPrize   string         `json:"megamillions_prize,omitempty"`   // Mega Millions prize amount
	MegaplierWinners    int            `json:"megaplier_winners,omitempty"`    // Number of winners with Megaplier
	MegaplierPrize      map[string]int `json:"megaplier_prize,omitempty"`      // Prize amounts with all multipliers (2x, 3x, 4x, 5x, 10x)
}
