package handler

import (
	"context"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/service"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/utils"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/newsletter"
)

type newsletterHandler struct {
	newsletter.UnimplementedNewsletterServiceServer

	newsletterService service.INewsletterService
}

func (nh *newsletterHandler) SubscribeNewsletter(ctx context.Context, request *newsletter.SubcribeNewsletterRequest) (*newsletter.SubcribeNewsletterResponse, error) {
	validationErrors, err := utils.CheckValidation(request)
	if err != nil {
		return nil, err
	}
	if validationErrors != nil {
		return &newsletter.SubcribeNewsletterResponse{
			Base: utils.ValidationErrorResponse(validationErrors),
		}, nil
	}

	res, err := nh.newsletterService.SubscribeNewsletter(ctx, request)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewNewsletterHandler(newsletterService service.INewsletterService) *newsletterHandler {
	return &newsletterHandler{
		newsletterService: newsletterService,
	}
}
