package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepositoryImpl struct {
	collection       *mongo.Collection
	ledgerCollection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ProductRepository {
	var p model.ProductModel
	var l model.StockLedgerModel
	return &ProductRepositoryImpl{
		collection:       db.Collection(p.CollectionName()),
		ledgerCollection: db.Collection(l.CollectionName()),
	}
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product *entity.Product) error {
	id, err := getNextSequence(ctx, r.collection.Database(), "product_id")
	if err != nil {
		return err
	}

	doc := model.ProductModel{
		ID:          id,
		CategoryID:  int64(product.CategoryID),
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "slug") {
				return apperror.ErrDuplicatedProduct
			}
			if strings.Contains(errMsg, "sku") {
				return apperror.ErrDuplicatedProductSku
			}
			return apperror.ErrDuplicatedKey
		}
		return err
	}

	product.ID = uint(id)

	return nil
}

func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	var model model.ProductModel

	err := r.collection.FindOne(ctx, bson.M{"_id": int64(id)}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}

	product := entity.Product{
		ID:          uint(model.ID),
		CategoryID:  uint(model.CategoryID),
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		Price:       model.Price,
		Stock:       model.Stock,
		SKU:         model.SKU,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return &product, nil
}

func (r *ProductRepositoryImpl) FindByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	var models []model.ProductModel
	idsInt64 := make([]int64, len(ids))
	for i, id := range ids {
		idsInt64[i] = int64(id)
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": idsInt64,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	products := make([]entity.Product, len(models))
	for i, m := range models {
		products[i] = entity.Product{
			ID:          uint(m.ID),
			CategoryID:  uint(m.CategoryID),
			Name:        m.Name,
			Slug:        m.Slug,
			Description: m.Description,
			Price:       m.Price,
			Stock:       m.Stock,
			SKU:         m.SKU,
			IsActive:    m.IsActive,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}

	return products, nil
}

func (r *ProductRepositoryImpl) FindAll(ctx context.Context, filter *entity.ProductFilter) ([]entity.Product, int64, error) {
	var models []model.ProductModel
	var total int64

	query := bson.M{
		"is_active": true,
	}

	if filter.Search != "" {
		query["name"] = bson.M{
			"$regex":   filter.Search,
			"$options": "i", // "i" = case-insensitive (mirip LIKE %search% di SQL)
		}
	}

	if filter.CategoryID != 0 {
		query["category_id"] = int64(filter.CategoryID)
	}

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []entity.Product{}, 0, nil
	}

	offSet := int64((filter.Page - 1) * filter.Limit)
	limit := int64(filter.Limit)

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(offSet).
		SetSort(bson.D{{Key: "name", Value: 1}}) // Order By name ASC (1 = ASC, -1 = DESC)

	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &models); err != nil {
		return nil, 0, err
	}

	products := make([]entity.Product, 0, len(models))
	for _, m := range models {
		products = append(products, entity.Product{
			ID:          uint(m.ID),
			CategoryID:  uint(m.CategoryID),
			Name:        m.Name,
			Slug:        m.Slug,
			Description: m.Description,
			Price:       m.Price,
			Stock:       m.Stock,
			SKU:         m.SKU,
			IsActive:    m.IsActive,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}

	return products, total, nil
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entity.Product) error {
	now := time.Now()
	filter := bson.M{"_id": int64(product.ID)}

	doc := model.ProductModel{
		ID:          int64(product.ID),
		CategoryID:  int64(product.CategoryID),
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   now,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "slug") {
				return apperror.ErrDuplicatedProduct
			}
			if strings.Contains(errMsg, "sku") {
				return apperror.ErrDuplicatedProductSku
			}
			return apperror.ErrDuplicatedKey
		}
		return err
	}

	if result.MatchedCount == 0 {
		return apperror.ErrRecordNotFound
	}

	product.UpdatedAt = now

	return nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uint) error {
	filter := bson.M{
		"_id":       int64(id),
		"is_active": true,
	}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return apperror.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepositoryImpl) BulkDecreaseStock(ctx context.Context, checkoutID string, items []entity.BulkDecreaseStock) error {
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		ledger := model.StockLedgerModel{
			ID:         checkoutID + ":decrease",
			CheckoutID: checkoutID,
			Operation:  "decrease",
			Items:      toLedgerItems(items),
			CreatedAt:  time.Now(),
		}
		if _, err := r.ledgerCollection.InsertOne(sessCtx, ledger); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, nil
			}
			return nil, err
		}
		now := time.Now()
		for _, item := range items {
			filter := bson.M{
				"_id": int64(item.ProductID),
				"stock": bson.M{
					"$gte": item.Quantity,
				},
			}
			update := bson.M{
				"$inc": bson.M{
					"stock": -item.Quantity,
				},
				"$set": bson.M{
					"updated_at": now,
				},
			}

			result, err := r.collection.UpdateOne(sessCtx, filter, update)
			if err != nil {
				return nil, err
			}
			if result.MatchedCount == 0 {
				count, _ := r.collection.CountDocuments(sessCtx, bson.M{"_id": item.ProductID})
				if count == 0 {
					return nil, apperror.ErrRecordNotFound
				}
				return nil, apperror.ErrInsufficientStock
			}
		}
		return nil, nil
	})

	return err
}

func (r *ProductRepositoryImpl) BulkRestoreStock(ctx context.Context, checkoutID string) error {
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		restoreID := checkoutID + ":restore"
		var existing model.StockLedgerModel
		err := r.ledgerCollection.FindOne(sessCtx, bson.M{"_id": restoreID}).Decode(&existing)
		if err == nil {
			return nil, nil
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}

		var decreaseLedger model.StockLedgerModel
		err = r.ledgerCollection.FindOne(sessCtx, bson.M{"_id": checkoutID + ":decrease"}).Decode(&decreaseLedger)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, nil
			}
			return nil, err
		}

		now := time.Now()
		for _, item := range decreaseLedger.Items {
			filter := bson.M{
				"_id": int64(item.ProductID),
			}

			update := bson.M{
				"$inc": bson.M{
					"stock": item.Quantity,
				},
				"$set": bson.M{
					"updated_at": now,
				},
			}

			if _, err := r.collection.UpdateOne(sessCtx, filter, update); err != nil {
				return nil, err
			}
		}

		restoreLedger := model.StockLedgerModel{
			ID:         restoreID,
			CheckoutID: checkoutID,
			Operation:  "restore",
			Items:      decreaseLedger.Items,
			CreatedAt:  now,
		}
		_, err = r.ledgerCollection.InsertOne(sessCtx, restoreLedger)
		return nil, err
	})

	return err
}

func (r *ProductRepositoryImpl) AdjustStock(ctx context.Context, productID uint, quantity int) error {
	filter := bson.M{
		"_id": int64(productID),
	}

	update := bson.M{
		"$set": bson.M{
			"stock":      quantity,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return apperror.ErrRecordNotFound
	}

	return nil
}

func toLedgerItems(items []entity.BulkDecreaseStock) []model.StockLedgerItem {
	ledgerItems := make([]model.StockLedgerItem, len(items))
	for i, item := range items {
		ledgerItems[i] = model.StockLedgerItem{
			ProductID: int64(item.ProductID),
			Quantity:  item.Quantity,
		}
	}
	return ledgerItems
}
