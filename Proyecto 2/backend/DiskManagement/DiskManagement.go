package DiskManagement

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
	"time"
)

// Función para leer el MBR desde un archivo binario
func ReadMBR(path string) {
	// Abrir el archivo binario
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	// Crear una variable para almacenar el MBR
	var mbr Structs.MRB

	// Leer el MBR desde el archivo
	err = Utilities.ReadObject(file, &mbr, 0) // Leer desde la posición 0
	if err != nil {
		fmt.Println("Error al leer el MBR:", err)
		return
	}

	// Imprimir el MBR
	Structs.PrintMBR(mbr)
}

// Estructura para representar una partición en JSON
type PartitionInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
	Status string `json:"status"`
}

// Función para leer el MBR desde un archivo binario y devolver las particiones
func ListPartitions(path string) ([]PartitionInfo, error) {
	// Abrir el archivo binario
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Crear una variable para almacenar el MBR
	var mbr Structs.MRB

	// Leer el MBR desde el archivo
	err = Utilities.ReadObject(file, &mbr, 0) // Leer desde la posición 0
	if err != nil {
		return nil, fmt.Errorf("Error al leer el MBR: %v", err)
	}

	// Crear una lista de particiones basada en el MBR
	var partitions []PartitionInfo
	for _, partition := range mbr.Partitions {
		if partition.Size > 0 { // Solo agregar si la partición tiene un tamaño
			// Limpiar el nombre para eliminar caracteres nulos
			partitionName := strings.TrimRight(string(partition.Name[:]), "\x00")

			partitions = append(partitions, PartitionInfo{
				Name:   partitionName,
				Type:   strings.TrimRight(string(partition.Type[:]), "\x00"),
				Start:  partition.Start,
				Size:   partition.Size,
				Status: strings.TrimRight(string(partition.Status[:]), "\x00"),
			})
		}
	}

	return partitions, nil
}

// Estructura para representar una partición montada
type MountedPartition struct {
	Path     string
	Name     string
	ID       string
	Status   byte // 0: no montada, 1: montada
	LoggedIn bool // true: usuario ha iniciado sesión, false: no ha iniciado sesión
}

// Mapa para almacenar las particiones montadas, organizadas por disco
var mountedPartitions = make(map[string][]MountedPartition)

// Función para imprimir las particiones montadas
func PrintMountedPartitions() {
	fmt.Println("Particiones montadas:")

	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}

	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			loginStatus := "No"
			if partition.LoggedIn {
				loginStatus = "Sí"
			}
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s, Status: %c, LoggedIn: %s\n",
				partition.Name, partition.ID, partition.Path, partition.Status, loginStatus)
		}
	}
	fmt.Println("")
}

// Función para obtener las particiones montadas
func GetMountedPartitions() map[string][]MountedPartition {
	return mountedPartitions
}

// Función para marcar una partición como logueada
func MarkPartitionAsLoggedIn(id string) {
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = true
				fmt.Printf("Partición con ID %s marcada como logueada.\n", id)
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

// Función para limpiar las particiones montadas
func CleanMountedPartitions() {
	mountedPartitions = make(map[string][]MountedPartition)
}

// Función Mkdisk optimizada para escribir bloques de ceros
func Mkdisk(size int, fit string, unit string, path string) {
	fmt.Println("======INICIO MKDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Fit:", fit)
	fmt.Println("Unit:", unit)
	fmt.Println("Path:", path)

	// Validar fit bf/ff/wf
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Println("Error: Fit debe ser bf, wf o ff")
		return
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Println("Error: Size debe ser mayor a 0")
		return
	}

	// Validar unit k - m
	if unit != "k" && unit != "m" {
		fmt.Println("Error: Las unidades válidas son k o m")
		return
	}

	// Crear el archivo
	err := Utilities.CreateFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Asignar tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return
	}

	// Optimización: Escribir grandes bloques de ceros
	blockSize := 1024 * 1024             // Bloques de 1MB
	zeroBlock := make([]byte, blockSize) // Crear un bloque de ceros

	remainingSize := size

	for remainingSize > 0 {
		if remainingSize < blockSize {
			// Escribe lo que queda si es menor que el tamaño del bloque
			zeroBlock = make([]byte, remainingSize)
		}
		_, err := file.Write(zeroBlock)
		if err != nil {
			fmt.Println("Error escribiendo ceros:", err)
			return
		}
		remainingSize -= blockSize
	}

	// Crear el MBR
	var newMRB Structs.MRB
	newMRB.MbrSize = int32(size)
	newMRB.Signature = rand.Int31() // Número aleatorio rand.Int31() genera solo números no negativos
	copy(newMRB.Fit[:], fit)

	// Obtener la fecha actual en formato YYYY-MM-DD
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")
	copy(newMRB.CreationDate[:], formattedDate)

	// Escribir el MBR en el archivo
	if err := Utilities.WriteObject(file, newMRB, 0); err != nil {
		return
	}

	// Leer el archivo y verificar el MBR
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	// Imprimir el MBR
	Structs.PrintMBR(TempMBR)

	// Cerrar el archivo
	defer file.Close()

	fmt.Println("======FIN MKDISK======")
}

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)

	// Validar fit (b/w/f)
	if fit != "b" && fit != "f" && fit != "w" {
		fmt.Println("Error: Fit must be 'b', 'f', or 'w'")
		return
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	// Validar unit (b/k/m)
	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Println("Error: Unit must be 'b', 'k', or 'm'")
		return
	}

	// Ajustar el tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return
	}

	// Imprimir el objeto MBR
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	// Validaciones de las particiones
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			totalPartitions++
			usedSpace += TempMBR.Partitions[i].Size

			if TempMBR.Partitions[i].Type[0] == 'p' {
				primaryCount++
			} else if TempMBR.Partitions[i].Type[0] == 'e' {
				extendedCount++
			}
		}
	}

	// Validar que no se exceda el número máximo de particiones primarias y extendidas
	if totalPartitions >= 4 {
		fmt.Println("Error: No se pueden crear más de 4 particiones primarias o extendidas en total.")
		return
	}

	// Validar que solo haya una partición extendida
	if type_ == "e" && extendedCount > 0 {
		fmt.Println("Error: Solo se permite una partición extendida por disco.")
		return
	}

	// Validar que no se pueda crear una partición lógica sin una extendida
	if type_ == "l" && extendedCount == 0 {
		fmt.Println("Error: No se puede crear una partición lógica sin una partición extendida.")
		return
	}

	// Validar que el tamaño de la nueva partición no exceda el tamaño del disco
	if usedSpace+int32(size) > TempMBR.MbrSize {
		fmt.Println("Error: No hay suficiente espacio en el disco para crear esta partición.")
		return
	}

	// Determinar la posición de inicio de la nueva partición
	var gap int32 = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.Partitions[totalPartitions-1].Start + TempMBR.Partitions[totalPartitions-1].Size
	}

	// Encontrar una posición vacía para la nueva partición
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				// Crear partición primaria o extendida
				TempMBR.Partitions[i].Size = int32(size)
				TempMBR.Partitions[i].Start = gap
				copy(TempMBR.Partitions[i].Name[:], name)
				copy(TempMBR.Partitions[i].Fit[:], fit)
				copy(TempMBR.Partitions[i].Status[:], "0")
				copy(TempMBR.Partitions[i].Type[:], type_)
				TempMBR.Partitions[i].Correlative = int32(totalPartitions + 1)

				if type_ == "e" {
					// Inicializar el primer EBR en la partición extendida
					ebrStart := gap // El primer EBR se coloca al inicio de la partición extendida
					ebr := Structs.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					Utilities.WriteObject(file, ebr, int64(ebrStart))
				}

				break
			}
		}
	}

	// Manejar la creación de particiones lógicas dentro de una partición extendida
	if type_ == "l" {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				// Calcular la posición de inicio de la nueva partición lógica
				newEBRPos := ebr.PartStart + ebr.PartSize                    // El nuevo EBR se coloca después de la partición lógica anterior
				logicalPartitionStart := newEBRPos + int32(binary.Size(ebr)) // El inicio de la partición lógica es justo después del EBR

				// Ajustar el siguiente EBR
				ebr.PartNext = newEBRPos
				Utilities.WriteObject(file, ebr, int64(ebrPos))

				// Crear y escribir el nuevo EBR
				newEBR := Structs.EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				Utilities.WriteObject(file, newEBR, int64(newEBRPos))

				// Imprimir el nuevo EBR creado
				fmt.Println("Nuevo EBR creado:")
				Structs.PrintEBR(newEBR)
				fmt.Println("")

				// Imprimir todos los EBRs en la partición extendida
				fmt.Println("Imprimiendo todos los EBRs en la partición extendida:")
				ebrPos = TempMBR.Partitions[i].Start
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					Structs.PrintEBR(ebr)
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				break
			}
		}
		fmt.Println("")
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return
	}

	var TempMBR2 Structs.MRB
	// Leer el objeto nuevamente para verificar
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file after writing")
		return
	}

	// Imprimir el objeto MBR actualizado
	Structs.PrintMBR(TempMBR2)

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Println("======FIN FDISK======")
	fmt.Println("")

}

// Función para eliminar particiones
func DeletePartition(path string, name string, delete_ string) {
	fmt.Println("======Start DELETE PARTITION======")
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Delete type:", delete_)

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return
	}

	// Buscar la partición por nombre
	found := false
	for i := 0; i < 4; i++ {
		// Limpiar los caracteres nulos al final del nombre de la partición
		partitionName := strings.TrimRight(string(TempMBR.Partitions[i].Name[:]), "\x00")
		if partitionName == name {
			found = true

			// Si es una partición extendida, eliminar las particiones lógicas dentro de ella
			if TempMBR.Partitions[i].Type[0] == 'e' {
				fmt.Println("Eliminando particiones lógicas dentro de la partición extendida...")
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					// Detener el bucle si el EBR está vacío
					if ebr.PartStart == 0 && ebr.PartSize == 0 {
						fmt.Println("EBR vacío encontrado, deteniendo la búsqueda.")
						break
					}
					// Depuración: Mostrar el EBR leído
					fmt.Println("EBR leído antes de eliminar:")
					Structs.PrintEBR(ebr)

					// Eliminar partición lógica
					if delete_ == "fast" {
						ebr = Structs.EBR{}                             // Resetear el EBR manualmente
						Utilities.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
					} else if delete_ == "full" {
						Utilities.FillWithZeros(file, ebr.PartStart, ebr.PartSize)
						ebr = Structs.EBR{}                             // Resetear el EBR manualmente
						Utilities.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
					}

					// Depuración: Mostrar el EBR después de eliminar
					fmt.Println("EBR después de eliminar:")
					Structs.PrintEBR(ebr)

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}

			// Proceder a eliminar la partición (extendida, primaria o lógica)
			if delete_ == "fast" {
				// Eliminar rápido: Resetear manualmente los campos de la partición
				TempMBR.Partitions[i] = Structs.Partition{} // Resetear la partición manualmente
				fmt.Println("Partición eliminada en modo Fast.")
			} else if delete_ == "full" {
				// Eliminar completamente: Resetear manualmente y sobrescribir con '\0'
				start := TempMBR.Partitions[i].Start
				size := TempMBR.Partitions[i].Size
				TempMBR.Partitions[i] = Structs.Partition{} // Resetear la partición manualmente
				// Escribir '\0' en el espacio de la partición en el disco
				Utilities.FillWithZeros(file, start, size)
				fmt.Println("Partición eliminada en modo Full.")

				// Leer y verificar si el área está llena de ceros
				Utilities.VerifyZeros(file, start, size)
			}
			break
		}
	}

	if !found {
		// Buscar particiones lógicas si no se encontró en el MBR
		fmt.Println("Buscando en particiones lógicas dentro de las extendidas...")
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' { // Solo buscar dentro de particiones extendidas
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}

					// Depuración: Mostrar el EBR leído
					fmt.Println("EBR leído:")
					Structs.PrintEBR(ebr)

					logicalName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if logicalName == name {
						found = true
						// Eliminar la partición lógica
						if delete_ == "fast" {
							ebr = Structs.EBR{}                             // Resetear el EBR manualmente
							Utilities.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
							fmt.Println("Partición lógica eliminada en modo Fast.")
						} else if delete_ == "full" {
							Utilities.FillWithZeros(file, ebr.PartStart, ebr.PartSize)
							ebr = Structs.EBR{}                             // Resetear el EBR manualmente
							Utilities.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
							Utilities.VerifyZeros(file, ebr.PartStart, ebr.PartSize)
							fmt.Println("Partición lógica eliminada en modo Full.")
						}
						break
					}

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		fmt.Println("Error: No se encontró la partición con el nombre:", name)
		return
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return
	}

	// Leer el MBR actualizado y mostrarlo
	fmt.Println("MBR actualizado después de la eliminación:")
	Structs.PrintMBR(TempMBR)

	// Si es una partición extendida, mostrar los EBRs actualizados
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'e' {
			fmt.Println("Imprimiendo EBRs actualizados en la partición extendida:")
			ebrPos := TempMBR.Partitions[i].Start
			var ebr Structs.EBR
			for {
				err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					fmt.Println("Error al leer EBR:", err)
					break
				}
				// Detener el bucle si el EBR está vacío
				if ebr.PartStart == 0 && ebr.PartSize == 0 {
					fmt.Println("EBR vacío encontrado, deteniendo la búsqueda.")
					break
				}
				// Depuración: Imprimir cada EBR leído
				fmt.Println("EBR leído después de actualización:")
				Structs.PrintEBR(ebr)
				if ebr.PartNext == -1 {
					break
				}
				ebrPos = ebr.PartNext
			}
		}
	}

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Println("======FIN DELETE PARTITION======")
}

func ModifyPartition(path string, name string, add int, unit string) error {
	fmt.Println("======Start MODIFY PARTITION======")
	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return err
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return err
	}

	// Imprimir MBR antes de modificar
	fmt.Println("MBR antes de la modificación:")
	Structs.PrintMBR(TempMBR)

	// Buscar la partición por nombre
	var foundPartition *Structs.Partition
	var partitionType byte

	// Revisar si la partición es primaria o extendida
	for i := 0; i < 4; i++ {
		partitionName := strings.TrimRight(string(TempMBR.Partitions[i].Name[:]), "\x00")
		if partitionName == name {
			foundPartition = &TempMBR.Partitions[i]
			partitionType = TempMBR.Partitions[i].Type[0]
			break
		}
	}

	// Si no se encuentra en las primarias/extendidas, buscar en las particiones lógicas
	if foundPartition == nil {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					if err := Utilities.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
						fmt.Println("Error al leer EBR:", err)
						return err
					}

					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if ebrName == name {
						partitionType = 'l' // Partición lógica
						foundPartition = &Structs.Partition{
							Start: ebr.PartStart,
							Size:  ebr.PartSize,
						}
						break
					}

					// Continuar buscando el siguiente EBR
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
				if foundPartition != nil {
					break
				}
			}
		}
	}

	// Verificar si la partición fue encontrada
	if foundPartition == nil {
		fmt.Println("Error: No se encontró la partición con el nombre:", name)
		return nil // Salir si no se encuentra la partición
	}

	// Convertir unidades a bytes
	var addBytes int
	if unit == "k" {
		addBytes = add * 1024
	} else if unit == "m" {
		addBytes = add * 1024 * 1024
	} else {
		fmt.Println("Error: Unidad desconocida, debe ser 'k' o 'm'")
		return nil // Salir si la unidad no es válida
	}

	// Flag para saber si continuar o no
	var shouldModify = true

	// Comprobar si es posible agregar o quitar espacio
	if add > 0 {
		// Agregar espacio: verificar si hay suficiente espacio libre después de la partición
		nextPartitionStart := foundPartition.Start + foundPartition.Size
		if partitionType == 'l' {
			// Para particiones lógicas, verificar con el siguiente EBR o el final de la partición extendida
			for i := 0; i < 4; i++ {
				if TempMBR.Partitions[i].Type[0] == 'e' {
					extendedPartitionEnd := TempMBR.Partitions[i].Start + TempMBR.Partitions[i].Size
					if nextPartitionStart+int32(addBytes) > extendedPartitionEnd {
						fmt.Println("Error: No hay suficiente espacio libre dentro de la partición extendida")
						shouldModify = false
					}
					break
				}
			}
		} else {
			// Para primarias o extendidas
			if nextPartitionStart+int32(addBytes) > TempMBR.MbrSize {
				fmt.Println("Error: No hay suficiente espacio libre después de la partición")
				shouldModify = false
			}
		}
	} else {
		// Quitar espacio: verificar que no se reduzca el tamaño por debajo de 0
		if foundPartition.Size+int32(addBytes) < 0 {
			fmt.Println("Error: No es posible reducir la partición por debajo de 0")
			shouldModify = false
		}
	}

	// Solo modificar si no hay errores
	if shouldModify {
		foundPartition.Size += int32(addBytes)
	} else {
		fmt.Println("No se realizaron modificaciones debido a un error.")
		return nil // Salir si hubo un error
	}

	// Si es una partición lógica, sobrescribir el EBR
	if partitionType == 'l' {
		ebrPos := foundPartition.Start
		var ebr Structs.EBR
		if err := Utilities.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
			fmt.Println("Error al leer EBR:", err)
			return err
		}

		// Actualizar el tamaño en el EBR y escribirlo de nuevo
		ebr.PartSize = foundPartition.Size
		if err := Utilities.WriteObject(file, ebr, int64(ebrPos)); err != nil {
			fmt.Println("Error al escribir el EBR actualizado:", err)
			return err
		}

		// Imprimir el EBR modificado
		fmt.Println("EBR modificado:")
		Structs.PrintEBR(ebr)
	}

	// Sobrescribir el MBR actualizado
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error al escribir el MBR actualizado:", err)
		return err
	}

	// Imprimir el MBR modificado
	fmt.Println("MBR después de la modificación:")
	Structs.PrintMBR(TempMBR)

	fmt.Println("======END MODIFY PARTITION======")
	return nil
}

// Función para montar particiones
func Mount(path string, name string) {
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", path)
		return
	}
	defer file.Close()

	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	fmt.Printf("Buscando partición con nombre: '%s'\n", name)

	partitionFound := false
	var partition Structs.Partition
	var partitionIndex int

	// Convertir el nombre a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(name))

	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'p' && bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: Partición no encontrada o no es una partición primaria")
		return
	}

	// Verificar si la partición ya está montada
	if partition.Status[0] == '1' {
		fmt.Println("Error: La partición ya está montada")
		return
	}

	//fmt.Printf("Partición encontrada: '%s' en posición %d\n", string(partition.Name[:]), partitionIndex+1)

	// Generar el ID de la partición
	diskID := generateDiskID(path)

	// Verificar si ya se ha montado alguna partición de este disco
	mountedPartitionsInDisk := mountedPartitions[diskID]
	var letter byte

	if len(mountedPartitionsInDisk) == 0 {
		// Es un nuevo disco, asignar la siguiente letra disponible
		if len(mountedPartitions) == 0 {
			letter = 'a'
		} else {
			lastDiskID := getLastDiskID()
			lastLetter := mountedPartitions[lastDiskID][0].ID[len(mountedPartitions[lastDiskID][0].ID)-1]
			letter = lastLetter + 1
		}
	} else {
		// Utilizar la misma letra que las otras particiones montadas en el mismo disco
		letter = mountedPartitionsInDisk[0].ID[len(mountedPartitionsInDisk[0].ID)-1]
	}

	// Incrementar el número para esta partición
	carnet := "202401234" // Cambiar su carnet aquí
	lastTwoDigits := carnet[len(carnet)-2:]
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, partitionIndex+1, letter)

	// Actualizar el estado de la partición a montada y asignar el ID
	partition.Status[0] = '1'
	copy(partition.Id[:], partitionID)
	TempMBR.Partitions[partitionIndex] = partition
	mountedPartitions[diskID] = append(mountedPartitions[diskID], MountedPartition{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Escribir el MBR actualizado al archivo
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		return
	}

	fmt.Printf("Partición montada con ID: %s\n", partitionID)

	fmt.Println("")
	// Imprimir el MBR actualizado
	fmt.Println("MBR actualizado:")
	Structs.PrintMBR(TempMBR)
	fmt.Println("")

	// Imprimir las particiones montadas (solo estan mientras dure la sesion de la consola)
	PrintMountedPartitions()
}

func Unmount(id string) {
	fmt.Println("Desmontando partición con ID:", id)

	// Buscar la partición montada por ID
	var partitionFound *MountedPartition
	var diskID string
	var partitionIndex int

	for disk, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				partitionFound = &partitions[i]
				diskID = disk
				partitionIndex = i
				break
			}
		}
		if partitionFound != nil {
			break
		}
	}

	// Si no se encuentra la partición, mostrar un error
	if partitionFound == nil {
		fmt.Println("Error: No se encontró una partición montada con el ID proporcionado:", id)
		return
	}

	// Abrir el archivo del disco correspondiente
	file, err := Utilities.OpenFile(partitionFound.Path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", partitionFound.Path)
		return
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	// Buscar la partición en el MBR utilizando el nombre
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(partitionFound.Name))
	partitionUpdated := false

	for i := 0; i < 4; i++ {
		if bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			// Cambiar el estado de la partición de montada ('1') a desmontada ('0')
			TempMBR.Partitions[i].Status[0] = '0'
			// Borrar el ID de la partición
			copy(TempMBR.Partitions[i].Id[:], "")
			partitionUpdated = true
			break
		}
	}

	if !partitionUpdated {
		fmt.Println("Error: No se pudo encontrar la partición en el MBR para desmontar")
		return
	}

	// Sobrescribir el MBR actualizado al archivo
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		return
	}

	// Eliminar la partición de la lista de particiones montadas
	mountedPartitions[diskID] = append(mountedPartitions[diskID][:partitionIndex], mountedPartitions[diskID][partitionIndex+1:]...)

	// Si ya no hay particiones montadas en este disco, eliminar el disco de la lista
	if len(mountedPartitions[diskID]) == 0 {
		delete(mountedPartitions, diskID)
	}

	fmt.Println("Partición desmontada con éxito.")
	PrintMountedPartitions() // Mostrar las particiones montadas restantes
}

// Función para obtener el ID del último disco montado
func getLastDiskID() string {
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

func generateDiskID(path string) string {
	return strings.ToLower(path)
}
