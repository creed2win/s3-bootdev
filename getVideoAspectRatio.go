package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

// get aspect ratio from video using exec ffprobe command
func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type Stream struct {
		Width               int    `json:"width"`
		Height              int    `json:"height"`
		DisplayAspectRation string `json:"display_aspect_ratio"`
	}

	type FFProbeOutput struct {
		Streams []Stream `json:"streams"`
	}

	var probeData FFProbeOutput
	err = json.Unmarshal(stdout.Bytes(), &probeData)
	if err != nil {
		return "error umarshaling data to get aspect ration", err
	}

	if len(probeData.Streams) > 0 {
		aspectRatio := probeData.Streams[0].DisplayAspectRation
		if aspectRatio == "16:9" {
			return "landscape", nil
		} else if aspectRatio == "9:16" {
			return "portrait", nil
		}

		return "other", nil
	}

	return "", nil
}
