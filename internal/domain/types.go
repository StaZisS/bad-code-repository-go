package domain

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleCourier UserRole = "courier"
)

func (r UserRole) Valid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleCourier:
		return true
	default:
		return false
	}
}

type DeliveryStatus string

const (
	StatusPlanned    DeliveryStatus = "planned"
	StatusInProgress DeliveryStatus = "in_progress"
	StatusCompleted  DeliveryStatus = "completed"
	StatusCancelled  DeliveryStatus = "cancelled"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Name         string
	Role         UserRole
	CreatedAt    time.Time
}

type Product struct {
	ID     int64
	Name   string
	Weight float64
	Length float64
	Width  float64
	Height float64
}

func (p Product) Volume() float64 {
	return (p.Length * p.Width * p.Height) / 1_000_000
}

type Vehicle struct {
	ID           int64
	Brand        string
	LicensePlate string
	MaxWeight    float64
	MaxVolume    float64
}

type Delivery struct {
	ID             int64
	CourierID      int64
	VehicleID      int64
	CreatedByID    int64
	DeliveryDate   time.Time
	TimeStart      time.Time
	TimeEnd        time.Time
	Status         DeliveryStatus
	DeliveryPoints []DeliveryPoint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type DeliveryPoint struct {
	ID        int64
	Sequence  int
	Latitude  float64
	Longitude float64
	Products  []DeliveryPointProduct
}

type DeliveryPointProduct struct {
	ProductID int64
	Quantity  int
}
