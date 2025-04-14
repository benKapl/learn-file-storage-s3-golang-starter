package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const uploadLimit = 10 << 30
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user ID from context", nil)
		return
	}

	fmt.Println("uploading video", videoID, "by user", userID)

	// Check video and authorization
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video does not exist", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Invalid operation", err)
		return
	}

	// Parse video file
	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse formfile code", err)
		return
	}
	defer file.Close()

	//Get media type
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type, only MP4 is allowed", nil)
		return
	}

	// Save on tempFile file on disk
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create temp file", err)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	if _, err = io.Copy(tempFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not write file to disk", err)
		return
	}

	// Reset the tempFile's file pointer to the beginning
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}

	// Put object to s3
	key := getAssetPath(mediaType)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &key,
		Body:        tempFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not put object to S3", err)
		return
	}

	// Update video URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
	video.VideoURL = &url
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

}
