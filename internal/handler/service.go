package handler

import (
	"context"
	"fmt"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/utils"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/service"
)

type serviceHandler struct {
	service.UnimplementedHelloWorldServiceServer
}

func (sh *serviceHandler) HelloWorld(ctx context.Context, request *service.HelloWorldRequest) (*service.HelloWorldResponse, error) {
	validationErrors, err := utils.CheckValidation(request)
	if err != nil {
		return nil, err
	}
	if validationErrors != nil {
		return &service.HelloWorldResponse{
			Base: utils.ValidationErrorResponse(validationErrors),
		}, nil
	}

	return &service.HelloWorldResponse{
		Message: fmt.Sprintf("Hello %s", request.Name),
		Base:    utils.SuccessResponse("Success"),
	}, nil
}

func NewServiceHandler() *serviceHandler {
	return &serviceHandler{}
}
