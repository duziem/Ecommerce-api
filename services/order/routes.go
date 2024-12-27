package order

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/duziem/ecommerce_proj/services/auth"
	"github.com/duziem/ecommerce_proj/types"
	"github.com/duziem/ecommerce_proj/utils"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.OrderStore
	userStore types.UserStore
}

func NewHandler(store types.OrderStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// get a list of orders for a user
	router.HandleFunc("/orders", auth.WithJWTAuth(h.handleOrders, h.userStore)).Methods(http.MethodGet)
	// cancel an order
	router.HandleFunc("/orders/{orderID}", auth.WithJWTAuth(h.cancelOrderStatusUpdate, h.userStore)).Methods(http.MethodPatch)

	// admin routes
	// update the status of an order
	router.HandleFunc("/admin/orders/{orderID}", auth.WithJWTAuth(auth.WithAdminRole(h.handleOrderStatusUpdate, h.userStore), h.userStore)).Methods(http.MethodPatch)
}

func (h *Handler) cancelOrderStatusUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["orderID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing order ID"))
		return
	}

	orderID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid order ID"))
		return
	}

	order, err := h.store.GetOrderByID(orderID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if order.Status != "pending" {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("invalid order status"))
		return
	}

	err = h.store.UpdateOrderStatus(order, "cancelled")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"message": "order updated successfully"})
}

func (h *Handler) handleOrderStatusUpdate(w http.ResponseWriter, r *http.Request) {
	var orderPayload types.UpdateOrderStatusPayload

	// Parse the request and populate the update product payload
	if err := utils.ParseJSON(r, &orderPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Validate the payload (if needed)
	if err := utils.Validate.Struct(orderPayload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	vars := mux.Vars(r)
	str, ok := vars["orderID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing order ID"))
		return
	}

	orderID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid order ID"))
		return
	}

	order, err := h.store.GetOrderByID(orderID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.UpdateOrderStatus(order, orderPayload.Status)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"message": "order updated successfully"})
}

func (h *Handler) handleOrders(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	orders, err := h.store.GetOrders(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, orders)
}
