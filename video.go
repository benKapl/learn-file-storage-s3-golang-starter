package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}

	var output struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return "", fmt.Errorf("could not parse ffprobe output: %v", err)
	}

	if len(output.Streams) == 0 {
		return "", errors.New("no video streams found")
	}

	height := output.Streams[0].Height
	width := output.Streams[0].Width

	ratio := float64(width) / float64(height)

	switch {
	case isHorizontalRatio(ratio):
		return "16:9", nil
	case isVerticalRatio(ratio):
		return "9:16", nil
	default:
		return "other", nil
	}
}

func isHorizontalRatio(value float64) bool {
	// Horizontal ratio is 16/9
	lowerBound := 1.77
	upperBound := 1.78
	return value > lowerBound && value < upperBound
}

func isVerticalRatio(value float64) bool {
	// Horizontal ratio is 16/9
	lowerBound := 0.562
	upperBound := 0.564
	return value > lowerBound && value < upperBound
}

func processVideoForFastStart(filePath string) (string, error) {
	out := filePath + ".processing"

	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-c", "copy",
		"-movflags", "faststart",
		"-f", "mp4", out)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v", err)
	}
	return out, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	object, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return object.URL, nil
}
