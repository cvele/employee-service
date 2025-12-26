package data

import (
	"testing"
	"time"

	eventsv1 "employee-service/api/events/v1"
	"employee-service/internal/biz"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestToProtoEmployeeData(t *testing.T) {
	tests := []struct {
		name     string
		employee *biz.Employee
		wantNil  bool
	}{
		{
			name:     "nil employee",
			employee: nil,
			wantNil:  true,
		},
		{
			name: "valid employee",
			employee: &biz.Employee{
				ID:        uuid.New(),
				Emails:    []string{"test@example.com", "secondary@example.com"},
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantNil: false,
		},
		{
			name: "employee with nil emails",
			employee: &biz.Employee{
				ID:        uuid.New(),
				Emails:    nil,
				FirstName: "Jane",
				LastName:  "Smith",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoEmployeeData(tt.employee)

			if tt.wantNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.employee.ID.String(), result.Id)
			assert.Equal(t, tt.employee.FirstName, result.FirstName)
			assert.Equal(t, tt.employee.LastName, result.LastName)

			// Should have empty array, not nil
			assert.NotNil(t, result.Emails)

			if tt.employee.Emails != nil {
				assert.Equal(t, tt.employee.Emails, result.Emails)
			} else {
				assert.Empty(t, result.Emails)
			}
		})
	}
}

func TestEmployeeCreatedEventContract(t *testing.T) {
	// Test that EmployeeCreatedEvent can be marshaled and unmarshaled
	employee := &biz.Employee{
		ID:        uuid.New(),
		Emails:    []string{"test@example.com"},
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	event := &eventsv1.EmployeeCreatedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_CREATED,
			TenantId:  "tenant-123",
			UserId:    "user-456",
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
	}

	// Marshal
	data, err := proto.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal
	var decoded eventsv1.EmployeeCreatedEvent
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, event.Event.EventId, decoded.Event.EventId)
	assert.Equal(t, event.Event.EventType, decoded.Event.EventType)
	assert.Equal(t, event.Event.TenantId, decoded.Event.TenantId)
	assert.Equal(t, event.Event.UserId, decoded.Event.UserId)
	assert.Equal(t, event.Event.Employee.Id, decoded.Event.Employee.Id)
	assert.Equal(t, event.Event.Employee.Emails, decoded.Event.Employee.Emails)
}

func TestEmployeeUpdatedEventContract(t *testing.T) {
	employee := &biz.Employee{
		ID:        uuid.New(),
		Emails:    []string{"updated@example.com"},
		FirstName: "Jane",
		LastName:  "Updated",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	event := &eventsv1.EmployeeUpdatedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_UPDATED,
			TenantId:  "tenant-123",
			UserId:    "user-456",
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
		UpdatedFields: []string{"email", "last_name"},
	}

	// Marshal
	data, err := proto.Marshal(event)
	assert.NoError(t, err)

	// Unmarshal
	var decoded eventsv1.EmployeeUpdatedEvent
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify updated fields
	assert.Equal(t, event.UpdatedFields, decoded.UpdatedFields)
}

func TestEmployeeDeletedEventContract(t *testing.T) {
	employee := &biz.Employee{
		ID:        uuid.New(),
		Emails:    []string{"deleted@example.com"},
		FirstName: "Deleted",
		LastName:  "User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	event := &eventsv1.EmployeeDeletedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_DELETED,
			TenantId:  "tenant-123",
			UserId:    "user-456",
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
	}

	// Marshal
	data, err := proto.Marshal(event)
	assert.NoError(t, err)

	// Unmarshal
	var decoded eventsv1.EmployeeDeletedEvent
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify event type
	assert.Equal(t, eventsv1.EventType_EVENT_TYPE_DELETED, decoded.Event.EventType)
}

func TestEmployeeMergedEventContract(t *testing.T) {
	employee := &biz.Employee{
		ID:        uuid.New(),
		Emails:    []string{"primary@example.com", "merged@example.com"},
		FirstName: "Merged",
		LastName:  "Employee",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	event := &eventsv1.EmployeeMergedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_MERGED,
			TenantId:  "tenant-123",
			UserId:    "user-456",
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
		MergedFromEmail: "secondary@example.com",
	}

	// Marshal
	data, err := proto.Marshal(event)
	assert.NoError(t, err)

	// Unmarshal
	var decoded eventsv1.EmployeeMergedEvent
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify merged_from_email field
	assert.Equal(t, event.MergedFromEmail, decoded.MergedFromEmail)
	assert.Equal(t, "secondary@example.com", decoded.MergedFromEmail)
}

func TestEventTypeEnum(t *testing.T) {
	// Test that all event types are defined
	assert.Equal(t, "EVENT_TYPE_UNSPECIFIED", eventsv1.EventType_EVENT_TYPE_UNSPECIFIED.String())
	assert.Equal(t, "EVENT_TYPE_CREATED", eventsv1.EventType_EVENT_TYPE_CREATED.String())
	assert.Equal(t, "EVENT_TYPE_UPDATED", eventsv1.EventType_EVENT_TYPE_UPDATED.String())
	assert.Equal(t, "EVENT_TYPE_DELETED", eventsv1.EventType_EVENT_TYPE_DELETED.String())
	assert.Equal(t, "EVENT_TYPE_MERGED", eventsv1.EventType_EVENT_TYPE_MERGED.String())
}

func TestRequiredFields(t *testing.T) {
	// Test that events can be created with minimum required fields
	event := &eventsv1.EmployeeCreatedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_CREATED,
			TenantId:  "tenant-123",
		},
	}

	data, err := proto.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestSubjectConstants(t *testing.T) {
	// Verify subject naming convention
	assert.Equal(t, "employees.v1.created", SubjectEmployeeCreated)
	assert.Equal(t, "employees.v1.updated", SubjectEmployeeUpdated)
	assert.Equal(t, "employees.v1.deleted", SubjectEmployeeDeleted)
	assert.Equal(t, "employees.v1.merged", SubjectEmployeeMerged)
}
