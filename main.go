package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run . <input.jpg> <output.jpg>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Read the input image
	imageData, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Process the image
	result := RemoveGPSFromJPEG(imageData)

	// Handle errors
	if result.Error != nil {
		fmt.Printf("Error processing image: %v\n", result.Error)
		os.Exit(1)
	}

	// Check results and decide what to write
	var dataToWrite []byte
	if result.GPSRemoved {
		fmt.Println("✓ GPS data found and removed")
		dataToWrite = result.ProcessedData
	} else if !result.HasEXIF {
		fmt.Println("ℹ No EXIF data found - saving original image")
		dataToWrite = imageData
	} else if !result.HasGPS {
		fmt.Println("ℹ EXIF data found but no GPS coordinates - saving original image")
		dataToWrite = imageData
	}

	// Write output
	if err := os.WriteFile(outputPath, dataToWrite, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Image saved to: %s\n", outputPath)
}
