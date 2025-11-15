package service_test

import (
	"math"
	"testing"

	"courier_managment_system_go/internal/service"
	"courier_managment_system_go/internal/testutil"
)

func TestRouteServiceCalculations(t *testing.T) {
	suite := testutil.NewAppTestSuite(t)
	rs := suite.RouteService

	t.Run("long distance route should exceed 600km", func(t *testing.T) {
		result, err := rs.Calculate([]service.RoutePoint{
			{Latitude: 55.7558, Longitude: 37.6176},
			{Latitude: 59.9311, Longitude: 30.3609},
		})
		if err != nil {
			t.Fatalf("calculate failed: %v", err)
		}
		if math.Abs(result.DistanceKm-635) > 20 {
			t.Fatalf("unexpected distance: %v", result.DistanceKm)
		}
	})

	t.Run("short distance route should be small", func(t *testing.T) {
		result, err := rs.Calculate([]service.RoutePoint{
			{Latitude: 55.7558, Longitude: 37.6176},
			{Latitude: 55.7600, Longitude: 37.6200},
		})
		if err != nil {
			t.Fatalf("calculate failed: %v", err)
		}
		if result.DistanceKm <= 0 {
			t.Fatalf("expected positive distance, got %v", result.DistanceKm)
		}
	})
}

func TestDeliveryTimeValidation(t *testing.T) {
	suite := testutil.NewAppTestSuite(t)
	product := suite.CreateProduct("Товар", 10, 10, 10, 10)
	vehicle := suite.CreateVehicle("Ford", "А123БВ", 1000, 15)

	t.Run("delivery validation should pass for short route with sufficient time", func(t *testing.T) {
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.Points = []service.DeliveryPointInput{
			{Sequence: 1, Latitude: 55.7558, Longitude: 37.6176, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
			{Sequence: 2, Latitude: 55.7600, Longitude: 37.6200, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
		}
		if _, err := suite.DeliveryService.CreateDelivery(input, suite.ManagerUser.ID); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("delivery validation should fail for long route with insufficient time", func(t *testing.T) {
		input := suite.DeliveryInput(product.ID, vehicle.ID, suite.CourierUser.ID, suite.FutureDate(5))
		input.TimeEnd = suite.TimeOfDay(9, 30)
		input.Points = []service.DeliveryPointInput{
			{Sequence: 1, Latitude: 55.7558, Longitude: 37.6176, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
			{Sequence: 2, Latitude: 59.9311, Longitude: 30.3609, Products: []service.DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}}},
		}
		if _, err := suite.DeliveryService.CreateDelivery(input, suite.ManagerUser.ID); err == nil {
			t.Fatalf("expected validation error for insufficient time")
		}
	})
}
