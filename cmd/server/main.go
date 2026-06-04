package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gametime/order-state-machine/internal/api"
	"github.com/gametime/order-state-machine/internal/order"
	"github.com/gametime/order-state-machine/internal/payment"
)

func main() {
	stub := payment.NewStubFromEnv(
		envBool("PAYMENT_AUTHORIZE_FAIL"),
		envBool("COMPLETE_FAIL"),
		envBool("VOID_FAIL"),
	)

	store := order.NewMemoryStore()
	svc := order.NewService(store, stub)
	handler := api.NewHandler(svc)

	mux := http.NewServeMux()
	handler.Register(mux)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}
