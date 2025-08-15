package api

import (
	"fafnir/shared/pkg/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	UserService *Service
}

func NewUserHandler(userService *Service) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) ServeUserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/info/{id}", h.getUserInfo)

	return r
}

func (h *UserHandler) getUserInfo(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userId)

	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"errors": "Invalid user ID"})
		return
	}

	userInfo := h.UserService.GetUserInfo(id)

	utils.WriteJSON(w, http.StatusOK, userInfo)
}
