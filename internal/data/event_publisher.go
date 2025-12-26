package data

import (
	"context"

	"employee-service/internal/biz"
	eventsv1 "employee-service/api/events/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NATS subject constants for versioned event types
const (
	SubjectEmployeeCreated = "employees.v1.created"
	SubjectEmployeeUpdated = "employees.v1.updated"
	SubjectEmployeeDeleted = "employees.v1.deleted"
	SubjectEmployeeMerged  = "employees.v1.merged"
)

// EventPublisher publishes events to NATS using Protocol Buffers
type EventPublisher struct {
	nc  *nats.Conn
	log *log.Helper
}

// NewEventPublisher creates a new event publisher
// Note: subjectPrefix is no longer used as we use specific subjects per event type
func NewEventPublisher(nc *nats.Conn, subjectPrefix string, logger log.Logger) *EventPublisher {
	return &EventPublisher{
		nc:  nc,
		log: log.NewHelper(logger),
	}
}

// toProtoEmployeeData converts biz.Employee to proto EmployeeData
func toProtoEmployeeData(emp *biz.Employee) *eventsv1.EmployeeData {
	if emp == nil {
		return nil
	}

	secondaryEmails := emp.SecondaryEmails
	if secondaryEmails == nil {
		secondaryEmails = []string{}
	}

	return &eventsv1.EmployeeData{
		Id:              emp.ID.String(),
		Email:           emp.Email,
		SecondaryEmails: secondaryEmails,
		FirstName:       emp.FirstName,
		LastName:        emp.LastName,
		CreatedAt:       timestamppb.New(emp.CreatedAt),
		UpdatedAt:       timestamppb.New(emp.UpdatedAt),
	}
}

// PublishEmployeeCreated publishes an employee created event
func (p *EventPublisher) PublishEmployeeCreated(
	ctx context.Context,
	tenantID, userID string,
	employee *biz.Employee,
) error {
	if p == nil || p.nc == nil {
		// NATS not configured, skip publishing
		return nil
	}

	event := &eventsv1.EmployeeCreatedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_CREATED,
			TenantId:  tenantID,
			Timestamp: timestamppb.Now(),
			UserId:    userID,
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
	}

	return p.publishProtoEvent(SubjectEmployeeCreated, event)
}

// PublishEmployeeUpdated publishes an employee updated event
func (p *EventPublisher) PublishEmployeeUpdated(
	ctx context.Context,
	tenantID, userID string,
	employee *biz.Employee,
	updatedFields []string,
) error {
	if p == nil || p.nc == nil {
		// NATS not configured, skip publishing
		return nil
	}

	if updatedFields == nil {
		updatedFields = []string{}
	}

	event := &eventsv1.EmployeeUpdatedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_UPDATED,
			TenantId:  tenantID,
			Timestamp: timestamppb.Now(),
			UserId:    userID,
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
		UpdatedFields: updatedFields,
	}

	return p.publishProtoEvent(SubjectEmployeeUpdated, event)
}

// PublishEmployeeDeleted publishes an employee deleted event
func (p *EventPublisher) PublishEmployeeDeleted(
	ctx context.Context,
	tenantID, userID string,
	employee *biz.Employee,
) error {
	if p == nil || p.nc == nil {
		// NATS not configured, skip publishing
		return nil
	}

	event := &eventsv1.EmployeeDeletedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_DELETED,
			TenantId:  tenantID,
			Timestamp: timestamppb.Now(),
			UserId:    userID,
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
	}

	return p.publishProtoEvent(SubjectEmployeeDeleted, event)
}

// PublishEmployeeMerged publishes an employee merged event
func (p *EventPublisher) PublishEmployeeMerged(
	ctx context.Context,
	tenantID, userID string,
	employee *biz.Employee,
	mergedFromEmail string,
) error {
	if p == nil || p.nc == nil {
		// NATS not configured, skip publishing
		return nil
	}

	event := &eventsv1.EmployeeMergedEvent{
		Event: &eventsv1.EmployeeEvent{
			EventId:   uuid.New().String(),
			EventType: eventsv1.EventType_EVENT_TYPE_MERGED,
			TenantId:  tenantID,
			Timestamp: timestamppb.Now(),
			UserId:    userID,
			Employee:  toProtoEmployeeData(employee),
			Metadata:  map[string]string{},
		},
		MergedFromEmail: mergedFromEmail,
	}

	return p.publishProtoEvent(SubjectEmployeeMerged, event)
}

// publishProtoEvent marshals and publishes a protobuf message to NATS
func (p *EventPublisher) publishProtoEvent(subject string, msg proto.Message) error {
	// Marshal event to Protocol Buffers
	data, err := proto.Marshal(msg)
	if err != nil {
		p.log.Errorf("failed to marshal proto event: %v", err)
		return err
	}

	// Publish to NATS (best-effort)
	if err := p.nc.Publish(subject, data); err != nil {
		p.log.Errorf("failed to publish event to NATS subject %s: %v", subject, err)
		return err
	}

	p.log.Infof("published event to subject: %s", subject)
	return nil
}
