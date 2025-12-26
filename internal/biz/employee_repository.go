package biz

import (
	"context"

	"github.com/google/uuid"
)

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
