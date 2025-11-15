package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"courier_managment_system_go/internal/testutil"
)

type productPayload struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func TestProductController(t *testing.T) {
	t.Run("get all products should return list of products", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)

		rr := suite.Get("/products", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) == 0 {
			t.Fatalf("expected products, got %s", rr.Body.String())
		}
		if resp[0]["name"] != "Тестовый товар" {
			t.Fatalf("expected name, got %v", resp[0]["name"])
		}
		if resp[0]["weight"].(float64) != 1.5 {
			t.Fatalf("expected weight 1.5, got %v", resp[0]["weight"])
		}
	})

	t.Run("get all products without auth should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/products", "")
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create product as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := productPayload{
			Name:   "Новый товар",
			Weight: 2.5,
			Length: 15.0,
			Width:  12.0,
			Height: 8.0,
		}
		rr := suite.Post("/products", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusCreated)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["name"] != "Новый товар" {
			t.Fatalf("unexpected product: %v", resp)
		}
	})

	t.Run("create product as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := productPayload{
			Name:   "Новый товар",
			Weight: 2.5,
			Length: 15.0,
			Width:  12.0,
			Height: 8.0,
		}
		rr := suite.Post("/products", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create product as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := productPayload{
			Name:   "Новый товар",
			Weight: 2.5,
			Length: 15.0,
			Width:  12.0,
			Height: 8.0,
		}
		rr := suite.Post("/products", payload, suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create product with invalid data should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := productPayload{
			Name:   "",
			Weight: -1.0,
			Length: 0.0,
			Width:  -5.0,
			Height: 0.0,
		}
		rr := suite.Post("/products", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update product as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		payload := productPayload{
			Name:   "Обновленный товар",
			Weight: 3.0,
			Length: 20.0,
			Width:  15.0,
			Height: 10.0,
		}
		rr := suite.Put("/products/"+fmt.Sprintf("%d", product.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["name"] != "Обновленный товар" {
			t.Fatalf("unexpected response: %v", resp)
		}
	})

	t.Run("update product as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		payload := productPayload{
			Name:   "Обновленный товар",
			Weight: 3.0,
			Length: 20.0,
			Width:  15.0,
			Height: 10.0,
		}
		rr := suite.Put("/products/"+fmt.Sprintf("%d", product.ID), payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("update non-existent product should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := productPayload{
			Name:   "Обновленный товар",
			Weight: 3.0,
			Length: 20.0,
			Width:  15.0,
			Height: 10.0,
		}
		rr := suite.Put("/products/999", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("delete product as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		rr := suite.Delete("/products/"+fmt.Sprintf("%d", product.ID), suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusNoContent)
	})

	t.Run("delete product as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		rr := suite.Delete("/products/"+fmt.Sprintf("%d", product.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("delete product as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		rr := suite.Delete("/products/"+fmt.Sprintf("%d", product.ID), suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("delete non-existent product should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Delete("/products/999", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("get products with manager token should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		rr := suite.Get("/products", suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})

	t.Run("get products with courier token should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		suite.CreateProduct("Тестовый товар", 1.5, 10, 10, 10)
		rr := suite.Get("/products", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})
}
