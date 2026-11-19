package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"colegio-backend/database"
	"colegio-backend/handlers"
)

func main() {
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	archivoDB := os.Getenv("DB_FILE")
	if archivoDB == "" {
		archivoDB = "colegio.db"
	}

	// 1. Inicializar la base de datos SQLite y poblar con datos iniciales
	log.Printf("Iniciando conexión con SQLite (%s)...", archivoDB)
	db, err := database.InitDB(archivoDB)
	if err != nil {
		log.Fatalf("Error crítico al iniciar SQLite: %v", err)
	}
	defer db.Close()
	log.Println("✓ Base de datos SQLite lista y conectada correctamente.")

	// 2. Definir multiplexor de rutas
	mux := http.NewServeMux()

	// Endpoints REST de la API del Colegio
	mux.HandleFunc("/api/alumnos", handlers.HandleAlumnos)
	mux.HandleFunc("/api/alumnos/", handlers.HandleAlumnoPorID)
	mux.HandleFunc("/api/notas", handlers.HandleActualizarNotas)
	mux.HandleFunc("/api/metricas", handlers.HandleMetricas)

	// Endpoint de comprobación de salud (Health Check)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","servicio":"API Colegio - Go + SQLite"}`)
	})

	// 3. Aplicar middleware CORS para permitir peticiones desde React y Angular
	handlerConCORS := handlers.MiddlewareCORS(mux)

	// 4. Mensaje estético de bienvenida en la consola
	fmt.Println("================================================================")
	fmt.Println("   SISTEMA DE GESTIÓN ACADÉMICA - BACKEND EN GO (GOLANG)        ")
	fmt.Println("================================================================")
	fmt.Printf("✓ Servidor HTTP activo en:       http://localhost:%s\n", puerto)
	fmt.Printf("✓ Archivo de base de datos:     %s (SQLite)\n", archivoDB)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("Endpoints disponibles para el Frontend (React / Angular):")
	fmt.Printf("  • GET    http://localhost:%s/api/alumnos             (Listar todos)\n", puerto)
	fmt.Printf("  • GET    http://localhost:%s/api/alumnos?dni=72458912 (Buscar por DNI)\n", puerto)
	fmt.Printf("  • POST   http://localhost:%s/api/alumnos             (Matricular alumno)\n", puerto)
	fmt.Printf("  • PUT    http://localhost:%s/api/notas               (Calificar alumno)\n", puerto)
	fmt.Printf("  • DELETE http://localhost:%s/api/alumnos/{id}        (Eliminar alumno)\n", puerto)
	fmt.Printf("  • GET    http://localhost:%s/api/metricas            (Métricas generales)\n", puerto)
	fmt.Println("================================================================")
	fmt.Println("Esperando peticiones del frontend... Presiona Ctrl+C para salir.")

	// 5. Iniciar servidor HTTP
	if err := http.ListenAndServe(":"+puerto, handlerConCORS); err != nil {
		log.Fatalf("Error al ejecutar el servidor HTTP: %v", err)
	}
}
