package service

import (
	"context"

	v1 "employee-service/api/employee/v1"
	"employee-service/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmployeeService is an employee service.
type EmployeeService struct {
	v1.UnimplementedEmployeeServiceServer

	uc *biz.EmployeeUsecase
}

// NewEmployeeService creates a new employee service.
func NewEmployeeService(uc *biz.EmployeeUsecase) *EmployeeService {
	return &EmployeeService{uc: uc}
}

// toProtoEmployee converts biz.Employee to proto Employee
func toProtoEmployee(e *biz.Employee) *v1.Employee {
	if e == nil {
		return nil
	}

	secondaryEmails := e.SecondaryEmails
	if secondaryEmails == nil {
		secondaryEmails = []string{}
	}

	return &v1.Employee{
		Id:              e.ID.String(),
		Email:           e.Email,
		SecondaryEmails: secondaryEmails,
		FirstName:       e.FirstName,
		LastName:        e.LastName,
		CreatedAt:       timestamppb.New(e.CreatedAt),
		UpdatedAt:       timestamppb.New(e.UpdatedAt),
	}
}

// CreateEmployee creates a new employee.
func (s *EmployeeService) CreateEmployee(ctx context.Context, req *v1.CreateEmployeeRequest) (*v1.CreateEmployeeResponse, error) {
	employee := &biz.Employee{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	created, err := s.uc.CreateEmployee(ctx, employee)
	if err != nil {
		return nil, err
	}

	return &v1.CreateEmployeeResponse{
		Employee: toProtoEmployee(created),
	}, nil
}

// UpdateEmployee updates an existing employee.
func (s *EmployeeService) UpdateEmployee(ctx context.Context, req *v1.UpdateEmployeeRequest) (*v1.UpdateEmployeeResponse, error) {
	// Parse UUID from string
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest("INVALID_UUID", "invalid employee ID format")
	}

	employee := &biz.Employee{
		ID: id,
	}
	
	// Handle optional fields (pointers from proto optional)
	if req.Email != nil {
		employee.Email = *req.Email
	}
	if req.FirstName != nil {
		employee.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		employee.LastName = *req.LastName
	}

	updated, err := s.uc.UpdateEmployee(ctx, employee)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateEmployeeResponse{
		Employee: toProtoEmployee(updated),
	}, nil
}

// DeleteEmployee deletes an employee.
func (s *EmployeeService) DeleteEmployee(ctx context.Context, req *v1.DeleteEmployeeRequest) (*v1.DeleteEmployeeResponse, error) {
	// Parse UUID from string
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest("INVALID_UUID", "invalid employee ID format")
	}

	err = s.uc.DeleteEmployee(ctx, id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteEmployeeResponse{
		Success: true,
	}, nil
}

// GetEmployee gets an employee by ID.
func (s *EmployeeService) GetEmployee(ctx context.Context, req *v1.GetEmployeeRequest) (*v1.GetEmployeeResponse, error) {
	// Parse UUID from string
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest("INVALID_UUID", "invalid employee ID format")
	}

	employee, err := s.uc.GetEmployee(ctx, id)
	if err != nil {
		return nil, err
	}

	return &v1.GetEmployeeResponse{
		Employee: toProtoEmployee(employee),
	}, nil
}

// GetEmployeeByEmail gets an employee by email.
func (s *EmployeeService) GetEmployeeByEmail(ctx context.Context, req *v1.GetEmployeeByEmailRequest) (*v1.GetEmployeeByEmailResponse, error) {
	employee, err := s.uc.GetEmployeeByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.GetEmployeeByEmailResponse{
		Employee: toProtoEmployee(employee),
	}, nil
}

// ListEmployees lists employees with pagination and filtering.
func (s *EmployeeService) ListEmployees(ctx context.Context, req *v1.ListEmployeesRequest) (*v1.ListEmployeesResponse, error) {
	filter := &biz.ListFilter{}
	
	// Handle optional pagination fields (default to 0, business logic applies defaults)
	if req.Page != nil {
		filter.Page = *req.Page
	}
	if req.PageSize != nil {
		filter.PageSize = *req.PageSize
	}

	if req.CreatedAfter != nil {
		t := req.CreatedAfter.AsTime()
		filter.CreatedAfter = &t
	}
	if req.CreatedBefore != nil {
		t := req.CreatedBefore.AsTime()
		filter.CreatedBefore = &t
	}

	result, err := s.uc.ListEmployees(ctx, filter)
	if err != nil {
		return nil, err
	}

	employees := make([]*v1.Employee, len(result.Employees))
	for i, e := range result.Employees {
		employees[i] = toProtoEmployee(e)
	}

	return &v1.ListEmployeesResponse{
		Employees: employees,
		Total:     result.Total,
		Page:      filter.Page,     // Return actual page used (after defaults)
		PageSize:  filter.PageSize, // Return actual page_size used (after defaults)
	}, nil
}

// MergeEmployees merges two employees by email.
func (s *EmployeeService) MergeEmployees(ctx context.Context, req *v1.MergeEmployeesRequest) (*v1.MergeEmployeesResponse, error) {
	employee, err := s.uc.MergeEmployees(ctx, req.PrimaryEmail, req.SecondaryEmail)
	if err != nil {
		return nil, err
	}

	return &v1.MergeEmployeesResponse{
		Employee: toProtoEmployee(employee),
	}, nil
}

