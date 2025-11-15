package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"courier_managment_system_go/internal/testutil"
)

type vehiclePayload struct {
	Brand        string  `json:"brand"`
	LicensePlate string  `json:"licensePlate"`
	MaxWeight    float64 `json:"maxWeight"`
	MaxVolume    float64 `json:"maxVolume"`
}

func TestVehicleController(t *testing.T) {
	t.Run("get all vehicles should return list of vehicles", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		rr := suite.Get("/vehicles", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) == 0 {
			t.Fatalf("expected vehicles, got %s", rr.Body.String())
		}
		if resp[0]["brand"] != "Ford Transit" {
			t.Fatalf("unexpected response: %v", resp[0])
		}
		if resp[0]["licensePlate"] != "А123БВ" {
			t.Fatalf("unexpected plate: %v", resp[0]["licensePlate"])
		}
	})

	t.Run("get all vehicles without auth should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/vehicles", "")
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create vehicle as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := vehiclePayload{
			Brand:        "Mercedes Sprinter",
			LicensePlate: "В456ГД",
			MaxWeight:    1500,
			MaxVolume:    20,
		}
		rr := suite.Post("/vehicles", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusCreated)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["brand"] != "Mercedes Sprinter" {
			t.Fatalf("unexpected response: %v", resp)
		}
	})

	t.Run("create vehicle as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := vehiclePayload{
			Brand:        "Mercedes Sprinter",
			LicensePlate: "В456ГД",
			MaxWeight:    1500,
			MaxVolume:    20,
		}
		rr := suite.Post("/vehicles", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create vehicle as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := vehiclePayload{
			Brand:        "Mercedes Sprinter",
			LicensePlate: "В456ГД",
			MaxWeight:    1500,
			MaxVolume:    20,
		}
		rr := suite.Post("/vehicles", payload, suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create vehicle with duplicate license plate should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		payload := vehiclePayload{
			Brand:        "Mercedes Sprinter",
			LicensePlate: "А123БВ",
			MaxWeight:    1500,
			MaxVolume:    20,
		}
		rr := suite.Post("/vehicles", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create vehicle with invalid data should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := vehiclePayload{
			Brand:        "",
			LicensePlate: "",
			MaxWeight:    -100,
			MaxVolume:    -10,
		}
		rr := suite.Post("/vehicles", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update vehicle as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		vehicle := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		payload := vehiclePayload{
			Brand:        "Updated Ford",
			LicensePlate: "Г789ЕЖ",
			MaxWeight:    2000,
			MaxVolume:    25,
		}
		rr := suite.Put("/vehicles/"+fmt.Sprintf("%d", vehicle.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})

	t.Run("update vehicle as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		vehicle := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		payload := vehiclePayload{
			Brand:        "Updated Ford",
			LicensePlate: "Г789ЕЖ",
			MaxWeight:    2000,
			MaxVolume:    25,
		}
		rr := suite.Put("/vehicles/"+fmt.Sprintf("%d", vehicle.ID), payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("update non-existent vehicle should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := vehiclePayload{
			Brand:        "Updated Ford",
			LicensePlate: "Г789ЕЖ",
			MaxWeight:    2000,
			MaxVolume:    25,
		}
		rr := suite.Put("/vehicles/999", payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update vehicle with duplicate license plate should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		first := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		second := suite.CreateVehicle("Mercedes", "В456ГД", 1500, 20)
		payload := vehiclePayload{
			Brand:        "Updated Mercedes",
			LicensePlate: first.LicensePlate,
			MaxWeight:    2000,
			MaxVolume:    25,
		}
		rr := suite.Put("/vehicles/"+fmt.Sprintf("%d", second.ID), payload, suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("delete vehicle as admin should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		vehicle := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		rr := suite.Delete("/vehicles/"+fmt.Sprintf("%d", vehicle.ID), suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusNoContent)
	})

	t.Run("delete vehicle as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		vehicle := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		rr := suite.Delete("/vehicles/"+fmt.Sprintf("%d", vehicle.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("delete non-existent vehicle should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Delete("/vehicles/999", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})
}
