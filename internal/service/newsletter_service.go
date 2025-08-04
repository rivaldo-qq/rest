package service

import (
	"context"
	"time"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/repository"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/utils"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/newsletter"
	"github.com/google/uuid"
)

type INewsletterService interface {
	SubscribeNewsletter(ctx context.Context, request *newsletter.SubcribeNewsletterRequest) (*newsletter.SubcribeNewsletterResponse, error)
}

type newsletterService struct {
	newsletterRepository repository.INewsletterRepository
}

func (ns *newsletterService) SubscribeNewsletter(ctx context.Context, request *newsletter.SubcribeNewsletterRequest) (*newsletter.SubcribeNewsletterResponse, error) {
	newsletterEntity, err := ns.newsletterRepository.GetNewsletterByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	if newsletterEntity != nil {
		return &newsletter.SubcribeNewsletterResponse{
			Base: utils.SuccessResponse("Subscribe newsletter success"),
		}, nil
	}

	newNewsletterEntity := entity.Newsletter{
		Id:        uuid.NewString(),
		FullName:  request.FullName,
		Email:     request.Email,
		CreatedAt: time.Now(),
		CreatedBy: "Public",
	}
	err = ns.newsletterRepository.CreateNewNewsletter(ctx, &newNewsletterEntity)
	if err != nil {
		return nil, err
	}

	return &newsletter.SubcribeNewsletterResponse{
		Base: utils.SuccessResponse("Subscribe newsletter success"),
	}, nil
}

func NewNewsletterService(newsletterRepository repository.INewsletterRepository) INewsletterService {
	return &newsletterService{
		newsletterRepository: newsletterRepository,
	}
}
