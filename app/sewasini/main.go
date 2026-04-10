package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	handler "sewasini/app/handler"
	authmiddleware "sewasini/app/sewasini/middleware"
	customvalidator "sewasini/app/sewasini/validator"
	"sewasini/database"
	repositorybooking "sewasini/repository/booking"
	repositorycategory "sewasini/repository/category"
	repositoryreview "sewasini/repository/review"
	repositoryroom "sewasini/repository/room"
	repositorytransaction "sewasini/repository/transaction"
	repositoryuser "sewasini/repository/user"
	servicebooking "sewasini/service/booking"
	servicecategory "sewasini/service/category"
	servicereview "sewasini/service/review"
	serviceroom "sewasini/service/room"
	servicetransaction "sewasini/service/transaction"
	serviceuser "sewasini/service/user"
)

func main() {
	loadEnv()

	database.InitDB()
	defer database.CloseDB()

	e := echo.New()
	e.HideBanner = false
	e.Validator = &customvalidator.CustomValidator{Validator: validator.New()}
	e.Use(middleware.Recover())
	e.Use(authmiddleware.RequestHitLogger())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	userRepo := repositoryuser.NewRepository(database.DB)
	userService := serviceuser.NewService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	roomRepo := repositoryroom.NewRepository(database.DB)
	roomService := serviceroom.NewService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)
	categoryRepo := repositorycategory.NewRepository(database.DB)
	categoryService := servicecategory.NewService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	reviewRepo := repositoryreview.NewRepository(database.DB)
	reviewService := servicereview.NewService(reviewRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)
	bookingRepo := repositorybooking.NewRepository(database.DB)
	bookingService := servicebooking.NewService(bookingRepo, roomRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)
	transactionRepo := repositorytransaction.NewRepository(database.DB)
	transactionService := servicetransaction.NewService(transactionRepo, bookingRepo, userRepo)
	paymentHandler := handler.NewPaymentHandler(transactionService)

	registerRoutes(e.Group("/api/v1"), userHandler, roomHandler, categoryHandler, reviewHandler, bookingHandler, paymentHandler)
	registerRoutes(e.Group("/api"), userHandler, roomHandler, categoryHandler, reviewHandler, bookingHandler, paymentHandler)

	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	addr := host + ":" + port
	if addr == ":" {
		addr = ":8080"
	}

	startErr := make(chan error, 1)
	go func() {
		log.Printf("sewasini server berjalan pada %s", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			startErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-startErr:
		e.Logger.Fatal(err)
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func registerRoutes(
	api *echo.Group,
	userHandler *handler.UserHandler,
	roomHandler *handler.RoomHandler,
	categoryHandler *handler.CategoryHandler,
	reviewHandler *handler.ReviewHandler,
	bookingHandler *handler.BookingHandler,
	paymentHandler *handler.PaymentHandler,
) {
	{
		usersGroup := api.Group("/users")
		{
			usersGroup.POST("/register", userHandler.RegisterUser)
			usersGroup.POST("/login", userHandler.LoginUser)
			usersGroup.GET("/login", userHandler.LoginUser)
			usersGroup.POST("/send-otp", userHandler.SendOTP)
			usersGroup.POST("/verify-otp", userHandler.VerifyOTP)

			protectedUsersGroup := usersGroup.Group("")
			protectedUsersGroup.Use(authmiddleware.BearerAuth())
			protectedUsersGroup.Use(authmiddleware.AdminOnly())
			{
				protectedUsersGroup.GET("", userHandler.ListUsers)
				protectedUsersGroup.GET("/:id", userHandler.GetUserByID)
				protectedUsersGroup.PUT("/:id", userHandler.UpdateUser)
				protectedUsersGroup.DELETE("/:id", userHandler.DeleteUser)
			}
		}

		roomsGroup := api.Group("/ruangan")
		{
			roomsGroup.GET("", roomHandler.ListRooms)
			roomsGroup.GET("/search", roomHandler.ListRooms)
			roomsGroup.GET("/filter", roomHandler.ListRooms)
			roomsGroup.GET("/:id", roomHandler.GetRoomByID)
		}

		categoriesGroup := api.Group("/categories")
		{
			categoriesGroup.GET("", categoryHandler.ListCategories)
			categoriesGroup.GET("/:id", categoryHandler.GetCategoryByID)

			adminCategoriesGroup := categoriesGroup.Group("")
			adminCategoriesGroup.Use(authmiddleware.BearerAuth())
			adminCategoriesGroup.Use(authmiddleware.AdminOnly())
			{
				adminCategoriesGroup.POST("", categoryHandler.CreateCategory)
				adminCategoriesGroup.PUT("/:id", categoryHandler.UpdateCategory)
				adminCategoriesGroup.DELETE("/:id", categoryHandler.DeleteCategory)
			}
		}

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", userHandler.RegisterUser)
			authGroup.POST("/login", userHandler.LoginUser)
			authGroup.POST("/verify-otp", userHandler.VerifyOTP)
			authGroup.POST("/resend-otp", userHandler.SendOTP)
			authGroup.POST("/forgot-password", userHandler.ForgotPassword)
			authGroup.POST("/reset-password", userHandler.ResetPassword)
		}

		kategoriGroup := api.Group("/kategori")
		{
			kategoriGroup.GET("", categoryHandler.ListCategories)
		}

		reviewsGroup := api.Group("/reviews")
		reviewsGroup.Use(authmiddleware.BearerAuth())
		{
			reviewsGroup.POST("", reviewHandler.CreateReview)
			reviewsGroup.GET("", reviewHandler.ListReviews)
			reviewsGroup.GET("/ruangan/:id", reviewHandler.ListReviewsByRoomID)
			reviewsGroup.GET("/:id", reviewHandler.GetReviewByID)
			reviewsGroup.PUT("/:id", reviewHandler.UpdateReview)
			reviewsGroup.DELETE("/:id", reviewHandler.DeleteReview)
		}

		bookingsGroup := api.Group("/bookings")
		bookingsGroup.Use(authmiddleware.BearerAuth())
		{
			bookingsGroup.POST("", bookingHandler.CreateBooking)
			bookingsGroup.GET("", bookingHandler.ListBookings)
			bookingsGroup.GET("/:id/status", bookingHandler.GetBookingStatus)
			bookingsGroup.GET("/:id", bookingHandler.GetBookingByID)
			bookingsGroup.PUT("/:id", bookingHandler.UpdateBooking)
			bookingsGroup.DELETE("/:id", bookingHandler.CancelBooking)
		}

		paymentsGroup := api.Group("/payments")
		{
			protectedPaymentsGroup := paymentsGroup.Group("")
			protectedPaymentsGroup.Use(authmiddleware.BearerAuth())
			protectedPaymentsGroup.POST("", paymentHandler.CreatePayment)
			protectedPaymentsGroup.GET("/invoice/:id", paymentHandler.GetInvoice)
			protectedPaymentsGroup.GET("/:id", paymentHandler.GetPayment)

			paymentsGroup.POST("/callback", paymentHandler.PaymentCallback)
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(authmiddleware.BearerAuth())
		adminGroup.Use(authmiddleware.AdminOnly())
		{
			adminUsersGroup := adminGroup.Group("/users")
			{
				adminUsersGroup.GET("", userHandler.ListUsers)
				adminUsersGroup.GET("/:id", userHandler.GetUserByID)
				adminUsersGroup.PUT("/:id", userHandler.UpdateUser)
				adminUsersGroup.DELETE("/:id", userHandler.DeleteUser)
			}

			adminRoomsGroup := adminGroup.Group("/ruangan")
			{
				adminRoomsGroup.POST("", roomHandler.CreateRoom)
				adminRoomsGroup.PUT("/:id", roomHandler.UpdateRoom)
				adminRoomsGroup.DELETE("/:id", roomHandler.DeleteRoom)
			}

			adminKategoriGroup := adminGroup.Group("/kategori")
			{
				adminKategoriGroup.POST("", categoryHandler.CreateCategory)
				adminKategoriGroup.PUT("/:id", categoryHandler.UpdateCategory)
				adminKategoriGroup.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			adminBookingsGroup := adminGroup.Group("/bookings")
			{
				adminBookingsGroup.GET("", bookingHandler.AdminListBookings)
				adminBookingsGroup.GET("/:id", bookingHandler.AdminGetBookingByID)
				adminBookingsGroup.PUT("/:id", bookingHandler.AdminUpdateBooking)
			}

			adminGroup.GET("/transactions", paymentHandler.AdminListTransactions)
			adminGroup.GET("/reports", paymentHandler.AdminReports)
			adminGroup.GET("/revenues", paymentHandler.AdminRevenues)
			adminGroup.GET("/dashboard", paymentHandler.AdminDashboard)
		}
	}
}

func loadEnv() {
	if err := godotenv.Load("../../.env"); err != nil {
		_ = godotenv.Load()
	}
}
