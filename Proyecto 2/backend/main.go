package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"proyecto1/DiskManagement"
	"syscall"

	"github.com/joho/godotenv"
)

// Middleware para habilitar CORS
func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	// Cargar las variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error al cargar el archivo .env")
	}
}

// Función para limpiar las particiones montadas
func CleanMountedPartitions() {
	DiskManagement.CleanMountedPartitions()
	fmt.Println("Particiones montadas limpiadas por finalización del programa.")
}

func main() {
	// Cargar el modo de la aplicación desde las variables de entorno
	mode := os.Getenv("ANALYZER_MODE")

	// Obtener el puerto desde las variables de entorno
	port := os.Getenv("PORT")

	// Imprimir el modo en el que se ejecuta el servidor
	fmt.Printf("Iniciando en modo: %s\n", mode)

	if mode == "development" {
		fmt.Println("Ejecución en modo de desarrollo")
	}

	// Capturar señales del sistema para limpiar antes de finalizar
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		fmt.Println("Recibida señal:", sig)
		CleanMountedPartitions()
		os.Exit(0)
	}()

	// Crear un multiplexor para manejar las rutas
	mux := http.NewServeMux()
	mux.HandleFunc("/api/mkdisk", MkDiskHandler)
	mux.HandleFunc("/api/fdisk", FdiskHandler)
	mux.HandleFunc("/api/mount", MountHandler)
	mux.HandleFunc("/api/unmount", UnmountHandler)
	mux.HandleFunc("/api/mkfs", MkfsHandler)
	mux.HandleFunc("/api/login", LoginHandler)
	mux.HandleFunc("/api/rep", RepHandler)
	mux.HandleFunc("/api/readmbr", ReadMBRHandler)

	// Iniciar el servidor con el middleware de CORS habilitado
	fmt.Printf("Servidor ejecutándose en el puerto %s\n", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, enableCors(mux)); err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}
