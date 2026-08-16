package productclient

import (
	"context"
	"fmt"

	"github.com/Mpayy/e-commerce/monolith/internal/product/entity"
	"github.com/Mpayy/e-commerce/monolith/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	productv1 "github.com/Mpayy/e-commerce/proto/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
type ProductGRPCClient struct {
	client productv1.ProductServiceClient
}

func NewProductGRPCClient(conn *grpc.ClientConn) usecase.ProductService {
	return &ProductGRPCClient{client: productv1.NewProductServiceClient(conn)}
}

func (c *ProductGRPCClient) GetByProductID(ctx context.Context, id uint) (*entity.Product, error) {
	resp, err := c.client.GetByID(ctx, &productv1.GetByIDRequest{Id: uint64(id)})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return nil, apperror.ErrProductNotFound
		default:
			return nil, fmt.Errorf("failed get product by id from grpc: %w", err)
		}
	}
	return &entity.Product{
		ID:          uint(resp.Product.Id),
		CategoryID:  uint(resp.Product.CategoryId),
		Name:        resp.Product.Name,
		Slug:        resp.Product.Slug,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
		Stock:       int(resp.Product.Stock),
		SKU:         resp.Product.Sku,
		IsActive:    resp.Product.IsActive,
	}, nil
}

func (c *ProductGRPCClient) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	protoIDs := make([]uint64, len(ids))
	for i, id := range ids {
		protoIDs[i] = uint64(id)
	}

	resp, err := c.client.GetByIDs(ctx, &productv1.GetByIDsRequest{Ids: protoIDs})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return nil, apperror.ErrProductNotFound
		default:
			return nil, fmt.Errorf("failed get products by ids from grpc: %w", err)
		}
	}

	products := make([]entity.Product, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, entity.Product{
			ID:          uint(p.Id),
			CategoryID:  uint(p.CategoryId),
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
			Price:       p.Price,
			Stock:       int(p.Stock),
			SKU:         p.Sku,
			IsActive:    p.IsActive,
		})
	}
	return products, nil
}

func (c *ProductGRPCClient) BulkDecreaseStock(ctx context.Context, checkoutID string, items []entity.BulkDecreaseStock) error {
	protoItems := make([]*productv1.DecreaseStockRequest, len(items))
	for i, item := range items {
		protoItems[i] = &productv1.DecreaseStockRequest{
			ProductId: uint64(item.ProductID),
			Quantity:  int32(item.Quantity),
		}
	}
	_, err := c.client.BulkDecreaseStock(ctx, &productv1.BulkDecreaseStockRequest{
		CheckoutId: checkoutID,
		Items:      protoItems,
	})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return apperror.ErrProductNotFound
		case codes.FailedPrecondition:
			return apperror.ErrInsufficientStock
		default:
			return fmt.Errorf("failed decrease stock from grpc: %w", err)
		}
	}
	return nil
}

func (c *ProductGRPCClient) BulkRestoreStock(ctx context.Context, checkoutID string) error {
	_, err := c.client.BulkRestoreStock(ctx, &productv1.BulkRestoreStockRequest{CheckoutId: checkoutID})
	if err != nil {
		return fmt.Errorf("failed to restore stock from grpc: %w", err)
	}
	return nil
}
