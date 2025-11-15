package service

import (
	"strings"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

type VehicleService struct {
	store *storage.MemoryStore
}

func NewVehicleService(store *storage.MemoryStore) *VehicleService {
	return &VehicleService{store: store}
}

type VehicleInput struct {
	Brand        string
	LicensePlate string
	MaxWeight    float64
	MaxVolume    float64
}

func (s *VehicleService) ListVehicles() []VehicleDTO {
	vehicles := s.store.ListVehicles()
	res := make([]VehicleDTO, len(vehicles))
	for i, vehicle := range vehicles {
		res[i] = NewVehicleDTO(vehicle)
	}
	return res
}

func (s *VehicleService) CreateVehicle(input VehicleInput) (VehicleDTO, error) {
	if errs := validateVehicle(input); len(errs) > 0 {
		return VehicleDTO{}, ValidationErrors(errs)
	}
	vehicle, err := s.store.CreateVehicle(domain.Vehicle{
		Brand:        strings.TrimSpace(input.Brand),
		LicensePlate: strings.TrimSpace(input.LicensePlate),
		MaxWeight:    input.MaxWeight,
		MaxVolume:    input.MaxVolume,
	})
	if err != nil {
		if err == storage.ErrConflict {
			return VehicleDTO{}, ValidationErrors{{
				Field:   "license_plate",
				Message: "vehicle with this license plate already exists",
			}}
		}
		return VehicleDTO{}, err
	}
	return NewVehicleDTO(vehicle), nil
}

func (s *VehicleService) UpdateVehicle(id int64, input VehicleInput) (VehicleDTO, error) {
	if errs := validateVehicle(input); len(errs) > 0 {
		return VehicleDTO{}, ValidationErrors(errs)
	}
	existing, ok := s.store.GetVehicle(id)
	if !ok {
		return VehicleDTO{}, ErrNotFound
	}
	existing.Brand = strings.TrimSpace(input.Brand)
	existing.LicensePlate = strings.TrimSpace(input.LicensePlate)
	existing.MaxWeight = input.MaxWeight
	existing.MaxVolume = input.MaxVolume
	updated, err := s.store.UpdateVehicle(*existing)
	if err != nil {
		if err == storage.ErrConflict {
			return VehicleDTO{}, ValidationErrors{{
				Field:   "license_plate",
				Message: "vehicle with this license plate already exists",
			}}
		}
		if err == storage.ErrNotFound {
			return VehicleDTO{}, ErrNotFound
		}
		return VehicleDTO{}, err
	}
	return NewVehicleDTO(updated), nil
}

func (s *VehicleService) DeleteVehicle(id int64) error {
	if err := s.store.DeleteVehicle(id); err != nil {
		if err == storage.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func validateVehicle(input VehicleInput) []ValidationError {
	var errs []ValidationError
	if strings.TrimSpace(input.Brand) == "" {
		errs = append(errs, ValidationError{Field: "brand", Message: "brand is required"})
	}
	if strings.TrimSpace(input.LicensePlate) == "" {
		errs = append(errs, ValidationError{Field: "license_plate", Message: "license plate is required"})
	}
	if input.MaxWeight <= 0 {
		errs = append(errs, ValidationError{Field: "max_weight", Message: "max weight must be positive"})
	}
	if input.MaxVolume <= 0 {
		errs = append(errs, ValidationError{Field: "max_volume", Message: "max volume must be positive"})
	}
	return errs
}
