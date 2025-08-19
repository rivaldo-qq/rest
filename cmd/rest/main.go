package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/handler"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/repository"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/service"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	storage "github.com/supabase-community/storage-go"
)

var storageClient *storage.Client

func init() {
	supabaseUrl := "https://lqskpaecrquwwsezlwcb.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imxxc2twYWVjcnF1d3dzZXpsd2NiIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1Mjg3MzYwNCwiZXhwIjoyMDY4NDQ5NjA0fQ.b7iOyA5lRdV-Q11PuPDrTnsW9ho45kk1D9TzK_aAqEU" // ⚠️ pakai service role di server
	storageClient = storage.NewClient(supabaseUrl, supabaseKey, nil)
}

func handleGetFileName(c *fiber.Ctx) error {
	fileNameParam := c.Params("filename")

	// coba download dari Supabase Storage
	data, err := storageClient.DownloadFile("cikalbakalstorage", fileNameParam)
	if err != nil {
		// kalau error 404 (file nggak ada di bucket)
		if strings.Contains(err.Error(), "Not Found") {
			return c.Status(http.StatusNotFound).SendString("Not Found")
		}
		log.Println(err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}
	mimeType := http.DetectContentType(data)

	c.Set("Content-Type", mimeType)
	return c.Send(data)
}

func main() {
	godotenv.Load()
	ctx := context.Background()
	app := fiber.New()

	db := database.ConnectDB(ctx, os.Getenv("DB_URI"))
	orderRepository := repository.NewOrderRepository(db)
	webhookService := service.NewWebhookService(orderRepository)
	webhookHandler := handler.NewWebhookHandler(webhookService)

	app.Use(cors.New())
	app.Get("/storage/product/:filename", handleGetFileName)
	app.Post("/product/upload", handler.UploadProductImageHandler)
	app.Post("/webhook/xendit/invoice", webhookHandler.ReceiveInvoice)

	app.Listen(":3000")
}
