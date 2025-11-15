package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine          *gin.Engine
	authService     *service.AuthService
	userService     *service.UserService
	productService  *service.ProductService
	vehicleService  *service.VehicleService
	deliveryService *service.DeliveryService
	courierService  *service.CourierService
	routeService    *service.RouteService
}

func NewServer(
	auth *service.AuthService,
	users *service.UserService,
	products *service.ProductService,
	vehicles *service.VehicleService,
	deliveries *service.DeliveryService,
	couriers *service.CourierService,
	routes *service.RouteService,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	s := &Server{
		engine:          engine,
		authService:     auth,
		userService:     users,
		productService:  products,
		vehicleService:  vehicles,
		deliveryService: deliveries,
		courierService:  couriers,
		routeService:    routes,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) registerRoutes() {
	s.engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	s.engine.POST("/auth/login", s.handleLogin)

	protected := s.engine.Group("/")
	protected.Use(AuthMiddleware(s.authService))

	protected.GET("/products", s.handleListProducts)
	protected.POST("/products", RequireRoles(domain.RoleAdmin), s.handleCreateProduct)
	protected.PUT("/products/:id", RequireRoles(domain.RoleAdmin), s.handleUpdateProduct)
	protected.DELETE("/products/:id", RequireRoles(domain.RoleAdmin), s.handleDeleteProduct)

	protected.GET("/vehicles", s.handleListVehicles)
	adminVehicles := protected.Group("/vehicles")
	adminVehicles.Use(RequireRoles(domain.RoleAdmin))
	adminVehicles.POST("", s.handleCreateVehicle)
	adminVehicles.PUT("/:id", s.handleUpdateVehicle)
	adminVehicles.DELETE("/:id", s.handleDeleteVehicle)

	usersGroup := protected.Group("/users")
	usersGroup.Use(RequireRoles(domain.RoleAdmin))
	usersGroup.GET("", s.handleListUsers)
	usersGroup.POST("", s.handleCreateUser)
	usersGroup.PUT("/:id", s.handleUpdateUser)
	usersGroup.DELETE("/:id", s.handleDeleteUser)

	deliveries := protected.Group("/deliveries")
	deliveries.GET("/:id", s.handleGetDelivery)
	managerDeliveries := deliveries.Group("")
	managerDeliveries.Use(RequireRoles(domain.RoleManager))
	managerDeliveries.GET("", s.handleListDeliveries)
	managerDeliveries.POST("", s.handleCreateDelivery)
	managerDeliveries.PUT("/:id", s.handleUpdateDelivery)
	managerDeliveries.DELETE("/:id", s.handleDeleteDelivery)
	managerDeliveries.POST("/generate", s.handleGenerateDeliveries)

	routeGroup := protected.Group("/routes")
	routeGroup.Use(RequireRoles(domain.RoleManager))
	routeGroup.POST("/calculate", s.handleCalculateRoute)

	courierGroup := protected.Group("/courier")
	courierGroup.Use(RequireRoles(domain.RoleCourier))
	courierGroup.GET("/deliveries", s.handleCourierDeliveries)
	courierGroup.GET("/deliveries/:id", s.handleCourierDeliveryDetail)
}

func (s *Server) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	token, user, err := s.authService.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to login"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (s *Server) handleListUsers(c *gin.Context) {
	var rolePtr *domain.UserRole
	role := c.Query("role")
	if role != "" {
		roleValue := domain.UserRole(role)
		if !roleValue.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role"})
			return
		}
		rolePtr = &roleValue
	}
	c.JSON(http.StatusOK, s.userService.ListUsers(rolePtr))
}

func (s *Server) handleCreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.userService.CreateUser(service.CreateUserInput{
		Login:    req.Login,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
	})
	handleResponse(c, dto, err, http.StatusCreated)
}

func (s *Server) handleUpdateUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var req userUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.userService.UpdateUser(id, service.UpdateUserInput{
		Name:     req.Name,
		Login:    req.Login,
		Role:     req.Role,
		Password: req.Password,
	})
	handleResponse(c, dto, err, http.StatusOK)
}

func (s *Server) handleDeleteUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	err = s.userService.DeleteUser(id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, s.productService.ListProducts())
}

func (s *Server) handleCreateProduct(c *gin.Context) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.productService.CreateProduct(req.toInput())
	handleResponse(c, dto, err, http.StatusCreated)
}

func (s *Server) handleUpdateProduct(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.productService.UpdateProduct(id, req.toInput())
	handleResponse(c, dto, err, http.StatusOK)
}

func (s *Server) handleDeleteProduct(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	err = s.productService.DeleteProduct(id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleListVehicles(c *gin.Context) {
	c.JSON(http.StatusOK, s.vehicleService.ListVehicles())
}

func (s *Server) handleCreateVehicle(c *gin.Context) {
	var req vehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.vehicleService.CreateVehicle(req.toInput())
	handleResponse(c, dto, err, http.StatusCreated)
}

func (s *Server) handleUpdateVehicle(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var req vehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	dto, err := s.vehicleService.UpdateVehicle(id, req.toInput())
	handleResponse(c, dto, err, http.StatusOK)
}

func (s *Server) handleDeleteVehicle(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	err = s.vehicleService.DeleteVehicle(id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleListDeliveries(c *gin.Context) {
	var filter service.DeliveryFilter
	if dateStr := c.Query("date"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			filter.Date = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid date"})
			return
		}
	}
	if courierStr := c.Query("courier_id"); courierStr != "" {
		if id, err := strconv.ParseInt(courierStr, 10, 64); err == nil {
			filter.CourierID = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid courier id"})
			return
		}
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status := domain.DeliveryStatus(statusStr)
		filter.Status = &status
	}
	c.JSON(http.StatusOK, s.deliveryService.ListDeliveries(filter))
}

func (s *Server) handleCreateDelivery(c *gin.Context) {
	var req deliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := MustUser(c)
	dto, err := s.deliveryService.CreateDelivery(input, user.ID)
	handleResponse(c, dto, err, http.StatusCreated)
}

func (s *Server) handleUpdateDelivery(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var req deliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	dto, err := s.deliveryService.UpdateDelivery(id, input)
	handleResponse(c, dto, err, http.StatusOK)
}

func (s *Server) handleDeleteDelivery(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	if err := s.deliveryService.DeleteDelivery(id); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleGetDelivery(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	dto, err := s.deliveryService.GetDelivery(id)
	handleResponse(c, dto, err, http.StatusOK)
}

func (s *Server) handleGenerateDeliveries(c *gin.Context) {
	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := MustUser(c)
	result, err := s.deliveryService.GenerateDeliveries(input, user.ID)
	handleResponse(c, result, err, http.StatusOK)
}

func (s *Server) handleCalculateRoute(c *gin.Context) {
	var req routeCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	if len(req.Points) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "need at least two points"})
		return
	}
	points := make([]service.RoutePoint, len(req.Points))
	for i, point := range req.Points {
		points[i] = service.RoutePoint(point)
	}
	result, err := s.routeService.Calculate(points)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) handleCourierDeliveries(c *gin.Context) {
	user := MustUser(c)
	var date, dateFrom, dateTo *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid date"})
			return
		}
	}
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if t, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid date_from"})
			return
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if t, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid date_to"})
			return
		}
	}
	var status *domain.DeliveryStatus
	if statusStr := c.Query("status"); statusStr != "" {
		value := domain.DeliveryStatus(statusStr)
		status = &value
	}
	deliveries, err := s.courierService.GetCourierDeliveries(user.ID, date, status, dateFrom, dateTo)
	handleResponse(c, deliveries, err, http.StatusOK)
}

func (s *Server) handleCourierDeliveryDetail(c *gin.Context) {
	user := MustUser(c)
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	dto, err := s.courierService.GetCourierDeliveryByID(user.ID, id)
	handleResponse(c, dto, err, http.StatusOK)
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func handleResponse[T any](c *gin.Context, payload T, err error, successStatus int) {
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(successStatus, payload)
}

func handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case service.ValidationErrors:
		c.JSON(http.StatusBadRequest, gin.H{"message": "validation failed", "details": e})
		return
	default:
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "not found"})
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"message": "forbidden"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userRequest struct {
	Login    string          `json:"login" binding:"required"`
	Password string          `json:"password" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

type userUpdateRequest struct {
	Name     *string          `json:"name"`
	Login    *string          `json:"login"`
	Role     *domain.UserRole `json:"role"`
	Password *string          `json:"password"`
}

type productRequest struct {
	Name   string  `json:"name" binding:"required"`
	Weight float64 `json:"weight" binding:"required"`
	Length float64 `json:"length" binding:"required"`
	Width  float64 `json:"width" binding:"required"`
	Height float64 `json:"height" binding:"required"`
}

func (r productRequest) toInput() service.ProductInput {
	return service.ProductInput{
		Name:   r.Name,
		Weight: r.Weight,
		Length: r.Length,
		Width:  r.Width,
		Height: r.Height,
	}
}

type vehicleRequest struct {
	Brand        string  `json:"brand" binding:"required"`
	LicensePlate string  `json:"licensePlate" binding:"required"`
	MaxWeight    float64 `json:"maxWeight" binding:"required"`
	MaxVolume    float64 `json:"maxVolume" binding:"required"`
}

func (r vehicleRequest) toInput() service.VehicleInput {
	return service.VehicleInput{
		Brand:        r.Brand,
		LicensePlate: strings.ToUpper(strings.TrimSpace(r.LicensePlate)),
		MaxWeight:    r.MaxWeight,
		MaxVolume:    r.MaxVolume,
	}
}

type deliveryPointProductRequest struct {
	ProductID int64 `json:"productId" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required"`
}

type deliveryPointRequest struct {
	Sequence  *int                          `json:"sequence"`
	Latitude  float64                       `json:"latitude" binding:"required"`
	Longitude float64                       `json:"longitude" binding:"required"`
	Products  []deliveryPointProductRequest `json:"products" binding:"required"`
}

type deliveryRequest struct {
	CourierID    int64                  `json:"courierId" binding:"required"`
	VehicleID    int64                  `json:"vehicleId" binding:"required"`
	DeliveryDate string                 `json:"deliveryDate" binding:"required"`
	TimeStart    string                 `json:"timeStart" binding:"required"`
	TimeEnd      string                 `json:"timeEnd" binding:"required"`
	Points       []deliveryPointRequest `json:"points" binding:"required"`
}

func (r deliveryRequest) toInput() (service.DeliveryInput, error) {
	date, err := time.Parse("2006-01-02", r.DeliveryDate)
	if err != nil {
		return service.DeliveryInput{}, err
	}
	start, err := parseTimeOfDay(r.TimeStart)
	if err != nil {
		return service.DeliveryInput{}, err
	}
	end, err := parseTimeOfDay(r.TimeEnd)
	if err != nil {
		return service.DeliveryInput{}, err
	}
	points := make([]service.DeliveryPointInput, len(r.Points))
	for i, point := range r.Points {
		products := make([]service.DeliveryPointProductInput, len(point.Products))
		for j, product := range point.Products {
			products[j] = service.DeliveryPointProductInput{
				ProductID: product.ProductID,
				Quantity:  product.Quantity,
			}
		}
		sequence := i + 1
		if point.Sequence != nil {
			sequence = *point.Sequence
		}
		points[i] = service.DeliveryPointInput{
			Sequence:  sequence,
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
			Products:  products,
		}
	}
	return service.DeliveryInput{
		CourierID:    r.CourierID,
		VehicleID:    r.VehicleID,
		DeliveryDate: date,
		TimeStart:    start,
		TimeEnd:      end,
		Points:       points,
	}, nil
}

type generateRequest struct {
	DeliveryData map[string][]generateRouteRequest `json:"deliveryData" binding:"required"`
}

type generateRouteRequest struct {
	Route    []routePointRequest           `json:"route" binding:"required"`
	Products []deliveryPointProductRequest `json:"products" binding:"required"`
}

func (r generateRequest) toInput() (service.GenerateDeliveriesInput, error) {
	if len(r.DeliveryData) == 0 {
		return service.GenerateDeliveriesInput{}, errors.New("delivery_data cannot be empty")
	}
	result := make(map[time.Time][]service.RouteWithProductsInput)
	for key, routes := range r.DeliveryData {
		date, err := time.Parse("2006-01-02", key)
		if err != nil {
			return service.GenerateDeliveriesInput{}, err
		}
		items := make([]service.RouteWithProductsInput, 0, len(routes))
		for _, route := range routes {
			points := make([]service.RoutePoint, len(route.Route))
			for i, point := range route.Route {
				points[i] = service.RoutePoint{
					Latitude:  point.Latitude,
					Longitude: point.Longitude,
				}
			}
			products := make([]service.DeliveryPointProductInput, len(route.Products))
			for i, product := range route.Products {
				products[i] = service.DeliveryPointProductInput{
					ProductID: product.ProductID,
					Quantity:  product.Quantity,
				}
			}
			items = append(items, service.RouteWithProductsInput{
				Route:    points,
				Products: products,
			})
		}
		result[date] = items
	}
	return service.GenerateDeliveriesInput{DeliveryData: result}, nil
}

type routeCalculationRequest struct {
	Points []routePointRequest `json:"points" binding:"required"`
}

type routePointRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func parseTimeOfDay(value string) (time.Time, error) {
	layouts := []string{"15:04:05", "15:04"}
	var err error
	for _, layout := range layouts {
		if t, e := time.Parse(layout, value); e == nil {
			return t, nil
		} else {
			err = e
		}
	}
	return time.Time{}, err
}
