package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse request", err)
		return
	}

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't retrieve file", err)
		return
	}

	defer file.Close()

	mediaType := fileHeader.Header.Get("Content-Type")
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read file", err)
		return
	}

	vidMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get video metadata", err)
		return
	}

	*vidMetadata.ThumbnailURL = fmt.Sprintf(
		"data:%s;base64, %s",
		mediaType,
		base64.StdEncoding.EncodeToString(data),
	)
	vidMetadata.UpdatedAt = time.Now()
	fmt.Println(vidMetadata)
	err = cfg.db.UpdateVideo(vidMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video", err)
		return
	}

	fmt.Println(vidMetadata)

	respondWithJSON(w, http.StatusOK, vidMetadata)
}
