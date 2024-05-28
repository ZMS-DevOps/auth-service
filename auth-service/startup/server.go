package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/api"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/persistence"
	"github.com/mmmajder/zms-devops-auth-service/startup/config"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type Server struct {
	config *config.Config
	router *mux.Router
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
		router: mux.NewRouter(),
	}
}

func (server *Server) Start() {
	server.setupHandlers()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", server.config.Port), server.router))
}

func (server *Server) setupHandlers() {
	keycloakService := server.initKeycloakService()
	emailService := server.initEmailService()

	mongoClient := server.initMongoClient()
	verificationStore := server.initVerificationStore(mongoClient)
	authService := server.initAuthService(verificationStore, keycloakService)
	authHandler := server.initAuthHandler(authService, emailService)
	authHandler.Init(server.router)
}

func (server *Server) initKeycloakService() *application.KeycloakService {
	return application.NewKeycloakService(&http.Client{}, server.config.IdentityProviderHost)
}

func (server *Server) initEmailService() *application.EmailService {
	return application.NewEmailService()
}

func (server *Server) initAuthService(store domain.VerificationStore, keycloakService *application.KeycloakService) *application.AuthService {
	return application.NewAuthService(store, &http.Client{}, keycloakService)
}

func (server *Server) initAuthHandler(authService *application.AuthService, emailService *application.EmailService) *api.AuthHandler {
	return api.NewAuthHandler(authService, emailService)
}

func (server *Server) initMongoClient() *mongo.Client {
	client, err := persistence.GetClient(server.config.DBUsername, server.config.DBPassword, server.config.DBHost, server.config.DBPort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (server *Server) initVerificationStore(client *mongo.Client) domain.VerificationStore {
	return persistence.NewVerificationMongoDBStore(client)
}
