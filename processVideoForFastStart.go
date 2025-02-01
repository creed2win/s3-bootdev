package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func processVideoForFastStart(filePath string) (string, error) {
	// Create output path with "-faststart" suffix
	ext := filepath.Ext(filePath)
	base := filePath[:len(filePath)-len(ext)]
	outputPath := base + "-faststart" + ext

	// Execute ffmpeg command to move moov atom to start
	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-c", "copy",
		"-movflags", "faststart",
		"-f", "mp4",
		outputPath,
	)

	// Run the command and capture any errors
	if err := cmd.Run(); err != nil {
		fmt.Println("error while processing:", err)
		return "", err
	}

	return outputPath, nil
}
