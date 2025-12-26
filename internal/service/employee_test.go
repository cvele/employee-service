package service

import (
	"context"
	"testing"
	"time"

	v1 "github.com/cvele/employee-service/api/employee/v1"
	"github.com/cvele/employee-service/internal/biz"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewEmployeeService(t *testing.T) {
	// Create a minimal usecase (nil is ok for this test)
	uc := &biz.EmployeeUsecase{}
	service := NewEmployeeService(uc)
	
	assert.NotNil(t, service)
	assert.NotNil(t, service.uc)
}

func TestToProtoEmployee(t *testing.T) {
	t.Run("nil employee", func(t *testing.T) {
		result := toProtoEmployee(nil)
		assert.Nil(t, result)
	})

	t.Run("valid employee", func(t *testing.T) {
		now := time.Now()
		id := uuid.New()
		employee := &biz.Employee{
			ID:        id,
			Emails:    []string{"test@example.com", "secondary@example.com"},
			FirstName: "John",
			LastName:  "Doe",
			CreatedAt: now,
			UpdatedAt: now,
		}

		result := toProtoEmployee(employee)
		
		assert.NotNil(t, result)
		assert.Equal(t, id.String(), result.Id)
		assert.Equal(t, []string{"test@example.com", "secondary@example.com"}, result.Emails)
		assert.Equal(t, "John", result.FirstName)
		assert.Equal(t, "Doe", result.LastName)
		assert.NotNil(t, result.CreatedAt)
		assert.NotNil(t, result.UpdatedAt)
	})

	t.Run("employee with nil emails", func(t *testing.T) {
		employee := &biz.Employee{
			ID:        uuid.New(),
			Emails:    nil,
			FirstName: "Jane",
			LastName:  "Smith",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := toProtoEmployee(employee)
		
		assert.NotNil(t, result)
		assert.NotNil(t, result.Emails)
		assert.Empty(t, result.Emails)
	})
}

func TestUpdateEmployee_UUIDValidation(t *testing.T) {
	uc := &biz.EmployeeUsecase{}
	service := NewEmployeeService(uc)

	firstName := "Jane"
	
	// Test invalid UUID
	resp, err := service.UpdateEmployee(context.Background(), &v1.UpdateEmployeeRequest{
		Id:        "invalid-uuid",
		FirstName: &firstName,
	})
	
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "INVALID_UUID")
}

func TestDeleteEmployee_UUIDValidation(t *testing.T) {
	uc := &biz.EmployeeUsecase{}
	service := NewEmployeeService(uc)

	// Test invalid UUID
	resp, err := service.DeleteEmployee(context.Background(), &v1.DeleteEmployeeRequest{
		Id: "invalid-uuid",
	})
	
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "INVALID_UUID")
}

func TestGetEmployee_UUIDValidation(t *testing.T) {
	uc := &biz.EmployeeUsecase{}
	service := NewEmployeeService(uc)

	// Test invalid UUID
	resp, err := service.GetEmployee(context.Background(), &v1.GetEmployeeRequest{
		Id: "invalid-uuid",
	})
	
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "INVALID_UUID")
}

func TestListEmployees_FilterHandling(t *testing.T) {
	t.Run("handles nil pagination", func(t *testing.T) {
		service := &EmployeeService{uc: &biz.EmployeeUsecase{}}
		
		// This should not panic - pagination fields are optional
		req := &v1.ListEmployeesRequest{
			Page:     nil,
			PageSize: nil,
		}
		
		// We expect it to fail (no repo configured) but not panic
		_, err := service.ListEmployees(context.Background(), req)
		// Error is expected because there's no real usecase/repo configured
		_ = err
	})

	t.Run("handles pagination values", func(t *testing.T) {
		service := &EmployeeService{uc: &biz.EmployeeUsecase{}}
		
		page := int32(1)
		pageSize := int32(20)
		
		req := &v1.ListEmployeesRequest{
			Page:     &page,
			PageSize: &pageSize,
		}
		
		// We expect it to fail (no repo configured) but not panic
		_, err := service.ListEmployees(context.Background(), req)
		// Error is expected because there's no real usecase/repo configured
		_ = err
	})
}

func TestEmployeeServiceMethods_ExistenceCheck(t *testing.T) {
	// This test verifies that all required methods exist on the service
	// and implements the required interface
	service := &EmployeeService{}
	
	var _ v1.EmployeeServiceServer = service
	
	// Verify methods can be called (even if they fail)
	ctx := context.Background()
	
	// CreateEmployee
	_, err := service.CreateEmployee(ctx, &v1.CreateEmployeeRequest{})
	_ = err // Expected to fail with no uc configured
	
	// UpdateEmployee
	_, err = service.UpdateEmployee(ctx, &v1.UpdateEmployeeRequest{})
	_ = err
	
	// DeleteEmployee
	_, err = service.DeleteEmployee(ctx, &v1.DeleteEmployeeRequest{})
	_ = err
	
	// GetEmployee
	_, err = service.GetEmployee(ctx, &v1.GetEmployeeRequest{})
	_ = err
	
	// GetEmployeeByEmail
	_, err = service.GetEmployeeByEmail(ctx, &v1.GetEmployeeByEmailRequest{})
	_ = err
	
	// ListEmployees
	_, err = service.ListEmployees(ctx, &v1.ListEmployeesRequest{})
	_ = err
	
	// MergeEmployees
	_, err = service.MergeEmployees(ctx, &v1.MergeEmployeesRequest{})
	_ = err
}
