package main

// Importación de bibliotecas necesarias
import (
	"bufio"        // Para leer y escribir datos en buffers, útil para leer entradas de usuario desde la consola.
	"encoding/gob" // Para codificar (serializar) y decodificar (deserializar) datos en formato binario.
	"fmt"          // Para formatear y imprimir datos en la consola.
	"os"           // Para trabajar con el sistema operativo, como manipulación de archivos y directorios.
	"strconv"      // Convertir datos entre tipos básicos, especialmente de cadenas a números y viceversa.
)

const archivoBinario = "registros.mia" // Nombre del archivo binario donde se almacenarán los datos

// Definición de la estructura Profesor
type Profesor struct {
	ID_profesor int
	Nombre      string
	Apellido    string
}

// Definición de la estructura Estudiante
type Estudiante struct {
	ID_estudiante int
	Carnet        int
	Nombre        string
	Apellido      string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin) // Crea un nuevo escáner para leer la entrada del usuario desde la consola

	// Bucle principal para mostrar el menú y procesar las opciones del usuario
	for {
		fmt.Println("Bienvenido al Menú Principal")
		fmt.Println("1. Registro de profesor")
		fmt.Println("2. Registro de estudiante")
		fmt.Println("3. Ver Registros")
		fmt.Println("4. Salir")
		fmt.Print("Elija una opción: ")
		scanner.Scan()           // Lee una línea de entrada del usuario
		opcion := scanner.Text() // Obtiene el texto ingresado por el usuario

		// Procesa la opción seleccionada por el usuario
		switch opcion {
		case "1":
			profesor := registrarProfesor(scanner) // Registra un nuevo profesor
			escribirRegistro(profesor)             // Escribe el registro en el archivo binario
		case "2":
			estudiante := registrarEstudiante(scanner) // Registra un nuevo estudiante
			escribirRegistro(estudiante)               // Escribe el registro en el archivo binario
		case "3":
			verRegistros() // Muestra los registros almacenados en el archivo binario
		case "4":
			fmt.Println("Saliendo...")
			return // Sale del programa
		default:
			fmt.Println("Opción no válida. Intente nuevamente.")
		}
	}
}

// Función para registrar un nuevo profesor
func registrarProfesor(scanner *bufio.Scanner) Profesor {
	var profesor Profesor

	// Solicita el ID del profesor y verifica que sea un número entero
	for {
		fmt.Print("Ingrese ID del Profesor: ")
		scanner.Scan()
		id, err := strconv.Atoi(scanner.Text()) // Convierte la entrada a entero
		if err != nil {
			fmt.Println("ID inválido. Por favor, ingrese un número entero.")
			continue // Si hay un error, vuelve a solicitar el ID
		}
		profesor.ID_profesor = id
		break // Sale del bucle si la entrada es válida
	}

	// Solicita el nombre del profesor
	fmt.Print("Ingrese Nombre: ")
	scanner.Scan()
	profesor.Nombre = scanner.Text()

	// Solicita el apellido del profesor
	fmt.Print("Ingrese Apellido: ")
	scanner.Scan()
	profesor.Apellido = scanner.Text()

	return profesor // Devuelve el objeto Profesor con los datos ingresados
}

// Función para registrar un nuevo estudiante
func registrarEstudiante(scanner *bufio.Scanner) Estudiante {
	var estudiante Estudiante

	// Solicita el ID del estudiante y verifica que sea un número entero
	for {
		fmt.Print("Ingrese ID del Estudiante: ")
		scanner.Scan()
		id, err := strconv.Atoi(scanner.Text()) // Convierte la entrada a entero
		if err != nil {
			fmt.Println("ID inválido. Por favor, ingrese un número entero.")
			continue
		}
		estudiante.ID_estudiante = id
		break
	}

	// Solicita el carnet del estudiante y verifica que sea un número entero
	for {
		fmt.Print("Ingrese Carnet: ")
		scanner.Scan()
		carnet, err := strconv.Atoi(scanner.Text()) // Convierte la entrada a entero
		if err != nil {
			fmt.Println("Carnet inválido. Por favor, ingrese un número entero.")
			continue
		}
		estudiante.Carnet = carnet
		break
	}

	// Solicita el nombre del estudiante
	fmt.Print("Ingrese Nombre: ")
	scanner.Scan()
	estudiante.Nombre = scanner.Text()

	// Solicita el apellido del estudiante
	fmt.Print("Ingrese Apellido: ")
	scanner.Scan()
	estudiante.Apellido = scanner.Text()

	return estudiante // Devuelve el objeto Estudiante con los datos ingresados
}

// Función para escribir un registro en el archivo binario
func escribirRegistro(data interface{}) {
	// Verifica si el archivo existe
	fileInfo, err := os.Stat(archivoBinario)
	if os.IsNotExist(err) {
		// Si el archivo no existe, créalo
		file, err := os.Create(archivoBinario)
		if err != nil {
			fmt.Println("Error al crear el archivo:", err)
			return
		}
		file.Close()
	} else {
		fmt.Printf("El archivo %s existe, tamaño: %d bytes\n", archivoBinario, fileInfo.Size())
	}

	// Abre el archivo en modo de escritura
	file, err := os.OpenFile(archivoBinario, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close() // Asegura que el archivo se cierre al final de la función

	// Codifica (serializa) los datos y los escribe en el archivo
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("Error al escribir en el archivo:", err)
	}
}

// Func para ver los registros almacenados en el archivo binario
func verRegistros() {
	// Verifica si el archivo existe antes de intentar abrirlo
	_, err := os.Stat(archivoBinario)
	if os.IsNotExist(err) {
		fmt.Println("No hay registros disponibles.")
		return
	}

	// Abre el archivo en modo de lectura
	file, err := os.Open(archivoBinario)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close() // Asegura que el archivo se cierre al final de la función

	decoder := gob.NewDecoder(file) // Crea un nuevo decodificador gob
	fmt.Println("\nProfesores y Estudiantes Registrados:")

	// Bucle para decodificar y mostrar los registros del archivo
	for {
		var profesor Profesor
		if err := decoder.Decode(&profesor); err == nil {
			fmt.Printf("Profesor - ID: %d, Nombre: %s, Apellido: %s\n", profesor.ID_profesor, profesor.Nombre, profesor.Apellido)
			continue
		}

		var estudiante Estudiante
		if err := decoder.Decode(&estudiante); err == nil {
			fmt.Printf("Estudiante - ID: %d, Carnet: %d, Nombre: %s, Apellido: %s\n", estudiante.ID_estudiante, estudiante.Carnet, estudiante.Nombre, estudiante.Apellido)
			continue
		}

		break // Sale del bucle si no hay más registros que decodificar
	}
}
