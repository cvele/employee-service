package data

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"employee-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringArray is a custom type for handling string arrays in PostgreSQL
type StringArray []string

// Scan implements sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// Value implements driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(a)
}

// EmployeeModel is the GORM model for Employee
type EmployeeModel struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey"`
	TenantID        string      `gorm:"type:varchar(255);not null;index:idx_tenant_email,priority:1;index:idx_tenant_id"`
	Email           string      `gorm:"type:varchar(255);not null;index:idx_tenant_email,unique,priority:2"`
	SecondaryEmails StringArray `gorm:"type:jsonb"`
	FirstName       string      `gorm:"type:varchar(255);not null"`
	LastName        string      `gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time   `gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (EmployeeModel) TableName() string {
	return "employees"
}

// ToEntity converts EmployeeModel to biz.Employee
func (m *EmployeeModel) ToEntity() *biz.Employee {
	return &biz.Employee{
		ID:              m.ID,
		TenantID:        m.TenantID,
		Email:           m.Email,
		SecondaryEmails: m.SecondaryEmails,
		FirstName:       m.FirstName,
		LastName:        m.LastName,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// FromEntity converts biz.Employee to EmployeeModel
func FromEntity(e *biz.Employee) *EmployeeModel {
	return &EmployeeModel{
		ID:              e.ID,
		TenantID:        e.TenantID,
		Email:           e.Email,
		SecondaryEmails: e.SecondaryEmails,
		FirstName:       e.FirstName,
		LastName:        e.LastName,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

type employeeRepo struct {
	data *Data
	log  *log.Helper
}

// NewEmployeeRepo creates a new employee repository.
func NewEmployeeRepo(data *Data, logger log.Logger) biz.EmployeeRepo {
	return &employeeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetEventPublisher returns the event publisher
func (r *employeeRepo) GetEventPublisher() biz.EventPublisher {
	if r.data.publisher == nil {
		return nil
	}
	return r.data.publisher
}

// Create creates a new employee in the database.
func (r *employeeRepo) Create(ctx context.Context, tenantID string, employee *biz.Employee) (*biz.Employee, error) {
	model := FromEntity(employee)
	model.TenantID = tenantID
	
	// Generate UUID if not set
	if model.ID == uuid.Nil {
		model.ID = uuid.New()
	}

	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

// Update updates an existing employee in the database.
func (r *employeeRepo) Update(ctx context.Context, tenantID string, employee *biz.Employee) (*biz.Employee, error) {
	model := FromEntity(employee)
	
	// Build update map with only non-empty fields
	updateFields := make(map[string]interface{})
	
	// Always update timestamp
	updateFields["updated_at"] = time.Now()
	
	// Only update email if provided
	if model.Email != "" {
		updateFields["email"] = model.Email
	}
	
	// Only update first name if provided
	if model.FirstName != "" {
		updateFields["first_name"] = model.FirstName
	}
	
	// Only update last name if provided
	if model.LastName != "" {
		updateFields["last_name"] = model.LastName
	}
	
	// Secondary emails can be updated (even if empty array)
	updateFields["secondary_emails"] = model.SecondaryEmails
	
	result := r.data.db.WithContext(ctx).
		Model(&EmployeeModel{}).
		Where("id = ? AND tenant_id = ?", employee.ID, tenantID).
		Updates(updateFields)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, biz.ErrEmployeeNotFound
	}

	// Fetch updated record
	return r.GetByID(ctx, tenantID, employee.ID)
}

// Delete deletes an employee from the database.
func (r *employeeRepo) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	result := r.data.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&EmployeeModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return biz.ErrEmployeeNotFound
	}

	return nil
}

// GetByID retrieves an employee by ID within tenant.
func (r *employeeRepo) GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*biz.Employee, error) {
	var model EmployeeModel
	
	err := r.data.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&model).Error

	if err == gorm.ErrRecordNotFound {
		return nil, biz.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

// GetByEmail retrieves an employee by email within tenant.
func (r *employeeRepo) GetByEmail(ctx context.Context, tenantID string, email string) (*biz.Employee, error) {
	var model EmployeeModel
	
	err := r.data.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&model).Error

	if err == gorm.ErrRecordNotFound {
		return nil, biz.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

// List retrieves employees with pagination and filtering within tenant.
func (r *employeeRepo) List(ctx context.Context, tenantID string, filter *biz.ListFilter) (*biz.ListResult, error) {
	var models []EmployeeModel
	var total int64

	query := r.data.db.WithContext(ctx).
		Model(&EmployeeModel{}).
		Where("tenant_id = ?", tenantID)

	// Apply date filters
	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", filter.CreatedBefore)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.
		Offset(int(offset)).
		Limit(int(filter.PageSize)).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	// Convert to entities
	employees := make([]*biz.Employee, len(models))
	for i, model := range models {
		employees[i] = model.ToEntity()
	}

	return &biz.ListResult{
		Employees: employees,
		Total:     total,
	}, nil
}

// CheckEmailExists checks if an email exists within tenant.
func (r *employeeRepo) CheckEmailExists(ctx context.Context, tenantID string, email string) (bool, error) {
	var count int64
	
	err := r.data.db.WithContext(ctx).
		Model(&EmployeeModel{}).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MergeEmployees merges two employees by adding secondary email to primary and deleting secondary.
func (r *employeeRepo) MergeEmployees(ctx context.Context, tenantID string, primaryEmail string, secondaryEmail string) (*biz.Employee, error) {
	var result *biz.Employee
	
	// Start transaction
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get primary employee
		var primaryModel EmployeeModel
		if err := tx.Where("email = ? AND tenant_id = ?", primaryEmail, tenantID).First(&primaryModel).Error; err != nil {
			return err
		}

		// Get secondary employee
		var secondaryModel EmployeeModel
		if err := tx.Where("email = ? AND tenant_id = ?", secondaryEmail, tenantID).First(&secondaryModel).Error; err != nil {
			return err
		}

		// Add secondary email to primary's secondary_emails if not already there
		secondaryEmails := primaryModel.SecondaryEmails
		if secondaryEmails == nil {
			secondaryEmails = []string{}
		}
		
		// Check if secondary email already exists in array
		found := false
		for _, e := range secondaryEmails {
			if e == secondaryEmail {
				found = true
				break
			}
		}
		
		if !found {
			secondaryEmails = append(secondaryEmails, secondaryEmail)
		}

		// Also add any secondary emails from the secondary employee
		for _, e := range secondaryModel.SecondaryEmails {
			found := false
			for _, existing := range secondaryEmails {
				if existing == e {
					found = true
					break
				}
			}
			if !found {
				secondaryEmails = append(secondaryEmails, e)
			}
		}

		// Update primary employee
		if err := tx.Model(&primaryModel).
			Where("id = ? AND tenant_id = ?", primaryModel.ID, tenantID).
			Update("secondary_emails", secondaryEmails).Error; err != nil {
			return err
		}

		// Delete secondary employee
		if err := tx.Where("id = ? AND tenant_id = ?", secondaryModel.ID, tenantID).
			Delete(&EmployeeModel{}).Error; err != nil {
			return err
		}

		// Reload primary with updated data
		if err := tx.Where("id = ? AND tenant_id = ?", primaryModel.ID, tenantID).First(&primaryModel).Error; err != nil {
			return err
		}

		result = primaryModel.ToEntity()
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

