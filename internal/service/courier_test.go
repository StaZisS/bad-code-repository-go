package service

import (
	"testing"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

func TestCourierServiceFiltersByDate(t *testing.T) {
	store := storage.NewMemoryStore()
	courier := createUser(t, store, "courier-test", domain.RoleCourier)
	manager := createUser(t, store, "manager-test", domain.RoleManager)
	product := createProduct(t, store, "box", 5)
	vehicle := createVehicle(t, store, 500, 10)

	deliveryService := NewDeliveryService(store, NewRouteService(40))
	date := time.Now().UTC().AddDate(0, 0, 5)

	input := DeliveryInput{
		CourierID:    courier.ID,
		VehicleID:    vehicle.ID,
		DeliveryDate: date,
		TimeStart:    parseTime(t, "08:00"),
		TimeEnd:      parseTime(t, "12:00"),
		Points: []DeliveryPointInput{
			{
				Sequence:  1,
				Latitude:  55.75,
				Longitude: 37.61,
				Products:  []DeliveryPointProductInput{{ProductID: product.ID, Quantity: 1}},
			},
			{
				Sequence:  2,
				Latitude:  55.76,
				Longitude: 37.62,
				Products:  []DeliveryPointProductInput{{ProductID: product.ID, Quantity: 2}},
			},
		},
	}
	if _, err := deliveryService.CreateDelivery(input, manager.ID); err != nil {
		t.Fatalf("failed to create delivery: %v", err)
	}

	// Another delivery on different date should not appear when filtering by date
	input.DeliveryDate = date.AddDate(0, 0, 1)
	if _, err := deliveryService.CreateDelivery(input, manager.ID); err != nil {
		t.Fatalf("failed to create delivery: %v", err)
	}

	service := NewCourierService(store)
	filterDate := date
	deliveries, err := service.GetCourierDeliveries(courier.ID, &filterDate, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to get courier deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected 1 delivery for date %s, got %d", filterDate, len(deliveries))
	}
	if deliveries[0].DeliveryDate != filterDate.Format("2006-01-02") {
		t.Fatalf("unexpected delivery date: %s", deliveries[0].DeliveryDate)
	}
}
