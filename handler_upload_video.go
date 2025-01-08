package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	vidMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get video metadata", err)
		return
	}

	if vidMetadata.UserID != userID {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"User is not authorized to upload this video",
			err,
		)
	}

	const uploadLimit = 1 << 30
	err = r.ParseMultipartForm(uploadLimit)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse request", err)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't retrieve file", err)
		return
	}

	defer file.Close()
	contentType := fileHeader.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't retrieve the media type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Incorrect media type", err)
		return
	}

	f, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create temp file", err)
		return
	}

	defer os.Remove(f.Name())
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save file", err)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to reposition file pointer",
			err,
		)
		return
	}

	processedFile, err := processVideoForFastStart(f.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process video", err)
		return
	}

	ratio, err := getVideoAspectRatio(processedFile)
	aspect := ""
	switch ratio {
	case "16:9":
		aspect = "landscape"
	case "9:16":
		aspect = "portrait"
	default:
		aspect = "other"
	}

	fastStartF, err := os.Open(processedFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to open processed video", err)
		return
	}

	defer fastStartF.Close()

	fi, err := fastStartF.Stat()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get file indo", err)
		return
	}

	key := filepath.Join(aspect, getAssetsPath(mediaType))
	length := fi.Size()
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:        &cfg.s3Bucket,
		Body:          fastStartF,
		Key:           &key,
		ContentType:   &mediaType,
		ContentLength: &length,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload video to S3", err)
		return
	}

	videoURL := fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
	vidMetadata.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(vidMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video", err)
		return
	}

	vidMetadata, err = cfg.dbVideoToSignedVideo(vidMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to presign video URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vidMetadata)
}
