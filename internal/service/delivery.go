package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

type DeliveryService struct {
	store *storage.MemoryStore
	route *RouteService
}

func NewDeliveryService(store *storage.MemoryStore, route *RouteService) *DeliveryService {
	return &DeliveryService{
		store: store,
		route: route,
	}
}

type DeliveryPointProductInput struct {
	ProductID int64
	Quantity  int
}

type DeliveryPointInput struct {
	Sequence  int
	Latitude  float64
	Longitude float64
	Products  []DeliveryPointProductInput
}

type DeliveryInput struct {
	CourierID    int64
	VehicleID    int64
	DeliveryDate time.Time
	TimeStart    time.Time
	TimeEnd      time.Time
	Points       []DeliveryPointInput
}

type DeliveryFilter struct {
	Date      *time.Time
	CourierID *int64
	Status    *domain.DeliveryStatus
}

type RouteWithProductsInput struct {
	Route    []RoutePoint
	Products []DeliveryPointProductInput
}

type GenerateDeliveriesInput struct {
	DeliveryData map[time.Time][]RouteWithProductsInput
}

func (s *DeliveryService) GetDelivery(id int64) (DeliveryDTO, error) {
	delivery, ok := s.store.GetDelivery(id)
	if !ok {
		return DeliveryDTO{}, ErrNotFound
	}
	users, vehicles, products := s.snapshot()
	return NewDeliveryDTO(*delivery, users, vehicles, products), nil
}

func (s *DeliveryService) ListDeliveries(filter DeliveryFilter) []DeliveryDTO {
	deliveries := s.store.ListDeliveries()
	users, vehicles, products := s.snapshot()
	var filtered []DeliveryDTO
	for _, delivery := range deliveries {
		if filter.Date != nil && !sameDay(delivery.DeliveryDate, *filter.Date) {
			continue
		}
		if filter.CourierID != nil && delivery.CourierID != *filter.CourierID {
			continue
		}
		if filter.Status != nil && delivery.Status != *filter.Status {
			continue
		}
		filtered = append(filtered, NewDeliveryDTO(delivery, users, vehicles, products))
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DeliveryDate < filtered[j].DeliveryDate
	})
	return filtered
}

func (s *DeliveryService) CreateDelivery(input DeliveryInput, createdByID int64) (DeliveryDTO, error) {
	delivery, err := s.buildDeliveryEntity(input, createdByID, nil)
	if err != nil {
		return DeliveryDTO{}, err
	}
	created, err := s.store.CreateDelivery(delivery)
	if err != nil {
		return DeliveryDTO{}, err
	}
	users, vehicles, products := s.snapshot()
	return NewDeliveryDTO(created, users, vehicles, products), nil
}

func (s *DeliveryService) UpdateDelivery(id int64, input DeliveryInput) (DeliveryDTO, error) {
	existing, ok := s.store.GetDelivery(id)
	if !ok {
		return DeliveryDTO{}, ErrNotFound
	}
	if !existing.DeliveryDate.After(time.Now().AddDate(0, 0, 3)) {
		return DeliveryDTO{}, ValidationErrors{{Field: "delivery_date", Message: "Редактирование запрещено за 3 дня до доставки"}}
	}
	delivery, err := s.buildDeliveryEntity(input, existing.CreatedByID, existing)
	if err != nil {
		return DeliveryDTO{}, err
	}
	delivery.ID = existing.ID
	updated, err := s.store.UpdateDelivery(delivery)
	if err != nil {
		return DeliveryDTO{}, err
	}
	users, vehicles, products := s.snapshot()
	return NewDeliveryDTO(updated, users, vehicles, products), nil
}

func (s *DeliveryService) DeleteDelivery(id int64) error {
	delivery, ok := s.store.GetDelivery(id)
	if !ok {
		return ErrNotFound
	}
	if !delivery.DeliveryDate.After(time.Now().AddDate(0, 0, 3)) {
		return ValidationErrors{{Field: "delivery_date", Message: "Удаление запрещено за 3 дня до доставки"}}
	}
	return s.store.DeleteDelivery(id)
}

func (s *DeliveryService) GenerateDeliveries(input GenerateDeliveriesInput, createdByID int64) (GenerateDeliveriesResponse, error) {
	if len(input.DeliveryData) == 0 {
		return GenerateDeliveriesResponse{}, ValidationErrors{{Field: "delivery_data", Message: "generation data is required"}}
	}
	usersList := s.store.ListUsers(nil)
	var couriers []domain.User
	for _, user := range usersList {
		if user.Role == domain.RoleCourier {
			couriers = append(couriers, user)
		}
	}
	vehicles := s.store.ListVehicles()
	if len(couriers) == 0 || len(vehicles) == 0 {
		return GenerateDeliveriesResponse{}, ValidationErrors{{Field: "delivery_data", Message: "need at least one courier and one vehicle"}}
	}
	courierIdx := 0
	vehicleIdx := 0
	result := GenerateDeliveriesResponse{
		ByDate: make(map[string]GenerationResultByDate),
	}
	for date, routes := range input.DeliveryData {
		var deliveries []DeliveryDTO
		var warnings []string
		for i, route := range routes {
			if len(route.Route) == 0 {
				warnings = append(warnings, fmt.Sprintf("date %s route %d skipped due to insufficient points", date.Format("2006-01-02"), i+1))
				continue
			}
			points := make([]DeliveryPointInput, len(route.Route))
			for idx, point := range route.Route {
				points[idx] = DeliveryPointInput{
					Sequence:  idx + 1,
					Latitude:  point.Latitude,
					Longitude: point.Longitude,
					Products:  route.Products,
				}
			}
			start := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC).Add(time.Duration(i) * 2 * time.Hour)
			end := start.Add(3 * time.Hour)
			input := DeliveryInput{
				CourierID:    couriers[courierIdx%len(couriers)].ID,
				VehicleID:    vehicles[vehicleIdx%len(vehicles)].ID,
				DeliveryDate: date,
				TimeStart:    start,
				TimeEnd:      end,
				Points:       points,
			}
			delivery, err := s.CreateDelivery(input, createdByID)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("date %s route %d skipped: %v", date.Format("2006-01-02"), i+1, err))
				continue
			}
			deliveries = append(deliveries, delivery)
			courierIdx++
			vehicleIdx++
		}
		result.ByDate[date.Format("2006-01-02")] = GenerationResultByDate{
			GeneratedCount: len(deliveries),
			Deliveries:     deliveries,
			Warnings:       warnings,
		}
		result.TotalGenerated += len(deliveries)
	}
	return result, nil
}

func (s *DeliveryService) validateDeliveryInput(input DeliveryInput) []ValidationError {
	var errs []ValidationError
	if input.CourierID <= 0 {
		errs = append(errs, ValidationError{Field: "courier_id", Message: "courier is required"})
	}
	if input.VehicleID <= 0 {
		errs = append(errs, ValidationError{Field: "vehicle_id", Message: "vehicle is required"})
	}
	if input.DeliveryDate.IsZero() {
		errs = append(errs, ValidationError{Field: "delivery_date", Message: "delivery date is required"})
	} else {
		today := time.Now().UTC().Truncate(24 * time.Hour)
		if input.DeliveryDate.Before(today) {
			errs = append(errs, ValidationError{Field: "delivery_date", Message: "delivery date must be in the future"})
		}
	}
	if input.TimeStart.IsZero() || input.TimeEnd.IsZero() {
		errs = append(errs, ValidationError{Field: "time_window", Message: "start and end time are required"})
	} else if !input.TimeEnd.After(input.TimeStart) {
		errs = append(errs, ValidationError{Field: "time_window", Message: "end time must be after start time"})
	}
	if len(input.Points) == 0 {
		errs = append(errs, ValidationError{Field: "points", Message: "at least one delivery point is required"})
	}
	return errs
}

func (s *DeliveryService) buildPoints(points []DeliveryPointInput) ([]domain.DeliveryPoint, float64, float64, error) {
	sort.Slice(points, func(i, j int) bool {
		return points[i].Sequence < points[j].Sequence
	})
	var totalWeight float64
	var totalVolume float64
	result := make([]domain.DeliveryPoint, len(points))
	for i, point := range points {
		if point.Sequence <= 0 {
			return nil, 0, 0, ValidationErrors{{Field: fmt.Sprintf("points[%d].sequence", i), Message: "sequence must be positive"}}
		}
		domainPoint := domain.DeliveryPoint{
			Sequence:  point.Sequence,
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
		}
		if len(point.Products) == 0 {
			return nil, 0, 0, ValidationErrors{{Field: fmt.Sprintf("points[%d].products", i), Message: "products are required"}}
		}
		domainPoint.Products = make([]domain.DeliveryPointProduct, len(point.Products))
		for j, product := range point.Products {
			if product.Quantity <= 0 {
				return nil, 0, 0, ValidationErrors{{Field: fmt.Sprintf("points[%d].products[%d].quantity", i, j), Message: "quantity must be positive"}}
			}
			entity, ok := s.store.GetProduct(product.ProductID)
			if !ok {
				return nil, 0, 0, ValidationErrors{{Field: fmt.Sprintf("points[%d].products[%d].product_id", i, j), Message: "product not found"}}
			}
			totalWeight += entity.Weight * float64(product.Quantity)
			totalVolume += entity.Volume() * float64(product.Quantity)
			domainPoint.Products[j] = domain.DeliveryPointProduct{
				ProductID: product.ProductID,
				Quantity:  product.Quantity,
			}
		}
		result[i] = domainPoint
	}
	return result, totalWeight, totalVolume, nil
}

func (s *DeliveryService) validateRouteDuration(points []DeliveryPointInput, start, end time.Time) error {
	if len(points) < 2 {
		return nil
	}
	routePoints := make([]RoutePoint, len(points))
	for i, point := range points {
		routePoints[i] = RoutePoint{
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
		}
	}
	result, err := s.route.Calculate(routePoints)
	if err != nil {
		return err
	}
	windowMinutes := int(end.Sub(start).Minutes())
	if windowMinutes <= 0 {
		return ValidationErrors{{Field: "time_window", Message: "invalid time window"}}
	}
	if result.DurationMinutes > windowMinutes {
		return ValidationErrors{{Field: "time_window", Message: "Недостаточно времени для маршрута"}}
	}
	return nil
}

func (s *DeliveryService) snapshot() (map[int64]domain.User, map[int64]domain.Vehicle, map[int64]domain.Product) {
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

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func combineDateTime(date time.Time, tod time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), tod.Hour(), tod.Minute(), tod.Second(), 0, time.UTC)
}

func (s *DeliveryService) validateOverlapCapacity(vehicle domain.Vehicle, delivery domain.Delivery, weight, volume float64, skipID int64) error {
	deliveries := s.store.ListDeliveries()
	var overlaps []domain.Delivery
	totalWeight := weight
	totalVolume := volume
	for _, existing := range deliveries {
		if existing.ID == skipID {
			continue
		}
		if existing.VehicleID != vehicle.ID {
			continue
		}
		if !sameDay(existing.DeliveryDate, delivery.DeliveryDate) {
			continue
		}
		if existing.Status == domain.StatusCompleted || existing.Status == domain.StatusCancelled {
			continue
		}
		if !timesOverlap(existing.TimeStart, existing.TimeEnd, delivery.TimeStart, delivery.TimeEnd) {
			continue
		}
		overlaps = append(overlaps, existing)
		w, v := s.deliveryMetrics(existing)
		totalWeight += w
		totalVolume += v
	}
	if len(overlaps) == 0 {
		return nil
	}
	if totalWeight > vehicle.MaxWeight+0.0001 {
		return ValidationErrors{{Field: "points", Message: fmt.Sprintf("Превышена грузоподъемность машины в период, пересекающиеся доставки: %s", joinDeliveryIDs(overlaps))}}
	}
	if totalVolume > vehicle.MaxVolume+0.0001 {
		return ValidationErrors{{Field: "points", Message: fmt.Sprintf("Превышен объем машины в период, пересекающиеся доставки: %s", joinDeliveryIDs(overlaps))}}
	}
	return nil
}

func (s *DeliveryService) deliveryMetrics(delivery domain.Delivery) (float64, float64) {
	var weight float64
	var volume float64
	for _, point := range delivery.DeliveryPoints {
		for _, product := range point.Products {
			if entity, ok := s.store.GetProduct(product.ProductID); ok {
				weight += entity.Weight * float64(product.Quantity)
				volume += entity.Volume() * float64(product.Quantity)
			}
		}
	}
	return weight, volume
}

func timesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func joinDeliveryIDs(deliveries []domain.Delivery) string {
	if len(deliveries) == 0 {
		return ""
	}
	ids := make([]string, len(deliveries))
	for i, delivery := range deliveries {
		ids[i] = fmt.Sprintf("#%d", delivery.ID)
	}
	return strings.Join(ids, ", ")
}

func (s *DeliveryService) buildDeliveryEntity(input DeliveryInput, createdByID int64, existing *domain.Delivery) (domain.Delivery, error) {
	if errs := s.validateDeliveryInput(input); len(errs) > 0 {
		return domain.Delivery{}, ValidationErrors(errs)
	}
	vehicle, ok := s.store.GetVehicle(input.VehicleID)
	if !ok {
		return domain.Delivery{}, ValidationErrors{{Field: "vehicle_id", Message: "vehicle not found"}}
	}
	courier, ok := s.store.GetUser(input.CourierID)
	if !ok {
		return domain.Delivery{}, ValidationErrors{{Field: "courier_id", Message: "courier not found"}}
	}
	if courier.Role != domain.RoleCourier {
		return domain.Delivery{}, ValidationErrors{{Field: "courier_id", Message: "user is not a courier"}}
	}
	points, totalWeight, totalVolume, err := s.buildPoints(input.Points)
	if err != nil {
		return domain.Delivery{}, err
	}
	if totalWeight > vehicle.MaxWeight+0.0001 {
		return domain.Delivery{}, ValidationErrors{{Field: "points", Message: "Превышена грузоподъемность машины"}}
	}
	if totalVolume > vehicle.MaxVolume+0.0001 {
		return domain.Delivery{}, ValidationErrors{{Field: "points", Message: "Превышен объем машины"}}
	}
	date := time.Date(input.DeliveryDate.Year(), input.DeliveryDate.Month(), input.DeliveryDate.Day(), 0, 0, 0, 0, time.UTC)
	start := combineDateTime(date, input.TimeStart)
	end := combineDateTime(date, input.TimeEnd)
	if err := s.validateRouteDuration(input.Points, start, end); err != nil {
		return domain.Delivery{}, err
	}
	delivery := domain.Delivery{
		CourierID:      input.CourierID,
		VehicleID:      input.VehicleID,
		CreatedByID:    createdByID,
		DeliveryDate:   date,
		TimeStart:      start,
		TimeEnd:        end,
		Status:         domain.StatusPlanned,
		DeliveryPoints: points,
	}
	if existing != nil {
		delivery.ID = existing.ID
		delivery.CreatedAt = existing.CreatedAt
		delivery.Status = existing.Status
	}
	skipID := int64(0)
	if existing != nil {
		skipID = existing.ID
	}
	if err := s.validateOverlapCapacity(*vehicle, delivery, totalWeight, totalVolume, skipID); err != nil {
		return domain.Delivery{}, err
	}
	return delivery, nil
}
