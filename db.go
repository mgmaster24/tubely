package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		fmt.Println("Video URL is nil")
		return video, nil
	}

	urlParts := strings.Split(*video.VideoURL, ",")
	if len(urlParts) != 2 {
		return video, fmt.Errorf("Incorrect number of url parts")
	}

	bucket := urlParts[0]
	key := urlParts[1]

	url, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Hour*24)
	if err != nil {
		return video, err
	}

	video.VideoURL = &url
	return video, nil
}
