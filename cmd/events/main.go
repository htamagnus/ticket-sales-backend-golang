package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	httpHandler "github.com/htamagnus/ticket-sales-backend-golang/internal/events/infra/http"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/infra/repository"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/infra/service"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/usecase"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/domain"
)

var data domain.Data

func init() {
	// Carregar JSON na inicialização do programa
	file, err := os.Open("data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		panic(err)
	}
}

// @title Events API
// @version 1.0
// @description This is a server for managing events. Imersão Full Cycle
// @host localhost:8080
// @BasePath /
func main() {
	// Configuração do banco de dados
	db, err := sql.Open("mysql", "test_user:test_password@tcp(localhost:3306)/test_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repositório
	eventRepo, err := repository.NewMysqlEventRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	// URLs base específicas para cada parceiro
	partnerBaseURLs := map[int]string{
		1: "http://localhost:9000/api1",
		2: "http://localhost:9000/api2",
	}

	listEventsUseCase := usecase.NewListEventsUseCase(eventRepo)
	getEventUseCase := usecase.NewGetEventUseCase(eventRepo)
	createEventUseCase := usecase.NewCreateEventUseCase(eventRepo)
	partnerFactory := service.NewPartnerFactory(partnerBaseURLs)
	buyTicketsUseCase := usecase.NewBuyTicketsUseCase(eventRepo, partnerFactory)
	createSpotsUseCase := usecase.NewCreateSpotsUseCase(eventRepo)
	listSpotsUseCase := usecase.NewListSpotsUseCase(eventRepo)

	// Handlers HTTP
	eventsHandler := httpHandler.NewEventsHandler(
		listEventsUseCase,
		getEventUseCase,
		createEventUseCase,
		buyTicketsUseCase,
		createSpotsUseCase,
		listSpotsUseCase,
		&data,
	)

	router := httprouter.New()
	router.GET("/events", eventsHandler.ListEvents)
	router.GET("/events/:eventID", eventsHandler.GetEvent)
	router.GET("/events/:eventID/spots", eventsHandler.ListSpots)
	router.POST("/events/:eventID/spots/:spotID/reserve", eventsHandler.ReserveSpot)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Canal para escutar sinais do sistema operacional
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		// Recebido sinal de interrupção, iniciando o graceful shutdown
		log.Println("Recebido sinal de interrupção, iniciando o graceful shutdown...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Erro no graceful shutdown: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	// Iniciando o servidor HTTP
	log.Println("Servidor HTTP rodando na porta 8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar o servidor HTTP: %v\n", err)
	}

	<-idleConnsClosed
	log.Println("Servidor HTTP finalizado")
}
