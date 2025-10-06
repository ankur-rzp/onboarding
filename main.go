package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"onboarding-system/examples"
	"onboarding-system/internal/api"
	"onboarding-system/internal/config"
	"onboarding-system/internal/onboarding"
	"onboarding-system/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize onboarding service
	onboardingService := onboarding.NewService(store, cfg)

	// Auto-seed demo data if no graphs exist
	seedDemoDataIfNeeded(onboardingService)

	// Initialize API handlers
	handlers := api.NewHandlers(onboardingService)

	// Setup HTTP server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: handlers.Router(),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.Server.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// seedDemoDataIfNeeded seeds demo data if no graphs exist
func seedDemoDataIfNeeded(service *onboarding.Service) {
	ctx := context.Background()

	// Check if any graphs exist
	graphs, err := service.ListGraphs(ctx)
	if err != nil {
		log.Printf("Failed to check existing graphs: %v", err)
		return
	}

	// If no graphs exist, seed demo data
	if len(graphs) == 0 {
		log.Println("No graphs found, seeding demo data...")

		// Create unified onboarding graph
		unifiedGraph := examples.CreateUnifiedOnboardingGraph()
		if err := service.CreateGraph(ctx, unifiedGraph); err != nil {
			log.Printf("Failed to create unified graph: %v", err)
		} else {
			log.Printf("Created unified onboarding graph with ID: %s", unifiedGraph.ID)
		}

		// Create production onboarding graph
		productionGraph := examples.CreateProductionOnboardingGraph()
		if err := service.CreateGraph(ctx, productionGraph); err != nil {
			log.Printf("Failed to create production graph: %v", err)
		} else {
			log.Printf("Created production onboarding graph with ID: %s", productionGraph.ID)
		}

		log.Println("Demo data seeding completed!")
	} else {
		log.Printf("Found %d existing graphs, skipping seeding", len(graphs))
	}
}
