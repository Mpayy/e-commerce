//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/model"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
)

var testDB *mongo.Database

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:8", mongodb.WithReplicaSet("rs0"))
	if err != nil {
		log.Fatalf("failed to start mongodb container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	clientOpts := options.Client().
		ApplyURI(connStr).
		SetDirect(true)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	testDB = client.Database("test_db")
	if err := EnsureIndexes(ctx, testDB); err != nil {
		log.Fatalf("failed to ensure indexes: %v", err)
	}

	code := m.Run()

	client.Disconnect(context.Background())
	testcontainers.TerminateContainer(container)

	os.Exit(code)
}

func seedProduct(t *testing.T, db *mongo.Database, stock int) uint {
	ctx := context.Background()
	id, err := getNextSequence(ctx, db, "product_id")
	if err != nil {
		t.Fatalf("failed to get next sequence: %v", err)
	}

	productID := uint(id)

	product := model.ProductModel{
		ID:    id,
		Slug:  fmt.Sprintf("product-%d", productID),
		Stock: stock,
		SKU:   fmt.Sprintf("SKU-%d", productID),
	}

	_, err = db.Collection(product.CollectionName()).InsertOne(ctx, product)
	if err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	return productID
}

func fetchProduct(t *testing.T, db *mongo.Database, productID int64) model.ProductModel {
	ctx := context.Background()
	var product model.ProductModel

	err := db.Collection(product.CollectionName()).FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		t.Fatalf("failed to fetch product: %v", err)
	}
	return product
}

func TestDecreaseStockBulk_PartialFailureRollsBackAll(t *testing.T) {
	ctx := context.Background()
	repo := NewProductRepository(testDB)

	id1 := seedProduct(t, testDB, 10)
	id2 := seedProduct(t, testDB, 2)

	t.Cleanup(func() {
		filter := bson.M{
			"_id": bson.M{
				"$in": []int64{int64(id1), int64(id2)},
			},
		}
		_, err := testDB.Collection("products").DeleteMany(ctx, filter)
		if err != nil {
			t.Logf("cleanup failed: failed to delete test products: %v", err)
		}
	})

	items := []entity.BulkDecreaseStock{
		{ProductID: id1, Quantity: 5},
		{ProductID: id2, Quantity: 5},
	}

	err := repo.BulkDecreaseStock(ctx, "checkout-test-1", items)
	assert.ErrorIs(t, err, apperror.ErrInsufficientStock)

	product1 := fetchProduct(t, testDB, int64(id1))
	assert.Equal(t, 10, product1.Stock, "product 1 stock must remain intact, not included in the deduction")
}

func TestRestoreStockBulk_CalledTwice_OnlyRestoresOnce(t *testing.T) {
	repo := NewProductRepository(testDB)
	ctx := context.Background()

	id := seedProduct(t, testDB, 10)
	checkoutID := "checkout-test-2"

	t.Cleanup(func() {
		filter := bson.M{
			"_id": int64(id),
		}
		_, err := testDB.Collection("products").DeleteOne(ctx, filter)
		if err != nil {
			t.Logf("cleanup failed: failed to delete test products: %v", err)
		}
	})

	err := repo.BulkDecreaseStock(ctx, checkoutID, []entity.BulkDecreaseStock{{ProductID: id, Quantity: 3}})
	assert.NoError(t, err)

	err1 := repo.BulkRestoreStock(ctx, checkoutID)
	err2 := repo.BulkRestoreStock(ctx, checkoutID)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	product := fetchProduct(t, testDB, int64(id))
	assert.Equal(t, 10, product.Stock, "product 1 stock must back to 10")
}

func TestDecreaseStockBulk_ConcurrentOverlappingCheckouts(t *testing.T) {
	ctx := context.Background()
	repo := NewProductRepository(testDB)

	id1 := seedProduct(t, testDB, 10)
	id2 := seedProduct(t, testDB, 10)

	t.Cleanup(func() {
		filter := bson.M{
			"_id": bson.M{
				"$in": []int64{int64(id1), int64(id2)},
			},
		}
		_, err := testDB.Collection("products").DeleteMany(ctx, filter)
		if err != nil {
			t.Logf("cleanup failed: failed to delete test products: %v", err)
		}
	})

	startSignal := make(chan struct{})

	var g errgroup.Group

	g.Go(func() error {
		<-startSignal
		items := []entity.BulkDecreaseStock{
			{ProductID: id1, Quantity: 3},
			{ProductID: id2, Quantity: 4},
		}
		err := repo.BulkDecreaseStock(ctx, "checkout-A", items)
		return err
	})

	g.Go(func() error {
		<-startSignal
		items := []entity.BulkDecreaseStock{
			{ProductID: id2, Quantity: 2},
			{ProductID: id1, Quantity: 1},
		}
		err := repo.BulkDecreaseStock(ctx, "checkout-B", items)
		return err
	})

	close(startSignal)

	err := g.Wait()
	assert.NoError(t, err)

	p1 := fetchProduct(t, testDB, int64(id1))
	p2 := fetchProduct(t, testDB, int64(id2))

	assert.Equal(t, 6, p1.Stock, "produk 1 stok harus berkurang total 4 (10 - 3 - 1)")
	assert.Equal(t, 4, p2.Stock, "produk 2 stok harus berkurang total 6 (10 - 4 - 2)")
}
