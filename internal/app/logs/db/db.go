package db

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/logs/models"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var log = logger.NewLogger("apollo.db")

type Options struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	AuthDatabase string
}

type DatabaseClient interface {
	Close() error
	InsertLogs(ctx context.Context, logsEntry models.InvocationLogsEntry) error
	GetLogs(ctx context.Context, functionUuid string, functionVerstion string) ([]models.InvocationLogsEntry, error)
}

type databaseClient struct {
	database string
	client   *mongo.Client
}

func NewDatabaseClient(ctx context.Context, opts Options) (DatabaseClient, error) {
	log.Infof("connecting to database server: %s:%d", opts.Host, opts.Port)

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", opts.Username, opts.Password, opts.Host, opts.Port, opts.AuthDatabase)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &databaseClient{
		database: opts.Database,
		client:   client,
	}, nil
}

// Close closes the database connection.
func (d *databaseClient) Close() error {
	log.Infof("closing database connection")

	if d.client == nil {
		return nil
	}
	return d.client.Disconnect(context.Background())
}

func (d *databaseClient) InsertLogs(ctx context.Context, logsEntry models.InvocationLogsEntry) error {
	collection := d.client.Database(d.database).Collection("logs")
	_, err := collection.InsertOne(ctx, logsEntry)
	if err != nil {
		return fmt.Errorf("failed to insert logs entry: %w", err)
	}
	return nil
}

func (d *databaseClient) GetLogs(ctx context.Context, functionUuid string, functionVerstion string) ([]models.InvocationLogsEntry, error) {
	collection := d.client.Database(d.database).Collection("logs")
	cursor, err := collection.Find(ctx, bson.D{{Key: "functionUuid", Value: functionUuid}, {Key: "functionVersion", Value: functionVerstion}})
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	var logs []models.InvocationLogsEntry
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return logs, nil
}
