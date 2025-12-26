package data

import (
	"time"

	"employee-service/internal/biz"

	"github.com/google/uuid"
)

// EmployeeEmailModel is the GORM model for employee emails
type EmployeeEmailModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_employee_emails_employee_id"`
	TenantID   string    `gorm:"type:varchar(255);not null;index:idx_employee_emails_tenant_email,unique,priority:1"`
	Email      string    `gorm:"type:varchar(255);not null;index:idx_employee_emails_tenant_email,unique,priority:2"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// TableName overrides the table name
func (EmployeeEmailModel) TableName() string {
	return "employee_emails"
}

// EmployeeModel is the GORM model for Employee
type EmployeeModel struct {
	ID        uuid.UUID            `gorm:"type:uuid;primaryKey"`
	TenantID  string               `gorm:"type:varchar(255);not null;index:idx_tenant_id"`
	FirstName string               `gorm:"type:varchar(255);not null"`
	LastName  string               `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time            `gorm:"autoCreateTime"`
	UpdatedAt time.Time            `gorm:"autoUpdateTime"`
	Emails    []EmployeeEmailModel `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE"`
}

// TableName overrides the table name
func (EmployeeModel) TableName() string {
	return "employees"
}

// ToEntity converts EmployeeModel to biz.Employee
func (m *EmployeeModel) ToEntity() *biz.Employee {
	emails := make([]string, len(m.Emails))
	for i, emailModel := range m.Emails {
		emails[i] = emailModel.Email
	}

	return &biz.Employee{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Emails:    emails,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromEntity converts biz.Employee to EmployeeModel
func FromEntity(e *biz.Employee) *EmployeeModel {
	emailModels := make([]EmployeeEmailModel, len(e.Emails))
	for i, email := range e.Emails {
		emailModels[i] = EmployeeEmailModel{
			EmployeeID: e.ID,
			TenantID:   e.TenantID,
			Email:      email,
		}
	}

	return &EmployeeModel{
		ID:        e.ID,
		TenantID:  e.TenantID,
		FirstName: e.FirstName,
		LastName:  e.LastName,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		Emails:    emailModels,
	}
}
