package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	eventsv1 "employee-service/api/events/v1"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

var (
	natsURL string
)

func init() {
	flag.StringVar(&natsURL, "nats", "nats://localhost:4222", "NATS server URL")
}

func main() {
	flag.Parse()

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Printf("✓ Connected to NATS at %s", natsURL)
	log.Println()
	log.Println("Subscribing to employee event subjects:")
	log.Println("  - employees.v1.created")
	log.Println("  - employees.v1.updated")
	log.Println("  - employees.v1.deleted")
	log.Println("  - employees.v1.merged")
	log.Println()

	// Subscribe to employee created events
	_, err = nc.Subscribe("employees.v1.created", func(msg *nats.Msg) {
		var event eventsv1.EmployeeCreatedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("✗ Error unmarshaling created event: %v", err)
			return
		}
		printEvent("CREATED", event.Event)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to created events: %v", err)
	}

	// Subscribe to employee updated events
	_, err = nc.Subscribe("employees.v1.updated", func(msg *nats.Msg) {
		var event eventsv1.EmployeeUpdatedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("✗ Error unmarshaling updated event: %v", err)
			return
		}
		printEvent("UPDATED", event.Event)
		if len(event.UpdatedFields) > 0 {
			log.Printf("  Updated Fields: %v", event.UpdatedFields)
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to updated events: %v", err)
	}

	// Subscribe to employee deleted events
	_, err = nc.Subscribe("employees.v1.deleted", func(msg *nats.Msg) {
		var event eventsv1.EmployeeDeletedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("✗ Error unmarshaling deleted event: %v", err)
			return
		}
		printEvent("DELETED", event.Event)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to deleted events: %v", err)
	}

	// Subscribe to employee merged events
	_, err = nc.Subscribe("employees.v1.merged", func(msg *nats.Msg) {
		var event eventsv1.EmployeeMergedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("✗ Error unmarshaling merged event: %v", err)
			return
		}
		printEvent("MERGED", event.Event)
		if event.MergedFromEmail != "" {
			log.Printf("  Merged From: %s", event.MergedFromEmail)
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to merged events: %v", err)
	}

	log.Println("🎧 Listening for employee events...")
	log.Println("   Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println()
	log.Println("Shutting down consumer...")
}

func printEvent(eventType string, event *eventsv1.EmployeeEvent) {
	log.Println("========================================")
	log.Printf("📨 %s Event", eventType)
	log.Println("========================================")
	log.Printf("Event ID:   %s", event.EventId)
	log.Printf("Event Type: %s", event.EventType.String())
	log.Printf("Tenant ID:  %s", event.TenantId)
	log.Printf("User ID:    %s", event.UserId)
	log.Printf("Timestamp:  %s", event.Timestamp.AsTime().Format("2006-01-02 15:04:05"))

	// Print employee data if present
	if event.Employee != nil {
		emp := event.Employee
		log.Println()
		log.Println("Employee Data:")
		log.Printf("  ID:         %s", emp.Id)
		log.Printf("  Email:      %s", emp.Email)
		log.Printf("  First Name: %s", emp.FirstName)
		log.Printf("  Last Name:  %s", emp.LastName)
		if len(emp.SecondaryEmails) > 0 {
			log.Printf("  Secondary Emails: %v", emp.SecondaryEmails)
		}
		log.Printf("  Created At: %s", emp.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
		log.Printf("  Updated At: %s", emp.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}

	// Print metadata if present
	if len(event.Metadata) > 0 {
		log.Println()
		log.Println("Metadata:")
		for key, value := range event.Metadata {
			log.Printf("  %s: %s", key, value)
		}
	}

	log.Println("========================================")
	log.Println()
}
