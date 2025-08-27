package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/adjust-contrast", adjustContrastHandler)
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
