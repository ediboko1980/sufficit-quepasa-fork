package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sufficit/sufficit-quepasa-fork/controllers"
	"github.com/sufficit/sufficit-quepasa-fork/models"
)

func main() {
	// Carregando variaveis de ambiente apartir de arquivo .env
	godotenv.Load()

	// Verifica se é necessario realizar alguma migração de base de dados
	err := models.MigrateToLatest()
	if err != nil {
		log.Fatalf("Database migration error: %s", err.Error())
	}

	// Inicializando serviço de controle do whatsapp
	// De forma assíncrona
	go models.QPWhatsAppStart()

	go func() {
		m := chi.NewRouter()
		m.Handle("/metrics", promhttp.Handler())
		host := fmt.Sprintf("%s:%s", os.Getenv("METRICS_HOST"), os.Getenv("METRICS_PORT"))

		log.Println("Starting Metrics Service")
		err := http.ListenAndServe(host, m)
		if err != nil {
			log.Fatal(err)
		}
	}()

	controllers.QPWebServerStart()
}
