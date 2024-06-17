package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/request"
	"net/http"
)

type UserHandler struct {
	service *application.UserService
}

func NewUserHandler(service *application.UserService) *UserHandler {
	return &UserHandler{
		service: service,
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
	res, err := handler.service.GetUser(r.Header.Get(domain.Authorization))

	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}

func (handler *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updatingUserRequest request.UpdatingUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updatingUserRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid updating user payload")
		return
	}

	if err := updatingUserRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := handler.service.UpdateUser(
		r.Header.Get(domain.Authorization),
		id,
		updatingUserRequest.FirstName,
		updatingUserRequest.LastName,
		updatingUserRequest.Address,
	)

	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (handler *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	group := mux.Vars(r)["group"]
	if group != domain.HostRole && group != domain.GuestRole {
		handleError(w, http.StatusBadRequest, domain.InvalidGroupErrorMessage)
		return
	}

	err := handler.service.DeleteUser(r.Header.Get(domain.Authorization), id, group)
	if err != nil {
		if err.Error() == domain.UserCouldNotBeDeletedErrorMessage {
			handleError(w, http.StatusPreconditionFailed, err.Error())
		} else {
			handleError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
}

func (handler *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updatePasswordRequest request.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&updatePasswordRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid resetting user password payload")
		return
	}

	if err := updatePasswordRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := handler.service.ResetPassword(r.Header.Get(domain.Authorization), id, updatePasswordRequest.Password)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (handler *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	res, err := handler.service.GetUserById(r.Header.Get(domain.Authorization), id)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}
