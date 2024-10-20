package Analyzer

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"proyecto1/DiskManagement"
	"proyecto1/FileSystem"
	"proyecto1/Structs"
	"proyecto1/User"
	"proyecto1/Utilities"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

// Helper para obtener el comando y sus parámetros de un input
func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input
}

// Función para analizar los comandos
func Analyze() {

	for {
		var input string
		fmt.Println("======================")
		fmt.Println("Ingrese comando: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		command, params := getCommandAndParams(input)

		fmt.Println("Comando: ", command, " - Parámetro: ", params)

		AnalyzeCommand(command, params)
	}
}

// Función que dirige los comandos hacia la función correcta según el comando
func AnalyzeCommand(command string, params string) {

	if strings.Contains(command, "mkdisk") {
		fn_mkdisk(params)
	} else if strings.Contains(command, "fdisk") {
		fn_fdisk(params)
	} else if strings.Contains(command, "mount") {
		fn_mount(params)
	} else if strings.Contains(command, "unmount") {
		fn_unmount(params)
	} else if strings.Contains(command, "mkfs") {
		fn_mkfs(params)
	} else if strings.Contains(command, "login") {
		fn_login(params)
	} else if strings.Contains(command, "rep") {
		fn_rep(params)
	} else if strings.Contains(command, "mkusr") {
		fn_mkusr(params)
	} else if strings.Contains(command, "readmbr") {
		fn_readmbr(params)
	} else {
		fmt.Println("Error: Comando inválido o no encontrado")
	}
}

// Función para crear un disco (mkdisk)
func fn_mkdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Encontrar las flags en el input
	matches := re.FindAllStringSubmatch(params, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrada")
		}
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: El tamaño debe ser mayor a 0")
		return
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: El ajuste debe ser 'bf', 'ff' o 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: La unidad debe ser 'k' o 'm'")
		return
	}

	if *path == "" {
		fmt.Println("Error: La ruta es requerida")
		return
	}

	// Llamar a la función que ejecuta el mkdisk
	DiskManagement.Mkdisk(*size, *fit, *unit, *path)
}

// Funcion FDISK
func fn_fdisk(input string) {
	// Definir flags
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "m", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "", "Ajuste")
	delete_ := fs.String("delete", "", "Eliminar partición (Fast/Full)")

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones para la opción -delete
	if *delete_ != "" {
		if *path == "" || *name == "" {
			fmt.Println("Error: Para eliminar una partición, se requiere 'path' y 'name'.")
			return
		}
		// Llamar a la función que elimina la partición
		DiskManagement.DeletePartition(*path, *name, *delete_)
		return
	}

	// Validaciones para la creación de particiones
	if *size <= 0 {
		fmt.Println("Error: El tamaño debe ser mayor a 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: La ruta es requerida")
		return
	}

	// Usar ajuste por defecto si no se proporciona
	if *fit == "" {
		*fit = "w"
	}

	// Validar fit
	if *fit != "b" && *fit != "f" && *fit != "w" {
		fmt.Println("Error: El ajuste debe ser 'b', 'f', o 'w'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: La unidad debe ser 'k' o 'm'")
		return
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: El tipo debe ser 'p', 'e', o 'l'")
		return
	}

	// Llamar a la función que ejecuta el fdisk
	DiskManagement.Fdisk(*size, *path, *name, *unit, *type_, *fit)
}

// Función para leer el MBR y listar particiones
func fn_readmbr(params string) {
	fs := flag.NewFlagSet("readmbr", flag.ExitOnError)
	path := fs.String("path", "", "Ruta del disco")

	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrada")
		}
	}

	if *path == "" {
		fmt.Println("Error: La ruta es requerida")
		return
	}

	// Llamar a la función para leer el MBR y mostrar las particiones
	DiskManagement.ReadMBR(*path)
}

// Función para montar particiones (fn_mount)
func fn_mount(params string) {
	// Definir flags
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	// Parsear los parámetros del input
	matches := re.FindAllStringSubmatch(params, -1)

	// Procesar los parámetros
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2]) // Convertir a minúsculas
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Name son obligatorios")
		return
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	lowercaseName := strings.ToLower(*name)
	DiskManagement.Mount(*path, lowercaseName)
}

// Función para desmontar particiones (fn_unmount)
func fn_unmount(params string) {
	// Definir flags
	fs := flag.NewFlagSet("unmount", flag.ExitOnError)
	id := fs.String("id", "", "ID de la partición a desmontar")

	// Parsear los parámetros del input
	matches := re.FindAllStringSubmatch(params, -1)

	// Procesar los parámetros
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2]) // Convertir a minúsculas
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *id == "" {
		fmt.Println("Error: ID es obligatorio")
		return
	}

	// Llamar a la función Unmount con el ID
	DiskManagement.Unmount(*id)
}

// Función para crear el sistema de archivos (fn_mkfs)
func fn_mkfs(input string) {
	// Definir flags
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "Id")
	type_ := fs.String("type", "", "Tipo")
	fs_ := fs.String("fs", "2fs", "Sistema de archivos")

	// Parsear los parámetros de entrada
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar los parámetros
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *id == "" {
		fmt.Println("Error: id es un parámetro obligatorio.")
		return
	}

	if *type_ == "" {
		fmt.Println("Error: type es un parámetro obligatorio.")
		return
	}

	// Validar el sistema de archivos: solo permitimos 2fs y 3fs
	if *fs_ != "2fs" && *fs_ != "3fs" {
		fmt.Println("Error: El sistema de archivos debe ser '2fs' (EXT2) o '3fs' (EXT3).")
		return
	}

	// Llamar a la función que crea el sistema de archivos
	FileSystem.Mkfs(*id, *type_, *fs_)
}

// Función para iniciar sesión (fn_login)
func fn_login(input string) {
	// Definir flags
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id de la partición")

	// Parsear los parámetros de entrada
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar los parámetros
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *user == "" || *pass == "" || *id == "" {
		fmt.Println("Error: Los campos user, pass e id son obligatorios.")
		return
	}

	// Llamar a la función de login
	User.Login(*user, *pass, *id)
}

// Función para crear usuario (fn_mkusr)
func fn_mkusr(input string) {
	// Definir flags
	fs := flag.NewFlagSet("mkusr", flag.ExitOnError)
	user := fs.String("user", "", "Nombre del usuario a crear")
	pass := fs.String("pass", "", "Contraseña del usuario")
	grp := fs.String("grp", "", "Grupo del usuario")

	// Parsear el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input para asignar valores a las flags
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"") // Limpiar el input de comillas

		switch flagName {
		case "user", "pass", "grp":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrada")
		}
	}

	// Validar que las flags requeridas no estén vacías
	if *user == "" || *pass == "" || *grp == "" {
		fmt.Println("Error: Todos los campos (user, pass, grp) son obligatorios.")
		return
	}

	// Verificar que los campos no excedan los 10 caracteres
	if len(*user) > 10 {
		fmt.Println("Error: El nombre de usuario excede el máximo de 10 caracteres.")
		return
	}
	if len(*pass) > 10 {
		fmt.Println("Error: La contraseña excede el máximo de 10 caracteres.")
		return
	}
	if len(*grp) > 10 {
		fmt.Println("Error: El grupo excede el máximo de 10 caracteres.")
		return
	}

	// Crear el nuevo usuario en el formato correcto
	newUser := fmt.Sprintf("2,U,%s,%s,%s", *grp, *user, *pass)

	// Llamar a la función que maneja la creación del usuario
	err := User.MkusrCommand("/users.txt", newUser)
	if err != nil {
		fmt.Println("Error al crear el usuario:", err)
		return
	}

	fmt.Println("Usuario creado con éxito:", *user)
}

// Función para generar reportes (fn_rep)
func fn_rep(input string) {
	// Definir flags
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	name := fs.String("name", "", "Nombre del reporte a generar (mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls)")
	path := fs.String("path", "", "Ruta donde se generará el reporte")
	id := fs.String("id", "", "ID de la partición")
	pathFileLs := fs.String("path_file_ls", "", "Nombre del archivo o carpeta para reportes file o ls") // Parámetro opcional

	// Parsear los parámetros de entrada
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")
		fs.Set(flagName, flagValue)
	}

	// Validar que los parámetros obligatorios estén presentes
	if *name == "" || *path == "" || *id == "" {
		fmt.Println("Error: 'name', 'path' y 'id' son parámetros obligatorios.")
		return
	}

	// Verificar si el disco está montado
	mounted := false
	var diskPath string
	for _, partitions := range DiskManagement.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == *id {
				mounted = true
				diskPath = partition.Path
				break
			}
		}
	}

	if !mounted {
		fmt.Println("Error: La partición con ID", *id, "no está montada.")
		return
	}

	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(*path)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error al crear la carpeta:", reportsDir)
		return
	}

	// Generar el reporte según el tipo solicitado
	switch *name {
	case "mbr":
		// Leer el MBR y generar el reporte
		GenerateMBRReport(diskPath, *path)

	case "disk":
		// Leer la información del disco y generar el reporte
		GenerateDiskReport(diskPath, *path)

	case "file", "ls":
		// Validar que se proporcione el path_file_ls para reportes file y ls
		if *pathFileLs == "" {
			fmt.Println("Error: 'path_file_ls' es obligatorio para los reportes 'file' y 'ls'.")
			return
		}
		// Lógica adicional para reportes file y ls
		fmt.Println("Generando reporte", *name, "con archivo/carpeta:", *pathFileLs)

	default:
		fmt.Println("Error: Tipo de reporte no válido.")
	}
}

// Helper para generar el reporte MBR
func GenerateMBRReport(diskPath, reportPath string) {
	// Abrir el archivo binario del disco montado
	file, err := Utilities.OpenFile(diskPath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
		return
	}
	defer file.Close()

	// Leer el MBR desde el archivo
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	// Leer y procesar los EBRs si hay particiones extendidas
	var ebrs []Structs.EBR
	for i := 0; i < 4; i++ {
		if string(TempMBR.Partitions[i].Type[:]) == "e" { // Partición extendida
			ebrPosition := TempMBR.Partitions[i].Start
			for ebrPosition != -1 {
				var tempEBR Structs.EBR
				if err := Utilities.ReadObject(file, &tempEBR, int64(ebrPosition)); err != nil {
					break
				}
				ebrs = append(ebrs, tempEBR)
				ebrPosition = tempEBR.PartNext
			}
		}
	}

	// Generar el reporte MBR
	if err := Utilities.GenerateMBRReport(TempMBR, ebrs, reportPath, file); err != nil {
		fmt.Println("Error al generar el reporte MBR:", err)
	} else {
		fmt.Println("Reporte MBR generado exitosamente en:", reportPath)

		// Renderizar el archivo .dot a .jpg
		renderDotToImage(reportPath)
	}
}

// Helper para generar el reporte DISK
func GenerateDiskReport(diskPath, reportPath string) {
	// Abrir el archivo binario del disco montado
	file, err := Utilities.OpenFile(diskPath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
		return
	}
	defer file.Close()

	// Leer el MBR desde el archivo
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	// Leer y procesar los EBRs si hay particiones extendidas
	var ebrs []Structs.EBR
	for i := 0; i < 4; i++ {
		if string(TempMBR.Partitions[i].Type[:]) == "e" { // Partición extendida
			ebrPosition := TempMBR.Partitions[i].Start
			for ebrPosition != -1 {
				var tempEBR Structs.EBR
				if err := Utilities.ReadObject(file, &tempEBR, int64(ebrPosition)); err != nil {
					break
				}
				ebrs = append(ebrs, tempEBR)
				ebrPosition = tempEBR.PartNext
			}
		}
	}

	// Generar el reporte DISK
	if err := Utilities.GenerateDiskReport(TempMBR, ebrs, reportPath, file, TempMBR.MbrSize); err != nil {
		fmt.Println("Error al generar el reporte DISK:", err)
	} else {
		fmt.Println("Reporte DISK generado exitosamente en:", reportPath)

		// Renderizar el archivo .dot a .jpg
		renderDotToImage(reportPath)
	}
}

// Helper para convertir un archivo .dot a una imagen .jpg
func renderDotToImage(reportPath string) {
	dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
	outputJpg := reportPath
	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error al renderizar el archivo .dot a imagen:", err)
	} else {
		fmt.Println("Imagen generada exitosamente en:", outputJpg)
	}
}
