package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/adjust-contrast", adjustContrastHandler)
	router.POST("/lottery-numbers", lotteryNumbersHandler)
	router.Run(":8080")
}

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

// lotteryNumbersHandler handles requests for lottery winning numbers
// This function receives a date and lottery type, then scrapes the website for results
func lotteryNumbersHandler(c *gin.Context) {
	var req LotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Scrape lottery numbers from the website
	result, err := scrapeLotteryNumbers(req.Date, req.Lottery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
