package service

import (
	"fmt"
	"time"

	"courier_managment_system_go/internal/domain"
)

type UserDTO struct {
	ID        int64           `json:"id"`
	Login     string          `json:"login"`
	Name      string          `json:"name"`
	Role      domain.UserRole `json:"role"`
	CreatedAt time.Time       `json:"createdAt"`
}

func NewUserDTO(user domain.User) UserDTO {
	return UserDTO{
		ID:        user.ID,
		Login:     user.Login,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

type ProductDTO struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Volume float64 `json:"volume"`
}

func NewProductDTO(product domain.Product) ProductDTO {
	return ProductDTO{
		ID:     product.ID,
		Name:   product.Name,
		Weight: product.Weight,
		Length: product.Length,
		Width:  product.Width,
		Height: product.Height,
		Volume: product.Volume(),
	}
}

type VehicleDTO struct {
	ID           int64   `json:"id"`
	Brand        string  `json:"brand"`
	LicensePlate string  `json:"licensePlate"`
	MaxWeight    float64 `json:"maxWeight"`
	MaxVolume    float64 `json:"maxVolume"`
}

func NewVehicleDTO(vehicle domain.Vehicle) VehicleDTO {
	return VehicleDTO{
		ID:           vehicle.ID,
		Brand:        vehicle.Brand,
		LicensePlate: vehicle.LicensePlate,
		MaxWeight:    vehicle.MaxWeight,
		MaxVolume:    vehicle.MaxVolume,
	}
}

type DeliveryPointProductDTO struct {
	Product  ProductDTO `json:"product"`
	Quantity int        `json:"quantity"`
}

type DeliveryPointDTO struct {
	ID        int64                     `json:"id"`
	Sequence  int                       `json:"sequence"`
	Latitude  float64                   `json:"latitude"`
	Longitude float64                   `json:"longitude"`
	Products  []DeliveryPointProductDTO `json:"products"`
}

type DeliveryDTO struct {
	ID             int64                 `json:"id"`
	DeliveryNumber string                `json:"deliveryNumber"`
	Courier        *UserDTO              `json:"courier,omitempty"`
	Vehicle        *VehicleDTO           `json:"vehicle,omitempty"`
	CreatedBy      UserDTO               `json:"createdBy"`
	DeliveryDate   string                `json:"deliveryDate"`
	TimeStart      string                `json:"timeStart"`
	TimeEnd        string                `json:"timeEnd"`
	Status         domain.DeliveryStatus `json:"status"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	DeliveryPoints []DeliveryPointDTO    `json:"deliveryPoints"`
	TotalWeight    float64               `json:"totalWeight"`
	TotalVolume    float64               `json:"totalVolume"`
	CanEdit        bool                  `json:"canEdit"`
}

func NewDeliveryDTO(
	delivery domain.Delivery,
	users map[int64]domain.User,
	vehicles map[int64]domain.Vehicle,
	products map[int64]domain.Product,
) DeliveryDTO {
	var courier *UserDTO
	if u, ok := users[delivery.CourierID]; ok {
		userCopy := NewUserDTO(u)
		courier = &userCopy
	}

	var vehicleDTO *VehicleDTO
	if v, ok := vehicles[delivery.VehicleID]; ok {
		veh := NewVehicleDTO(v)
		vehicleDTO = &veh
	} else if delivery.VehicleID == 0 {
		vehicleDTO = &VehicleDTO{
			Brand:        "Не назначена",
			LicensePlate: "",
		}
	}

	createdBy, ok := users[delivery.CreatedByID]
	if !ok {
		createdBy = domain.User{
			ID:   delivery.CreatedByID,
			Name: "unknown",
			Role: domain.RoleManager,
		}
	}
	points := make([]DeliveryPointDTO, len(delivery.DeliveryPoints))
	var totalWeight float64
	var totalVolume float64
	for i, point := range delivery.DeliveryPoints {
		dto := DeliveryPointDTO{
			ID:        point.ID,
			Sequence:  point.Sequence,
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
		}
		if len(point.Products) > 0 {
			dto.Products = make([]DeliveryPointProductDTO, len(point.Products))
			for j, product := range point.Products {
				p := products[product.ProductID]
				dto.Products[j] = DeliveryPointProductDTO{
					Product:  NewProductDTO(p),
					Quantity: product.Quantity,
				}
				totalWeight += p.Weight * float64(product.Quantity)
				totalVolume += p.Volume() * float64(product.Quantity)
			}
		}
		points[i] = dto
	}

	canEdit := delivery.DeliveryDate.After(time.Now().AddDate(0, 0, 3))

	return DeliveryDTO{
		ID:             delivery.ID,
		DeliveryNumber: fmt.Sprintf("DEL-%d-%03d", delivery.DeliveryDate.Year(), delivery.ID),
		Courier:        courier,
		Vehicle:        vehicleDTO,
		CreatedBy:      NewUserDTO(createdBy),
		DeliveryDate:   delivery.DeliveryDate.Format("2006-01-02"),
		TimeStart:      delivery.TimeStart.Format("15:04:05"),
		TimeEnd:        delivery.TimeEnd.Format("15:04:05"),
		Status:         delivery.Status,
		CreatedAt:      delivery.CreatedAt,
		UpdatedAt:      delivery.UpdatedAt,
		DeliveryPoints: points,
		TotalWeight:    totalWeight,
		TotalVolume:    totalVolume,
		CanEdit:        canEdit,
	}
}

type CourierDeliveryResponse struct {
	ID             int64                 `json:"id"`
	DeliveryNumber string                `json:"deliveryNumber"`
	DeliveryDate   string                `json:"deliveryDate"`
	TimeStart      string                `json:"timeStart"`
	TimeEnd        string                `json:"timeEnd"`
	Status         domain.DeliveryStatus `json:"status"`
	Vehicle        VehicleSummary        `json:"vehicle"`
	PointsCount    int                   `json:"pointsCount"`
	ProductsCount  int                   `json:"productsCount"`
	TotalWeight    float64               `json:"totalWeight"`
}

type VehicleSummary struct {
	Brand        string `json:"brand"`
	LicensePlate string `json:"licensePlate"`
}

type GenerateDeliveriesResponse struct {
	TotalGenerated int                               `json:"totalGenerated"`
	ByDate         map[string]GenerationResultByDate `json:"byDate"`
}

type GenerationResultByDate struct {
	GeneratedCount int           `json:"generatedCount"`
	Deliveries     []DeliveryDTO `json:"deliveries"`
	Warnings       []string      `json:"warnings,omitempty"`
}
