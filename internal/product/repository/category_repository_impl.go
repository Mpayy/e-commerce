package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Mpayy/e-commerce/internal/product/entity"
	"github.com/Mpayy/e-commerce/internal/product/model"
	"github.com/Mpayy/e-commerce/pkg/apperror"
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

	// disini untuk memperbaharui nilai id dari struct category berdasarkan nilai yang baru saja di insert
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
	var category entity.Category
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &category, nil
}
