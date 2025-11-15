package service_test

import (
	"fmt"
	"testing"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/service"
	"courier_managment_system_go/internal/testutil"
)

func TestVehicleCapacityValidation(t *testing.T) {
	suite := testutil.NewAppTestSuite(t)

	createHeavyProduct := func(weight float64) service.ProductDTO {
		return suite.CreateProduct("Тяжелый товар", weight, 100, 100, 100)
	}
	createBulkyProduct := func(side float64) service.ProductDTO {
		return suite.CreateProduct("Объемный товар", 10, side, side, side)
	}
	createSmallVehicle := func() service.VehicleDTO {
		return suite.CreateVehicle("Small Truck", fmt.Sprintf("SMALL-%d", time.Now().UnixNano()), 1000, 10)
	}

	t.Run("delivery should succeed when vehicle has sufficient capacity", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := suite.CreateProduct("Легкий товар", 1.5, 10, 10, 10)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.Points[0].Products[0].Quantity = 10
		if _, err := suite.DeliveryService.CreateDelivery(input, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("delivery should fail when exceeding vehicle weight capacity", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createHeavyProduct(600)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.Points[0].Products[0].Quantity = 3
		if _, err := suite.DeliveryService.CreateDelivery(input, suite.ManagerUser.ID); err == nil {
			t.Fatalf("expected failure due to weight")
		}
	})

	t.Run("delivery should fail when exceeding vehicle volume capacity", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createBulkyProduct(200)
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.Points[0].Products[0].Quantity = 2
		if _, err := suite.DeliveryService.CreateDelivery(input, suite.ManagerUser.ID); err == nil {
			t.Fatalf("expected failure due to volume")
		}
	})

	t.Run("delivery should fail when combined with existing deliveries exceeding capacity", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createHeavyProduct(600)
		date := suite.FutureDate(5)
		first := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		first.TimeEnd = suite.TimeOfDay(12, 0)
		first.Points[0].Products[0].Quantity = 1
		if _, err := suite.DeliveryService.CreateDelivery(first, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected first delivery success: %v", err)
		}
		second := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		second.TimeStart = suite.TimeOfDay(13, 0)
		second.TimeEnd = suite.TimeOfDay(16, 0)
		second.Points = []service.DeliveryPointInput{
			{Sequence: 1, Latitude: 55.7600, Longitude: 37.6200, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
			{Sequence: 2, Latitude: 55.7700, Longitude: 37.6300, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
		}
		if _, err := suite.DeliveryService.CreateDelivery(second, suite.ManagerUser.ID); err == nil {
			t.Fatalf("expected failure due to total weight")
		}
	})

	t.Run("delivery should succeed when time periods do not overlap", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createHeavyProduct(500)
		date := suite.FutureDate(5)
		first := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		first.TimeEnd = suite.TimeOfDay(12, 0)
		if _, err := suite.DeliveryService.CreateDelivery(first, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success: %v", err)
		}
		second := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		second.TimeStart = suite.TimeOfDay(13, 0)
		second.TimeEnd = suite.TimeOfDay(16, 0)
		if _, err := suite.DeliveryService.CreateDelivery(second, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success: %v", err)
		}
	})

	t.Run("delivery should fail when time periods overlap and exceed capacity", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createHeavyProduct(500)
		date := suite.FutureDate(5)
		first := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		first.TimeEnd = suite.TimeOfDay(13, 0)
		if _, err := suite.DeliveryService.CreateDelivery(first, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success: %v", err)
		}
		second := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		second.TimeStart = suite.TimeOfDay(12, 0)
		second.TimeEnd = suite.TimeOfDay(16, 0)
		second.Points = []service.DeliveryPointInput{
			{Sequence: 1, Latitude: 55.7600, Longitude: 37.6200, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
			{Sequence: 2, Latitude: 55.7700, Longitude: 37.6300, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
		}
		if _, err := suite.DeliveryService.CreateDelivery(second, suite.ManagerUser.ID); err == nil {
			t.Fatalf("expected failure due to overlapping capacity")
		}
	})

	t.Run("completed deliveries should not affect capacity validation", func(t *testing.T) {
		vehicle := createSmallVehicle()
		product := createHeavyProduct(500)
		date := suite.FutureDate(5)
		first := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		first.TimeEnd = suite.TimeOfDay(13, 0)
		dto, err := suite.DeliveryService.CreateDelivery(first, suite.ManagerUser.ID)
		if err != nil {
			t.Fatalf("expected success: %v", err)
		}
		if err := suite.Store.SetDeliveryStatus(dto.ID, domain.StatusCompleted); err != nil {
			t.Fatalf("failed to mark completed: %v", err)
		}
		second := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, date)
		second.TimeStart = suite.TimeOfDay(12, 0)
		second.TimeEnd = suite.TimeOfDay(16, 0)
		if _, err := suite.DeliveryService.CreateDelivery(second, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success: %v", err)
		}
	})
}
