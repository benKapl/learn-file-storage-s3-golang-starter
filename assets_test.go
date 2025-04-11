package main

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func setupTestAPIConfig(t *testing.T) apiConfig {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	assetsRoot := os.Getenv("ASSETS_ROOT")
	if assetsRoot == "" {
		t.Fatal("ASSETS_ROOT environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		t.Fatal("PORT environment variable is not set")
	}

	return apiConfig{
		assetsRoot: assetsRoot,
		port:       port,
	}
}

func TestGetAssetPath(t *testing.T) {
	testCases := []struct {
		videoID   uuid.UUID
		mediaType string
		expected  string
	}{
		{
			videoID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mediaType: "image/jpeg",
			expected:  "550e8400-e29b-41d4-a716-446655440000.jpeg",
		},
		{
			videoID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mediaType: "image/png",
			expected:  "550e8400-e29b-41d4-a716-446655440000.png",
		},
		{
			videoID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mediaType: "video/mp4",
			expected:  "550e8400-e29b-41d4-a716-446655440000.mp4",
		},
		{
			videoID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mediaType: "invalid",
			expected:  "550e8400-e29b-41d4-a716-446655440000.bin",
		},
	}

	for _, tc := range testCases {
		actual := getAssetPath(tc.videoID, tc.mediaType)
		if actual != tc.expected {
			t.Errorf("got %s but expected %s for videoID %s and mediaType %s",
				actual, tc.expected, tc.videoID, tc.mediaType)
		}
	}
}

func TestGetAssetURL(t *testing.T) {
	cfg := setupTestAPIConfig(t)

	testCases := []struct {
		assetPath string
		expected  string
	}{
		{
			assetPath: "test.jpg",
			expected:  "http://localhost:" + cfg.port + "/assets/test.jpg",
		},
		{
			assetPath: "folder/test.mp4",
			expected:  "http://localhost:" + cfg.port + "/assets/folder/test.mp4",
		},
	}

	for _, tc := range testCases {
		actual := cfg.getAssetURL(tc.assetPath)
		if actual != tc.expected {
			t.Errorf("got %s but expected %s for assetPath %s",
				actual, tc.expected, tc.assetPath)
		}
	}
}

func TestMediaTypeToExt(t *testing.T) {
	testCases := []struct {
		mediaType string
		expected  string
	}{
		{
			mediaType: "image/jpeg",
			expected:  ".jpeg",
		},
		{
			mediaType: "image/png",
			expected:  ".png",
		},
		{
			mediaType: "video/mp4",
			expected:  ".mp4",
		},
		{
			mediaType: "invalid",
			expected:  ".bin",
		},
	}

	for _, tc := range testCases {
		actual := mediaTypeToExt(tc.mediaType)
		if actual != tc.expected {
			t.Errorf("got %s but expected %s for mediaType %s",
				actual, tc.expected, tc.mediaType)
		}
	}
}

func TestIsImage(t *testing.T) {
	testCases := []struct {
		mimeType string
		expected bool
	}{
		{
			mimeType: "image/jpeg",
			expected: true,
		},
		{
			mimeType: "image/png",
			expected: true,
		},
		{
			mimeType: "video/mp4",
			expected: false,
		},
		{
			mimeType: "application/pdf",
			expected: false,
		},
	}

	for _, tc := range testCases {
		actual := isImage(tc.mimeType)
		if actual != tc.expected {
			t.Errorf("got %t but expected %t for mime type %s", actual, tc.expected, tc.mimeType)
		}
	}
}
