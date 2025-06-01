package services

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

var Connector *gorm.DB

type DBService struct {
	config *ConnectorConfig
	db     *gorm.DB
}

type ConnectorConfig struct {
	Host     string
	Port     string
	User     string
	DBName   string
	Password string
}

func NewDBService() *DBService {
	return &DBService{
		config: buildConnectorConfig(),
	}
}

func buildConnectorConfig() *ConnectorConfig {
	return &ConnectorConfig{
		Host:     config.EnvDBHost(),
		Port:     config.EnvDBPort(),
		User:     config.EnvDBUser(),
		Password: config.EnvDBPassword(),
		DBName:   config.EnvDBName(),
	}
}

func (s *DBService) getConnectionURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		s.config.Host,
		s.config.Port,
		s.config.User,
		s.config.DBName,
		s.config.Password,
	)
}

func OpenConnection() *errors.AppError {
	service := NewDBService()
	if err := service.connect(); err != nil {
		return err
	}

	service.configureConnection()
	go service.startHealthCheck()

	return nil
}

func (s *DBService) connect() *errors.AppError {
	dbURL := s.getConnectionURL()
	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		return errors.DatabaseError(err.Error())
	}

	s.db = db
	Connector = db
	return nil
}

func (s *DBService) configureConnection() {
	s.db.SingularTable(true)
	s.db.LogMode(!isProduction())
	s.db.DB().SetConnMaxLifetime(10 * time.Second)
	s.db.DB().SetMaxIdleConns(30)
}

func isProduction() bool {
	return config.EnvironmentConfig() == entities.Environment.Production
}

func (s *DBService) startHealthCheck() {
	intervals := []time.Duration{3, 3, 15, 30, 60}
	dbURL := s.getConnectionURL()

	for {
		time.Sleep(60 * time.Second)
		if err := s.db.DB().Ping(); err != nil {
			s.handleReconnection(intervals, dbURL)
		}
	}
}

func (s *DBService) handleReconnection(intervals []time.Duration, dbURL string) {
	for i := 0; i < len(intervals); i++ {
		if success := s.attemptReconnect(dbURL); success {
			return
		}
		time.Sleep(intervals[i] * time.Second)
		if i == len(intervals)-1 {
			i--
		}
	}
}

func (s *DBService) attemptReconnect(dbURL string) bool {
	err := RetryHandler(3, func() (bool, error) {
		db, err := gorm.Open("postgres", dbURL)
		if err != nil {
			return false, err
		}
		s.db = db
		Connector = db
		return true, nil
	})
	return err == nil
}

func RetryHandler(attempts int, fn func() (bool, error)) error {
	ok, err := fn()
	if ok && err == nil {
		return nil
	}
	if attempts > 1 {
		return RetryHandler(attempts-1, fn)
	}
	return err
}

func RunMigrations() {
	// Define migrations here
}
