package api

import (
	"encoding/json"
	"errors"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/request"
	"github.com/mmmajder/zms-devops-auth-service/util"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net/http"
)

type UserHandler struct {
	service       *application.UserService
	traceProvider *sdktrace.TracerProvider
	loki          promtail.Client
}

func NewUserHandler(service *application.UserService, traceProvider *sdktrace.TracerProvider, loki promtail.Client) *UserHandler {
	return &UserHandler{
		service:       service,
		traceProvider: traceProvider,
		loki:          loki,
	}
}

func (handler *UserHandler) Init(router *mux.Router) {
	router.HandleFunc(domain.UserContextPath, handler.GetUser).Methods(http.MethodGet)
	router.HandleFunc(domain.UserContextPath+"/{id}", handler.GetUserById).Methods(http.MethodGet)
	router.HandleFunc(domain.UserContextPath+"/{id}", handler.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc(domain.UserContextPath+"/{id}/reset-password", handler.ResetPassword).Methods(http.MethodPut)
	router.HandleFunc(domain.UserContextPath+"/{id}/{group}", handler.DeleteUser).Methods(http.MethodDelete)
}

func (handler *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "get-user-get")
	defer func() { span.End() }()

	res, err := handler.service.GetUser(r.Header.Get(domain.Authorization), span, handler.loki)

	if err != nil {
		util.HttpTraceError(err, "Failed to get user", span, handler.loki, "GetUser", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}

func (handler *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "update-user-put")
	defer func() { span.End() }()

	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("invalid user id"), "Invalid user id", span, handler.loki, "UpdateUser", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updatingUserRequest request.UpdatingUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updatingUserRequest); err != nil {
		util.HttpTraceError(err, "Invalid updating user payload", span, handler.loki, "UpdateUser", "")
		handleError(w, http.StatusBadRequest, "Invalid updating user payload")
		return
	}

	if err := updatingUserRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "Invalid request data", span, handler.loki, "UpdateUser", "")
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := handler.service.UpdateUser(
		r.Header.Get(domain.Authorization),
		id,
		updatingUserRequest.FirstName,
		updatingUserRequest.LastName,
		updatingUserRequest.Address,
		span,
		handler.loki,
	)

	if err != nil {
		util.HttpTraceError(err, "Failed to update user", span, handler.loki, "UpdateUser", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (handler *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "delete-user-delete")
	defer func() { span.End() }()
	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("invalid user id"), "Invalid user id", span, handler.loki, "DeleteUser", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	group := mux.Vars(r)["group"]
	if group != domain.HostRole && group != domain.GuestRole {
		util.HttpTraceError(errors.New("invalid group"), "Invalid group", span, handler.loki, "DeleteUser", "")
		handleError(w, http.StatusBadRequest, domain.InvalidGroupErrorMessage)
		return
	}

	err := handler.service.DeleteUser(r.Header.Get(domain.Authorization), id, group, span, handler.loki)
	if err != nil {
		if err.Error() == domain.UserCouldNotBeDeletedErrorMessage {
			util.HttpTraceError(err, "User could not be deleted", span, handler.loki, "DeleteUser", "")
			handleError(w, http.StatusPreconditionFailed, err.Error())
		} else {
			util.HttpTraceError(err, "User could not be deleted", span, handler.loki, "DeleteUser", "")
			handleError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
}

func (handler *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "reset-password-put")
	defer func() { span.End() }()
	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("invalid user id"), "Invalid user id", span, handler.loki, "ResetPassword", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updatePasswordRequest request.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&updatePasswordRequest); err != nil {
		util.HttpTraceError(err, "Invalid payload", span, handler.loki, "ResetPassword", "")
		handleError(w, http.StatusBadRequest, "Invalid resetting user password payload")
		return
	}

	if err := updatePasswordRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "Invalid request data", span, handler.loki, "ResetPassword", "")
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := handler.service.ResetPassword(r.Header.Get(domain.Authorization), id, updatePasswordRequest.Password, span, handler.loki)
	if err != nil {
		util.HttpTraceError(err, "Failed to reset password", span, handler.loki, "ResetPassword", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (handler *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "get-user-by-id-get")
	defer func() { span.End() }()
	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("invalid user id"), "Invalid user id", span, handler.loki, "GetUserById", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	res, err := handler.service.GetUserById(r.Header.Get(domain.Authorization), id, span, handler.loki)
	if err != nil {
		util.HttpTraceError(err, "Failed to get user by id", span, handler.loki, "GetUserById", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}
