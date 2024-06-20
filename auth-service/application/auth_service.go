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

func (service *AuthService) Login(email string, password string, span trace.Span, loki promtail.Client) (*dto.LoginDTO, error) {
	util.HttpTraceInfo("Establishing connection to the keycloak for logging in...", span, loki, "Login", "")
	responseBody, err := service.KeycloakService.LoginKeycloakUser(email, password, span, loki)
	if err != nil {
		return nil, err
	}

	loginDTO := &dto.LoginDTO{}
	if err := json.NewDecoder(responseBody).Decode(loginDTO); err != nil {
		util.HttpTraceError(err, "can't decode login payload", span, service.loki, "Login", "")
		return nil, errors.New("can't decode login payload")
	}
	if !service.userIsEnabled(loginDTO, span, loki) {
		util.HttpTraceError(err, domain.UserNotVerifiedErrorMessage, span, service.loki, "Login", "")
		return nil, errors.New(domain.UserNotVerifiedErrorMessage)
	}
	return loginDTO, nil
}

func (service *AuthService) SignUp(email, firstName, lastName, password, address, group string, span trace.Span, loki promtail.Client) (domain.Verification, error) {
	util.HttpTraceInfo("Signing up...", span, loki, "SignUp", "")
	requestBody := dto.NewKeycloakDTO(email, firstName, lastName, password, address, group)

	loginDTO, err := service.getAdminLoginDTO(span, loki)
	if err != nil {
		util.HttpTraceError(err, "cannot log as keycloak admin", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	util.HttpTraceInfo("Establishing connection to the keycloak for signing up...", span, loki, "SignUp", "")
	userId, err := service.KeycloakService.CreateKeycloakUser(requestBody, domain.BearerSchema+loginDTO.AccessToken, span, loki)
	if err != nil {
		util.HttpTraceError(err, "cannot create user", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	verification, err := service.saveVerification(userId, firstName, lastName, address, span, loki)
	if err != nil {
		util.HttpTraceError(err, "cannot save verification code", span, service.loki, "Login", "")
		return domain.Verification{}, err
	}

	service.produceOnUserCreatedNotification(userId, group, span, loki)

	return *verification, nil
}

func (service *AuthService) VerifyUser(verificationId primitive.ObjectID, userId string, securityCode int, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Verifying user...", span, loki, "VerifyUser", "")
	verification, err := service.store.Get(verificationId)
	if err != nil {
		return err
	}

	if _, err := service.canVerifyUser(verification, userId, securityCode); err != nil {
		return err
	}

	loginDTO, err := service.getAdminLoginDTO(span, loki)
	if err != nil {
		return err
	}

	if _, err := service.KeycloakService.UpdateKeycloakUser(domain.BearerSchema+loginDTO.AccessToken, userId, dto.NewUpdateKeycloakUserDTO(userId, verification.FirstName, verification.LastName, verification.Address), span, loki); err != nil {
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

func (service *AuthService) getAdminLoginDTO(span trace.Span, loki promtail.Client) (*dto.LoginDTO, error) {
	responseBody, err := service.KeycloakService.LoginKeycloakUser(domain.AdminUsername, domain.AdminPassword, span, loki)
	if err != nil {
		return nil, err
	}

	loginDTO := &dto.LoginDTO{}
	if err := json.NewDecoder(responseBody).Decode(loginDTO); err != nil {
		return nil, errors.New("can't decode admin login payload")
	}

	return loginDTO, nil
}

func (service *AuthService) UpdateVerificationCode(verificationId string, span trace.Span, loki promtail.Client) (domain.Verification, error) {
	util.HttpTraceInfo("Updating verification code...", span, loki, "UpdateVerificationCode", "")
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

func (service *AuthService) userIsEnabled(loginDTO *dto.LoginDTO, span trace.Span, loki promtail.Client) bool {
	responseBody, err := service.KeycloakService.GetKeycloakUser(domain.BearerSchema+loginDTO.AccessToken, span, loki)
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

func (service *AuthService) saveVerification(userId string, firstName string, lastName string, address string, span trace.Span, loki promtail.Client) (*domain.Verification, error) {
	util.HttpTraceInfo("Saving verification...", span, loki, "saveVerification", "")
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

func (service *AuthService) produceOnUserCreatedNotification(userId string, role string, span trace.Span, loki promtail.Client) {
	util.HttpTraceInfo("Producing on user created notification...", span, loki, "produceOnUserCreatedNotification", "")
	var topic = "user.created"
	util.HttpTraceInfo("Producing notification to subscribed clients", span, loki, "produceOnUserCreatedNotification", "")
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
