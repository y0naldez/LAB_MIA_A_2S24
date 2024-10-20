package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"proyecto1/Analyzer"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/User"
	"strings"
)

// Estructura para los parámetros del ReadMBR
type ReadMBRParams struct {
	Path string `json:"path"`
}

// Handler para leer el MBR y devolver las particiones
func ReadMBRHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params ReadMBRParams

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		// Validaciones
		if params.Path == "" {
			http.Error(w, "La ruta es requerida", http.StatusBadRequest)
			return
		}

		// Leer el MBR y obtener las particiones
		partitions, err := DiskManagement.ListPartitions(params.Path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error al leer las particiones: %v", err), http.StatusInternalServerError)
			return
		}

		// Responder con las particiones en formato JSON
		json.NewEncoder(w).Encode(partitions)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Estructura para los parámetros de mkdisk
type MkDiskParams struct {
	Size int    `json:"size"`
	Fit  string `json:"fit"`
	Unit string `json:"unit"`
	Path string `json:"path"`
}

// Handler para el comando mkdisk
func MkDiskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params MkDiskParams

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		// Validaciones
		if params.Size <= 0 {
			http.Error(w, "El tamaño debe ser mayor a 0", http.StatusBadRequest)
			return
		}

		if params.Fit != "bf" && params.Fit != "ff" && params.Fit != "wf" {
			http.Error(w, "El ajuste debe ser 'bf', 'ff' o 'wf'", http.StatusBadRequest)
			return
		}

		if params.Unit != "k" && params.Unit != "m" {
			http.Error(w, "La unidad debe ser 'k' o 'm'", http.StatusBadRequest)
			return
		}

		if params.Path == "" {
			http.Error(w, "La ruta es requerida", http.StatusBadRequest)
			return
		}

		// Llamar a la función que ejecuta el mkdisk
		DiskManagement.Mkdisk(params.Size, params.Fit, params.Unit, params.Path)

		// Responder con éxito
		response := map[string]string{
			"message": "Disco creado exitosamente",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Estructura para los parámetros de fdisk
type FdiskParams struct {
	Size   int    `json:"size"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Type   string `json:"type"`
	Fit    string `json:"fit"`
	Delete string `json:"delete"`
	Add    int    `json:"add"`
}

// Handler para el comando fdisk
func FdiskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params FdiskParams

		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		if params.Delete != "" {
			if params.Path == "" || params.Name == "" {
				http.Error(w, "Para eliminar una partición, se requiere 'path' y 'name'.", http.StatusBadRequest)
				return
			}
			lowercaseName := strings.ToLower(params.Name)
			DiskManagement.DeletePartition(params.Path, lowercaseName, params.Delete)

			response := map[string]string{
				"message": "Partición eliminada exitosamente",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Si hay modificación de espacio
		if params.Add != 0 {
			if params.Path == "" || params.Name == "" {
				http.Error(w, "Para modificar una partición, se requiere 'path' y 'name'.", http.StatusBadRequest)
				return
			}

			lowercaseName := strings.ToLower(params.Name)
			err := DiskManagement.ModifyPartition(params.Path, lowercaseName, params.Add, params.Unit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest) // Manejo de error
				return
			}

			response := map[string]string{
				"message": "Espacio de la partición modificado exitosamente",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Creación de particiones
		if params.Size <= 0 {
			http.Error(w, "El tamaño debe ser mayor a 0", http.StatusBadRequest)
			return
		}

		if params.Path == "" {
			http.Error(w, "La ruta es requerida", http.StatusBadRequest)
			return
		}

		if params.Unit != "k" && params.Unit != "m" {
			http.Error(w, "La unidad debe ser 'k' o 'm'", http.StatusBadRequest)
			return
		}

		if params.Type != "p" && params.Type != "e" && params.Type != "l" {
			http.Error(w, "El tipo debe ser 'p', 'e', o 'l'", http.StatusBadRequest)
			return
		}

		if params.Fit != "b" && params.Fit != "f" && params.Fit != "w" {
			http.Error(w, "El ajuste debe ser 'b', 'f', o 'w'", http.StatusBadRequest)
			return
		}

		if params.Fit == "" {
			params.Fit = "w"
		}

		lowercaseName := strings.ToLower(params.Name)
		DiskManagement.Fdisk(params.Size, params.Path, lowercaseName, params.Unit, params.Type, params.Fit)

		response := map[string]string{
			"message": "Partición creada exitosamente",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Estructura para los parámetros de mount
type MountParams struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// Handler para el comando mount
func MountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var params MountParams

	// Decodificar el cuerpo JSON de la solicitud
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Error al procesar la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validaciones
	if params.Path == "" || params.Name == "" {
		http.Error(w, "El path y el nombre son obligatorios", http.StatusBadRequest)
		return
	}

	// Convertir el nombre a minúsculas y eliminar cualquier comilla adicional
	lowercaseName := strings.ToLower(strings.Trim(params.Name, "\""))

	// Imprimir el nombre procesado (opcional, para depurar)
	//fmt.Println("Montando partición con nombre en minúsculas (sin comillas):", lowercaseName)

	// Llamar a la función que ejecuta el mount (sin capturar un valor de retorno)
	DiskManagement.Mount(params.Path, lowercaseName)

	// Responder con éxito
	response := map[string]string{
		"message": "Partición montada exitosamente",
	}
	json.NewEncoder(w).Encode(response)
}

// Estructura para los parámetros de unmount
type UnmountParams struct {
	ID string `json:"id"`
}

// Handler para el comando unmount
func UnmountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var params UnmountParams

	// Decodificar el cuerpo JSON de la solicitud
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Error al procesar la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validaciones
	if params.ID == "" {
		http.Error(w, "El ID es obligatorio", http.StatusBadRequest)
		return
	}

	// Imprimir el ID procesado (opcional, para depuración)
	//fmt.Println("Desmontando partición con ID:", params.ID)

	// Llamar a la función que ejecuta el unmount
	DiskManagement.Unmount(params.ID)

	// Responder con éxito
	response := map[string]string{
		"message": "Partición desmontada exitosamente",
	}
	json.NewEncoder(w).Encode(response)
}

// Estructura para los parámetros de mkfs
type MkfsParams struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	FS   string `json:"fs"`
}

// Handler para el comando mkfs
func MkfsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params MkfsParams

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		// Validaciones
		if params.ID == "" {
			http.Error(w, "El ID es obligatorio", http.StatusBadRequest)
			return
		}

		if params.Type == "" {
			http.Error(w, "El tipo es obligatorio", http.StatusBadRequest)
			return
		}

		// Asignar el valor por defecto para el sistema de archivos (2fs si no se especifica)
		if params.FS == "" {
			params.FS = "2fs"
		}

		// Validar el sistema de archivos (solo permitimos 2fs y 3fs)
		if params.FS != "2fs" && params.FS != "3fs" {
			http.Error(w, "El sistema de archivos debe ser '2fs' (EXT2) o '3fs' (EXT3)", http.StatusBadRequest)
			return
		}

		// Llamar a la función que ejecuta el mkfs
		FileSystem.Mkfs(params.ID, params.Type, params.FS)

		// Responder con éxito
		response := map[string]string{
			"message": "Sistema de archivos creado exitosamente",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Estructura para los parámetros de login
type LoginParams struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	ID   string `json:"id"`
}

// Handler para el comando login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.Method == http.MethodPost {
		var params LoginParams

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error al procesar la solicitud"})
			return
		}

		// Validaciones
		if params.User == "" || params.Pass == "" || params.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Los campos usuario, contraseña y id son obligatorios"})
			return
		}

		// Llamar a la función que ejecuta el login
		loginMessage, loginError := User.Login(params.User, params.Pass, params.ID)

		// Verificar si el login fue exitoso o no y responder en consecuencia
		if loginError == nil {
			response := map[string]string{
				"message": loginMessage,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			// Enviar mensaje de error como JSON con código 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": loginError.Error()})
		}
	} else {
		// Método no permitido
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
	}
}

// Estructura para los parámetros de rep
type RepParams struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	ID         string `json:"id"`
	PathFileLs string `json:"path_file_ls,omitempty"`
}

// Handler para el comando rep
func RepHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var params RepParams

		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		// Validar los parámetros obligatorios
		if params.Name == "" || params.Path == "" || params.ID == "" {
			http.Error(w, "'name', 'path' y 'id' son obligatorios", http.StatusBadRequest)
			return
		}

		// Verificar si la partición está montada
		mounted := false
		var diskPath string
		for _, partitions := range DiskManagement.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == params.ID {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
		}

		if !mounted {
			http.Error(w, "La partición con ID "+params.ID+" no está montada", http.StatusBadRequest)
			return
		}

		// Crear la carpeta si no existe
		reportsDir := filepath.Dir(params.Path)
		err = os.MkdirAll(reportsDir, os.ModePerm)
		if err != nil {
			http.Error(w, "Error al crear la carpeta: "+reportsDir, http.StatusInternalServerError)
			return
		}

		// Generar el reporte según el tipo solicitado
		switch params.Name {
		case "mbr":
			Analyzer.GenerateMBRReport(diskPath, params.Path)

		case "disk":
			Analyzer.GenerateDiskReport(diskPath, params.Path)

		case "file", "ls":
			// Validar que se proporcione el path_file_ls para reportes file y ls
			if params.PathFileLs == "" {
				http.Error(w, "'path_file_ls' es obligatorio para los reportes 'file' y 'ls'", http.StatusBadRequest)
				return
			}
			// Lógica adicional para generar reportes de archivo o lista
			fmt.Println("Generando reporte", params.Name, "con archivo/carpeta:", params.PathFileLs)

		default:
			http.Error(w, "Tipo de reporte no válido", http.StatusBadRequest)
			return
		}

		// Responder con éxito
		response := map[string]string{
			"message": "Reporte generado exitosamente",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
