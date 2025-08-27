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
