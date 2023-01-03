package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"time"
	"user/service"
	"user/service/api"
	"user/service/config"
	"user/service/producer"
	"user/service/repository"
	"user/service/utility"
)

// init is invoked before main()
func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
	log.Info("Initializing environment variables")
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}
}

//InitializeService create a new instance of server service
func InitializeService(cfg config.Config, ctx context.Context) (*service.Service, error) {
	log.Info("Initializing user service")
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	utility := utility.New(cfg.SecretKey)
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoHost))
	err := mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Error while connecting to MongoDB instance")
		return nil, err
	}
	userCollection := mongoClient.Database("users_collection").Collection("users")
	// Clean db on start if DEV mode
	if cfg.Mode == "DEV" {
		err = userCollection.Drop(ctx)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	repo := repository.New(mongoClient, utility)
	broker, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.KafkaServer})
	if err != nil {
		fmt.Println("Failed to create producer due to ", err)
		os.Exit(1)
	}
	kafkaProducer := producer.New(broker, cfg.KafkaTopic)
	s := service.New(repo, kafkaProducer)
	log.Info("Created account service")
	return s, nil
}

// runServer runs gRPC server and HTTP gateway
func runServer() error {
	log.Info("Starting gRPC server")
	ctx := context.Background()
	cfg := config.New(ctx)
	if cfg.Mode == "DEV" {
		log.Info("Waiting for Kafka instance to be ready...")
		time.Sleep(18 * time.Second)
	}
	s, err := InitializeService(*cfg, ctx)
	if err != nil {
		return err
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GrpcPort))
	if err != nil {
		return err
	}

	server := grpc.NewServer()

	api.RegisterUserServiceServer(server, s)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Info("shutting down gRPC server...")

			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Printf("starting gRPC server on port %s...\n", cfg.GrpcPort)
	return server.Serve(listen)
}

func main() {
	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
