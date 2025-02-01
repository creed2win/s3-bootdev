package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
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
	const maxMemory = 10 << 20 // 10 MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get file", err)
		return
	}

	mediaType := header.Header.Get("Content-Type")
	fmt.Println("mediaType parsed by hand: ", mediaType)

	mtString, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", err)
		return
	}
	fmt.Println("mediaType from mime.ParseMediaType: ", mtString)

	if mtString != "image/jpeg" && mtString != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", nil)
		return
	}

	defer file.Close()

	video, err := cfg.db.GetVideo(videoID) // check if the video exists
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Not the owner of the video", nil)
		return
	}

	updatedVideo := video

	// save the file to disk
	fileExtension := func(mediaType string) string {
		switch mediaType {
		case "image/jpeg":
			return "jpg"
		case "image/png":
			return "png"
		default:
			return "bin"
		}
	}

	filePath := path.Join(cfg.assetsRoot, fmt.Sprintf("%v.%v", videoID, fileExtension(mediaType)))
	newFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create file", err)
		return
	}

	io.Copy(newFile, file)

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate random name", err)
		return
	}
	randomName := hex.EncodeToString(randomBytes)

	tURL := fmt.Sprintf("http://localhost:%v/assets/%v.%v", cfg.port, randomName, fileExtension(mediaType))

	updatedVideo.ThumbnailURL = &tURL

	err = cfg.db.UpdateVideo(updatedVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedVideo)
}
