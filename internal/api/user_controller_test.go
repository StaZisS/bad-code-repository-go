package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"courier_managment_system_go/internal/testutil"
)

type userPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type userUpdatePayload struct {
	Login    *string `json:"login,omitempty"`
	Name     *string `json:"name,omitempty"`
	Role     *string `json:"role,omitempty"`
	Password *string `json:"password,omitempty"`
}

func TestUserController(t *testing.T) {
	t.Run("get all users as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/users", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 3 {
			t.Fatalf("expected 3 users, got %d (%s)", len(resp), rr.Body.String())
		}
	})

	t.Run("get all users as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/users", suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get all users as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/users", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get all users without auth should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/users", "")
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get users filtered by role should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/users?role=courier", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 courier, got %d", len(resp))
		}
		if resp[0]["role"] != "courier" {
			t.Fatalf("expected courier role, got %v", resp[0]["role"])
		}
	})

	t.Run("create user as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "newcourier",
			Password: "password123",
			Name:     "Новый Курьер",
			Role:     "courier",
		}
		rr := suite.Post("/users", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusCreated)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["login"] != "newcourier" {
			t.Fatalf("unexpected response: %v", resp)
		}
	})

	t.Run("create user as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "newcourier",
			Password: "password123",
			Name:     "Новый Курьер",
			Role:     "courier",
		}
		rr := suite.Post("/users", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create user with duplicate login should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "admin",
			Password: "password123",
			Name:     "Другой Админ",
			Role:     "admin",
		}
		rr := suite.Post("/users", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create user with invalid data should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "",
			Password: "",
			Name:     "",
			Role:     "courier",
		}
		rr := suite.Post("/users", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create manager user should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "newmanager",
			Password: "password123",
			Name:     "Новый Менеджер",
			Role:     "manager",
		}
		rr := suite.Post("/users", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusCreated)
	})

	t.Run("update user as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		name := "Обновленное Имя"
		login := "updatedcourier"
		role := "manager"
		password := "newpassword"
		payload := userUpdatePayload{
			Name:     &name,
			Login:    &login,
			Role:     &role,
			Password: &password,
		}
		rr := suite.Put("/users/"+fmt.Sprintf("%d", suite.CourierUser.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["login"] != "updatedcourier" {
			t.Fatalf("unexpected response: %v", resp)
		}
		if resp["role"] != "manager" {
			t.Fatalf("unexpected role: %v", resp["role"])
		}
	})

	t.Run("update user as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		name := "Обновленное Имя"
		payload := userUpdatePayload{Name: &name}
		rr := suite.Put("/users/"+fmt.Sprintf("%d", suite.CourierUser.ID), payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("update user with duplicate login should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		login := "admin"
		payload := userUpdatePayload{Login: &login}
		rr := suite.Put("/users/"+fmt.Sprintf("%d", suite.CourierUser.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update non-existent user should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		name := "Обновленное Имя"
		payload := userUpdatePayload{Name: &name}
		rr := suite.Put("/users/999", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update user partial data should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		name := "Только Новое Имя"
		payload := userUpdatePayload{Name: &name}
		rr := suite.Put("/users/"+fmt.Sprintf("%d", suite.CourierUser.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["name"] != name {
			t.Fatalf("unexpected response: %v", resp)
		}
	})

	t.Run("delete user as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := userPayload{
			Login:    "todelete",
			Password: "password123",
			Name:     "Для Удаления",
			Role:     "courier",
		}
		createResp := suite.Post("/users", payload, suite.AdminToken)
		suite.ExpectStatus(createResp, http.StatusCreated)
		var created map[string]interface{}
		suite.Decode(createResp, &created)
		id := int64(created["id"].(float64))
		rr := suite.Delete("/users/"+fmt.Sprintf("%d", id), suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusNoContent)
	})

	t.Run("delete user as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Delete("/users/"+fmt.Sprintf("%d", suite.CourierUser.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("delete non-existent user should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Delete("/users/999", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})
}
