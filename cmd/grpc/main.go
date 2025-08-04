package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/grpcmiddleware"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/handler"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/repository"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/service"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/auth"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/cart"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/newsletter"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/order"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/product"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pkg/database"
	"github.com/joho/godotenv"
	gocache "github.com/patrickmn/go-cache"
	"github.com/xendit/xendit-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	xendit.Opt.SecretKey = os.Getenv("XENDIT_SECRET")

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Panicf("Error when listening %v", err)
	}

	db := database.ConnectDB(ctx, os.Getenv("DB_URI"))
	log.Println("Connected to database.")

	cacheService := gocache.New(time.Hour*24, time.Hour)

	authMiddleware := grpcmiddleware.NewAuthMiddleware(cacheService)

	authRepository := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepository, cacheService)
	authHandler := handler.NewAuthHandler(authService)

	productRepository := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepository)
	productHandler := handler.NewProductHandler(productService)

	cartRepository := repository.NewCartRepository(db)
	cartService := service.NewCartService(productRepository, cartRepository)
	cartHandler := handler.NewCartHandler(cartService)

	orderRepository := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(db, orderRepository, productRepository)
	orderHandler := handler.NewOrderHandler(orderService)

	newsletterRepository := repository.NewNewsletterRepository(db)
	newsletterService := service.NewNewsletterService(newsletterRepository)
	newsletterHandler := handler.NewNewsletterHandler(newsletterService)

	serv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.ErrorMiddleware,
			authMiddleware.Middleware,
		),
	)

	auth.RegisterAuthServiceServer(serv, authHandler)
	product.RegisterProductServiceServer(serv, productHandler)
	cart.RegisterCartServiceServer(serv, cartHandler)
	order.RegisterOrderServiceServer(serv, orderHandler)
	newsletter.RegisterNewsletterServiceServer(serv, newsletterHandler)

	if os.Getenv("ENVIRONMENT") == "dev" {
		reflection.Register(serv)
		log.Println("Reflection is registered.")
	}

	log.Println("Server is running on :50052 port.")
	if err := serv.Serve(lis); err != nil {
		log.Panicf("Server is error %v", err)
	}
}
