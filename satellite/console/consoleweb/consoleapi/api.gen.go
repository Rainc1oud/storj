// AUTOGENERATED BY private/apigen
// DO NOT EDIT.

package consoleapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"storj.io/storj/private/api"
	"storj.io/storj/satellite/console"
)

var ErrProjectsAPI = errs.Class("consoleapi projects api")

type ProjectManagementService interface {
	GetUserProjects(context.Context) ([]console.Project, api.HTTPError)
}

type Handler struct {
	log     *zap.Logger
	service ProjectManagementService
	auth    api.Auth
}

func NewProjectManagement(log *zap.Logger, service ProjectManagementService, router *mux.Router) *Handler {
	handler := &Handler{
		log:     log,
		service: service,
	}

	projectsRouter := router.PathPrefix("/api/v0/projects").Subrouter()
	projectsRouter.HandleFunc("/", handler.handleGetUserProjects).Methods("GET")

	return handler
}

func (h *Handler) handleGetUserProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer mon.Task()(&ctx)(&err)

	w.Header().Set("Content-Type", "application/json")

	err = h.auth.IsAuthenticated(r)
	if err != nil {
		api.ServeError(h.log, w, http.StatusUnauthorized, err)
		return
	}

	retVal, httpErr := h.service.GetUserProjects(ctx)
	if err != nil {
		api.ServeError(h.log, w, httpErr.Status, httpErr.Err)
		return
	}

	err = json.NewEncoder(w).Encode(retVal)
	if err != nil {
		h.log.Debug("failed to write json GetUserProjects response", zap.Error(ErrProjectsAPI.Wrap(err)))
	}
}
