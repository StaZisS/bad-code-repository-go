package storage

import (
	"errors"
	"sync"
	"time"

	"courier_managment_system_go/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type MemoryStore struct {
	mu sync.RWMutex

	users      map[int64]*domain.User
	products   map[int64]*domain.Product
	vehicles   map[int64]*domain.Vehicle
	deliveries map[int64]*domain.Delivery

	nextUserID          int64
	nextProductID       int64
	nextVehicleID       int64
	nextDeliveryID      int64
	nextDeliveryPointID int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:      make(map[int64]*domain.User),
		products:   make(map[int64]*domain.Product),
		vehicles:   make(map[int64]*domain.Vehicle),
		deliveries: make(map[int64]*domain.Delivery),
	}
}

func (s *MemoryStore) CreateUser(user domain.User) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.users {
		if existing.Login == user.Login {
			return domain.User{}, ErrConflict
		}
	}

	s.nextUserID++
	user.ID = s.nextUserID
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	copy := user
	s.users[user.ID] = &copy
	return user, nil
}

func (s *MemoryStore) GetUserByLogin(login string) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.Login == login {
			copy := *user
			return &copy, true
		}
	}
	return nil, false
}

func (s *MemoryStore) GetUser(id int64) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return nil, false
	}
	copy := *user
	return &copy, true
}

func (s *MemoryStore) ListUsers(role *domain.UserRole) []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		if role != nil && user.Role != *role {
			continue
		}
		res = append(res, *user)
	}
	return res
}

func (s *MemoryStore) UpdateUser(user domain.User) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.users[user.ID]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	for _, other := range s.users {
		if other.Login == user.Login && other.ID != user.ID {
			return domain.User{}, ErrConflict
		}
	}
	existing.Login = user.Login
	existing.Name = user.Name
	existing.Role = user.Role
	if user.PasswordHash != "" {
		existing.PasswordHash = user.PasswordHash
	}
	copy := *existing
	return copy, nil
}

func (s *MemoryStore) DeleteUser(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return ErrNotFound
	}
	delete(s.users, id)
	return nil
}

func (s *MemoryStore) CreateProduct(product domain.Product) (domain.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextProductID++
	product.ID = s.nextProductID
	copy := product
	s.products[product.ID] = &copy
	return product, nil
}

func (s *MemoryStore) UpdateProduct(product domain.Product) (domain.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.products[product.ID]
	if !ok {
		return domain.Product{}, ErrNotFound
	}
	existing.Name = product.Name
	existing.Weight = product.Weight
	existing.Length = product.Length
	existing.Width = product.Width
	existing.Height = product.Height
	copy := *existing
	return copy, nil
}

func (s *MemoryStore) GetProduct(id int64) (*domain.Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, ok := s.products[id]
	if !ok {
		return nil, false
	}
	copy := *product
	return &copy, true
}

func (s *MemoryStore) ListProducts() []domain.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]domain.Product, 0, len(s.products))
	for _, product := range s.products {
		res = append(res, *product)
	}
	return res
}

func (s *MemoryStore) DeleteProduct(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[id]; !ok {
		return ErrNotFound
	}
	delete(s.products, id)
	return nil
}

func (s *MemoryStore) CreateVehicle(vehicle domain.Vehicle) (domain.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.vehicles {
		if existing.LicensePlate == vehicle.LicensePlate {
			return domain.Vehicle{}, ErrConflict
		}
	}
	s.nextVehicleID++
	vehicle.ID = s.nextVehicleID
	copy := vehicle
	s.vehicles[vehicle.ID] = &copy
	return vehicle, nil
}

func (s *MemoryStore) UpdateVehicle(vehicle domain.Vehicle) (domain.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.vehicles[vehicle.ID]
	if !ok {
		return domain.Vehicle{}, ErrNotFound
	}
	for _, other := range s.vehicles {
		if other.LicensePlate == vehicle.LicensePlate && other.ID != vehicle.ID {
			return domain.Vehicle{}, ErrConflict
		}
	}
	existing.Brand = vehicle.Brand
	existing.LicensePlate = vehicle.LicensePlate
	existing.MaxWeight = vehicle.MaxWeight
	existing.MaxVolume = vehicle.MaxVolume
	copy := *existing
	return copy, nil
}

func (s *MemoryStore) GetVehicle(id int64) (*domain.Vehicle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicle, ok := s.vehicles[id]
	if !ok {
		return nil, false
	}
	copy := *vehicle
	return &copy, true
}

func (s *MemoryStore) ListVehicles() []domain.Vehicle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]domain.Vehicle, 0, len(s.vehicles))
	for _, vehicle := range s.vehicles {
		res = append(res, *vehicle)
	}
	return res
}

func (s *MemoryStore) DeleteVehicle(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vehicles[id]; !ok {
		return ErrNotFound
	}
	delete(s.vehicles, id)
	return nil
}

func (s *MemoryStore) CreateDelivery(delivery domain.Delivery) (domain.Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextDeliveryID++
	delivery.ID = s.nextDeliveryID
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = time.Now().UTC()
	}
	delivery.UpdatedAt = delivery.CreatedAt
	for i := range delivery.DeliveryPoints {
		s.nextDeliveryPointID++
		delivery.DeliveryPoints[i].ID = s.nextDeliveryPointID
	}
	copy := delivery
	s.deliveries[delivery.ID] = &copy
	return delivery, nil
}

func (s *MemoryStore) UpdateDelivery(delivery domain.Delivery) (domain.Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.deliveries[delivery.ID]
	if !ok {
		return domain.Delivery{}, ErrNotFound
	}
	delivery.UpdatedAt = time.Now().UTC()
	for i := range delivery.DeliveryPoints {
		s.nextDeliveryPointID++
		delivery.DeliveryPoints[i].ID = s.nextDeliveryPointID
	}
	copy := delivery
	s.deliveries[delivery.ID] = &copy
	return delivery, nil
}

func (s *MemoryStore) GetDelivery(id int64) (*domain.Delivery, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	delivery, ok := s.deliveries[id]
	if !ok {
		return nil, false
	}
	copy := *delivery
	copy.DeliveryPoints = clonePoints(delivery.DeliveryPoints)
	return &copy, true
}

func (s *MemoryStore) ListDeliveries() []domain.Delivery {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]domain.Delivery, 0, len(s.deliveries))
	for _, delivery := range s.deliveries {
		item := *delivery
		item.DeliveryPoints = clonePoints(delivery.DeliveryPoints)
		res = append(res, item)
	}
	return res
}

func (s *MemoryStore) DeleteDelivery(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deliveries[id]; !ok {
		return ErrNotFound
	}
	delete(s.deliveries, id)
	return nil
}

func (s *MemoryStore) SetDeliveryStatus(id int64, status domain.DeliveryStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delivery, ok := s.deliveries[id]
	if !ok {
		return ErrNotFound
	}
	delivery.Status = status
	delivery.UpdatedAt = time.Now().UTC()
	return nil
}

func clonePoints(points []domain.DeliveryPoint) []domain.DeliveryPoint {
	if len(points) == 0 {
		return nil
	}
	res := make([]domain.DeliveryPoint, len(points))
	for i, point := range points {
		pointCopy := point
		if len(point.Products) > 0 {
			pointCopy.Products = make([]domain.DeliveryPointProduct, len(point.Products))
			copy(pointCopy.Products, point.Products)
		}
		res[i] = pointCopy
	}
	return res
}
