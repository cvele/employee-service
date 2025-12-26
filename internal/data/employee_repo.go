package data

import (
	"context"
	"time"

	"employee-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	// Generate UUID if not set
	if employee.ID == uuid.Nil {
		employee.ID = uuid.New()
	}

	model := FromEntity(employee)
	model.TenantID = tenantID

	// Use transaction to create employee and emails
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create employee record
		if err := tx.Create(&EmployeeModel{
			ID:        model.ID,
			TenantID:  model.TenantID,
			FirstName: model.FirstName,
			LastName:  model.LastName,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		// Create email records
		for _, emailModel := range model.Emails {
			emailModel.EmployeeID = model.ID
			emailModel.TenantID = tenantID
			if err := tx.Create(&emailModel).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch and return the created employee with emails
	return r.GetByID(ctx, tenantID, employee.ID)
}

// Update updates an existing employee in the database.
func (r *employeeRepo) Update(ctx context.Context, tenantID string, employee *biz.Employee) (*biz.Employee, error) {
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Build update map with only non-empty fields
		updateFields := make(map[string]interface{})

		// Always update timestamp
		updateFields["updated_at"] = time.Now()

		// Only update first name if provided
		if employee.FirstName != "" {
			updateFields["first_name"] = employee.FirstName
		}

		// Only update last name if provided
		if employee.LastName != "" {
			updateFields["last_name"] = employee.LastName
		}

		// Update employee record
		result := tx.Model(&EmployeeModel{}).
			Where("id = ? AND tenant_id = ?", employee.ID, tenantID).
			Updates(updateFields)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return biz.ErrEmployeeNotFound
		}

		// Update emails if provided
		if len(employee.Emails) > 0 {
			// Delete existing emails
			if err := tx.Where("employee_id = ? AND tenant_id = ?", employee.ID, tenantID).
				Delete(&EmployeeEmailModel{}).Error; err != nil {
				return err
			}

			// Insert new emails
			for _, email := range employee.Emails {
				emailModel := EmployeeEmailModel{
					EmployeeID: employee.ID,
					TenantID:   tenantID,
					Email:      email,
				}
				if err := tx.Create(&emailModel).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
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
		Preload("Emails").
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
	var emailModel EmployeeEmailModel

	// Find the email record first
	err := r.data.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&emailModel).Error

	if err == gorm.ErrRecordNotFound {
		return nil, biz.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, err
	}

	// Fetch the employee with all emails
	return r.GetByID(ctx, tenantID, emailModel.EmployeeID)
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

	// Apply pagination and preload emails
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.
		Preload("Emails").
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
		Model(&EmployeeEmailModel{}).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MergeEmployees merges two employees by transferring all emails from secondary to primary.
func (r *employeeRepo) MergeEmployees(ctx context.Context, tenantID string, primaryEmail string, secondaryEmail string) (*biz.Employee, error) {
	var result *biz.Employee

	// Start transaction
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get primary employee email record
		var primaryEmailModel EmployeeEmailModel
		if err := tx.Where("email = ? AND tenant_id = ?", primaryEmail, tenantID).First(&primaryEmailModel).Error; err != nil {
			return err
		}

		// Get secondary employee email record
		var secondaryEmailModel EmployeeEmailModel
		if err := tx.Where("email = ? AND tenant_id = ?", secondaryEmail, tenantID).First(&secondaryEmailModel).Error; err != nil {
			return err
		}

		primaryEmployeeID := primaryEmailModel.EmployeeID
		secondaryEmployeeID := secondaryEmailModel.EmployeeID

		// Transfer all emails from secondary employee to primary employee
		if err := tx.Model(&EmployeeEmailModel{}).
			Where("employee_id = ? AND tenant_id = ?", secondaryEmployeeID, tenantID).
			Update("employee_id", primaryEmployeeID).Error; err != nil {
			return err
		}

		// Delete secondary employee record
		if err := tx.Where("id = ? AND tenant_id = ?", secondaryEmployeeID, tenantID).
			Delete(&EmployeeModel{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch the merged employee with all emails
	primaryEmailModel := EmployeeEmailModel{}
	if err := r.data.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", primaryEmail, tenantID).
		First(&primaryEmailModel).Error; err != nil {
		return nil, err
	}

	result, err = r.GetByID(ctx, tenantID, primaryEmailModel.EmployeeID)
	if err != nil {
		return nil, err
	}

	return result, nil
}
