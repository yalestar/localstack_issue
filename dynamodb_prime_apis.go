package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	av "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

var (
	client    *dynamodb.Client
	tableName string
)

func init() {
	client, _ = NewDynamoDBClient()
	tableName = "local-prime-integration-service-prime-apis"
}

type PrimeAPI struct {
	Id      int    `yaml:"id" json:"id" dynamodbav:"prime_api_id"`
	ApiName string `yaml:"api_name" json:"api_name" dynamodbav:"prime_api_name"`
	Enabled bool   `yaml:"enabled" json:"enabled" dynamodbav:"enabled"`
}

func NewDynamoDBClient() (*dynamodb.Client, error) {
	DynamoDBEndpoint := viper.GetString("DYNAMODB_ENDPOINT")

	var client *dynamodb.Client
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	fmt.Println("in LOCAL mode")
	client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.EndpointResolver = dynamodb.EndpointResolverFromURL(DynamoDBEndpoint)
	})
	return client, nil
}

func AllPrimeAPIs(ctx context.Context) (*[]PrimeAPI, error) {
	out, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:      &tableName,
		ConsistentRead: aws.Bool(true),
	},
	)

	if err != nil {
		return nil, err
	}

	var primeAPIS []PrimeAPI
	err = av.UnmarshalListOfMaps(out.Items, &primeAPIS)
	if err != nil {
		return nil, err
	}

	return &primeAPIS, nil
}

// SeedPrimeAPIs inserts (upserts, really) these Prime APIs into DynamoDB
// so we can toggle their status as needed
func seedThoseAPIs() error {

	seedsYml := "prime_api_seeds.yml"

	var primeAPIs = make(map[string][]PrimeAPI)
	yf, err := os.ReadFile(seedsYml)

	if err != nil {
		log.Printf("Unable to read prime API configuration: %s", err.Error())
		return err
	}

	err = yaml.Unmarshal(yf, &primeAPIs)
	if err != nil {
		log.Printf("Unable to unmarshal prime API configuration: %s", err.Error())
		return err
	}

	for _, api := range primeAPIs["prime_apis"] {
		err := insertPrimeApiItem(&api)
		if err != nil {
			log.Printf("Unable to insert prime API configuration: %s", err.Error())
			return err
		}
	}

	return nil
}

func insertPrimeApiItem(primeAPI *PrimeAPI) error {
	marhsalledAPI, err := av.MarshalMap(primeAPI)

	if err != nil {
		return err
	}

	input := dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      marhsalledAPI,
	}

	_, err = client.PutItem(context.Background(), &input)
	if err != nil {
		return err
	}

	fmt.Printf("Seeded PrimeAPI %d %s: (enabled %t)", primeAPI.Id, primeAPI.ApiName,
		primeAPI.Enabled)
	return nil
}

func SetEnabledTo(ctx context.Context, primeApiId int, enabled bool) error {

	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"prime_api_id": &types.AttributeValueMemberN{Value: strconv.Itoa(primeApiId)},
		},
		UpdateExpression: aws.String("set enabled = :newVal"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newVal": &types.AttributeValueMemberBOOL{
				Value: enabled,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
