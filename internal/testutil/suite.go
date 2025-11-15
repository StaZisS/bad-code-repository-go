package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "courier_managment_system_go/internal/api"
	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/service"
	"courier_managment_system_go/internal/storage"

	"github.com/gin-gonic/gin"
)

type AppTestSuite struct {
	T *testing.T

	Store           *storage.MemoryStore
	UserService     *service.UserService
	ProductService  *service.ProductService
	VehicleService  *service.VehicleService
	DeliveryService *service.DeliveryService
	CourierService  *service.CourierService
	AuthService     *service.AuthService
	RouteService    *service.RouteService

	Server *api.Server
	Router *gin.Engine

	AdminUser   domain.User
	ManagerUser domain.User
	CourierUser domain.User

	AdminToken   string
	ManagerToken string
	CourierToken string
}

func NewAppTestSuite(t *testing.T) *AppTestSuite {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := storage.NewMemoryStore()
	route := service.NewRouteService(40)
	userService := service.NewUserService(store)
	productService := service.NewProductService(store)
	vehicleService := service.NewVehicleService(store)
	deliveryService := service.NewDeliveryService(store, route)
	courierService := service.NewCourierService(store)
	authService := service.NewAuthService(store, "test-secret", 24*time.Hour)

	server := api.NewServer(
		authService,
		userService,
		productService,
		vehicleService,
		deliveryService,
		courierService,
		route,
	)

	suite := &AppTestSuite{
		T:               t,
		Store:           store,
		UserService:     userService,
		ProductService:  productService,
		VehicleService:  vehicleService,
		DeliveryService: deliveryService,
		CourierService:  courierService,
		AuthService:     authService,
		RouteService:    route,
		Server:          server,
		Router:          server.Engine(),
	}
	suite.bootstrapUsers()
	return suite
}

func (s *AppTestSuite) bootstrapUsers() {
	if err := s.UserService.EnsureAdminUser("admin", "admin123", "Системный администратор"); err != nil {
		s.T.Fatalf("failed to seed admin: %v", err)
	}

	admin, ok := s.Store.GetUserByLogin("admin")
	if !ok {
		s.T.Fatalf("admin not found")
	}
	s.AdminUser = *admin

	if _, err := s.UserService.CreateUser(service.CreateUserInput{
		Login:    "manager",
		Password: "password",
		Name:     "Менеджер",
		Role:     domain.RoleManager,
	}); err != nil {
		s.T.Fatalf("failed to seed manager: %v", err)
	}
	if _, err := s.UserService.CreateUser(service.CreateUserInput{
		Login:    "courier",
		Password: "password",
		Name:     "Курьер",
		Role:     domain.RoleCourier,
	}); err != nil {
		s.T.Fatalf("failed to seed courier: %v", err)
	}

	manager, ok := s.Store.GetUserByLogin("manager")
	if !ok {
		s.T.Fatalf("manager not found")
	}
	courier, ok := s.Store.GetUserByLogin("courier")
	if !ok {
		s.T.Fatalf("courier not found")
	}

	s.ManagerUser = *manager
	s.CourierUser = *courier

	s.AdminToken = s.mustLogin("admin", "admin123")
	s.ManagerToken = s.mustLogin("manager", "password")
	s.CourierToken = s.mustLogin("courier", "password")
}

func (s *AppTestSuite) mustLogin(login, password string) string {
	token, _, err := s.AuthService.Login(login, password)
	if err != nil {
		s.T.Fatalf("login failed for %s: %v", login, err)
	}
	return token
}

func (s *AppTestSuite) DoRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			s.T.Fatalf("failed to encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func (s *AppTestSuite) Get(path, token string) *httptest.ResponseRecorder {
	return s.DoRequest(http.MethodGet, path, nil, token)
}

func (s *AppTestSuite) Post(path string, body interface{}, token string) *httptest.ResponseRecorder {
	return s.DoRequest(http.MethodPost, path, body, token)
}

func (s *AppTestSuite) Put(path string, body interface{}, token string) *httptest.ResponseRecorder {
	return s.DoRequest(http.MethodPut, path, body, token)
}

func (s *AppTestSuite) Delete(path string, token string) *httptest.ResponseRecorder {
	return s.DoRequest(http.MethodDelete, path, nil, token)
}

func (s *AppTestSuite) Decode(rr *httptest.ResponseRecorder, target interface{}) {
	s.T.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		s.T.Fatalf("failed to decode response: %v (%s)", err, rr.Body.String())
	}
}

func (s *AppTestSuite) ExpectStatus(rr *httptest.ResponseRecorder, status int) {
	s.T.Helper()
	if rr.Code != status {
		s.T.Fatalf("expected status %d, got %d: %s", status, rr.Code, rr.Body.String())
	}
}

func (s *AppTestSuite) CreateProduct(name string, weight, length, width, height float64) service.ProductDTO {
	dto, err := s.ProductService.CreateProduct(service.ProductInput{
		Name:   name,
		Weight: weight,
		Length: length,
		Width:  width,
		Height: height,
	})
	if err != nil {
		s.T.Fatalf("failed to create product: %v", err)
	}
	return dto
}

func (s *AppTestSuite) CreateVehicle(brand, plate string, maxWeight, maxVolume float64) service.VehicleDTO {
	dto, err := s.VehicleService.CreateVehicle(service.VehicleInput{
		Brand:        brand,
		LicensePlate: plate,
		MaxWeight:    maxWeight,
		MaxVolume:    maxVolume,
	})
	if err != nil {
		s.T.Fatalf("failed to create vehicle: %v", err)
	}
	return dto
}

func (s *AppTestSuite) FutureDate(days int) time.Time {
	return time.Now().UTC().AddDate(0, 0, days)
}

func (s *AppTestSuite) TimeOfDay(hour, minute int) time.Time {
	return time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
}

func (s *AppTestSuite) DeliveryInput(productID, vehicleID, courierID int64, date time.Time) service.DeliveryInput {
	return service.DeliveryInput{
		CourierID:    courierID,
		VehicleID:    vehicleID,
		DeliveryDate: date,
		TimeStart:    s.TimeOfDay(9, 0),
		TimeEnd:      s.TimeOfDay(18, 0),
		Points: []service.DeliveryPointInput{
			{
				Sequence:  1,
				Latitude:  55.7558,
				Longitude: 37.6176,
				Products: []service.DeliveryPointProductInput{
					{ProductID: productID, Quantity: 1},
				},
			},
		},
	}
}

func (s *AppTestSuite) MustCreateDelivery(input service.DeliveryInput) domain.Delivery {
	dto, err := s.DeliveryService.CreateDelivery(input, s.ManagerUser.ID)
	if err != nil {
		s.T.Fatalf("failed to create delivery: %v", err)
	}
	delivery, ok := s.Store.GetDelivery(dto.ID)
	if !ok {
		s.T.Fatalf("delivery not found after creation")
	}
	return *delivery
}

func (s *AppTestSuite) SetDeliveryStatus(id int64, status domain.DeliveryStatus) {
	if err := s.Store.SetDeliveryStatus(id, status); err != nil {
		s.T.Fatalf("failed to set delivery status: %v", err)
	}
}
