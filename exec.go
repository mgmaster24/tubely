package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath,
	)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type ffprobeOutput struct {
		Streams []struct {
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"streams"`
	}

	var out ffprobeOutput
	err = json.Unmarshal(buffer.Bytes(), &out)
	if err != nil {
		return "", err
	}

	ratio := math.Round((out.Streams[0].Width/out.Streams[0].Height)*100) / 100
	if ratio == 0.56 {
		return "9:16", nil
	} else if ratio == 1.78 {
		return "16:9", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFile := fmt.Sprintf("%s.processing", filePath)
	cmd := exec.Command(
		"ffmpeg",
		"-i",
		filePath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		outputFile,
	)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outputFile, nil
}
