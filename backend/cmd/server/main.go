package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simrs-asoy/backend/internal/config"
	"simrs-asoy/backend/internal/delivery/http/handler"
	"simrs-asoy/backend/internal/delivery/http/router"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/usecase"
)

func main() {
	log.Println("Starting SIMRS-ASOY Backend Server...")

	// 1. Config
	cfg := config.LoadConfig()

	// 2. Database Connection
	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Printf("Connected to PostgreSQL database: %s", cfg.DBName)

	// 3. Repositories
	userRepo := repository.NewUserRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)
	clinicalRepo := repository.NewClinicalRepository(db)
	pharmacyRepo := repository.NewPharmacyRepository(db)
	ancillaryRepo := repository.NewAncillaryRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	masterRepo := repository.NewMasterRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// 4. Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, auditRepo, tokenRepo, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo, auditRepo)
	patientUsecase := usecase.NewPatientUsecase(patientRepo, auditRepo)
	visitUsecase := usecase.NewVisitUsecase(visitRepo, patientRepo, auditRepo)
	appointmentUsecase := usecase.NewAppointmentUsecase(appointmentRepo, visitRepo, patientRepo, auditRepo)
	clinicalUsecase := usecase.NewClinicalUsecase(clinicalRepo, visitRepo, patientRepo, auditRepo)
	pharmacyUsecase := usecase.NewPharmacyUsecase(pharmacyRepo, visitRepo, auditRepo)
	ancillaryUsecase := usecase.NewAncillaryUsecase(ancillaryRepo, visitRepo, auditRepo)
	billingUsecase := usecase.NewBillingUsecase(billingRepo, visitRepo, pharmacyRepo, ancillaryRepo, clinicalRepo, auditRepo)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, auditRepo)
	masterUsecase := usecase.NewMasterUsecase(masterRepo)
	auditUsecase := usecase.NewAuditUsecase(auditRepo)

	// 5. Handlers
	authHandler := handler.NewAuthHandler(authUsecase, userUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	patientHandler := handler.NewPatientHandler(patientUsecase)
	visitHandler := handler.NewVisitHandler(visitUsecase)
	appointmentHandler := handler.NewAppointmentHandler(appointmentUsecase)
	clinicalHandler := handler.NewClinicalHandler(clinicalUsecase)
	pharmacyHandler := handler.NewPharmacyHandler(pharmacyUsecase)
	ancillaryHandler := handler.NewAncillaryHandler(ancillaryUsecase)
	billingHandler := handler.NewBillingHandler(billingUsecase)
	inventoryHandler := handler.NewInventoryHandler(inventoryUsecase)
	masterHandler := handler.NewMasterHandler(masterUsecase)
	auditHandler := handler.NewAuditHandler(auditUsecase)

	// 6. Router Setup
	r := router.SetupRouter(router.RouterParams{
		Config:             cfg,
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		PatientHandler:     patientHandler,
		VisitHandler:       visitHandler,
		AppointmentHandler: appointmentHandler,
		ClinicalHandler:    clinicalHandler,
		PharmacyHandler:    pharmacyHandler,
		AncillaryHandler:   ancillaryHandler,
		BillingHandler:     billingHandler,
		InventoryHandler:   inventoryHandler,
		MasterHandler:      masterHandler,
		AuditHandler:       auditHandler,
	})

	// 7. Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("SIMRS-ASOY Backend listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down SIMRS-ASOY Backend...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("SIMRS-ASOY Backend exited properly")
}
