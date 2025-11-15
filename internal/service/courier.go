package service

import (
	"errors"
	"fmt"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

var ErrNotCourier = errors.New("user is not a courier")

type CourierService struct {
	store *storage.MemoryStore
}

func NewCourierService(store *storage.MemoryStore) *CourierService {
	return &CourierService{store: store}
}

func (s *CourierService) GetCourierDeliveries(
	courierID int64,
	date *time.Time,
	status *domain.DeliveryStatus,
	dateFrom *time.Time,
	dateTo *time.Time,
) ([]CourierDeliveryResponse, error) {
	user, ok := s.store.GetUser(courierID)
	if !ok {
		return nil, ErrNotFound
	}
	if user.Role != domain.RoleCourier {
		return nil, ErrNotCourier
	}
	deliveries := s.store.ListDeliveries()
	var responses []CourierDeliveryResponse
	for _, delivery := range deliveries {
		if delivery.CourierID != courierID {
			continue
		}
		if date != nil && !sameDay(delivery.DeliveryDate, *date) {
			continue
		}
		if status != nil && delivery.Status != *status {
			continue
		}
		if dateFrom != nil && delivery.DeliveryDate.Before(*dateFrom) {
			continue
		}
		if dateTo != nil && delivery.DeliveryDate.After(*dateTo) {
			continue
		}
		responses = append(responses, s.courierDTO(delivery))
	}
	return responses, nil
}

func (s *CourierService) GetCourierDeliveryByID(courierID, deliveryID int64) (DeliveryDTO, error) {
	delivery, ok := s.store.GetDelivery(deliveryID)
	if !ok {
		return DeliveryDTO{}, ErrNotFound
	}
	if delivery.CourierID != courierID {
		return DeliveryDTO{}, ValidationErrors{{Field: "delivery", Message: "доставка недоступна"}}
	}
	users, vehicles, products := s.snapshot()
	return NewDeliveryDTO(*delivery, users, vehicles, products), nil
}

func (s *CourierService) courierDTO(delivery domain.Delivery) CourierDeliveryResponse {
	vehicle, _ := s.store.GetVehicle(delivery.VehicleID)
	vehicleSummary := VehicleSummary{
		Brand:        "Не назначена",
		LicensePlate: "",
	}
	if vehicle != nil {
		vehicleSummary.Brand = vehicle.Brand
		vehicleSummary.LicensePlate = vehicle.LicensePlate
	}
	var productsCount int
	var totalWeight float64
	for _, point := range delivery.DeliveryPoints {
		for _, product := range point.Products {
			entity, _ := s.store.GetProduct(product.ProductID)
			productsCount += product.Quantity
			totalWeight += entity.Weight * float64(product.Quantity)
		}
	}
	return CourierDeliveryResponse{
		ID:             delivery.ID,
		DeliveryNumber: timeDeliveryNumber(delivery),
		DeliveryDate:   delivery.DeliveryDate.Format("2006-01-02"),
		TimeStart:      delivery.TimeStart.Format("15:04:05"),
		TimeEnd:        delivery.TimeEnd.Format("15:04:05"),
		Status:         delivery.Status,
		Vehicle:        vehicleSummary,
		PointsCount:    len(delivery.DeliveryPoints),
		ProductsCount:  productsCount,
		TotalWeight:    totalWeight,
	}
}

func (s *CourierService) snapshot() (map[int64]domain.User, map[int64]domain.Vehicle, map[int64]domain.Product) {
	usersMap := make(map[int64]domain.User)
	for _, user := range s.store.ListUsers(nil) {
		usersMap[user.ID] = user
	}
	vehicleMap := make(map[int64]domain.Vehicle)
	for _, vehicle := range s.store.ListVehicles() {
		vehicleMap[vehicle.ID] = vehicle
	}
	productMap := make(map[int64]domain.Product)
	for _, product := range s.store.ListProducts() {
		productMap[product.ID] = product
	}
	return usersMap, vehicleMap, productMap
}

func timeDeliveryNumber(delivery domain.Delivery) string {
	return fmt.Sprintf("DEL-%d-%03d", delivery.DeliveryDate.Year(), delivery.ID)
}
