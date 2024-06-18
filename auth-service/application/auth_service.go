package application

import (
	"encoding/json"
	"errors"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"github.com/mmmajder/zms-devops-auth-service/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type AuthService struct {
	store           domain.VerificationStore
	HttpClient      *http.Client
	KeycloakService *KeycloakService
	producer        *kafka.Producer
	loki            promtail.Client
}

func NewAuthService(store domain.VerificationStore, httpClient *http.Client, keycloakService *KeycloakService, producer *kafka.Producer, loki promtail.Client) *AuthService {
	return &AuthService{
		store:           store,
		HttpClient:      httpClient,
		KeycloakService: keycloakService,
		producer:        producer,
		loki:            loki,
	}
}

func (service *AuthService) Login(email string, password string, span trace.Span) (*dto.LoginDTO, error) {
	span.AddEvent("Establishing connection to the keycloak for logging in...")
	responseBody, err := service.KeycloakService.LoginKeycloakUser(email, password)
	if err != nil {
		return nil, err
	}

	loginDTO := &dto.LoginDTO{}
	if err := json.NewDecoder(responseBody).Decode(loginDTO); err != nil {
		util.HttpTraceError(err, "can't decode login payload", span, service.loki, "Login", "")
		return nil, errors.New("can't decode login payload")
	}
	if !service.userIsEnabled(loginDTO) {
		util.HttpTraceError(err, domain.UserNotVerifiedErrorMessage, span, service.loki, "Login", "")
		return nil, errors.New(domain.UserNotVerifiedErrorMessage)
	}
	return loginDTO, nil
}

func (service *AuthService) SignUp(email, firstName, lastName, password, address, group string, span trace.Span) (domain.Verification, error) {
	requestBody := dto.NewKeycloakDTO(email, firstName, lastName, password, address, group)

	loginDTO, err := service.getAdminLoginDTO()
	if err != nil {
		util.HttpTraceError(err, "cannot log as keycloak admin", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	span.AddEvent("Establishing connection to the keycloak for signing up...")
	userId, err := service.KeycloakService.CreateKeycloakUser(requestBody, domain.BearerSchema+loginDTO.AccessToken)
	if err != nil {
		util.HttpTraceError(err, "cannot create user", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	verification, err := service.saveVerification(userId, firstName, lastName, address)
	if err != nil {
		util.HttpTraceError(err, "cannot save verification code", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	service.produceOnUserCreatedNotification(userId, group, span)

	return *verification, nil
}

func (service *AuthService) VerifyUser(verificationId primitive.ObjectID, userId string, securityCode int) error {
	verification, err := service.store.Get(verificationId)
	if err != nil {
		return err
	}

	if _, err := service.canVerifyUser(verification, userId, securityCode); err != nil {
		return err
	}

	loginDTO, err := service.getAdminLoginDTO()
	if err != nil {
		return err
	}

	if _, err := service.KeycloakService.UpdateKeycloakUser(domain.BearerSchema+loginDTO.AccessToken, userId, dto.NewUpdateKeycloakUserDTO(userId, verification.FirstName, verification.LastName, verification.Address)); err != nil {
		return err
	}

	if err := service.store.Delete(verificationId); err != nil {
		return err
	}
	return nil
}

func (service *AuthService) canVerifyUser(verification *domain.Verification, userId string, securityCode int) (bool, error) {
	if !service.userIsCorrectForVerificationCode(verification.UserId, userId) {
		return false, errors.New("incorrect user for chosen verification")
	}
	if !service.codesAreSame(verification.Code, securityCode) {
		return false, errors.New("verification code incorrect")
	}
	return true, nil
}

func (service *AuthService) getAdminLoginDTO() (*dto.LoginDTO, error) {
	responseBody, err := service.KeycloakService.LoginKeycloakUser(domain.AdminUsername, domain.AdminPassword)
	if err != nil {
		return nil, err
	}

	loginDTO := &dto.LoginDTO{}
	if err := json.NewDecoder(responseBody).Decode(loginDTO); err != nil {
		return nil, errors.New("can't decode admin login payload")
	}

	return loginDTO, nil
}

func (service *AuthService) UpdateVerificationCode(verificationId string) (domain.Verification, error) {
	verificationPrimitiveObjectId, err := primitive.ObjectIDFromHex(verificationId)
	if err != nil {
		return domain.Verification{}, err
	}
	verification, err := service.store.Get(verificationPrimitiveObjectId)
	verification.Code = service.generateRandomVerificationCode()
	err = service.store.Update(verificationPrimitiveObjectId, verification)
	if err != nil {
		return *verification, err
	}

	return *verification, nil
}

func (service *AuthService) codesAreSame(verificationCode, inputCode int) bool {
	return verificationCode == inputCode
}

func (service *AuthService) userIsCorrectForVerificationCode(verificationUserId, authenticatedUserId string) bool {
	return verificationUserId == authenticatedUserId
}

func (service *AuthService) generateRandomVerificationCode() int {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	rng := rand.New(source)
	minNumber := 1000
	maxNumber := 9999

	return rng.Intn(maxNumber-minNumber+1) + minNumber
}

func (service *AuthService) userIsEnabled(loginDTO *dto.LoginDTO) bool {
	responseBody, err := service.KeycloakService.GetKeycloakUser(domain.BearerSchema + loginDTO.AccessToken)
	if err != nil {
		return false
	}
	userDTO := &dto.UserDTO{}
	if err := json.NewDecoder(responseBody).Decode(userDTO); err != nil {
		return false
	}

	if !userDTO.EmailVerified {
		return false
	}
	return true
}

func (service *AuthService) saveVerification(userId string, firstName string, lastName string, address string) (*domain.Verification, error) {
	verification := &domain.Verification{
		UserId:    userId,
		FirstName: firstName,
		LastName:  lastName,
		Address:   address,
		Code:      service.generateRandomVerificationCode(),
	}

	verificationId, err := service.store.Insert(verification)
	if err != nil {
		return &domain.Verification{}, err
	}
	insertedVerificationDto := dto.VerificationDTO{}
	insertedVerificationDto.Id = verificationId.Hex()
	insertedVerificationDto.UserId = userId

	return verification, nil
}

func (service *AuthService) produceOnUserCreatedNotification(userId string, role string, span trace.Span) {
	var topic = "user.created"
	span.AddEvent("Producing notification to subscribed clients")
	notificationDTO := dto.UserCreatedNotificationDTO{
		UserId: userId,
		Role:   role,
	}
	message, _ := json.Marshal(notificationDTO)
	err := service.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	if err != nil {
		log.Fatalf("Failed to produce message: %s", err)
	}

	service.producer.Flush(4 * 1000)
}
