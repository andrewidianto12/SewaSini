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
	repositoryroom "sewasini/repository/room"
	repositoryuser "sewasini/repository/user"
	servicebooking "sewasini/service/booking"
	serviceroom "sewasini/service/room"
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
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")

	userRepo := repositoryuser.NewRepository(database.DB)
	userService := serviceuser.NewService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	roomRepo := repositoryroom.NewRepository(database.DB)
	roomService := serviceroom.NewService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)
	bookingRepo := repositorybooking.NewRepository(database.DB)
	bookingService := servicebooking.NewService(bookingRepo, roomRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)

	{
		usersGroup := api.Group("/users")
		{
			usersGroup.POST("/register", userHandler.RegisterUser)
			usersGroup.POST("/login", userHandler.LoginUser)
			usersGroup.POST("/send-otp", userHandler.SendOTP)
			usersGroup.POST("/verify-otp", userHandler.VerifyOTP)

			protectedUsersGroup := usersGroup.Group("")
			protectedUsersGroup.Use(authmiddleware.BearerAuth())
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
			roomsGroup.GET("/:id", roomHandler.GetRoomByID)
		}

		bookingsGroup := api.Group("/bookings")
		bookingsGroup.Use(authmiddleware.BearerAuth())
		{
			bookingsGroup.POST("", bookingHandler.CreateBooking)
		}
	}

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

func loadEnv() {
	if err := godotenv.Load("../../.env"); err != nil {
		_ = godotenv.Load()
	}
}
