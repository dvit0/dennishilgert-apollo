package db

import (
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/frontend/models"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var log = logger.NewLogger("apollo.db")

type Options struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SslMode  bool
	Timezone string
}

type DatabaseClient interface {
	Close() error
	Migrate() error
	CreateKernel(kernel *models.Kernel) error
	GetKernel(name string, version string, architecture string) (*models.Kernel, error)
	ListKernels() ([]models.Kernel, error)
	DeleteKernel(name string, version string, architecture string) error
	CreateRuntime(runtime *models.Runtime) error
	GetRuntime(name string, version string, architecture string) (*models.Runtime, error)
	ListRuntimes() ([]models.Runtime, error)
	DeleteRuntime(name string, version string, architecture string) error
	CreateFunction(function *models.Function) error
	GetFunction(uuid string) (*models.Function, error)
	ListFunctions() ([]models.Function, error)
	UpdateFunction(function *models.Function) error
	DeleteFunction(uuid string) error
	CreateHttpTrigger(trigger *models.HttpTrigger) error
	GetHttpTrigger(urlId string) (*models.HttpTrigger, error)
	GetHttpTriggerByFunctionUuid(functionUuid string) (*models.HttpTrigger, error)
	DeleteHttpTrigger(urlId string) error
}

type databaseClient struct {
	db *gorm.DB
}

// NewDatabaseClient creates a new database client.
func NewDatabaseClient(opts Options) (DatabaseClient, error) {
	log.Infof("connecting to database server: %s:%d", opts.Host, opts.Port)

	sslMode := "disable"
	if opts.SslMode {
		sslMode = "enable"
	}
	timezone := "Europe/Berlin"
	if opts.Timezone != "" {
		timezone = opts.Timezone
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s", opts.Host, opts.Port, opts.Username, opts.Password, opts.Database, sslMode, timezone)
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &databaseClient{
		db: gormDb,
	}, nil
}

// Close closes the database connection.
func (d *databaseClient) Close() error {
	log.Infof("closing database connection")
	sqlDb, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDb.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

// Migrate migrates the database schema.
func (d *databaseClient) Migrate() error {
	log.Infof("migrating database schema")
	if err := d.db.AutoMigrate(
		&models.Kernel{},
		&models.Runtime{},
		&models.Function{},
		&models.HttpTrigger{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

// CreateKernel creates a new kernel.
func (d *databaseClient) CreateKernel(kernel *models.Kernel) error {
	if err := d.db.Create(kernel).Error; err != nil {
		return fmt.Errorf("failed to create kernel: %w", err)
	}
	return nil
}

// GetKernel retrieves a kernel by its name and version.
func (d *databaseClient) GetKernel(name string, version string, architecture string) (*models.Kernel, error) {
	var kernel models.Kernel
	if err := d.db.First(&kernel, "name = ? AND version = ? AND architecture = ?", name, version, architecture).Error; err != nil {
		return nil, fmt.Errorf("failed to get kernel: %w", err)
	}
	return &kernel, nil
}

// ListKernels retrieves all kernels.
func (d *databaseClient) ListKernels() ([]models.Kernel, error) {
	var kernels []models.Kernel
	if err := d.db.Find(&kernels).Error; err != nil {
		return nil, fmt.Errorf("failed to get kernels: %w", err)
	}
	return kernels, nil
}

// DeleteKernel deletes a kernel by its name and version.
func (d *databaseClient) DeleteKernel(name string, version string, architecture string) error {
	if err := d.db.Delete(&models.Kernel{}, "name = ? AND version = ? AND architecture = ?", name, version, architecture).Error; err != nil {
		return fmt.Errorf("failed to delete kernel: %w", err)
	}
	return nil
}

// CreateRuntime creates a new runtime.
func (d *databaseClient) CreateRuntime(runtime *models.Runtime) error {
	if err := d.db.Create(runtime).Error; err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}
	return nil
}

// GetRuntime retrieves a runtime by its name and version.
func (d *databaseClient) GetRuntime(name string, version string, architecture string) (*models.Runtime, error) {
	var runtime models.Runtime
	if err := d.db.First(&runtime, "name = ? AND version = ? AND kernel_architecture = ?", name, version, architecture).Error; err != nil {
		return nil, fmt.Errorf("failed to get runtime: %w", err)
	}
	return &runtime, nil
}

// ListRuntimes retrieves all runtimes.
func (d *databaseClient) ListRuntimes() ([]models.Runtime, error) {
	var runtimes []models.Runtime
	if err := d.db.Find(&runtimes).Error; err != nil {
		return nil, fmt.Errorf("failed to get runtimes: %w", err)
	}
	return runtimes, nil
}

// DeleteRuntime deletes a runtime by its name and version.
func (d *databaseClient) DeleteRuntime(name string, version string, architecture string) error {
	if err := d.db.Delete(&models.Runtime{}, "name = ? AND version = ? AND kernel_architecture = ?", name, version, architecture).Error; err != nil {
		return fmt.Errorf("failed to delete runtime: %w", err)
	}
	return nil
}

// CreateFunction creates a new function.
func (d *databaseClient) CreateFunction(function *models.Function) error {
	if err := d.db.Create(function).Error; err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}
	return nil
}

// GetFunction retrieves a function by its uuid.
func (d *databaseClient) GetFunction(uuid string) (*models.Function, error) {
	var function models.Function
	if err := d.db.Preload("Runtime").First(&function, "uuid = ?", uuid).Error; err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}
	return &function, nil
}

// ListFunctions retrieves all functions.
func (d *databaseClient) ListFunctions() ([]models.Function, error) {
	var functions []models.Function
	if err := d.db.Preload("Runtime").Find(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}
	return functions, nil
}

// UpdateFunction updates a function.
func (d *databaseClient) UpdateFunction(function *models.Function) error {
	if err := d.db.Save(function).Error; err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

// DeleteFunction deletes a function by its uuid.
func (d *databaseClient) DeleteFunction(uuid string) error {
	if err := d.db.Delete(&models.Function{}, "uuid = ?", uuid).Error; err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	return nil
}

func (d *databaseClient) CreateHttpTrigger(trigger *models.HttpTrigger) error {
	if err := d.db.Create(trigger).Error; err != nil {
		return fmt.Errorf("failed to create http trigger: %w", err)
	}
	return nil
}

func (d *databaseClient) GetHttpTrigger(urlId string) (*models.HttpTrigger, error) {
	var trigger models.HttpTrigger
	if err := d.db.First(&trigger, "url_id = ?", urlId).Error; err != nil {
		return nil, fmt.Errorf("failed to get http trigger: %w", err)
	}
	return &trigger, nil
}

func (d *databaseClient) GetHttpTriggerByFunctionUuid(functionUuid string) (*models.HttpTrigger, error) {
	var trigger models.HttpTrigger
	if err := d.db.First(&trigger, "function_uuid = ?", functionUuid).Error; err != nil {
		return nil, fmt.Errorf("failed to get http trigger: %w", err)
	}
	return &trigger, nil
}

func (d *databaseClient) DeleteHttpTrigger(urlId string) error {
	if err := d.db.Delete(&models.HttpTrigger{}, "url_id = ?", urlId).Error; err != nil {
		return fmt.Errorf("failed to delete http trigger: %w", err)
	}
	return nil
}
