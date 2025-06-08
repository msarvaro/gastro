package service

import (
	"context"
	"fmt"
	"log"
	"restaurant-management/internal/domain/supplier"
)

type SupplierService struct {
	supplierRepo supplier.Repository
}

func NewSupplierService(supplierRepo supplier.Repository) supplier.Service {
	return &SupplierService{
		supplierRepo: supplierRepo,
	}
}

func (s *SupplierService) GetAll(ctx context.Context, businessID int) ([]supplier.Supplier, error) {
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	suppliers, err := s.supplierRepo.GetAll(ctx, businessID)
	if err != nil {
		log.Printf("Error retrieving suppliers for business %d: %v", businessID, err)
		return nil, err
	}

	if suppliers == nil {
		suppliers = []supplier.Supplier{}
	}

	return suppliers, nil
}

func (s *SupplierService) GetByID(ctx context.Context, id int, businessID int) (*supplier.Supplier, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid supplier ID: %d", id)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	supplier, err := s.supplierRepo.GetByID(ctx, id, businessID)
	if err != nil {
		log.Printf("Error retrieving supplier %d for business %d: %v", id, businessID, err)
		return nil, err
	}

	return supplier, nil
}

func (s *SupplierService) Create(ctx context.Context, supplierReq supplier.CreateSupplierRequest, businessID int) (*supplier.Supplier, error) {
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Validate required fields
	if supplierReq.Name == "" {
		return nil, fmt.Errorf("supplier name is required")
	}
	if len(supplierReq.Categories) == 0 {
		return nil, fmt.Errorf("supplier categories are required")
	}

	// Set default status if not provided
	if supplierReq.Status == "" {
		supplierReq.Status = "active"
	}

	createdSupplier, err := s.supplierRepo.Create(ctx, supplierReq, businessID)
	if err != nil {
		log.Printf("Error creating supplier for business %d: %v", businessID, err)
		return nil, err
	}

	log.Printf("Successfully created supplier %d (%s) for business %d", createdSupplier.ID, createdSupplier.Name, businessID)
	return createdSupplier, nil
}

func (s *SupplierService) Update(ctx context.Context, id int, supplierReq supplier.UpdateSupplierRequest, businessID int) (*supplier.Supplier, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid supplier ID: %d", id)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Validate that at least one field is being updated
	if supplierReq.Name == "" && len(supplierReq.Categories) == 0 && 
		supplierReq.Phone == "" && supplierReq.Email == "" && 
		supplierReq.Address == "" && supplierReq.Status == "" {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	updatedSupplier, err := s.supplierRepo.Update(ctx, id, supplierReq, businessID)
	if err != nil {
		log.Printf("Error updating supplier %d for business %d: %v", id, businessID, err)
		return nil, err
	}

	log.Printf("Successfully updated supplier %d for business %d", id, businessID)
	return updatedSupplier, nil
}

func (s *SupplierService) Delete(ctx context.Context, id int, businessID int) error {
	if id <= 0 {
		return fmt.Errorf("invalid supplier ID: %d", id)
	}
	if businessID <= 0 {
		return fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Check if supplier exists before deleting
	_, err := s.supplierRepo.GetByID(ctx, id, businessID)
	if err != nil {
		log.Printf("Supplier %d not found for business %d: %v", id, businessID, err)
		return fmt.Errorf("supplier not found")
	}

	err = s.supplierRepo.Delete(ctx, id, businessID)
	if err != nil {
		log.Printf("Error deleting supplier %d for business %d: %v", id, businessID, err)
		return err
	}

	log.Printf("Successfully deleted supplier %d for business %d", id, businessID)
	return nil
} 