package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity"
	jwtentity "github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity/jwt"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/repository"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/utils"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/product"
	
	storage "github.com/supabase-community/storage-go"
)

func init() {
	supabaseUrl := "https://lqskpaecrquwwsezlwcb.supabase.co/storage/v1"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imxxc2twYWVjcnF1d3dzZXpsd2NiIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1Mjg3MzYwNCwiZXhwIjoyMDY4NDQ5NjA0fQ.b7iOyA5lRdV-Q11PuPDrTnsW9ho45kk1D9TzK_aAqEU" // ⚠️ pakai service role di server
	storageClient = storage.NewClient(supabaseUrl, supabaseKey, nil)
}

var storageClient *storage.Client

type IProductService interface {
	CreateProduct(ctx context.Context, request *product.CreateProductRequest) (*product.CreateProductResponse, error)
	DetailProduct(ctx context.Context, request *product.DetailProductRequest) (*product.DetailProductResponse, error)
	EditProduct(ctx context.Context, request *product.EditProductRequest) (*product.EditProductResponse, error)
	DeleteProduct(ctx context.Context, request *product.DeleteProductRequest) (*product.DeleteProductResponse, error)
	ListProduct(ctx context.Context, request *product.ListProductRequest) (*product.ListProductResponse, error)
	ListProductAdmin(ctx context.Context, request *product.ListProductAdminRequest) (*product.ListProductAdminResponse, error)
	HighlightProducts(ctx context.Context, request *product.HighlightProductsRequest) (*product.HighlightProductsResponse, error)
}

type productService struct {
	productRepository repository.IProductRepository
}

func (ps *productService) CreateProduct(ctx context.Context, request *product.CreateProductRequest) (*product.CreateProductResponse, error) {
	claims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != entity.UserRoleAdmin {
		return nil, utils.UnauthenticatedResponse()
	}

	// cek juga apakah image nya ada ?
	// cek apakah file ada di Supabase
	

	// File ditemukan

	
	err = ps.productRepository.CreateNewProduct(ctx)
	if err != nil {
		return nil, err
	}

	return &product.CreateProductResponse{
		Base: utils.SuccessResponse("Product is created"),
		
	}, nil
}

func (ps *productService) DetailProduct(ctx context.Context, request *product.DetailProductRequest) (*product.DetailProductResponse, error) {
	// queyr ke db dengan data id
	
	// apabila null, kita return not found
	if productEntity == nil {
		return &product.DetailProductResponse{
			Base: utils.NotFoundResponse("Product not found"),
		}, nil
	}

	return &product.DetailProductResponse{
		Base:        utils.SuccessResponse("Get product detail success"),
		Id:          productEntity.Id,
		Name:        productEntity.Name,
		Description: productEntity.Description,
		Price:       productEntity.Price,
		ImageUrl:    fmt.Sprintf("%s/product/%s", os.Getenv("STORAGE_SERVICE_URL"), productEntity.ImageFileName),
	}, nil
}

func (ps *productService) EditProduct(ctx context.Context, request *product.EditProductRequest) (*product.EditProductResponse, error) {
	claims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != entity.UserRoleAdmin {
		return nil, utils.UnauthenticatedResponse()
	}

	productEntity, err := ps.productRepository.GetProductById(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	if productEntity == nil {
		return &product.EditProductResponse{
			Base: utils.NotFoundResponse("Product not found"),
		}, nil
	}

	if productEntity.ImageFileName != request.ImageFileName {
		newImagePath := filepath.Join("storage", "product", request.ImageFileName)
		_, err = os.Stat(newImagePath)
		if err != nil {
			if os.IsNotExist(err) {
				return &product.EditProductResponse{
					Base: utils.BadRequestResponse("Image not found"),
				}, nil
			}

			return nil, err
		}

		oldImagePath := filepath.Join("storage", "product",)
		err = os.Remove(oldImagePath)
		if err != nil {
			return nil, err
		}
	}

	newProduct := entity.Product{
		Id:            request.Id,
		Name:          request.Name,
		Description:   request.Description,
		Price:         request.Price,
		ImageFileName: request.ImageFileName,
		UpdatedAt:     time.Now(),
		UpdatedBy:     &claims.FullName,
	}

	err = ps.productRepository.UpdateProduct(ctx, &newProduct)
	if err != nil {
		return nil, err
	}

	return &product.EditProductResponse{
		Base: utils.SuccessResponse("Edit product success"),
		Id:   request.Id,
	}, nil
}

func (ps *productService) DeleteProduct(ctx context.Context, request *product.DeleteProductRequest) (*product.DeleteProductResponse, error) {
	claims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != entity.UserRoleAdmin {
		return nil, utils.UnauthenticatedResponse()
	}

	
	if productEntity == nil {
		return &product.DeleteProductResponse{
			Base: utils.NotFoundResponse("Product not found"),
		}, nil
	}

	err = ps.productRepository.DeleteProduct(ctx, request.Id, time.Now(), claims.FullName)
	if err != nil {
		return nil, err
	}

	return &product.DeleteProductResponse{
		Base: utils.SuccessResponse("Delete product success"),
	}, nil
}

func (ps *productService) ListProduct(ctx context.Context, request *product.ListProductRequest) (*product.ListProductResponse, error) {
	products, paginationResponse, err := ps.productRepository.GetProductsPagination(ctx, request.Pagination)
	if err != nil {
		return nil, err
	}

	var data []*product.ListProductResponseItem = make([]*product.ListProductResponseItem, 0)
	for _, prod := range products {
		data = append(data, &product.ListProductResponseItem{
			Id:          prod.Id,
			Name:        prod.Name,
			Description: prod.Description,
			Price:       prod.Price,
			ImageUrl:    fmt.Sprintf("%s/product/%s", os.Getenv("STORAGE_SERVICE_URL"), prod.ImageFileName),
		})
	}

	return &product.ListProductResponse{
		Base:       utils.SuccessResponse("Get list product success"),
		Pagination: paginationResponse,
		Data:       data,
	}, nil
}

func (ps *productService) ListProductAdmin(ctx context.Context, request *product.ListProductAdminRequest) (*product.ListProductAdminResponse, error) {
	claims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != entity.UserRoleAdmin {
		return nil, utils.UnauthenticatedResponse()
	}

	products, paginationResponse, err := ps.productRepository.GetProductsPaginationAdmin(ctx, request.Pagination)
	if err != nil {
		return nil, err
	}

	var data []*product.ListProductAdminResponseItem = make([]*product.ListProductAdminResponseItem, 0)
	for _, prod := range products {
		data = append(data, &product.ListProductAdminResponseItem{
			Id:          prod.Id,
			Name:        prod.Name,
			Description: prod.Description,
			Price:       prod.Price,
			ImageUrl:    fmt.Sprintf("%s/product/%s", os.Getenv("STORAGE_SERVICE_URL"), prod.ImageFileName),
		})
	}

	return &product.ListProductAdminResponse{
		Base:       utils.SuccessResponse("Get list product admin success"),
		Pagination: paginationResponse,
		Data:       data,
	}, nil
}

func (ps *productService) HighlightProducts(ctx context.Context, request *product.HighlightProductsRequest) (*product.HighlightProductsResponse, error) {
	products, err := ps.productRepository.GetProductHighlight(ctx)
	if err != nil {
		return nil, err
	}

	var data []*product.HighlightProductsResponseItem = make([]*product.HighlightProductsResponseItem, 0)
	for _, prod := range products {
		data = append(data, &product.HighlightProductsResponseItem{
			Id:          prod.Id,
			Name:        prod.Name,
			Description: prod.Description,
			Price:       prod.Price,
			ImageUrl:    fmt.Sprintf("%s/product/%s", os.Getenv("STORAGE_SERVICE_URL"), prod.ImageFileName),
		})
	}

	return &product.HighlightProductsResponse{
		Base: utils.SuccessResponse("Get highlight products success"),
		Data: data,
	}, nil
}

func NewProductService(productRepository repository.IProductRepository) IProductService {
	return &productService{
		productRepository: productRepository,
	}
}
