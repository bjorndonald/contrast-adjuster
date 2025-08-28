package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	
	// Add CORS middleware for cross-origin requests
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Existing contrast adjustment route
	router.POST("/adjust-contrast", adjustContrastHandler)
	
	// New lottery scraping route
	router.POST("/lottery-numbers", lotteryNumbersHandler)
	
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

// lotteryNumbersHandler handles lottery number scraping requests
func lotteryNumbersHandler(c *gin.Context) {
	var req LotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := validateLotteryRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Scrape the lottery numbers
	response, err := scrapeLotteryNumbers(req.Date, req.LotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
