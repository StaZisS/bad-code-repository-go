package main

import (
	"log"
	"os"
	"time"

	api "courier_managment_system_go/internal/api"
	"courier_managment_system_go/internal/service"
	"courier_managment_system_go/internal/storage"
)

func main() {
	store := storage.NewMemoryStore()

	userService := service.NewUserService(store)
	if err := userService.EnsureAdminUser("admin", "admin123", "System Admin"); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	productService := service.NewProductService(store)
	vehicleService := service.NewVehicleService(store)
	routeService := service.NewRouteService(40)
	deliveryService := service.NewDeliveryService(store, routeService)
	courierService := service.NewCourierService(store)
	authService := service.NewAuthService(store, getEnv("JWT_SECRET", "dev-secret"), 24*time.Hour)

	server := api.NewServer(
		authService,
		userService,
		productService,
		vehicleService,
		deliveryService,
		courierService,
		routeService,
	)

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("starting courier management API on %s", addr)
	if err := server.Engine().Run(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
