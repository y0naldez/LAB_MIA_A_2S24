package Analyzer

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"proyecto1/DiskManagement"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

//input := "mkdisk -size=3000 -unit=K -fit=BF -path=/home/bang/Disks/disk1.bin"

/*
parts[0] es "mkdisk"
*/

func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input

	/*Después de procesar la entrada:
	command será "mkdisk".
	params será "-size=3000 -unit=K -fit=BF -path=/home/bang/Disks/disk1.bin".*/
}

func Analyze() {

	for true {
		var input string
		fmt.Println("======================")
		fmt.Println("Ingrese comando: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		command, params := getCommandAndParams(input)

		fmt.Println("Comando: ", command, " - ", "Parametro: ", params)

		AnalyzeCommnad(command, params)

		//mkdisk -size=3000 -unit=K -fit=BF -path="/home/bang/Disks/disk1.bin"
	}
}

func AnalyzeCommnad(command string, params string) {

	if strings.Contains(command, "mkdisk") {
		fn_mkdisk(params)
	} else if strings.Contains(command, "fdisk") {
		fn_fdisk(params)
	} else {
		fmt.Println("Error: Commando invalido o no encontrado")
	}

}

func fn_mkdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse flag
	fs.Parse(os.Args[1:])

	// Encontrar la flag en el input
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	/*
			Primera Iteración :
		    flagName es "size".
		    flagValue es "3000".
		    El switch encuentra que "size" es un flag reconocido, por lo que se ejecuta fs.Set("size", "3000").
		    Esto asigna el valor 3000 al flag size.

	*/

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// LLamamos a la funcion
	DiskManagement.Mkdisk(*size, *fit, *unit, *path)
}

func fn_fdisk(input string) {
	// Definir flags
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "m", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "", "Ajuste") // Dejar fit vacío por defecto

	// Parsear los flags
	fs.Parse(os.Args[1:])

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// Si no se proporcionó un fit, usar el valor predeterminado "w"
	if *fit == "" {
		*fit = "w"
	}

	// Validar fit (b/w/f)
	if *fit != "b" && *fit != "f" && *fit != "w" {
		fmt.Println("Error: Fit must be 'b', 'f', or 'w'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return
	}

	// Llamar a la función
	DiskManagement.Fdisk(*size, *path, *name, *unit, *type_, *fit)
}
