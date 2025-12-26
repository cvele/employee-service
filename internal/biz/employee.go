package biz

import (
	"time"

	v1 "github.com/cvele/employee-service/api/employee/v1"

	"github.com/go-kratos/kratos/v2/errors"
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
	Page          int32
	PageSize      int32
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// ListResult represents paginated list result
type ListResult struct {
	Employees []*Employee
	Total     int64
}
