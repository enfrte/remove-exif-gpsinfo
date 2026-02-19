package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// removeGPSHandler handles multipart uploads and returns a JPEG with
// GPS EXIF tags removed (or the original image if no GPS data was found).
//
// It expects a POST request with `Content-Type: multipart/form-data` and
// a single file field named `image`. On success the response body contains
// the JPEG bytes and the header `Content-Type: image/jpeg`.
func removeGPSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	// limit the size to something reasonable (e.g. 20MB)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// read image bytes so we can reuse the data for both processing and
	// returning the original when no GPS removal is needed.
	imgBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
		return
	}

	result := RemoveGPSFromJPEG(imgBytes)
	if result.Error != nil {
		msg := result.Error.Error()
		if msg == "failed to parse JPEG: EOF" || msg == "failed to parse JPEG: unexpected EOF" {
			http.Error(w, msg, http.StatusBadRequest)
		} else {
			http.Error(w, msg, http.StatusInternalServerError)
		}
		return
	}

	// if GPS wasn't removed we don't need to echo the image back; the
	// caller still has the original.  Use a 204/no-content response and
	// include headers so the client can distinguish all the cases.  A
	// separate `X-GPS-Removed` header is useful even when we send bytes so
	// callers don't have to diff the payload.
	if !result.GPSRemoved {
		w.Header().Set("X-Has-EXIF", fmt.Sprintf("%t", result.HasEXIF))
		w.Header().Set("X-Has-GPS", fmt.Sprintf("%t", result.HasGPS))
		w.Header().Set("X-GPS-Removed", "false")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-GPS-Removed", "true")
	w.Write(result.ProcessedData)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/remove-gps", removeGPSHandler)

	log.Printf("starting server on :%s\n", port)
	if err := http.ListenAndServe(
		fmt.Sprintf(":%s", port), nil,
	); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
