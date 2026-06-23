package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadHandler handles multipart file uploads and saves them locally.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// السماح بملفات حتى حجم 10 ميغابايت
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// إنشاء المجلد تلقائياً
	os.MkdirAll("./uploads", os.ModePerm)

	// توليد اسم فريد لتجنب التعارض
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
	dst, err := os.Create(filepath.Join("./uploads", filename))
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	json.NewEncoder(w).Encode(map[string]string{
		"url": "/uploads/" + filename,
	})
}
