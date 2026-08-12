package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Mpayy/e-commerce/internal/product/entity"
	"github.com/Mpayy/e-commerce/internal/product/model"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepositoryImpl struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ProductRepository {
	var p model.ProductModel
	return &ProductRepositoryImpl{
		collection: db.Collection(p.CollectionName()),
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

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperror.ErrNotFound
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
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
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
	filter := bson.M{"_id": product.ID}

	update := bson.M{
		"$set": bson.M{
			"category_id": product.CategoryID,
			"name":        product.Name,
			"slug":        product.Slug,
			"description": product.Description,
			"price":       product.Price,
			"stock":       product.Stock,
			"sku":         product.SKU,
			"is_active":   product.IsActive,
			"updated_at":  now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
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
		return apperror.ErrNotFound
	}

	product.UpdatedAt = now

	return nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uint) error {
	filter := bson.M{
		"_id":       id, // langsung int64
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
		return apperror.ErrNotFound
	}
	return nil
}

func (r *ProductRepositoryImpl) DecreaseStock(ctx context.Context, productID uint, quantity int) error {
	filter := bson.M{
		"_id": productID,
		"stock": bson.M{
			"$gte": quantity, // mengambil product yang stoknya >= quantity
		},
	}

	update := bson.M{
		"$inc": bson.M{ // menggunakan $inc untuk update nilai stock kalau quantity positif maka stock bertambah kalau negatif maka stock berkurang
			"stock": -quantity,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		count, err := r.collection.CountDocuments(ctx, bson.M{"_id": productID})
		if err != nil {
			return err
		}
		if count == 0 {
			return apperror.ErrNotFound
		}
		return apperror.ErrInsufficientStock
	}

	return nil
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
		return apperror.ErrNotFound
	}

	return nil
}
