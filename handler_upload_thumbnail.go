package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentTypeHeader := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentTypeHeader)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", nil)
		return
	}
	if mediaType != "image/png" && mediaType != "image/jpeg" {
		respondWithError(w, http.StatusBadRequest, "Wrong media type", nil)
		return
	}
	
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
		
	extension := strings.Split(mediaType, "/")[1]
	filename := fmt.Sprintf("%s.%s", videoID, extension)	
	path := filepath.Join(cfg.assetsRoot, filename)

	newfile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create file", err)
		return
	}
	defer newfile.Close()
	
	_, err = io.Copy(newfile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy file", err)
		return
	}

	url := fmt.Sprintf("http://localhost:%s/%s", cfg.port, path)
	video.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
