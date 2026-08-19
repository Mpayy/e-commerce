package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CategoryRepositoryImpl struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *mongo.Database) CategoryRepository {
	var c model.CategoryModel
	return &CategoryRepositoryImpl{
		collection: db.Collection(c.CollectionName()),
	}
}

func (r *CategoryRepositoryImpl) Create(ctx context.Context, category *entity.Category) error {
	id, err := getNextSequence(ctx, r.collection.Database(), "category_id")
	if err != nil {
		return err
	}

	doc := model.CategoryModel{
		ID:        id,
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperror.ErrDuplicatedKey
		}
		return err
	}

	category.ID = uint(id)

	return nil
}

func (r *CategoryRepositoryImpl) FindAll(ctx context.Context) ([]entity.Category, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []model.CategoryModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	categories := make([]entity.Category, 0, len(models))
	for _, m := range models {
		categories = append(categories, entity.Category{
			ID:        uint(m.ID),
			Name:      m.Name,
			Slug:      m.Slug,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.Category, error) {
	var model model.CategoryModel
	err := r.collection.FindOne(ctx, bson.M{"_id": int64(id)}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}

	category := entity.Category{
		ID:        uint(model.ID),
		Name:      model.Name,
		Slug:      model.Slug,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return &category, nil
}
