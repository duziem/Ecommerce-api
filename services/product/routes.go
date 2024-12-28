package product

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/duziem/ecommerce_proj/services/auth"
	"github.com/duziem/ecommerce_proj/types"
	"github.com/duziem/ecommerce_proj/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.ProductStore
	userStore types.UserStore
}

func NewHandler(store types.ProductStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// get list of products
	router.HandleFunc("/products", auth.WithJWTAuth(h.handleGetProducts, h.userStore)).Methods(http.MethodGet)

	// get a single product
	router.HandleFunc("/products/{productID}", auth.WithJWTAuth(h.handleGetProduct, h.userStore)).Methods(http.MethodGet)

	// admin routes
	// create a product
	router.HandleFunc("/admin/products", auth.WithJWTAuth(auth.WithAdminRole(h.handleCreateProduct, h.userStore), h.userStore)).Methods(http.MethodPost)
	// update a product
	router.HandleFunc("/admin/products/{productID}", auth.WithJWTAuth(auth.WithAdminRole(h.handleUpdateProduct, h.userStore), h.userStore)).Methods(http.MethodPatch)
	// delete a product
	router.HandleFunc("/admin/products/{productID}", auth.WithJWTAuth(auth.WithAdminRole(h.handleDeleteProduct, h.userStore), h.userStore)).Methods(http.MethodDelete)

	// delete products
	router.HandleFunc("/admin/products", auth.WithJWTAuth(auth.WithAdminRole(h.handleDeleteProducts, h.userStore), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleDeleteProducts(w http.ResponseWriter, r *http.Request) {
	var productPayload types.DeleteProductsPayload
	if err := utils.ParseJSON(r, &productPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(productPayload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	_, err := h.store.GetProductsByID(productPayload.Ids)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}

	// delete products
	err = h.store.DeleteProducts(productPayload.Ids)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"message":   "products deleted successfully",
		"deletedAt": time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.GetProducts()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, products)
}

func (h *Handler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["productID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	productID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	product, err := h.store.GetProductByID(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	var productPayload types.UpdateProductPayload

	// Parse the request and populate the update product payload
	if err := utils.ParseJSON(r, &productPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Validate the payload (if needed)
	if err := utils.Validate.Struct(productPayload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	vars := mux.Vars(r)
	str, ok := vars["productID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	productID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	product, err := h.store.GetProductByID(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Update only provided fields
	if productPayload.Name != nil {
		product.Name = *productPayload.Name
	}
	if productPayload.Description != nil {
		product.Description = *productPayload.Description
	}
	if productPayload.Image != nil {
		product.Image = *productPayload.Image
	}
	if productPayload.Price != nil {
		product.Price = *productPayload.Price
	}
	if productPayload.Quantity != nil {
		product.Quantity = *productPayload.Quantity
	}

	err = h.store.UpdateProduct(*product) // Pass the updated Product object (dereferencing the pointer)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, product) // Return the updated product
}

func (h *Handler) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["productID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	productID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	product, err := h.store.GetProductByID(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if product == nil { // Check if the product was found
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("product not found"))
		return
	}

	err = h.store.DeleteProduct(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"message":   "product deleted successfully",
		"deletedAt": time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var product types.CreateProductPayload
	if err := utils.ParseJSON(r, &product); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(product); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	err := h.store.CreateProduct(product)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, product)
}
