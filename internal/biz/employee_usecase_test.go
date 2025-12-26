package biz

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEmployeeRepo is a mock implementation of EmployeeRepo
type MockEmployeeRepo struct {
	mock.Mock
}

func (m *MockEmployeeRepo) Create(ctx context.Context, tenantID string, employee *Employee) (*Employee, error) {
	args := m.Called(ctx, tenantID, employee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockEmployeeRepo) Update(ctx context.Context, tenantID string, employee *Employee) (*Employee, error) {
	args := m.Called(ctx, tenantID, employee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockEmployeeRepo) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockEmployeeRepo) GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*Employee, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockEmployeeRepo) GetByEmail(ctx context.Context, tenantID string, email string) (*Employee, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockEmployeeRepo) CheckEmailExists(ctx context.Context, tenantID string, email string) (bool, error) {
	args := m.Called(ctx, tenantID, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockEmployeeRepo) List(ctx context.Context, tenantID string, filter *ListFilter) (*ListResult, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ListResult), args.Error(1)
}

func (m *MockEmployeeRepo) MergeEmployees(ctx context.Context, tenantID string, primaryEmail string, secondaryEmail string) (*Employee, error) {
	args := m.Called(ctx, tenantID, primaryEmail, secondaryEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockEmployeeRepo) GetEventPublisher() EventPublisher {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(EventPublisher)
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishEmployeeCreated(ctx context.Context, tenantID, userID string, employee *Employee) error {
	args := m.Called(ctx, tenantID, userID, employee)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishEmployeeUpdated(ctx context.Context, tenantID, userID string, employee *Employee, updatedFields []string) error {
	args := m.Called(ctx, tenantID, userID, employee, updatedFields)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishEmployeeDeleted(ctx context.Context, tenantID, userID string, employee *Employee) error {
	args := m.Called(ctx, tenantID, userID, employee)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishEmployeeMerged(ctx context.Context, tenantID, userID string, employee *Employee, mergedFromEmail string) error {
	args := m.Called(ctx, tenantID, userID, employee, mergedFromEmail)
	return args.Error(0)
}

func setupUsecase() (*EmployeeUsecase, *MockEmployeeRepo) {
	repo := new(MockEmployeeRepo)
	// Create a simple no-op logger with io.Discard
	logger := log.NewHelper(log.NewStdLogger(io.Discard))
	uc := &EmployeeUsecase{
		repo: repo,
		log:  logger,
	}
	return uc, repo
}

func TestNewEmployeeUsecase(t *testing.T) {
	repo := new(MockEmployeeRepo)
	logger := log.NewStdLogger(io.Discard)
	uc := NewEmployeeUsecase(repo, logger)
	
	assert.NotNil(t, uc)
	assert.NotNil(t, uc.repo)
	assert.NotNil(t, uc.log)
}

func TestCreateEmployee(t *testing.T) {
	tests := []struct {
		name        string
		employee    *Employee
		setupMock   func(*MockEmployeeRepo, *MockEventPublisher)
		wantErr     bool
		errExpected error
	}{
		{
			name: "successful creation",
			employee: &Employee{
				Emails:    []string{"test@example.com"},
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "test@example.com").Return(false, nil)
				created := &Employee{
					ID:        uuid.New(),
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repo.On("Create", mock.Anything, "tenant-123", mock.Anything).Return(created, nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeCreated", mock.Anything, "tenant-123", "user-456", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "no emails provided",
			employee: &Employee{
				Emails:    []string{},
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				// No expectations - should fail early
			},
			wantErr:     true,
			errExpected: ErrInvalidEmail,
		},
		{
			name: "email already exists",
			employee: &Employee{
				Emails:    []string{"existing@example.com"},
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "existing@example.com").Return(true, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeAlreadyExists,
		},
		{
			name: "repository error",
			employee: &Employee{
				Emails:    []string{"test@example.com"},
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "test@example.com").Return(false, nil)
				repo.On("Create", mock.Anything, "tenant-123", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "event publish failure (non-fatal)",
			employee: &Employee{
				Emails:    []string{"test@example.com"},
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "test@example.com").Return(false, nil)
				created := &Employee{
					ID:        uuid.New(),
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repo.On("Create", mock.Anything, "tenant-123", mock.Anything).Return(created, nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeCreated", mock.Anything, "tenant-123", "user-456", mock.Anything).Return(errors.New("event error"))
			},
			wantErr: false, // Event publish errors are non-fatal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			pub := new(MockEventPublisher)
			
			if tt.setupMock != nil {
				tt.setupMock(repo, pub)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")
			ctx = WithUserID(ctx, "user-456")

			result, err := uc.CreateEmployee(ctx, tt.employee)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "tenant-123", result.TenantID)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	existingID := uuid.New()
	
	tests := []struct {
		name        string
		employee    *Employee
		setupMock   func(*MockEmployeeRepo, *MockEventPublisher)
		wantErr     bool
		errExpected error
	}{
		{
			name: "successful update",
			employee: &Employee{
				ID:        existingID,
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				existing := &Employee{
					ID:        existingID,
					Emails:    []string{"old@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", existingID).Return(existing, nil)
				
				updated := &Employee{
					ID:        existingID,
					Emails:    []string{"old@example.com"},
					FirstName: "Jane",
					LastName:  "Smith",
					TenantID:  "tenant-123",
					UpdatedAt: time.Now(),
				}
				repo.On("Update", mock.Anything, "tenant-123", mock.Anything).Return(updated, nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeUpdated", mock.Anything, "tenant-123", "user-456", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "employee not found",
			employee: &Employee{
				ID: existingID,
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, "tenant-123", existingID).Return(nil, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeNotFound,
		},
		{
			name: "update with new email",
			employee: &Employee{
				ID:     existingID,
				Emails: []string{"new@example.com"},
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				existing := &Employee{
					ID:        existingID,
					Emails:    []string{"old@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", existingID).Return(existing, nil)
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "new@example.com").Return(false, nil)
				
				updated := &Employee{
					ID:        existingID,
					Emails:    []string{"new@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
					UpdatedAt: time.Now(),
				}
				repo.On("Update", mock.Anything, "tenant-123", mock.Anything).Return(updated, nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeUpdated", mock.Anything, "tenant-123", "user-456", mock.Anything, []string{"emails"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "email already exists for another employee",
			employee: &Employee{
				ID:     existingID,
				Emails: []string{"taken@example.com"},
			},
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				existing := &Employee{
					ID:        existingID,
					Emails:    []string{"old@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", existingID).Return(existing, nil)
				repo.On("CheckEmailExists", mock.Anything, "tenant-123", "taken@example.com").Return(true, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			pub := new(MockEventPublisher)
			
			if tt.setupMock != nil {
				tt.setupMock(repo, pub)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")
			ctx = WithUserID(ctx, "user-456")

			result, err := uc.UpdateEmployee(ctx, tt.employee)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestDeleteEmployee(t *testing.T) {
	employeeID := uuid.New()
	
	tests := []struct {
		name        string
		id          uuid.UUID
		setupMock   func(*MockEmployeeRepo, *MockEventPublisher)
		wantErr     bool
		errExpected error
	}{
		{
			name: "successful deletion",
			id:   employeeID,
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				existing := &Employee{
					ID:        employeeID,
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", employeeID).Return(existing, nil)
				repo.On("Delete", mock.Anything, "tenant-123", employeeID).Return(nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeDeleted", mock.Anything, "tenant-123", "user-456", existing).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "employee not found",
			id:   employeeID,
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, "tenant-123", employeeID).Return(nil, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeNotFound,
		},
		{
			name: "repository error",
			id:   employeeID,
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				existing := &Employee{
					ID:        employeeID,
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", employeeID).Return(existing, nil)
				repo.On("Delete", mock.Anything, "tenant-123", employeeID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			pub := new(MockEventPublisher)
			
			if tt.setupMock != nil {
				tt.setupMock(repo, pub)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")
			ctx = WithUserID(ctx, "user-456")

			err := uc.DeleteEmployee(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestGetEmployee(t *testing.T) {
	employeeID := uuid.New()
	
	tests := []struct {
		name        string
		id          uuid.UUID
		setupMock   func(*MockEmployeeRepo)
		wantErr     bool
		errExpected error
	}{
		{
			name: "successful get",
			id:   employeeID,
			setupMock: func(repo *MockEmployeeRepo) {
				employee := &Employee{
					ID:        employeeID,
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByID", mock.Anything, "tenant-123", employeeID).Return(employee, nil)
			},
			wantErr: false,
		},
		{
			name: "employee not found",
			id:   employeeID,
			setupMock: func(repo *MockEmployeeRepo) {
				repo.On("GetByID", mock.Anything, "tenant-123", employeeID).Return(nil, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")

			result, err := uc.GetEmployee(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestGetEmployeeByEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		setupMock   func(*MockEmployeeRepo)
		wantErr     bool
		errExpected error
	}{
		{
			name:  "successful get",
			email: "test@example.com",
			setupMock: func(repo *MockEmployeeRepo) {
				employee := &Employee{
					ID:        uuid.New(),
					Emails:    []string{"test@example.com"},
					FirstName: "John",
					LastName:  "Doe",
					TenantID:  "tenant-123",
				}
				repo.On("GetByEmail", mock.Anything, "tenant-123", "test@example.com").Return(employee, nil)
			},
			wantErr: false,
		},
		{
			name:  "employee not found",
			email: "notfound@example.com",
			setupMock: func(repo *MockEmployeeRepo) {
				repo.On("GetByEmail", mock.Anything, "tenant-123", "notfound@example.com").Return(nil, nil)
			},
			wantErr:     true,
			errExpected: ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")

			result, err := uc.GetEmployeeByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestListEmployees(t *testing.T) {
	now := time.Now()
	before := now.Add(-24 * time.Hour)
	after := now.Add(24 * time.Hour)
	
	tests := []struct {
		name        string
		filter      *ListFilter
		setupMock   func(*MockEmployeeRepo)
		wantErr     bool
		errExpected error
	}{
		{
			name: "successful list with defaults",
			filter: &ListFilter{},
			setupMock: func(repo *MockEmployeeRepo) {
				result := &ListResult{
					Employees: []*Employee{
						{ID: uuid.New(), Emails: []string{"test1@example.com"}},
						{ID: uuid.New(), Emails: []string{"test2@example.com"}},
					},
					Total: 2,
				}
				repo.On("List", mock.Anything, "tenant-123", mock.Anything).Return(result, nil)
			},
			wantErr: false,
		},
		{
			name: "list with pagination",
			filter: &ListFilter{
				Page:     2,
				PageSize: 10,
			},
			setupMock: func(repo *MockEmployeeRepo) {
				result := &ListResult{
					Employees: []*Employee{},
					Total:     0,
				}
				repo.On("List", mock.Anything, "tenant-123", mock.Anything).Return(result, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid date range",
			filter: &ListFilter{
				CreatedAfter:  &after,
				CreatedBefore: &before,
			},
			setupMock: func(repo *MockEmployeeRepo) {
				// Should fail before calling repo
			},
			wantErr:     true,
			errExpected: ErrInvalidDateRange,
		},
		{
			name: "valid date range",
			filter: &ListFilter{
				CreatedAfter:  &before,
				CreatedBefore: &after,
			},
			setupMock: func(repo *MockEmployeeRepo) {
				result := &ListResult{
					Employees: []*Employee{},
					Total:     0,
				}
				repo.On("List", mock.Anything, "tenant-123", mock.Anything).Return(result, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")

			result, err := uc.ListEmployees(ctx, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errExpected != nil {
					assert.Equal(t, tt.errExpected, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				
				// Check pagination defaults were applied
				if tt.filter.Page == 0 {
					assert.Equal(t, int32(1), tt.filter.Page)
				}
				if tt.filter.PageSize == 0 {
					assert.Equal(t, int32(20), tt.filter.PageSize)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestMergeEmployees(t *testing.T) {
	primaryID := uuid.New()
	secondaryID := uuid.New()
	
	tests := []struct {
		name           string
		primaryEmail   string
		secondaryEmail string
		setupMock      func(*MockEmployeeRepo, *MockEventPublisher)
		wantErr        bool
		errContains    string
	}{
		{
			name:           "successful merge",
			primaryEmail:   "primary@example.com",
			secondaryEmail: "secondary@example.com",
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				primary := &Employee{
					ID:       primaryID,
					Emails:   []string{"primary@example.com"},
					TenantID: "tenant-123",
				}
				secondary := &Employee{
					ID:       secondaryID,
					Emails:   []string{"secondary@example.com"},
					TenantID: "tenant-123",
				}
				merged := &Employee{
					ID:       primaryID,
					Emails:   []string{"primary@example.com", "secondary@example.com"},
					TenantID: "tenant-123",
				}
				
				repo.On("GetByEmail", mock.Anything, "tenant-123", "primary@example.com").Return(primary, nil)
				repo.On("GetByEmail", mock.Anything, "tenant-123", "secondary@example.com").Return(secondary, nil)
				repo.On("MergeEmployees", mock.Anything, "tenant-123", "primary@example.com", "secondary@example.com").Return(merged, nil)
				repo.On("GetEventPublisher").Return(EventPublisher(pub))
				pub.On("PublishEmployeeMerged", mock.Anything, "tenant-123", "user-456", merged, "secondary@example.com").Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "same email provided",
			primaryEmail:   "same@example.com",
			secondaryEmail: "same@example.com",
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				// Should fail early
			},
			wantErr:     true,
			errContains: "INVALID_MERGE",
		},
		{
			name:           "primary not found",
			primaryEmail:   "notfound@example.com",
			secondaryEmail: "secondary@example.com",
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				repo.On("GetByEmail", mock.Anything, "tenant-123", "notfound@example.com").Return(nil, nil)
			},
			wantErr:     true,
			errContains: "PRIMARY_NOT_FOUND",
		},
		{
			name:           "secondary not found",
			primaryEmail:   "primary@example.com",
			secondaryEmail: "notfound@example.com",
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				primary := &Employee{
					ID:       primaryID,
					Emails:   []string{"primary@example.com"},
					TenantID: "tenant-123",
				}
				repo.On("GetByEmail", mock.Anything, "tenant-123", "primary@example.com").Return(primary, nil)
				repo.On("GetByEmail", mock.Anything, "tenant-123", "notfound@example.com").Return(nil, nil)
			},
			wantErr:     true,
			errContains: "SECONDARY_NOT_FOUND",
		},
		{
			name:           "same employee ID",
			primaryEmail:   "email1@example.com",
			secondaryEmail: "email2@example.com",
			setupMock: func(repo *MockEmployeeRepo, pub *MockEventPublisher) {
				sameEmployee := &Employee{
					ID:       primaryID,
					Emails:   []string{"email1@example.com", "email2@example.com"},
					TenantID: "tenant-123",
				}
				repo.On("GetByEmail", mock.Anything, "tenant-123", "email1@example.com").Return(sameEmployee, nil)
				repo.On("GetByEmail", mock.Anything, "tenant-123", "email2@example.com").Return(sameEmployee, nil)
			},
			wantErr:     true,
			errContains: "CANNOT_MERGE_SAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setupUsecase()
			pub := new(MockEventPublisher)
			
			if tt.setupMock != nil {
				tt.setupMock(repo, pub)
			}

			ctx := WithTenantID(context.Background(), "tenant-123")
			ctx = WithUserID(ctx, "user-456")

			result, err := uc.MergeEmployees(ctx, tt.primaryEmail, tt.secondaryEmail)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestMissingTenantID(t *testing.T) {
	uc, _ := setupUsecase()
	ctx := context.Background() // No tenant ID

	// Test all methods fail without tenant ID
	_, err := uc.CreateEmployee(ctx, &Employee{Emails: []string{"test@example.com"}})
	assert.Error(t, err)

	_, err = uc.UpdateEmployee(ctx, &Employee{ID: uuid.New()})
	assert.Error(t, err)

	err = uc.DeleteEmployee(ctx, uuid.New())
	assert.Error(t, err)

	_, err = uc.GetEmployee(ctx, uuid.New())
	assert.Error(t, err)

	_, err = uc.GetEmployeeByEmail(ctx, "test@example.com")
	assert.Error(t, err)

	_, err = uc.ListEmployees(ctx, &ListFilter{})
	assert.Error(t, err)

	_, err = uc.MergeEmployees(ctx, "primary@example.com", "secondary@example.com")
	assert.Error(t, err)
}

