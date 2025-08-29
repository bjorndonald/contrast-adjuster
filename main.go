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

	// New Mega Millions prize calculation demonstration route
	router.GET("/megamillions-demo", megaMillionsDemoHandler)

	// New Powerball ticket checking route
	router.POST("/check-powerball-ticket", checkPowerballTicketHandler)

	// New Mega Millions ticket checking route
	router.POST("/check-megamillions-ticket", checkMegaMillionsTicketHandler)

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
	// Create a response with Powerball prize information based on the official prize structure
	demoResponse := gin.H{
		"message": "Powerball Prize Calculation System",
		"prize_tiers": []gin.H{
			{"match": "5 white balls + Powerball", "prize": "Jackpot (varies)", "power_play_example": "All multipliers = Jackpot"},
			{"match": "5 white balls (no Powerball)", "prize": "$1,000,000", "power_play_example": "2x: $2,000,000, 3x: $3,000,000, 4x: $4,000,000, 5x: $5,000,000, 10x: $10,000,000"},
			{"match": "4 white balls + Powerball", "prize": "$50,000", "power_play_example": "2x: $100,000, 3x: $150,000, 4x: $200,000, 5x: $250,000, 10x: $500,000"},
			{"match": "4 white balls (no Powerball)", "prize": "$100", "power_play_example": "2x: $200, 3x: $300, 4x: $400, 5x: $500, 10x: $1,000"},
			{"match": "3 white balls + Powerball", "prize": "$100", "power_play_example": "2x: $200, 3x: $300, 4x: $400, 5x: $500, 10x: $1,000"},
			{"match": "3 white balls (no Powerball)", "prize": "$7", "power_play_example": "2x: $14, 3x: $21, 4x: $28, 5x: $35, 10x: $70"},
			{"match": "2 white balls + Powerball", "prize": "$7", "power_play_example": "2x: $14, 3x: $21, 4x: $28, 5x: $35, 10x: $70"},
			{"match": "1 white ball + Powerball", "prize": "$4", "power_play_example": "2x: $8, 3x: $12, 4x: $16, 5x: $20, 10x: $40"},
			{"match": "Powerball only", "prize": "$4", "power_play_example": "2x: $8, 3x: $12, 4x: $16, 5x: $20, 10x: $40"},
		},
		"power_play_multipliers": []int{2, 3, 4, 5, 10},
		"power_play_notes": []string{
			"All Power Play multipliers apply to all prizes",
			"Power Play costs an additional $1 per ticket",
			"Jackpot prizes remain Jackpot regardless of multiplier",
		},
		"usage": "Use the /check-powerball-ticket endpoint to check specific tickets",
	}

	c.JSON(http.StatusOK, demoResponse)
}

// MegaMillionsDemoHandler demonstrates the Mega Millions prize calculation system
// This handler shows examples of different ticket combinations and their prizes
func megaMillionsDemoHandler(c *gin.Context) {
	// Create a response with Mega Millions prize information based on the official prize structure
	demoResponse := gin.H{
		"message": "Mega Millions Prize Calculation System",
		"prize_tiers": []gin.H{
			{"match": "5 white balls + Mega Ball", "prize": "Jackpot (varies)", "megaplier_example": "All multipliers = Jackpot"},
			{"match": "5 white balls (no Mega Ball)", "prize": "$1,000,000", "megaplier_example": "2x = $2,000,000, 3x = $3,000,000, 4x = $4,000,000, 5x = $5,000,000, 10x = $10,000,000"},
			{"match": "4 white balls + Mega Ball", "prize": "$10,000", "megaplier_example": "2x = $20,000, 3x = $30,000, 4x = $40,000, 5x = $50,000, 10x = $100,000"},
			{"match": "4 white balls (no Mega Ball)", "prize": "$500", "megaplier_example": "2x = $1,000, 3x = $1,500, 4x = $2,000, 5x = $2,500, 10x = $5,000"},
			{"match": "3 white balls + Mega Ball", "prize": "$200", "megaplier_example": "2x = $400, 3x = $600, 4x = $800, 5x = $1,000, 10x = $2,000"},
			{"match": "3 white balls (no Mega Ball)", "prize": "$10", "megaplier_example": "2x = $20, 3x = $30, 4x = $40, 5x = $50, 10x = $100"},
			{"match": "2 white balls + Mega Ball", "prize": "$10", "megaplier_example": "2x = $20, 3x = $30, 4x = $40, 5x = $50, 10x = $100"},
			{"match": "1 white ball + Mega Ball", "prize": "$4", "megaplier_example": "2x = $8, 3x = $12, 4x = $16, 5x = $20, 10x = $40"},
			{"match": "Mega Ball only", "prize": "$2", "megaplier_example": "2x = $4, 3x = $6, 4x = $8, 5x = $10, 10x = $20"},
		},
		"megaplier_multipliers": []int{2, 3, 4, 5, 10},
		"megaplier_notes": []string{
			"All Megaplier multipliers apply to all prizes (unlike Power Play)",
			"Megaplier costs an additional $1 per ticket",
			"Jackpot prizes remain Jackpot regardless of multiplier",
		},
		"usage": "Use the /check-megamillions-ticket endpoint to check specific tickets",
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
	// For now, use a default estimated jackpot since we don't have it from the winning numbers
	estimatedJackpot := "$500 Million" // This would come from the lottery data in a real implementation
	ticketResult, err := checkPowerballTicket(
		req.WhiteBallNumbers,
		req.PowerballNumber,
		winningNumbersResponse.WinningNumbers,
		req.PowerPlayMultiplier,
		estimatedJackpot,
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

// MegaMillionsTicketRequest represents a request to check a Mega Millions ticket
// This struct contains the ticket numbers and Megaplier multiplier
type MegaMillionsTicketRequest struct {
	WhiteBallNumbers    []int  `json:"white_ball_numbers" binding:"required,len=5"`
	MegaBallNumber      int    `json:"mega_ball_number" binding:"required,min=1,max=25"`
	MegaplierMultiplier int    `json:"megaplier_multiplier"`                    // 0 = no Megaplier, 2,3,4,5,10 = multiplier
	WinningNumbersDate  string `json:"winning_numbers_date" binding:"required"` // MM/DD/YYYY format
}

// checkMegaMillionsTicketHandler handles requests to check Mega Millions tickets
// This handler validates tickets and calculates prizes based on winning numbers
func checkMegaMillionsTicketHandler(c *gin.Context) {
	var req MegaMillionsTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Validate white ball numbers (must be 1-70)
	for _, num := range req.WhiteBallNumbers {
		if num < 1 || num > 70 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("White ball numbers must be between 1 and 70, got: %d", num),
			})
			return
		}
	}

	// Validate Megaplier multiplier
	validMultipliers := map[int]bool{0: true, 2: true, 3: true, 4: true, 5: true, 10: true}
	if !validMultipliers[req.MegaplierMultiplier] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Megaplier multiplier must be 0, 2, 3, 4, 5, or 10",
		})
		return
	}

	// Get winning numbers for the specified date
	winningNumbersResponse, err := getLotteryWinningNumbers(req.WinningNumbersDate, "megamillions")
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
	ticketResult, err := checkMegaMillionsTicket(
		req.WhiteBallNumbers,
		req.MegaBallNumber,
		winningNumbersResponse.WinningNumbers,
		req.MegaplierMultiplier,
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
			"white_ball_numbers":   req.WhiteBallNumbers,
			"mega_ball_number":     req.MegaBallNumber,
			"megaplier_multiplier": req.MegaplierMultiplier,
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
			"mega_ball": winningNumbersResponse.WinningNumbers.MBall,
		},
		"result": gin.H{
			"is_winner":          ticketResult.IsWinner,
			"white_ball_matches": ticketResult.WhiteBallMatches,
			"has_mega_ball":      ticketResult.HasPowerball, // Reusing Powerball field for Mega Ball
			"prize_description":  ticketResult.PrizeDescription,
			"base_prize":         fmt.Sprintf("$%.2f", float64(ticketResult.BasePrize)/100),
			"megaplier_prize":    ticketResult.PowerPlayPrize, // Reusing Power Play field for Megaplier
			"total_prize":        fmt.Sprintf("$%.2f", float64(ticketResult.TotalPrize)/100),
		},
	}

	c.JSON(http.StatusOK, response)
}
