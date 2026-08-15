package grpc

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	productv1 "github.com/Mpayy/e-commerce/proto/product/v1"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductGRPCServer struct {
	productv1.UnimplementedProductServiceServer
	productUsecase usecase.ProductService
}

func NewProductGRPCServer(productService usecase.ProductService) *ProductGRPCServer {
	return &ProductGRPCServer{
		productUsecase: productService,
	}
}

func (s *ProductGRPCServer) GetByID(ctx context.Context, req *productv1.GetByIDRequest) (*productv1.GetByIDResponse, error) {
	product, err := s.productUsecase.GetByProductID(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, apperror.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
	}

	return &productv1.GetByIDResponse{
		Product: &productv1.Product{
			Id:         uint64(product.ID),
			CategoryId: uint64(product.CategoryID),
			Name:       product.Name,
			Price:      product.Price,
			Stock:      int32(product.Stock),
			Sku:        product.SKU,
			IsActive:   product.IsActive,
		},
	}, nil
}

func (s *ProductGRPCServer) GetByIDs(ctx context.Context, req *productv1.GetByIDsRequest) (*productv1.GetByIDsResponse, error) {
	ids := make([]uint, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = uint(id)
	}

	products, err := s.productUsecase.GetProductsByIDs(ctx, ids)
	if err != nil {
		if errors.Is(err, apperror.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
	}

	productRes := make([]*productv1.Product, 0, len(products))
	for _, product := range products {
		productRes = append(productRes, &productv1.Product{
			Id:          uint64(product.ID),
			CategoryId:  uint64(product.CategoryID),
			Name:        product.Name,
			Slug:        product.Slug,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int32(product.Stock),
			Sku:         product.SKU,
			IsActive:    product.IsActive,
		})
	}

	return &productv1.GetByIDsResponse{
		Products: productRes,
	}, nil
}

func (s *ProductGRPCServer) DecreaseStock(ctx context.Context, req *productv1.DecreaseStockRequest) (*productv1.DecreaseStockResponse, error) {
	err := s.productUsecase.DecreaseStock(ctx, uint(req.ProductId), int(req.Quantity))
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		case errors.Is(err, apperror.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, "insufficient stock")
		default:
			return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
		}
	}
	return &productv1.DecreaseStockResponse{}, nil
}
