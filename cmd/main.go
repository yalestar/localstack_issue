package main

import (
	"context"
	"github.com/cambiahealth/prime-integration-service/pkg/prime"
	"github.com/go-chi/chi/v5"

	"github.com/cambiahealth/janus-go/log"
	janusServer "github.com/cambiahealth/janus-go/server/v2"
	_ "github.com/cambiahealth/prime-integration-service/api/generated/swaggerui/statik"
	"github.com/cambiahealth/prime-integration-service/db"
	"github.com/spf13/viper"
)

type PrimeIntegrationService struct {
	logger         *log.JanusLogger
	dbService      db.DynamoDB
	chiMux         *chi.Mux
	apiEnvironment *prime.APIEnvironment
}

func main() {
	logger := log.New()

	srv := janusServer.New(logger, janusServer.HTTPOnly(true))
	dynamoService := db.NewDynamoDBService()

	if dynamoService == nil {
		logger.Panicf(context.Background(), "Unable to connect to DynamoDB")
		return
	}

	apiEnv := &prime.APIEnvironment{Environment: "production"}
	ps := &PrimeIntegrationService{
		logger:         logger,
		dbService:      dynamoService,
		chiMux:         srv.GetRouter(),
		apiEnvironment: apiEnv,
	}

	ps.logger.Infof(context.Background(), "DYNAMODB_ENDPOINT %s", viper.GetString("DYNAMODB_ENDPOINT"))
	ps.logger.Infof(context.Background(), "PRIME_APIS_TABLE_NAME is %s", viper.GetString("PRIME_APIS_TABLE_NAME"))

	err := ps.dbService.SeedPrimeAPIs()
	if err != nil {
		ps.logger.Panicf(context.Background(), "Unable to load seed data")
	}

	//ps.routes()
	ps.logger.Info(context.Background(), "Starting prime-integration-service server")

	srv.Run()
}
