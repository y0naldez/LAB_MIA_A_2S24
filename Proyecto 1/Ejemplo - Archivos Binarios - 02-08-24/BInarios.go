package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
)

const archivoBinario = "registros.mia"

// Estructura para almacenar la información de un profesor
type Profesor struct {
	ID_profesor int64
	Nombre      string
	Apellido    string
	Tipo        int64 // 1 para profesor
}

// Estructura para almacenar la información de un estudiante
type Estudiante struct {
	ID_estudiante int64
	Carnet        int64
	Nombre        string
	Apellido      string
	Tipo          int64 // 2 para estudiante
}

// Función para crear un nuevo archivo binario si no existe
func createFile() error {
	file, err := os.OpenFile(archivoBinario, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// Función para escribir la información de un profesor en el archivo binario
func escribirProfesor(profesor Profesor) error {
	file, err := os.OpenFile(archivoBinario, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// Escribir el tipo de registro (1 para profesor)
	err = binary.Write(writer, binary.LittleEndian, profesor.Tipo)
	if err != nil {
		return err
	}
	// Escribir el ID del profesor
	err = binary.Write(writer, binary.LittleEndian, profesor.ID_profesor)
	if err != nil {
		return err
	}

	// Escribir la longitud del nombre y luego el nombre
	err = binary.Write(writer, binary.LittleEndian, int64(len(profesor.Nombre)))
	if err != nil {
		return err
	}
	_, err = writer.WriteString(profesor.Nombre)
	if err != nil {
		return err
	}

	// Escribir la longitud del apellido y luego el apellido
	err = binary.Write(writer, binary.LittleEndian, int64(len(profesor.Apellido)))
	if err != nil {
		return err
	}
	_, err = writer.WriteString(profesor.Apellido)
	if err != nil {
		return err
	}

	return writer.Flush()
}

// Función para escribir la información de un estudiante en el archivo binario
func escribirEstudiante(estudiante Estudiante) error {
	file, err := os.OpenFile(archivoBinario, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// Escribir el tipo de registro (2 para estudiante)
	err = binary.Write(writer, binary.LittleEndian, estudiante.Tipo)
	if err != nil {
		return err
	}
	// Escribir el ID del estudiante
	err = binary.Write(writer, binary.LittleEndian, estudiante.ID_estudiante)
	if err != nil {
		return err
	}
	// Escribir el carnet del estudiante
	err = binary.Write(writer, binary.LittleEndian, estudiante.Carnet)
	if err != nil {
		return err
	}

	// Escribir la longitud del nombre y luego el nombre
	err = binary.Write(writer, binary.LittleEndian, int64(len(estudiante.Nombre)))
	if err != nil {
		return err
	}
	_, err = writer.WriteString(estudiante.Nombre)
	if err != nil {
		return err
	}

	// Escribir la longitud del apellido y luego el apellido
	err = binary.Write(writer, binary.LittleEndian, int64(len(estudiante.Apellido)))
	if err != nil {
		return err
	}
	_, err = writer.WriteString(estudiante.Apellido)
	if err != nil {
		return err
	}

	return writer.Flush()
}

// Función para registrar un nuevo profesor a través de la entrada del usuario
func registrarProfesor(scanner *bufio.Scanner) Profesor {
	var profesor Profesor

	// Solicitar y leer el ID del profesor
	for {
		fmt.Print("Ingrese ID del Profesor: ")
		scanner.Scan()
		id, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			fmt.Println("ID inválido. Por favor, ingrese un número entero.")
			continue
		}
		profesor.ID_profesor = id
		break
	}

	// Solicitar y leer el nombre del profesor
	fmt.Print("Ingrese Nombre: ")
	scanner.Scan()
	profesor.Nombre = scanner.Text()

	// Solicitar y leer el apellido del profesor
	fmt.Print("Ingrese Apellido: ")
	scanner.Scan()
	profesor.Apellido = scanner.Text()

	// Establecer el tipo a 1 (profesor)
	profesor.Tipo = 1

	return profesor
}

// Función para registrar un nuevo estudiante a través de la entrada del usuario
func registrarEstudiante(scanner *bufio.Scanner) Estudiante {
	var estudiante Estudiante

	// Solicitar y leer el ID del estudiante
	for {
		fmt.Print("Ingrese ID del Estudiante: ")
		scanner.Scan()
		id, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			fmt.Println("ID inválido. Por favor, ingrese un número entero.")
			continue
		}
		estudiante.ID_estudiante = id
		break
	}

	// Solicitar y leer el carnet del estudiante
	for {
		fmt.Print("Ingrese Carnet: ")
		scanner.Scan()
		carnet, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			fmt.Println("Carnet inválido. Por favor, ingrese un número entero.")
			continue
		}
		estudiante.Carnet = carnet
		break
	}

	// Solicitar y leer el nombre del estudiante
	fmt.Print("Ingrese Nombre: ")
	scanner.Scan()
	estudiante.Nombre = scanner.Text()

	// Solicitar y leer el apellido del estudiante
	fmt.Print("Ingrese Apellido: ")
	scanner.Scan()
	estudiante.Apellido = scanner.Text()

	// Establecer el tipo a 2 (estudiante)
	estudiante.Tipo = 2

	return estudiante
}

// Función para leer y mostrar todos los registros del archivo binario
func verRegistros() {
	file, err := os.Open(archivoBinario)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	for {
		var tipo int64
		// Leer el tipo de registro (1 para profesor, 2 para estudiante)
		if err := binary.Read(file, binary.LittleEndian, &tipo); err != nil {
			if err != io.EOF {
				fmt.Println("Error al leer el tipo:", err)
			}
			break
		}

		if tipo == 1 {
			// Leer y mostrar la información de un profesor
			var profesor Profesor
			binary.Read(file, binary.LittleEndian, &profesor.ID_profesor)
			var nombreLen int64
			binary.Read(file, binary.LittleEndian, &nombreLen)
			nombre := make([]byte, nombreLen)
			file.Read(nombre)
			profesor.Nombre = string(nombre)
			var apellidoLen int64
			binary.Read(file, binary.LittleEndian, &apellidoLen)
			apellido := make([]byte, apellidoLen)
			file.Read(apellido)
			profesor.Apellido = string(apellido)
			fmt.Printf("Profesor - ID: %d, Nombre: %s, Apellido: %s\n", profesor.ID_profesor, profesor.Nombre, profesor.Apellido)
		} else if tipo == 2 {
			// Leer y mostrar la información de un estudiante
			var estudiante Estudiante
			binary.Read(file, binary.LittleEndian, &estudiante.ID_estudiante)
			binary.Read(file, binary.LittleEndian, &estudiante.Carnet)
			var nombreLen int64
			binary.Read(file, binary.LittleEndian, &nombreLen)
			nombre := make([]byte, nombreLen)
			file.Read(nombre)
			estudiante.Nombre = string(nombre)
			var apellidoLen int64
			binary.Read(file, binary.LittleEndian, &apellidoLen)
			apellido := make([]byte, apellidoLen)
			file.Read(apellido)
			estudiante.Apellido = string(apellido)
			fmt.Printf("Estudiante - ID: %d, Carnet: %d, Nombre: %s, Apellido: %s\n", estudiante.ID_estudiante, estudiante.Carnet, estudiante.Nombre, estudiante.Apellido)
		}
	}
}

func main() {
	// Crear el archivo binario si no existe
	err := createFile()
	if err != nil {
		fmt.Println("El archivo ya existe.")
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Mostrar el menú principal
		fmt.Println("Bienvenido al Menú Principal")
		fmt.Println("1. Registro de profesor")
		fmt.Println("2. Registro de estudiante")
		fmt.Println("3. Ver Registros")
		fmt.Println("4. Salir")
		fmt.Print("Elija una opción: ")
		scanner.Scan()
		opcion := scanner.Text()

		switch opcion {
		case "1":
			// Registrar un nuevo profesor
			profesor := registrarProfesor(scanner)
			err := escribirProfesor(profesor)
			if err != nil {
				fmt.Println("Error al escribir el profesor:", err)
			}
		case "2":
			// Registrar un nuevo estudiante
			estudiante := registrarEstudiante(scanner)
			err := escribirEstudiante(estudiante)
			if err != nil {
				fmt.Println("Error al escribir el estudiante:", err)
			}
		case "3":
			// Ver todos los registros
			verRegistros()
		case "4":
			// Salir del programa
			fmt.Println("Saliendo...")
			return
		default:
			fmt.Println("Opción no válida. Intente nuevamente.")
		}
	}
}

