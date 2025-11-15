package service

import (
	"strings"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"
)

type ProductService struct {
	store *storage.MemoryStore
}

func NewProductService(store *storage.MemoryStore) *ProductService {
	return &ProductService{store: store}
}

type ProductInput struct {
	Name   string
	Weight float64
	Length float64
	Width  float64
	Height float64
}

func (s *ProductService) ListProducts() []ProductDTO {
	products := s.store.ListProducts()
	res := make([]ProductDTO, len(products))
	for i, product := range products {
		res[i] = NewProductDTO(product)
	}
	return res
}

func (s *ProductService) CreateProduct(input ProductInput) (ProductDTO, error) {
	if errs := validateProduct(input); len(errs) > 0 {
		return ProductDTO{}, ValidationErrors(errs)
	}
	product, err := s.store.CreateProduct(domain.Product{
		Name:   strings.TrimSpace(input.Name),
		Weight: input.Weight,
		Length: input.Length,
		Width:  input.Width,
		Height: input.Height,
	})
	if err != nil {
		return ProductDTO{}, err
	}
	return NewProductDTO(product), nil
}

func (s *ProductService) UpdateProduct(id int64, input ProductInput) (ProductDTO, error) {
	if errs := validateProduct(input); len(errs) > 0 {
		return ProductDTO{}, ValidationErrors(errs)
	}
	existing, ok := s.store.GetProduct(id)
	if !ok {
		return ProductDTO{}, ErrNotFound
	}
	existing.Name = strings.TrimSpace(input.Name)
	existing.Weight = input.Weight
	existing.Length = input.Length
	existing.Width = input.Width
	existing.Height = input.Height
	updated, err := s.store.UpdateProduct(*existing)
	if err != nil {
		if err == storage.ErrNotFound {
			return ProductDTO{}, ErrNotFound
		}
		return ProductDTO{}, err
	}
	return NewProductDTO(updated), nil
}

func (s *ProductService) DeleteProduct(id int64) error {
	if err := s.store.DeleteProduct(id); err != nil {
		if err == storage.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func validateProduct(input ProductInput) []ValidationError {
	var errs []ValidationError
	if strings.TrimSpace(input.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	}
	if input.Weight <= 0 {
		errs = append(errs, ValidationError{Field: "weight", Message: "weight must be positive"})
	}
	if input.Length <= 0 {
		errs = append(errs, ValidationError{Field: "length", Message: "length must be positive"})
	}
	if input.Width <= 0 {
		errs = append(errs, ValidationError{Field: "width", Message: "width must be positive"})
	}
	if input.Height <= 0 {
		errs = append(errs, ValidationError{Field: "height", Message: "height must be positive"})
	}
	return errs
}
