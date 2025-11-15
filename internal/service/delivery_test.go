package service

import (
	"strconv"
	"testing"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

func TestDeliveryServiceValidatesCapacity(t *testing.T) {
	store := storage.NewMemoryStore()
	courier := createUser(t, store, "courier", domain.RoleCourier)
	manager := createUser(t, store, "manager", domain.RoleManager)
	product := createProduct(t, store, "heavy", 70)
	vehicle := createVehicle(t, store, 100, 5)

	svc := NewDeliveryService(store, NewRouteService(40))
	input := DeliveryInput{
		CourierID:    courier.ID,
		VehicleID:    vehicle.ID,
		DeliveryDate: time.Now().AddDate(0, 0, 5),
		TimeStart:    parseTime(t, "09:00"),
		TimeEnd:      parseTime(t, "12:00"),
		Points: []DeliveryPointInput{
			{
				Sequence:  1,
				Latitude:  55.7558,
				Longitude: 37.6176,
				Products: []DeliveryPointProductInput{
					{ProductID: product.ID, Quantity: 1},
				},
			},
			{
				Sequence:  2,
				Latitude:  55.7570,
				Longitude: 37.6200,
				Products: []DeliveryPointProductInput{
					{ProductID: product.ID, Quantity: 1},
				},
			},
		},
	}

	_, err := svc.CreateDelivery(input, manager.ID)
	if err == nil {
		t.Fatal("expected capacity validation error")
	}
	if _, ok := err.(ValidationErrors); !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestDeliveryServiceValidatesRouteDuration(t *testing.T) {
	store := storage.NewMemoryStore()
	courier := createUser(t, store, "courier2", domain.RoleCourier)
	manager := createUser(t, store, "manager2", domain.RoleManager)
	product := createProduct(t, store, "light", 10)
	vehicle := createVehicle(t, store, 1000, 20)

	svc := NewDeliveryService(store, NewRouteService(50))
	input := DeliveryInput{
		CourierID:    courier.ID,
		VehicleID:    vehicle.ID,
		DeliveryDate: time.Now().AddDate(0, 0, 5),
		TimeStart:    parseTime(t, "09:00"),
		TimeEnd:      parseTime(t, "09:30"),
		Points: []DeliveryPointInput{
			{
				Sequence:  1,
				Latitude:  55.7558,
				Longitude: 37.6176,
				Products: []DeliveryPointProductInput{
					{ProductID: product.ID, Quantity: 1},
				},
			},
			{
				Sequence:  2,
				Latitude:  59.9311,
				Longitude: 30.3609,
				Products: []DeliveryPointProductInput{
					{ProductID: product.ID, Quantity: 1},
				},
			},
		},
	}

	_, err := svc.CreateDelivery(input, manager.ID)
	if err == nil {
		t.Fatal("expected duration validation error")
	}
	if _, ok := err.(ValidationErrors); !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func parseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	return parsed
}

func createUser(t *testing.T, store *storage.MemoryStore, login string, role domain.UserRole) domain.User {
	t.Helper()
	user, err := store.CreateUser(domain.User{
		Login:        login,
		PasswordHash: "hash",
		Name:         login,
		Role:         role,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func createProduct(t *testing.T, store *storage.MemoryStore, name string, weight float64) domain.Product {
	t.Helper()
	product, err := store.CreateProduct(domain.Product{
		Name:   name,
		Weight: weight,
		Length: 10,
		Width:  10,
		Height: 10,
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	return product
}

func createVehicle(t *testing.T, store *storage.MemoryStore, maxWeight, maxVolume float64) domain.Vehicle {
	t.Helper()
	vehicle, err := store.CreateVehicle(domain.Vehicle{
		Brand:        "Test",
		LicensePlate: strconv.FormatInt(time.Now().UnixNano(), 10),
		MaxWeight:    maxWeight,
		MaxVolume:    maxVolume,
	})
	if err != nil {
		t.Fatalf("failed to create vehicle: %v", err)
	}
	return vehicle
}
