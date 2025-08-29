package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Test the Powerball prize calculation system
	fmt.Println("Testing Powerball Prize Calculation System...")
	testPowerballPrizes()

	router := gin.Default()

	// Existing contrast adjustment route
	router.POST("/adjust-contrast", adjustContrastHandler)

	// New lottery winning numbers route
	router.POST("/lottery-winning-numbers", lotteryWinningNumbersHandler)

	// New lottery prize amounts route
	router.POST("/lottery-prize-amounts", lotteryPrizeAmountsHandler)

	// New Powerball prize calculation demonstration route
	router.GET("/powerball-demo", powerballDemoHandler)

	// New Powerball ticket checking route
	router.POST("/check-powerball-ticket", checkPowerballTicketHandler)

	// Add a simple health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "message": "Lottery API is running"})
	})

	router.Run(":8080")
}

// adjustContrastHandler handles contrast adjustment requests
func adjustContrastHandler(c *gin.Context) {
	var req ContrastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processedImage, err := processImage(req.ImageData, req.ContrastFactor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ContrastResponse{ProcessedImage: processedImage})
}

// lotteryWinningNumbersHandler handles requests for lottery winning numbers
// This handler processes POST requests to get winning numbers for a specific date and lottery type
func lotteryWinningNumbersHandler(c *gin.Context) {
	var req LotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Get winning numbers for the specified date and lottery type
	response, err := getLotteryWinningNumbers(req.Date, req.LotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Internal server error: %v", err),
		})
		return
	}

	// Return the response with appropriate HTTP status
	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// lotteryPrizeAmountsHandler handles requests for lottery prize amounts
// This handler processes POST requests to get prize information for a specific date and lottery type
func lotteryPrizeAmountsHandler(c *gin.Context) {
	var req PrizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Get prize amounts for the specified date and lottery type
	response, err := getLotteryPrizeAmounts(req.Date, req.LotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Internal server error: %v", err),
		})
		return
	}

	// Return the response with appropriate HTTP status
	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// powerballDemoHandler demonstrates the Powerball prize calculation system
// This handler shows examples of different ticket combinations and their prizes
func powerballDemoHandler(c *gin.Context) {
	// Create a response with Powerball prize information
	demoResponse := gin.H{
		"message": "Powerball Prize Calculation System",
		"prize_tiers": []gin.H{
			{"match": "5 white balls + Powerball", "prize": "Jackpot (varies)", "power_play_example": "2x = $2,000,000"},
			{"match": "5 white balls (no Powerball)", "prize": "$1,000,000", "power_play_example": "2x = $2,000,000"},
			{"match": "4 white balls + Powerball", "prize": "$50,000", "power_play_example": "2x = $100,000"},
			{"match": "4 white balls (no Powerball)", "prize": "$100", "power_play_example": "2x = $200"},
			{"match": "3 white balls + Powerball", "prize": "$100", "power_play_example": "2x = $200"},
			{"match": "3 white balls (no Powerball)", "prize": "$7", "power_play_example": "2x = $14"},
			{"match": "2 white balls + Powerball", "prize": "$7", "power_play_example": "2x = $14"},
			{"match": "1 white ball + Powerball", "prize": "$4", "power_play_example": "2x = $8"},
			{"match": "Powerball only", "prize": "$4", "power_play_example": "2x = $8"},
		},
		"power_play_multipliers": []int{2, 3, 4, 5, 10},
		"power_play_notes": []string{
			"10x multiplier only applies to prizes $150,000 or less",
			"10x multiplier defaults to 2x for larger prizes",
			"Power Play costs an additional $1 per ticket",
		},
		"usage": "Use the /check-powerball-ticket endpoint to check specific tickets",
	}

	c.JSON(http.StatusOK, demoResponse)
}

// PowerballTicketRequest represents a request to check a Powerball ticket
// This struct contains the ticket numbers and Power Play multiplier
type PowerballTicketRequest struct {
	WhiteBallNumbers    []int  `json:"white_ball_numbers" binding:"required,len=5"`
	PowerballNumber     int    `json:"powerball_number" binding:"required,min=1,max=26"`
	PowerPlayMultiplier int    `json:"power_play_multiplier"`                   // 0 = no Power Play, 2,3,4,5,10 = multiplier
	WinningNumbersDate  string `json:"winning_numbers_date" binding:"required"` // MM/DD/YYYY format
}

// checkPowerballTicketHandler handles requests to check Powerball tickets
// This handler validates tickets and calculates prizes based on winning numbers
func checkPowerballTicketHandler(c *gin.Context) {
	var req PowerballTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Validate white ball numbers (must be 1-69)
	for _, num := range req.WhiteBallNumbers {
		if num < 1 || num > 69 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("White ball numbers must be between 1 and 69, got: %d", num),
			})
			return
		}
	}

	// Validate Power Play multiplier
	validMultipliers := map[int]bool{0: true, 2: true, 3: true, 4: true, 5: true, 10: true}
	if !validMultipliers[req.PowerPlayMultiplier] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Power Play multiplier must be 0, 2, 3, 4, 5, or 10",
		})
		return
	}

	// Get winning numbers for the specified date
	winningNumbersResponse, err := getLotteryWinningNumbers(req.WinningNumbersDate, "powerball")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to get winning numbers: %v", err),
		})
		return
	}

	if !winningNumbersResponse.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   winningNumbersResponse.Error,
		})
		return
	}

	// Check the ticket
	ticketResult, err := checkPowerballTicket(
		req.WhiteBallNumbers,
		req.PowerballNumber,
		winningNumbersResponse.WinningNumbers,
		req.PowerPlayMultiplier,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to check ticket: %v", err),
		})
		return
	}

	// Format the response
	response := gin.H{
		"success": true,
		"ticket": gin.H{
			"white_ball_numbers":    req.WhiteBallNumbers,
			"powerball_number":      req.PowerballNumber,
			"power_play_multiplier": req.PowerPlayMultiplier,
		},
		"winning_numbers": gin.H{
			"date": req.WinningNumbersDate,
			"white_balls": []int{
				winningNumbersResponse.WinningNumbers.N1,
				winningNumbersResponse.WinningNumbers.N2,
				winningNumbersResponse.WinningNumbers.N3,
				winningNumbersResponse.WinningNumbers.N4,
				winningNumbersResponse.WinningNumbers.N5,
			},
			"powerball": winningNumbersResponse.WinningNumbers.MBall,
		},
		"result": gin.H{
			"is_winner":          ticketResult.IsWinner,
			"white_ball_matches": ticketResult.WhiteBallMatches,
			"has_powerball":      ticketResult.HasPowerball,
			"prize_description":  ticketResult.PrizeDescription,
			"base_prize":         fmt.Sprintf("$%.2f", float64(ticketResult.BasePrize)/100),
			"power_play_prize":   ticketResult.PowerPlayPrize,
			"total_prize":        fmt.Sprintf("$%.2f", float64(ticketResult.TotalPrize)/100),
		},
	}

	c.JSON(http.StatusOK, response)
}
