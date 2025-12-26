package data

import (
	"employee-service/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewEmployeeRepo)

// Data .
type Data struct {
	db        *gorm.DB
	nc        *nats.Conn
	publisher *EventPublisher
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)
	
	// Open database connection
	db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		logHelper.Errorf("failed to connect to database: %v", err)
		return nil, nil, err
	}

	logHelper.Info("database connected successfully")

	// Connect to NATS (optional)
	var nc *nats.Conn
	var publisher *EventPublisher
	
	if c.Nats != nil && c.Nats.Url != "" {
		nc, err = nats.Connect(c.Nats.Url,
			nats.MaxReconnects(-1), // Infinite reconnects
			nats.ReconnectWait(2*time.Second),
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				logHelper.Warnf("NATS disconnected: %v", err)
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				logHelper.Infof("NATS reconnected to %s", nc.ConnectedUrl())
			}),
		)
		if err != nil {
			logHelper.Warnf("failed to connect to NATS (continuing without events): %v", err)
			nc = nil
		} else {
			logHelper.Infof("connected to NATS at %s", c.Nats.Url)
			// Using versioned subjects (employees.v1.{created,updated,deleted,merged})
			publisher = NewEventPublisher(nc, "", logger)
		}
	} else {
		logHelper.Info("NATS not configured, events disabled")
	}

	cleanup := func() {
		if nc != nil {
			nc.Close()
			logHelper.Info("NATS connection closed")
		}
		
		sqlDB, err := db.DB()
		if err != nil {
			logHelper.Errorf("failed to get database instance: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			logHelper.Errorf("failed to close database: %v", err)
		}
		logHelper.Info("closing the data resources")
	}
	
	return &Data{db: db, nc: nc, publisher: publisher}, cleanup, nil
}
