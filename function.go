package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"strings"

	"golang.org/x/image/draw"
)

// changeContrast processes the image
func changeContrast(img image.Image, contrast float64) (*image.RGBA, error) {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	contrastFactor := float64(contrast)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := newImg.At(x, y).RGBA()

			rNorm := (float64(r>>8)/255.0-0.5)*contrastFactor + 0.5
			gNorm := (float64(g>>8)/255.0-0.5)*contrastFactor + 0.5
			bNorm := (float64(b>>8)/255.0-0.5)*contrastFactor + 0.5

			rFinal := uint8(math.Max(0, math.Min(1.0, rNorm)) * 255.0)
			gFinal := uint8(math.Max(0, math.Min(1.0, gNorm)) * 255.0)
			bFinal := uint8(math.Max(0, math.Min(1.0, bNorm)) * 255.0)

			newImg.Set(x, y, color.RGBA{rFinal, gFinal, bFinal, uint8(a >> 8)})
		}
	}
	return newImg, nil
}

// processImage takes a base64 string and contrast factor, returns a new base64 string
func processImage(base64Str string, contrast float64) (string, error) {
	// Determine image type (e.g., "image/jpeg") from base64 string prefix
	parts := strings.SplitN(base64Str, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 image format")
	}
	mimeType, base64Data := parts[0], parts[1]

	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(bytes.NewReader(decodedData))
	if err != nil {
		return "", err
	}

	processedImg, err := changeContrast(img, contrast)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if strings.Contains(mimeType, "jpeg") {
		err = jpeg.Encode(&buf, processedImg, nil)
	} else if strings.Contains(mimeType, "png") {
		err = png.Encode(&buf, processedImg)
	} else {
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	if err != nil {
		return "", err
	}

	return mimeType + "," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
