package biz

import (
	"context"
	"time"

	v1 "employee-service/api/employee/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

var (
	// ErrEmployeeNotFound is employee not found.
	ErrEmployeeNotFound = errors.NotFound(v1.ErrorReason_EMPLOYEE_NOT_FOUND.String(), "employee not found")
	// ErrEmployeeAlreadyExists is employee already exists.
	ErrEmployeeAlreadyExists = errors.BadRequest(v1.ErrorReason_EMPLOYEE_ALREADY_EXISTS.String(), "employee already exists")
	// ErrInvalidEmail is invalid email format.
	ErrInvalidEmail = errors.BadRequest(v1.ErrorReason_INVALID_EMAIL.String(), "invalid email format")
	// ErrInvalidEmployeeID is invalid employee ID.
	ErrInvalidEmployeeID = errors.BadRequest(v1.ErrorReason_INVALID_EMPLOYEE_ID.String(), "invalid employee ID")
	// ErrInvalidDateRange is invalid date range.
	ErrInvalidDateRange = errors.BadRequest(v1.ErrorReason_INVALID_DATE_RANGE.String(), "created_after must be before created_before")
	// ErrInvalidMerge is invalid merge request.
	ErrInvalidMerge = errors.BadRequest(v1.ErrorReason_INVALID_MERGE.String(), "primary and secondary emails must be different")
)

// Employee is an Employee domain model.
type Employee struct {
	ID        uuid.UUID
	TenantID  string
	Emails    []string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListFilter represents filtering options for listing employees
type ListFilter struct {
	Page         int32
	PageSize     int32
	CreatedAfter *time.Time
	CreatedBefore *time.Time
}

// ListResult represents paginated list result
type ListResult struct {
	Employees []*Employee
	Total     int64
}

// EventPublisher defines the interface for publishing events using Protocol Buffers
type EventPublisher interface {
	PublishEmployeeCreated(ctx context.Context, tenantID, userID string, employee *Employee) error
	PublishEmployeeUpdated(ctx context.Context, tenantID, userID string, employee *Employee, updatedFields []string) error
	PublishEmployeeDeleted(ctx context.Context, tenantID, userID string, employee *Employee) error
	PublishEmployeeMerged(ctx context.Context, tenantID, userID string, employee *Employee, mergedFromEmail string) error
}

// EmployeeRepo is an Employee repository interface.
type EmployeeRepo interface {
	Create(ctx context.Context, tenantID string, employee *Employee) (*Employee, error)
	Update(ctx context.Context, tenantID string, employee *Employee) (*Employee, error)
	Delete(ctx context.Context, tenantID string, id uuid.UUID) error
	GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*Employee, error)
	GetByEmail(ctx context.Context, tenantID string, email string) (*Employee, error)
	List(ctx context.Context, tenantID string, filter *ListFilter) (*ListResult, error)
	CheckEmailExists(ctx context.Context, tenantID string, email string) (bool, error)
	MergeEmployees(ctx context.Context, tenantID string, primaryEmail string, secondaryEmail string) (*Employee, error)
	GetEventPublisher() EventPublisher
}

// EmployeeUsecase is an Employee usecase.
type EmployeeUsecase struct {
	repo EmployeeRepo
	log  *log.Helper
}

// NewEmployeeUsecase creates a new Employee usecase.
func NewEmployeeUsecase(repo EmployeeRepo, logger log.Logger) *EmployeeUsecase {
	return &EmployeeUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateEmployee creates a new employee after checking email uniqueness within tenant.
func (uc *EmployeeUsecase) CreateEmployee(ctx context.Context, employee *Employee) (*Employee, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// Validate at least one email is provided
	if len(employee.Emails) == 0 {
		return nil, ErrInvalidEmail
	}

	uc.log.WithContext(ctx).Infof("CreateEmployee: tenant=%s, emails=%v", tenantID, employee.Emails)

	// Check if any email already exists in this tenant
	for _, email := range employee.Emails {
		exists, err := uc.repo.CheckEmailExists(ctx, tenantID, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmployeeAlreadyExists
		}
	}

	// Set tenant ID
	employee.TenantID = tenantID

	created, err := uc.repo.Create(ctx, tenantID, employee)
	if err != nil {
		return nil, err
	}

	// Publish event (best-effort)
	userID, _ := GetUserID(ctx)
	if publisher := uc.repo.GetEventPublisher(); publisher != nil {
		if err := publisher.PublishEmployeeCreated(ctx, tenantID, userID, created); err != nil {
			uc.log.Warnf("failed to publish employee.created event: %v", err)
		}
	}

	return created, nil
}

// UpdateEmployee updates an existing employee within tenant.
func (uc *EmployeeUsecase) UpdateEmployee(ctx context.Context, employee *Employee) (*Employee, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("UpdateEmployee: tenant=%s, id=%s", tenantID, employee.ID)

	// Verify employee exists in this tenant
	existing, err := uc.repo.GetByID(ctx, tenantID, employee.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrEmployeeNotFound
	}

	// Track which fields are being updated
	updatedFields := []string{}
	
	// Check if emails are being updated
	if len(employee.Emails) > 0 {
		// Check uniqueness for any new emails
		for _, email := range employee.Emails {
			// Skip if email already belongs to this employee
			alreadyOwned := false
			for _, existingEmail := range existing.Emails {
				if email == existingEmail {
					alreadyOwned = true
					break
				}
			}
			if alreadyOwned {
				continue
			}

			// Check if email exists for another employee
			exists, err := uc.repo.CheckEmailExists(ctx, tenantID, email)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrEmployeeAlreadyExists
			}
		}
		updatedFields = append(updatedFields, "emails")
	}

	if employee.FirstName != "" && employee.FirstName != existing.FirstName {
		updatedFields = append(updatedFields, "first_name")
	}
	if employee.LastName != "" && employee.LastName != existing.LastName {
		updatedFields = append(updatedFields, "last_name")
	}

	// Set tenant ID
	employee.TenantID = tenantID

	updated, err := uc.repo.Update(ctx, tenantID, employee)
	if err != nil {
		return nil, err
	}

	// Publish event (best-effort)
	userID, _ := GetUserID(ctx)
	if publisher := uc.repo.GetEventPublisher(); publisher != nil {
		if err := publisher.PublishEmployeeUpdated(ctx, tenantID, userID, updated, updatedFields); err != nil {
			uc.log.Warnf("failed to publish employee.updated event: %v", err)
		}
	}

	return updated, nil
}

// DeleteEmployee deletes an employee within tenant.
func (uc *EmployeeUsecase) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return err
	}

	uc.log.WithContext(ctx).Infof("DeleteEmployee: tenant=%s, id=%s", tenantID, id)

	// Verify employee exists in this tenant
	existing, err := uc.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrEmployeeNotFound
	}

	err = uc.repo.Delete(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Publish event with deleted employee info (best-effort)
	userID, _ := GetUserID(ctx)
	if publisher := uc.repo.GetEventPublisher(); publisher != nil {
		if err := publisher.PublishEmployeeDeleted(ctx, tenantID, userID, existing); err != nil {
			uc.log.Warnf("failed to publish employee.deleted event: %v", err)
		}
	}

	return nil
}

// GetEmployee gets an employee by ID within tenant.
func (uc *EmployeeUsecase) GetEmployee(ctx context.Context, id uuid.UUID) (*Employee, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("GetEmployee: tenant=%s, id=%s", tenantID, id)

	employee, err := uc.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	return employee, nil
}

// GetEmployeeByEmail gets an employee by email within tenant.
func (uc *EmployeeUsecase) GetEmployeeByEmail(ctx context.Context, email string) (*Employee, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("GetEmployeeByEmail: tenant=%s, email=%s", tenantID, email)

	employee, err := uc.repo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	return employee, nil
}

// ListEmployees lists employees with pagination and filtering within tenant.
func (uc *EmployeeUsecase) ListEmployees(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("ListEmployees: tenant=%s, page=%d, size=%d", tenantID, filter.Page, filter.PageSize)

	// Set default pagination values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// Business validation: date range check
	if filter.CreatedAfter != nil && filter.CreatedBefore != nil {
		if filter.CreatedAfter.After(*filter.CreatedBefore) {
			return nil, ErrInvalidDateRange
		}
	}

	return uc.repo.List(ctx, tenantID, filter)
}

// MergeEmployees merges two employees by email within tenant.
// All emails from the secondary employee are transferred to the primary employee.
func (uc *EmployeeUsecase) MergeEmployees(ctx context.Context, primaryEmail string, secondaryEmail string) (*Employee, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// Business validation: emails must be different
	if primaryEmail == secondaryEmail {
		return nil, ErrInvalidMerge
	}

	uc.log.WithContext(ctx).Infof("MergeEmployees: tenant=%s, primary=%s, secondary=%s", tenantID, primaryEmail, secondaryEmail)

	// Validate both emails exist in this tenant
	primary, err := uc.repo.GetByEmail(ctx, tenantID, primaryEmail)
	if err != nil {
		return nil, err
	}
	if primary == nil {
		return nil, errors.BadRequest("PRIMARY_NOT_FOUND", "primary employee not found")
	}

	secondary, err := uc.repo.GetByEmail(ctx, tenantID, secondaryEmail)
	if err != nil {
		return nil, err
	}
	if secondary == nil {
		return nil, errors.BadRequest("SECONDARY_NOT_FOUND", "secondary employee not found")
	}

	// Cannot merge the same employee
	if primary.ID == secondary.ID {
		return nil, errors.BadRequest("CANNOT_MERGE_SAME", "cannot merge employee with itself")
	}

	merged, err := uc.repo.MergeEmployees(ctx, tenantID, primaryEmail, secondaryEmail)
	if err != nil {
		return nil, err
	}

	// Publish event with merge information (best-effort)
	userID, _ := GetUserID(ctx)
	if publisher := uc.repo.GetEventPublisher(); publisher != nil {
		if err := publisher.PublishEmployeeMerged(ctx, tenantID, userID, merged, secondaryEmail); err != nil {
			uc.log.Warnf("failed to publish employee.merged event: %v", err)
		}
	}

	return merged, nil
}

