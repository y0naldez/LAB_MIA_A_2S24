package Utilities

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"proyecto1/Structs"
	"strings"
)

// Funcion para crear un archivo binario
func CreateFile(name string) error {
	//Se asegura que el archivo existe
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Err CreateFile dir==", err)
		return err
	}

	// Crear archivo
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Err CreateFile create==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

// Funcion para abrir un archivo binario ead/write mode
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

// Funcion para escribir un objecto en un archivo binario
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

// Funcion para leer un objeto de un archivo binario
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

// Función para generar el reporte del MBR y los EBRs en formato Graphviz y guardarlo en un archivo .dot
func GenerateMBRReport(mbr Structs.MRB, ebrs []Structs.EBR, outputPath string, file *os.File) error {
	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	// Crear el archivo .dot donde se generará el reporte
	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	// Iniciar el contenido del archivo en formato Graphviz (.dot)
	content := "digraph G {\n"
	content += "\tnode [fillcolor=lightyellow style=filled]\n"

	// Subgrafo del MBR
	content += fmt.Sprintf("\tsubgraph cluster_MBR {\n\t\tcolor=lightgrey fillcolor=lightblue label=\"MBR\nTamaño: %d\nFecha Creación: %s\nDisk Signature: %d\" style=filled\n",
		mbr.MbrSize, string(mbr.CreationDate[:]), mbr.Signature)

	// Recorrer las particiones del MBR en orden
	lastPartId := ""
	for i := 0; i < 4; i++ {
		part := mbr.Partitions[i]
		if part.Size > 0 { // Si la partición tiene un tamaño válido
			partName := strings.TrimRight(string(part.Name[:]), "\x00") // Limpiar el nombre de la partición
			partId := fmt.Sprintf("PART%d", i+1)
			content += fmt.Sprintf("\t\t%s [label=\"Partición %d\nStatus: %s\nType: %s\nFit: %s\nStart: %d\nSize: %d\nName: %s\" fillcolor=green shape=box style=filled]\n",
				partId, i+1, string(part.Status[:]), string(part.Type[:]), string(part.Fit[:]), part.Start, part.Size, partName)

			// Conectar la partición actual con la anterior de manera invisible para mantener el orden
			if lastPartId != "" {
				content += fmt.Sprintf("\t\t%s -> %s [style=invis]\n", lastPartId, partId)
			}
			lastPartId = partId

			// Si la partición es extendida, leer los EBRs
			if string(part.Type[:]) == "e" {
				content += fmt.Sprintf("\tsubgraph cluster_EBR%d {\n\t\tcolor=black fillcolor=lightpink label=\"Partición Extendida %d\" style=dashed\n", i+1, i+1)

				// Recolectamos todos los EBRs en orden
				ebrPos := part.Start
				var ebrList []Structs.EBR
				for {
					var ebr Structs.EBR
					err := ReadObject(file, &ebr, int64(ebrPos)) // Asegúrate de que la función ReadObject proviene de Utilities
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					ebrList = append(ebrList, ebr)

					// Si no hay más EBRs, salir del bucle
					if ebr.PartNext == -1 {
						break
					}

					// Mover a la siguiente posición de EBR
					ebrPos = ebr.PartNext
				}

				// Ahora agregamos los EBRs en orden correcto
				lastEbrId := ""
				for j, ebr := range ebrList {
					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00") // Limpiar el nombre del EBR
					ebrId := fmt.Sprintf("EBR%d", j+1)
					content += fmt.Sprintf("\t\t%s [label=\"EBR\nStart: %d\nSize: %d\nNext: %d\nName: %s\" fillcolor=lightpink shape=box style=filled]\n",
						ebrId, ebr.PartStart, ebr.PartSize, ebr.PartNext, ebrName)

					// Conectar el EBR actual con el anterior de manera invisible para mantener el orden
					if lastEbrId != "" {
						content += fmt.Sprintf("\t\t%s -> %s [style=invis]\n", lastEbrId, ebrId)
					}
					lastEbrId = ebrId
				}

				content += "\t}\n" // Cerrar el subgrafo de la partición extendida
			}
		}
	}

	content += "\t}\n" // Cerrar el subgrafo del MBR

	content += "}\n" // Cerrar el grafo principal

	// Escribir el contenido en el archivo .dot
	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte MBR generado exitosamente en:", dotFilePath)
	return nil
}
