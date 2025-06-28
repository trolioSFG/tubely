package main

import (
	"fmt"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"

	"io"
//	"encoding/base64"
	"mime"
	"os"
	"path/filepath"
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
		respondWithError(w, http.StatusBadRequest, "Unable to parse file", err)
		return
	}
	defer file.Close()

	media := header.Header.Get("Content-Type")
	fmt.Println("Content-Type:", media)

	/**
	Read later when copying to disk
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read file", err)
		return
	}
	**/

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "No auth", fmt.Errorf("userID mismatch"))
		return
	}

	/***
	thumb := thumbnail{
		data: data,
		mediaType: media,
	}
	videoThumbnails[videoID] = thumb

	thumbURL := "http://localhost:8091/api/thumbnails/" + videoID.String()
	video.ThumbnailURL = &thumbURL 
	***/

	/**
	encodedData := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", media, encodedData)
	***/
	extensions, err := mime.ExtensionsByType(media)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "MIME Type", err)
		return
	}

	// The returned extensions will each begin with a leading dot, as in ".html"
	ext := extensions[0]
	fileName := videoID.String() + ext
	fpath := filepath.Join(cfg.assetsRoot, fileName)
	dstFile, err := os.Create(fpath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Create file", err)
		return
	}

	written, err := io.Copy(dstFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Copy", err)
		return
	}

	fmt.Println(written, "bytes copied to file:", fpath)
	file.Close()


	dataURL := "http://localhost:8091/assets/" + fileName
	video.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
