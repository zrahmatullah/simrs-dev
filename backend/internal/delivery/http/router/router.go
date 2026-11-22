package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"simrs-asoy/backend/internal/config"
	"simrs-asoy/backend/internal/delivery/http/handler"
	"simrs-asoy/backend/internal/middleware"
)

type RouterParams struct {
	Config             *config.Config
	AuthHandler        *handler.AuthHandler
	UserHandler        *handler.UserHandler
	PatientHandler     *handler.PatientHandler
	VisitHandler       *handler.VisitHandler
	AppointmentHandler *handler.AppointmentHandler
	ClinicalHandler    *handler.ClinicalHandler
	PharmacyHandler    *handler.PharmacyHandler
	AncillaryHandler   *handler.AncillaryHandler
	BillingHandler     *handler.BillingHandler
	InventoryHandler   *handler.InventoryHandler
	MasterHandler      *handler.MasterHandler
	AuditHandler       *handler.AuditHandler
}

func SetupRouter(p RouterParams) *chi.Mux {
	r := chi.NewRouter()

	// OWASP Rate Limiters
	globalLimiter := middleware.NewIPRateLimiter(120, time.Minute)      // 120 req/min for general API
	loginLimiter := middleware.NewIPRateLimiter(5, time.Minute)         // 5 login attempts/min anti-brute force

	// Global Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders())                                 // OWASP Top 10 Security Headers
	r.Use(middleware.BodyLimit(2 << 20))                                // 2MB Max Payload limit
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimit(globalLimiter, ""))

	// Custom JSON 404 & 405 handlers to prevent plain-text "404 page not found"
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.RespondNotFound(w, "Endpoint tidak ditemukan: "+r.Method+" "+r.URL.Path)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.RespondError(w, http.StatusMethodNotAllowed, "Metode HTTP tidak diizinkan: "+r.Method+" "+r.URL.Path, nil)
	})

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.RespondSuccess(w, "SIMRS-ASOY Backend API is Healthy and Operational", map[string]string{
			"status":  "UP",
			"version": "1.0.0",
		})
	})

	// API v1 Routes
	r.Route("/api/v1", func(v1 chi.Router) {
		// Public Auth with Brute-Force Rate Limiting
		v1.With(middleware.RateLimit(loginLimiter, "Terlalu banyak percobaan login. Silakan tunggu 1 menit demi keamanan.")).
			Post("/auth/login", p.AuthHandler.Login)
		v1.Post("/auth/refresh", p.AuthHandler.RefreshToken)
		v1.Post("/auth/logout", p.AuthHandler.Logout)

		// Protected Routes
		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.Auth(p.Config))

			// Auth User Me
			auth.Get("/auth/me", p.AuthHandler.Me)

			// Master Data (Read-only for all operational roles)
			auth.Get("/master/roles", p.MasterHandler.GetRoles)
			auth.Get("/master/poli", p.MasterHandler.GetPoliUnits)
			auth.With(middleware.RequireRoles("SUPERADMIN")).Post("/master/poli", p.MasterHandler.CreatePoli)
			auth.Get("/master/faskes", p.MasterHandler.GetFaskes)
			auth.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR", "NURSE")).Post("/master/faskes", p.MasterHandler.CreateFaskes)
			auth.Get("/master/spesialis", p.MasterHandler.GetSpesialis)
			auth.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR", "NURSE")).Post("/master/spesialis", p.MasterHandler.CreateSpesialis)
			auth.Get("/master/guarantors", p.MasterHandler.GetGuarantors)
			auth.Get("/master/icd10", p.MasterHandler.GetICD10)
			auth.Get("/master/icd9", p.MasterHandler.GetICD9)
			auth.Get("/master/categories", p.MasterHandler.GetCategories)
			auth.Get("/master/promos", p.MasterHandler.GetPromos)
			auth.With(middleware.RequireRoles("SUPERADMIN", "CASHIER")).Post("/master/promos", p.MasterHandler.CreatePromo)
			auth.Get("/master/suppliers", p.MasterHandler.GetSuppliers)
			auth.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS", "PHARMACY")).Post("/master/suppliers", p.MasterHandler.CreateSupplier)
			auth.Get("/master/procedures", p.ClinicalHandler.GetMasterProcedures)

			// 1. SUPERADMIN: User & Role Management & Audit Trail
			auth.Group(func(admin chi.Router) {
				admin.Use(middleware.RequireRoles("SUPERADMIN"))
				admin.Get("/users", p.UserHandler.GetUsers)
				admin.Get("/users/{id}", p.UserHandler.GetUserByID)
				admin.Post("/users", p.UserHandler.CreateUser)
				admin.Put("/users/{id}", p.UserHandler.UpdateUser)
				admin.Delete("/users/{id}", p.UserHandler.DeleteUser)
				admin.Get("/audit/logs", p.AuditHandler.GetAuditLogs)
			})

			// 2. PATIENT RECORDS (Registration, Nurse, Doctor, Cashier, etc.)
			auth.Group(func(pat chi.Router) {
				pat.Use(middleware.RequireRoles("SUPERADMIN", "REGISTRATION", "CALL_CENTER", "NURSE", "DOCTOR", "CASHIER"))
				pat.Get("/patients", p.PatientHandler.GetPatients)
				pat.Get("/patients/{id}", p.PatientHandler.GetPatientByID)
				pat.Get("/patients/rm/{no_rm}", p.PatientHandler.GetPatientByNoRM)
				pat.Post("/patients", p.PatientHandler.CreatePatient)
				pat.Put("/patients/{id}", p.PatientHandler.UpdatePatient)
			})
			auth.With(middleware.RequireRoles("SUPERADMIN")).Delete("/patients/{id}", p.PatientHandler.DeletePatient)

			// 3. VISITS & QUEUE (Rawat Jalan / Antrean)
			auth.Group(func(vis chi.Router) {
				vis.Get("/visits", p.VisitHandler.GetVisits)
				vis.Get("/visits/metrics", p.VisitHandler.GetQueueMetrics)
				vis.Get("/visits/{id}", p.VisitHandler.GetVisitByID)
				vis.With(middleware.RequireRoles("SUPERADMIN", "REGISTRATION")).Post("/visits", p.VisitHandler.CreateVisit)
				vis.Patch("/visits/{id}/status", p.VisitHandler.UpdateStatus)
			})

			// 4. CALL CENTER: Appointments
			auth.Group(func(cc chi.Router) {
				cc.Use(middleware.RequireRoles("SUPERADMIN", "CALL_CENTER", "REGISTRATION", "DOCTOR"))
				cc.Get("/appointments", p.AppointmentHandler.GetAppointments)
				cc.Get("/appointments/{id}", p.AppointmentHandler.GetAppointmentByID)
				cc.Post("/appointments", p.AppointmentHandler.CreateAppointment)
				cc.Post("/appointments/{id}/reschedule", p.AppointmentHandler.Reschedule)
				cc.Post("/appointments/{id}/cancel", p.AppointmentHandler.Cancel)
				cc.Post("/appointments/{id}/checkin", p.AppointmentHandler.CheckIn)
			})

			// 5. NURSE: Vital Signs TTV, Triase & Asesmen
			auth.Group(func(nur chi.Router) {
				nur.Get("/clinical/vitals/visit/{visit_id}", p.ClinicalHandler.GetVitalSignsByVisit)
				nur.Get("/clinical/vitals/patient/{patient_id}", p.ClinicalHandler.GetVitalSignsByPatient)
				nur.With(middleware.RequireRoles("SUPERADMIN", "NURSE")).Post("/clinical/vitals", p.ClinicalHandler.CreateVitalSign)
				nur.Get("/clinical/assessments/visit/{visit_id}", p.ClinicalHandler.GetNursingAssessment)
				nur.With(middleware.RequireRoles("SUPERADMIN", "NURSE")).Post("/clinical/assessments", p.ClinicalHandler.CreateNursingAssessment)
			})

			// 6. DOCTOR: EMR SOAP, Procedures & Medical Certificates
			auth.Group(func(doc chi.Router) {
				doc.Get("/clinical/emr/visit/{visit_id}", p.ClinicalHandler.GetMedicalRecordByVisit)
				doc.Get("/clinical/emr/patient/{patient_id}", p.ClinicalHandler.GetMedicalRecordsByPatient)
				doc.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR")).Post("/clinical/emr", p.ClinicalHandler.CreateMedicalRecord)

				doc.Get("/clinical/procedures/visit/{visit_id}", p.ClinicalHandler.GetVisitProcedures)
				doc.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR")).Post("/clinical/procedures", p.ClinicalHandler.CreateVisitProcedure)
				doc.Get("/clinical/assessments/visit/{visit_id}", p.ClinicalHandler.GetNursingAssessment)

				doc.Get("/clinical/certificates/visit/{visit_id}", p.ClinicalHandler.GetMedicalCertificateByVisit)
				doc.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR")).Post("/clinical/certificates", p.ClinicalHandler.CreateMedicalCertificate)
			})

			// 7. PHARMACY: Prescriptions & Dispensing
			auth.Group(func(pha chi.Router) {
				pha.Get("/pharmacy/prescriptions", p.PharmacyHandler.GetPrescriptions)
				pha.Get("/pharmacy/prescriptions/{id}", p.PharmacyHandler.GetPrescriptionByID)
				pha.Get("/pharmacy/prescriptions/visit/{visit_id}", p.PharmacyHandler.GetPrescriptionByVisit)
				pha.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR")).Post("/pharmacy/prescriptions", p.PharmacyHandler.CreatePrescription)
				pha.With(middleware.RequireRoles("SUPERADMIN", "PHARMACY")).Post("/pharmacy/prescriptions/{id}/review", p.PharmacyHandler.ReviewPrescription)
				pha.With(middleware.RequireRoles("SUPERADMIN", "PHARMACY")).Post("/pharmacy/prescriptions/{id}/dispense", p.PharmacyHandler.DispensePrescription)
			})

			// 8. SUPPORT / PENUNJANG: Orders (Lab/Rad/Fisio) - support both /support and /ancillary
			registerOrderRoutes := func(prefix string) {
				auth.Group(func(sup chi.Router) {
					sup.Get("/"+prefix+"/orders", p.AncillaryHandler.GetOrders)
					sup.Get("/"+prefix+"/orders/{id}", p.AncillaryHandler.GetOrderByID)
					sup.Get("/"+prefix+"/orders/visit/{visit_id}", p.AncillaryHandler.GetOrdersByVisit)
					sup.With(middleware.RequireRoles("SUPERADMIN", "DOCTOR")).Post("/"+prefix+"/orders", p.AncillaryHandler.CreateOrder)
					sup.With(middleware.RequireRoles("SUPERADMIN", "SUPPORT")).Post("/"+prefix+"/orders/{id}/result", p.AncillaryHandler.UpdateResult)
					sup.With(middleware.RequireRoles("SUPERADMIN", "SUPPORT")).Put("/"+prefix+"/orders/{id}/result", p.AncillaryHandler.UpdateResult)
				})
			}
			registerOrderRoutes("support")
			registerOrderRoutes("ancillary")

			// 9. CASHIER & BILLING: Invoices & Shifts - support both /cashier and /billing
			auth.Group(func(csh chi.Router) {
				csh.Use(middleware.RequireRoles("SUPERADMIN", "CASHIER"))

				// Invoices
				csh.Get("/cashier/invoices", p.BillingHandler.GetInvoices)
				csh.Get("/billing/invoices", p.BillingHandler.GetInvoices)
				csh.Get("/cashier/invoices/{id}", p.BillingHandler.GetInvoiceByID)
				csh.Get("/billing/invoices/{id}", p.BillingHandler.GetInvoiceByID)
				csh.Get("/cashier/invoices/visit/{visit_id}", p.BillingHandler.GetInvoiceByVisit)
				csh.Get("/billing/invoices/visit/{visit_id}", p.BillingHandler.GetInvoiceByVisit)
				csh.Post("/cashier/invoices/generate/{visit_id}", p.BillingHandler.GenerateBillingFromVisit)
				csh.Post("/billing/invoices/generate/{visit_id}", p.BillingHandler.GenerateBillingFromVisit)
				csh.Post("/billing/generate/{visit_id}", p.BillingHandler.GenerateBillingFromVisit)
				csh.Post("/cashier/invoices", p.BillingHandler.CreateInvoice)
				csh.Post("/billing/invoices", p.BillingHandler.CreateInvoice)
				csh.Post("/cashier/invoices/{id}/pay", p.BillingHandler.ProcessPayment)
				csh.Post("/billing/invoices/{id}/pay", p.BillingHandler.ProcessPayment)

				// Shifts (supporting both /cashier/shift/... and /billing/shifts/...)
				csh.Get("/cashier/shift/active", p.BillingHandler.GetActiveShift)
				csh.Get("/billing/shifts/active", p.BillingHandler.GetActiveShift)
				csh.Post("/cashier/shift/open", p.BillingHandler.OpenShift)
				csh.Post("/billing/shifts/open", p.BillingHandler.OpenShift)
				csh.Post("/cashier/shift/{shift_id}/close", p.BillingHandler.CloseShift)
				csh.Post("/billing/shifts/{shift_id}/close", p.BillingHandler.CloseShift)
				csh.Get("/cashier/shift/history", p.BillingHandler.GetShiftHistory)
				csh.Get("/billing/shifts/history", p.BillingHandler.GetShiftHistory)
			})

			// 10. LOGISTICS & INVENTORY: Stock, PR, PO & GRN (supporting both /logistics and /inventory prefixes)
			registerInventoryRoutes := func(prefix string) {
				auth.Group(func(log chi.Router) {
					// Items
					log.Get("/"+prefix+"/items", p.InventoryHandler.GetItems)
					log.Get("/"+prefix+"/items/{id}", p.InventoryHandler.GetItemByID)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/items", p.InventoryHandler.CreateItem)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Put("/"+prefix+"/items/{id}", p.InventoryHandler.UpdateItem)

					// Purchase Requisitions (PR)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS", "PHARMACY")).Get("/"+prefix+"/pr", p.InventoryHandler.GetPRs)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS", "PHARMACY")).Post("/"+prefix+"/pr", p.InventoryHandler.CreatePR)

					// Purchase Orders (PO) - support both singular /po and plural /pos
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Get("/"+prefix+"/po", p.InventoryHandler.GetPOs)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Get("/"+prefix+"/pos", p.InventoryHandler.GetPOs)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/po", p.InventoryHandler.CreatePO)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/pos", p.InventoryHandler.CreatePO)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/po/{id}/receive", p.InventoryHandler.ReceivePO)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/pos/{id}/receive", p.InventoryHandler.ReceivePO)

					// Goods Receipt Notes (GRN)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Get("/"+prefix+"/grn", p.InventoryHandler.GetGRNs)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS")).Post("/"+prefix+"/grn", p.InventoryHandler.CreateGRN)

					// Stock Mutations
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS", "PHARMACY")).Post("/"+prefix+"/mutations", p.InventoryHandler.CreateStockMutation)
					log.With(middleware.RequireRoles("SUPERADMIN", "LOGISTICS", "PHARMACY")).Get("/"+prefix+"/mutations", p.InventoryHandler.GetStockMutations)
				})
			}
			registerInventoryRoutes("logistics")
			registerInventoryRoutes("inventory")
		})
	})

	return r
}
