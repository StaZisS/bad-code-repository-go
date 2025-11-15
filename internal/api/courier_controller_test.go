package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/service"
	"courier_managment_system_go/internal/testutil"
)

func TestCourierController(t *testing.T) {
	t.Run("get courier deliveries should return own deliveries", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.Points[0].Products[0].Quantity = 2
		suite.MustCreateDelivery(input)

		rr := suite.Get("/courier/deliveries", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d (%s)", len(resp), rr.Body.String())
		}
		if resp[0]["pointsCount"].(float64) != 1 || resp[0]["productsCount"].(float64) != 2 {
			t.Fatalf("unexpected summary: %v", resp[0])
		}
	})

	t.Run("get courier deliveries as admin should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/courier/deliveries", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get courier deliveries as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/courier/deliveries", suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get courier deliveries without auth should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/courier/deliveries", "")
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get courier deliveries with date filter should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		date := suite.FutureDate(5)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		suite.MustCreateDelivery(input)

		rr := suite.Get("/courier/deliveries?date="+date.Format("2006-01-02"), suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d", len(resp))
		}
	})

	t.Run("get courier deliveries with non-matching date filter should return empty", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		date := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
		rr := suite.Get("/courier/deliveries?date="+date, suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 0 {
			t.Fatalf("expected empty response, got %v", resp)
		}
	})

	t.Run("get courier deliveries with status filter should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get("/courier/deliveries?status=planned", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})

	t.Run("get courier deliveries with date range should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		date := suite.FutureDate(5)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		suite.MustCreateDelivery(input)

		from := date.AddDate(0, 0, -1).Format("2006-01-02")
		to := date.AddDate(0, 0, 1).Format("2006-01-02")
		rr := suite.Get(fmt.Sprintf("/courier/deliveries?date_from=%s&date_to=%s", from, to), suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d", len(resp))
		}
	})

	t.Run("get courier deliveries should not return other courier deliveries", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		other := suite.CreateProduct("Товар", 1, 10, 10, 10)
		another := suite.CreateVehicle("Mercedes", "В456ГД", 1000, 15)
		newCourier, err := suite.UserService.CreateUser(service.CreateUserInput{
			Login:    "othercourier",
			Password: "password",
			Name:     "Другой Курьер",
			Role:     domain.RoleCourier,
		})
		if err != nil {
			t.Fatalf("failed to create other courier: %v", err)
		}
		user, _ := suite.Store.GetUserByLogin(newCourier.Login)
		input := suite.DeliveryInput(other.ID, another.ID, user.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get("/courier/deliveries", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 0 {
			t.Fatalf("expected empty response, got %v", resp)
		}
	})

	t.Run("get courier delivery by id should return delivery details", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/courier/deliveries/%d", delivery.ID), suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})

	t.Run("get other courier delivery by id should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		newCourier, err := suite.UserService.CreateUser(service.CreateUserInput{
			Login:    "othercourier2",
			Password: "password",
			Name:     "Другой Курьер 2",
			Role:     domain.RoleCourier,
		})
		if err != nil {
			t.Fatalf("failed to create other courier: %v", err)
		}
		user, _ := suite.Store.GetUserByLogin(newCourier.Login)
		input := suite.DeliveryInput(product.ID, vehicle.ID, user.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/courier/deliveries/%d", delivery.ID), suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("get non-existent delivery should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/courier/deliveries/999", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("get courier delivery as admin should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/courier/deliveries/%d", delivery.ID), suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get courier delivery as manager should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/courier/deliveries/%d", delivery.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("courier should see correct vehicle information", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford Transit", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get("/courier/deliveries", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		vehicleInfo := resp[0]["vehicle"].(map[string]interface{})
		if vehicleInfo["brand"] != "Ford Transit" {
			t.Fatalf("unexpected vehicle: %v", vehicleInfo)
		}
	})

	t.Run("courier should see delivery with no vehicle assigned", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1, 10, 10, 10)
		delivery := domain.Delivery{
			CourierID:    suite.CourierUser.ID,
			VehicleID:    0,
			CreatedByID:  suite.ManagerUser.ID,
			DeliveryDate: suite.FutureDate(5),
			TimeStart:    suite.TimeOfDay(9, 0),
			TimeEnd:      suite.TimeOfDay(18, 0),
			Status:       domain.StatusPlanned,
			DeliveryPoints: []domain.DeliveryPoint{
				{
					Sequence:  1,
					Latitude:  55.7558,
					Longitude: 37.6176,
					Products: []domain.DeliveryPointProduct{
						{ProductID: product.ID, Quantity: 1},
					},
				},
			},
		}
		if _, err := suite.Store.CreateDelivery(delivery); err != nil {
			t.Fatalf("failed to create delivery: %v", err)
		}

		rr := suite.Get("/courier/deliveries", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		vehicleInfo := resp[0]["vehicle"].(map[string]interface{})
		if vehicleInfo["brand"] != "Не назначена" {
			t.Fatalf("unexpected vehicle info: %v", vehicleInfo)
		}
	})
}
