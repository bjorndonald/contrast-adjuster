package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	
	// Existing contrast adjustment route
	router.POST("/adjust-contrast", adjustContrastHandler)
	
	// New lottery winning numbers route
	router.POST("/lottery-winning-numbers", lotteryWinningNumbersHandler)
	
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
