package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// helper builds and executes a POST /remove-gps with the provided image bytes.
func doPostImage(t *testing.T, imageBytes []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("image", "file.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		t.Fatalf("failed to write image bytes: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/remove-gps", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	removeGPSHandler(rr, req)
	return rr
}

func TestHandler_GpsImage(t *testing.T) {
	data, err := os.ReadFile("gps-img.jpg")
	if err != nil {
		t.Fatalf("couldn't read sample gps image: %v", err)
	}

	rr := doPostImage(t, data)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("unexpected content type: %s", ct)
	}

	// returned bytes should differ because GPS is removed
	if bytes.Equal(rr.Body.Bytes(), data) {
		t.Errorf("response body identical to input; GPS may not have been removed")
	}
	if rr.Header().Get("X-GPS-Removed") != "true" {
		t.Errorf("expected X-GPS-Removed:true, got %s", rr.Header().Get("X-GPS-Removed"))
	}
}

func TestHandler_NoGpsImage(t *testing.T) {
	data, err := os.ReadFile("new.jpg")
	if err != nil {
		t.Fatalf("couldn't read sample non-gps image: %v", err)
	}

	rr := doPostImage(t, data)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for no-change image, got %d", rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "" {
		t.Errorf("no content type expected for 204 but got %s", ct)
	}

	if rr.Header().Get("X-GPS-Removed") != "false" {
		t.Errorf("expected X-GPS-Removed:false but got %s", rr.Header().Get("X-GPS-Removed"))
	}
	// we can't predict whether the uploaded image contains other EXIF
	// tags, but it should never advertise GPS data when we're in the
	// "no change" case.
	if h := rr.Header().Get("X-Has-EXIF"); h != "true" && h != "false" {
		t.Errorf("X-Has-EXIF header missing or invalid: %s", h)
	}
	if rr.Header().Get("X-Has-GPS") != "false" {
		t.Errorf("expected X-Has-GPS:false but got %s", rr.Header().Get("X-Has-GPS"))
	}
}

func TestHandler_MissingImage(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/remove-gps", nil)
	// no content-type set
	removeGPSHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing image, got %d", rr.Code)
	}
}

func TestHandler_InvalidData(t *testing.T) {
	rr := doPostImage(t, []byte("not a jpeg"))
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 400 or 500 for invalid data, got %d", rr.Code)
	}
}
