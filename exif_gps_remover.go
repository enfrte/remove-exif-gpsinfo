package main

import (
	"bytes"
	"fmt"
	"io"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
)

// RemoveGPSResult contains the result of GPS removal operation
type RemoveGPSResult struct {
	HasEXIF       bool   // Whether the image had EXIF data
	HasGPS        bool   // Whether GPS data was found
	GPSRemoved    bool   // Whether GPS data was removed
	ProcessedData []byte // The processed image data (nil if no processing needed)
	Error         error  // Any error that occurred
}

// RemoveGPSFromJPEG takes JPEG image data and removes GPS EXIF tags if present.
func RemoveGPSFromJPEG(imageData []byte) RemoveGPSResult {
	result := RemoveGPSResult{
		HasEXIF:    false,
		HasGPS:     false,
		GPSRemoved: false,
	}

	// Parse the JPEG
	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(imageData)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse JPEG: %w", err)
		return result
	}

	sl := intfc.(*jpegstructure.SegmentList)

	// Try to construct EXIF builder
	rootIb, err := sl.ConstructExifBuilder() // See tagIndex for types of tags if detected
	if err != nil {
		// No EXIF data found
		result.HasEXIF = false
		return result
	}

	result.HasEXIF = true

	// Try to get GPS IFD using GetOrCreateIbFromRootIb
	gpsIfdPath := "IFD/GPSInfo"
	gpsIb, err := exif.GetOrCreateIbFromRootIb(rootIb, gpsIfdPath)
	if err != nil {
		// No GPS IFD
		result.HasGPS = false
		return result
	}

	// Check if GPS has any tags
	gpsTags := gpsIb.Tags()
	if len(gpsTags) == 0 {
		result.HasGPS = false
		return result
	}

	result.HasGPS = true

	// Get IFD0 (parent of GPSInfo)
	ifd0Ib, err := exif.GetOrCreateIbFromRootIb(rootIb, "IFD")
	if err != nil {
		result.Error = fmt.Errorf("failed to get IFD0: %w", err)
		return result
	}

	// Remove GPS IFD by using ReplaceChildWithNew with empty builder
	// The GPS IFD tag ID is 0x8825
	gpsTagId := exifcommon.IfdGpsInfoStandardIfdIdentity.TagId()
	err = ifd0Ib.DeleteFirst(gpsTagId)
	if err != nil {
		result.Error = fmt.Errorf("failed to remove GPS IFD: %w", err)
		return result
	}

	// Set the modified EXIF back
	err = sl.SetExif(rootIb)
	if err != nil {
		result.Error = fmt.Errorf("failed to set EXIF: %w", err)
		return result
	}

	// Write to buffer
	var buf bytes.Buffer
	err = sl.Write(&buf)
	if err != nil {
		result.Error = fmt.Errorf("failed to write JPEG: %w", err)
		return result
	}

	result.ProcessedData = buf.Bytes()
	result.GPSRemoved = true

	return result
}

// RemoveGPSFromJPEGReader is a convenience function that works with io.Reader
func RemoveGPSFromJPEGReader(reader io.Reader) RemoveGPSResult {
	imageData, err := io.ReadAll(reader)
	if err != nil {
		return RemoveGPSResult{
			Error: fmt.Errorf("failed to read image data: %w", err),
		}
	}
	return RemoveGPSFromJPEG(imageData)
}
