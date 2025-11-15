package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"courier_managment_system_go/internal/testutil"
)

type deliveryPointProductPayload struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}

type deliveryPointPayload struct {
	Sequence  *int                          `json:"sequence,omitempty"`
	Latitude  float64                       `json:"latitude"`
	Longitude float64                       `json:"longitude"`
	Products  []deliveryPointProductPayload `json:"products"`
}

type deliveryPayload struct {
	CourierID    int64                  `json:"courierId"`
	VehicleID    int64                  `json:"vehicleId"`
	DeliveryDate string                 `json:"deliveryDate"`
	TimeStart    string                 `json:"timeStart"`
	TimeEnd      string                 `json:"timeEnd"`
	Points       []deliveryPointPayload `json:"points"`
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func ptrInt(v int) *int {
	value := v
	return &value
}

func defaultDeliveryPayload(suite *testutil.AppTestSuite, productID, vehicleID int64) deliveryPayload {
	return deliveryPayload{
		CourierID:    suite.CourierUser.ID,
		VehicleID:    vehicleID,
		DeliveryDate: formatDate(suite.FutureDate(5)),
		TimeStart:    "09:00",
		TimeEnd:      "18:00",
		Points: []deliveryPointPayload{
			{
				Sequence:  ptrInt(1),
				Latitude:  55.7558,
				Longitude: 37.6176,
				Products: []deliveryPointProductPayload{
					{ProductID: productID, Quantity: 3},
				},
			},
		},
	}
}

func TestDeliveryController(t *testing.T) {
	t.Run("get all deliveries as manager should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get("/deliveries", suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d (%s)", len(resp), rr.Body.String())
		}
	})

	t.Run("get all deliveries as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/deliveries", suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get all deliveries as admin should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		rr := suite.Get("/deliveries", suite.AdminToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("get deliveries with date filter should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		date := suite.FutureDate(5)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		suite.MustCreateDelivery(input)

		rr := suite.Get("/deliveries?date="+formatDate(date), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d", len(resp))
		}
	})

	t.Run("get deliveries with courier filter should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/deliveries?courier_id=%d", suite.CourierUser.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d", len(resp))
		}
	})

	t.Run("get deliveries with status filter should return filtered results", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		suite.MustCreateDelivery(input)

		rr := suite.Get("/deliveries?status=planned", suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp []map[string]interface{}
		suite.Decode(rr, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 delivery, got %d", len(resp))
		}
		if resp[0]["status"] != "planned" {
			t.Fatalf("unexpected status: %v", resp[0]["status"])
		}
	})

	t.Run("create delivery as manager should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := deliveryPayload{
			CourierID:    suite.CourierUser.ID,
			VehicleID:    vehicle.ID,
			DeliveryDate: formatDate(suite.FutureDate(5)),
			TimeStart:    "09:00",
			TimeEnd:      "18:00",
			Points: []deliveryPointPayload{
				{
					Sequence:  ptrInt(1),
					Latitude:  55.7558,
					Longitude: 37.6176,
					Products: []deliveryPointProductPayload{
						{ProductID: product.ID, Quantity: 3},
					},
				},
			},
		}
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusCreated)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["courier"].(map[string]interface{})["id"].(float64) != float64(suite.CourierUser.ID) {
			t.Fatalf("unexpected courier: %v", resp["courier"])
		}
	})

	t.Run("create delivery as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		rr := suite.Post("/deliveries", payload, suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create delivery with invalid courier role should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.CourierID = suite.AdminUser.ID
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create delivery with past date should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.DeliveryDate = formatDate(time.Now().AddDate(0, 0, -1))
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create delivery with invalid time should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.TimeStart = "18:00"
		payload.TimeEnd = "09:00"
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("get delivery by id should return delivery details", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		rr := suite.Get(fmt.Sprintf("/deliveries/%d", delivery.ID), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if int64(resp["id"].(float64)) != delivery.ID {
			t.Fatalf("unexpected delivery: %v", resp)
		}
	})

	t.Run("update delivery as manager should succeed when more than 3 days before", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.TimeStart = "10:00"
		payload.TimeEnd = "19:00"
		payload.Points[0].Latitude = 55.76
		payload.Points[0].Products[0].Quantity = 5
		rr := suite.Put(fmt.Sprintf("/deliveries/%d", delivery.ID), payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		if resp["timeStart"] != "10:00:00" {
			t.Fatalf("unexpected response: %v", resp)
		}
	})

	t.Run("update delivery less than 3 days before should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.DeliveryDate = formatDate(suite.FutureDate(1))
		create := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(create, http.StatusCreated)
		var created map[string]interface{}
		suite.Decode(create, &created)
		id := int64(created["id"].(float64))

		updatePayload := payload
		updatePayload.TimeStart = "10:00"
		rr := suite.Put(fmt.Sprintf("/deliveries/%d", id), updatePayload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("delete delivery as manager should succeed when more than 3 days before", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		create := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(create, http.StatusCreated)
		var created map[string]interface{}
		suite.Decode(create, &created)
		id := int64(created["id"].(float64))

		rr := suite.Delete(fmt.Sprintf("/deliveries/%d", id), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusNoContent)
	})

	t.Run("delete delivery less than 3 days before should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := defaultDeliveryPayload(suite, product.ID, vehicle.ID)
		payload.DeliveryDate = formatDate(suite.FutureDate(1))
		create := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(create, http.StatusCreated)
		var created map[string]interface{}
		suite.Decode(create, &created)
		id := int64(created["id"].(float64))

		rr := suite.Delete(fmt.Sprintf("/deliveries/%d", id), suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("generate deliveries as manager should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		suite.CreateVehicle("Ford", "GEN123", 1000, 15)
		payload := map[string]interface{}{
			"deliveryData": map[string]interface{}{
				formatDate(suite.FutureDate(5)): []map[string]interface{}{
					{
						"route": []map[string]interface{}{
							{"latitude": 55.7558, "longitude": 37.6176},
							{"latitude": 55.7600, "longitude": 37.6200},
						},
						"products": []map[string]interface{}{
							{"productId": product.ID, "quantity": 5},
						},
					},
				},
			},
		}
		rr := suite.Post("/deliveries/generate", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusOK)
	})

	t.Run("generate deliveries as courier should return 403", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		payload := map[string]interface{}{"deliveryData": map[string]interface{}{}}
		rr := suite.Post("/deliveries/generate", payload, suite.CourierToken)
		suite.ExpectStatus(rr, http.StatusForbidden)
	})

	t.Run("create delivery with insufficient time window should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := deliveryPayload{
			CourierID:    suite.CourierUser.ID,
			VehicleID:    vehicle.ID,
			DeliveryDate: formatDate(suite.FutureDate(5)),
			TimeStart:    "09:00",
			TimeEnd:      "09:30",
			Points: []deliveryPointPayload{
				{
					Sequence:  ptrInt(1),
					Latitude:  55.7558,
					Longitude: 37.6176,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
				{
					Sequence:  ptrInt(2),
					Latitude:  59.9311,
					Longitude: 30.3609,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
			},
		}
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("update delivery with insufficient time window should return 400", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		delivery := suite.MustCreateDelivery(input)

		payload := deliveryPayload{
			CourierID:    suite.CourierUser.ID,
			VehicleID:    vehicle.ID,
			DeliveryDate: formatDate(suite.FutureDate(6)),
			TimeStart:    "10:00",
			TimeEnd:      "10:30",
			Points: []deliveryPointPayload{
				{
					Sequence:  ptrInt(1),
					Latitude:  55.7558,
					Longitude: 37.6176,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
				{
					Sequence:  ptrInt(2),
					Latitude:  59.9311,
					Longitude: 30.3609,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
			},
		}
		rr := suite.Put(fmt.Sprintf("/deliveries/%d", delivery.ID), payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create delivery with sufficient time window should succeed", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		product := suite.CreateProduct("Товар", 1.0, 10, 10, 10)
		vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)
		payload := deliveryPayload{
			CourierID:    suite.CourierUser.ID,
			VehicleID:    vehicle.ID,
			DeliveryDate: formatDate(suite.FutureDate(5)),
			TimeStart:    "09:00",
			TimeEnd:      "18:00",
			Points: []deliveryPointPayload{
				{
					Sequence:  ptrInt(1),
					Latitude:  55.7558,
					Longitude: 37.6176,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
				{
					Sequence:  ptrInt(2),
					Latitude:  55.7600,
					Longitude: 37.6200,
					Products:  []deliveryPointProductPayload{{ProductID: product.ID, Quantity: 1}},
				},
			},
		}
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusCreated)
		var resp map[string]interface{}
		suite.Decode(rr, &resp)
		points := resp["deliveryPoints"].([]interface{})
		if len(points) != 2 {
			t.Fatalf("expected two points, got %d", len(points))
		}
	})

	t.Run("create delivery should fail when vehicle weight capacity exceeded", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		heavy := suite.CreateProduct("Тяжелый товар", 600, 100, 100, 100)
		vehicle := suite.CreateVehicle("Small", "SMALL123", 1000, 10)
		payload := defaultDeliveryPayload(suite, heavy.ID, vehicle.ID)
		payload.Points[0].Products[0].Quantity = 3
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})

	t.Run("create delivery should fail when vehicle volume capacity exceeded", func(t *testing.T) {
		suite := testutil.NewAppTestSuite(t)
		bulky := suite.CreateProduct("Объемный товар", 10, 200, 200, 200)
		vehicle := suite.CreateVehicle("Small", "SMALL789", 2000, 10)
		payload := defaultDeliveryPayload(suite, bulky.ID, vehicle.ID)
		payload.Points[0].Products[0].Quantity = 2
		rr := suite.Post("/deliveries", payload, suite.ManagerToken)
		suite.ExpectStatus(rr, http.StatusBadRequest)
	})
}
