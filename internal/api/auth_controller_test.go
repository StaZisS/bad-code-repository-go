package api_test

import (
	"net/http"
	"testing"

	"courier_managment_system_go/internal/testutil"
)

func TestAuthController(t *testing.T) {
	t.Run("login with valid credentials should return token and user info", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "admin",
			"password": "admin123",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusOK)

		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if _, ok := resp["token"].(string); !ok {
			t.Fatalf("expected token in response: %s", rr.Body.String())
		}
		user := resp["user"].(map[string]interface{})
		if int64(user["id"].(float64)) != suite.AdminUser.ID {
			t.Fatalf("expected admin id, got %v", user["id"])
		}
		if user["login"] != "admin" {
			t.Fatalf("expected login admin, got %v", user["login"])
		}
		if user["name"] != "Системный администратор" {
			t.Fatalf("expected admin name, got %v", user["name"])
		}
		if user["role"] != "admin" {
			t.Fatalf("expected role admin, got %v", user["role"])
		}
	})

	t.Run("login with invalid login should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "nonexistent",
			"password": "password",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("login with invalid password should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "admin",
			"password": "wrong",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("login with empty login should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "",
			"password": "admin123",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("login with empty password should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "admin",
			"password": "",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("login with manager credentials should return manager token", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "manager",
			"password": "password",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusOK)

		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		user := resp["user"].(map[string]interface{})
		if user["role"] != "manager" {
			t.Fatalf("expected role manager, got %v", user["role"])
		}
	})

	t.Run("login with courier credentials should return courier token", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		body := map[string]string{
			"login":    "courier",
			"password": "password",
		}
		rr := suite.Post("/auth/login", body, "")
		suite.ExpectStatus(rr, http.StatusOK)

		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		user := resp["user"].(map[string]interface{})
		if user["role"] != "courier" {
			t.Fatalf("expected role courier, got %v", user["role"])
		}
	})
}
