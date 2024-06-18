package api

import (
	"encoding/json"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/request"
	"github.com/mmmajder/zms-devops-auth-service/util"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net/http"
)

type AuthHandler struct {
	authService   *application.AuthService
	emailService  *application.EmailService
	traceProvider *sdktrace.TracerProvider
	loki          promtail.Client
}

func NewAuthHandler(authService *application.AuthService, emailService *application.EmailService, traceProvider *sdktrace.TracerProvider, loki promtail.Client) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		emailService:  emailService,
		traceProvider: traceProvider,
		loki:          loki,
	}
}

func (handler *AuthHandler) Init(router *mux.Router) {
	router.HandleFunc(domain.AuthContextPath+"/login", handler.Login).Methods(http.MethodPost)
	router.HandleFunc(domain.AuthContextPath+"/signup", handler.SignUp).Methods(http.MethodPost)
	router.HandleFunc(domain.AuthContextPath+"/verify", handler.VerifyUser).Methods(http.MethodPut)
	router.HandleFunc(domain.AuthContextPath+"/send-code-again", handler.SendVerificationCodeAgain).Methods(http.MethodPost)
	router.HandleFunc(domain.AuthContextPath+"/health", handler.GetHealthCheck).Methods(http.MethodGet)
}

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "login-post")
	defer func() { span.End() }()

	var loginRequest request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		util.HttpTraceError(err, "Cannot decode login payload", span, handler.loki, "Login", "")
		handleError(w, http.StatusBadRequest, "Cannot decode login payload")
		return
	}

	if err := loginRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "Invalid login payload", span, handler.loki, "Login", "")
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := handler.authService.Login(loginRequest.Email, loginRequest.Password, span)

	if err != nil {
		util.HttpTraceError(err, err.Error(), span, handler.loki, "Login", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}

func (handler *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "signup-post")
	defer func() { span.End() }()
	var registrationRequest request.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&registrationRequest); err != nil {
		util.HttpTraceError(err, "Cannot decode registration payload", span, handler.loki, "SignUp", "")
		handleError(w, http.StatusBadRequest, "Cannot decode registration payload")
		return
	}

	if err := registrationRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "Invalid registration payload", span, handler.loki, "SignUp", "")
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}
	verification, err := handler.authService.SignUp(
		registrationRequest.Email,
		registrationRequest.FirstName,
		registrationRequest.LastName,
		registrationRequest.Password,
		registrationRequest.Address,
		registrationRequest.Group,
		span,
	)

	if err != nil {
		util.HttpTraceError(err, err.Error(), span, handler.loki, "SignUp", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	span.AddEvent("Establishing connection to the email service...")
	handler.emailService.SendEmail(domain.SubjectVerifyUser, handler.emailService.GetVerificationCodeEmailBody(registrationRequest.Email, verification))

	writeResponse(w, http.StatusCreated, dto.VerificationDTO{Id: verification.Id.Hex(), UserId: verification.UserId})
}

func (handler *AuthHandler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	var verificationRequest request.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&verificationRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid verification user profile payload")
		return
	}

	if err := verificationRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := handler.authService.VerifyUser(
		verificationRequest.VerificationId,
		verificationRequest.UserId,
		verificationRequest.SecurityCode,
	)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) SendVerificationCodeAgain(w http.ResponseWriter, r *http.Request) {
	var sendCodeAgainRequest request.SendVerificationCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&sendCodeAgainRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid payload for send new verification code method")
		return
	}

	if err := sendCodeAgainRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	verification, err := handler.authService.UpdateVerificationCode(sendCodeAgainRequest.VerificationId)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	handler.emailService.SendEmail(domain.SubjectVerifyUser, handler.emailService.GetVerificationCodeEmailBody(sendCodeAgainRequest.UserEmail, verification))

	writeResponse(w, http.StatusOK, dto.VerificationDTO{Id: verification.Id.Hex(), UserId: verification.UserId})
}

func (handler *AuthHandler) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, domain.HealthCheckMessage)
}
