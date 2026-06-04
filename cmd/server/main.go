package main

import (
	"log"
	"net/http"

	"github.com/gametime/order-state-machine/internal/api"
	"github.com/gametime/order-state-machine/internal/order"
	"github.com/gametime/order-state-machine/internal/payment"
)

func main() {
	store := order.NewMemoryStore()
	svc := order.NewService(store, &payment.Stub{})
	handler := api.NewHandler(svc)

	mux := http.NewServeMux()
	handler.Register(mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
