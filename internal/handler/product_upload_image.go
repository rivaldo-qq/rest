package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	storage "github.com/supabase-community/storage-go"
)

var storageClient *storage.Client

func init() {
	supabaseUrl := "https://lqskpaecrquwwsezlwcb.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imxxc2twYWVjcnF1d3dzZXpsd2NiIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1Mjg3MzYwNCwiZXhwIjoyMDY4NDQ5NjA0fQ.b7iOyA5lRdV-Q11PuPDrTnsW9ho45kk1D9TzK_aAqEU"

	// PERBAIKAI 1: Tambahkan "/storage/v1" di akhir URL
	storageClient = storage.NewClient(supabaseUrl+"/storage/v1", supabaseKey, nil)
}

func UploadProductImageHandler(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "image data not found",
		})
	}

	// validasi extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "image extension is not allowed (jpg, jpeg, png, webp)",
		})
	}

	// validasi content type
	contentType := file.Header.Get("Content-Type")
	allowedContentType := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedContentType[contentType] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "content type is not allowed",
		})
	}

	// PERBAIKAN 3: Tambahkan kurung tutup yang hilang
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("product_%d%s", timestamp, filepath.Ext(file.Filename))

	// buka file sebagai io.Reader
	src, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to open file",
		})
	}
	defer src.Close()

	// PERBAIKAN 2: Gunakan nilai langsung, bukan pointer, untuk FileOptions
	opts := storage.FileOptions{
		ContentType: contentType,
		Upsert:      true,
	}

	// upload ke bucket "cikalbakalstorage"
	_, err = storageClient.UploadFile("cikalbakalstorage", fileName, src, opts)
	if err != nil {
		fmt.Println("upload error:", err) // Cek error detail di sini
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to upload file to storage",
		})
	}

	// SARAN: Gunakan fungsi bawaan untuk mendapatkan URL publik
	publicUrl := storageClient.GetPublicUrl("cikalbakalstorage", fileName, storage.UrlOptions{
        Download: false,
    }).SignedURL

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "upload success",
		"file_name": fileName,
		"url":       publicUrl,
	})
}
