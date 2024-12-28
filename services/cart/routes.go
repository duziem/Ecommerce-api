package cart

import (
	"fmt"
	"net/http"

	"github.com/duziem/ecommerce_proj/services/auth"
	"github.com/duziem/ecommerce_proj/types"
	"github.com/duziem/ecommerce_proj/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store      types.ProductStore
	orderStore types.OrderStore
	userStore  types.UserStore
}

func NewHandler(
	store types.ProductStore,
	orderStore types.OrderStore,
	userStore types.UserStore,
) *Handler {
	return &Handler{
		store:      store,
		orderStore: orderStore,
		userStore:  userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/cart/checkout", auth.WithJWTAuth(h.handleCheckout, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	var cart types.CartCheckoutPayload
	if err := utils.ParseJSON(r, &cart); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(cart); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	productIDs, err := getCartItemsIDs(cart.Items)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Use a transaction for atomicity
	tx, err := h.store.BeginTransaction()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err))
		return
	}
	defer tx.Rollback() // Ensure rollback on failure

	// Fetch products within the transaction to ensure consistency
	products, err := h.store.GetProductsByIDWithLock(tx, productIDs)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to fetch products: %v", err))
		return
	}

	// create a map of products for easier access
	productsMap := make(map[int]types.Product)
	for _, product := range products {
		productsMap[product.ID] = product
	}

	// Validate stock availability
	if err := checkIfCartIsInStock(cart.Items, productsMap); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Calculate total price
	totalPrice := calculateTotalPrice(cart.Items, productsMap)

	// Update product quantities
	if err := h.store.UpdateProductQuantities(tx, cart.Items); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update product quantities: %v", err))
		return
	}

	// Create order
	orderID, err := h.orderStore.CreateOrder(tx, types.Order{
		UserID:  userID,
		Total:   totalPrice,
		Status:  "pending",
		Address: cart.Address, // Use address from the payload
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create order: %v", err))
		return
	}

	// Create order items
	if err := h.orderStore.CreateOrderItems(tx, orderID, cart.Items, productsMap); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create order items: %v", err))
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err))
		return
	}

	// Respond with success
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"total_price": totalPrice,
		"order_id":    orderID,
	})
}
