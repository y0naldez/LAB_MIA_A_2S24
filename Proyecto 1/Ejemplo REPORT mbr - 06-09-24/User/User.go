package User

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/DiskManagement"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

func Login(user string, pass string, id string) {
	fmt.Println("======Start LOGIN======")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	fmt.Println("Id:", id)

	// Verificar si el usuario ya está logueado buscando en las particiones montadas
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var login bool = false

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id && partition.LoggedIn { // Verifica si ya está logueado
				fmt.Println("Ya existe un usuario logueado!")
				return
			}
			if partition.ID == id { // Encuentra la partición correcta
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No se encontró ninguna partición montada con el ID proporcionado")
		return
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return
	}
	defer file.Close()

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Imprimir el MBR
	Structs.PrintMBR(TempMBR)
	fmt.Println("-------------")

	var index int = -1
	// Iterar sobre las particiones del MBR para encontrar la correcta
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' {
					fmt.Println("Partition is mounted")
					index = i
				} else {
					fmt.Println("Partition is not mounted")
					return
				}
				break
			}
		}
	}

	if index != -1 {
		Structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Partition not found")
		return
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock desde el archivo binario
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return
	}

	// Buscar el archivo de usuarios /users.txt -> retorna índice del Inodo
	indexInode := InitSearch("/users.txt", file, tempSuperblock)

	var crrInode Structs.Inode
	// Leer el Inodo desde el archivo binario
	if err := Utilities.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Structs.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return
	}

	// Leer datos del archivo
	data := GetInodeFileData(crrInode, file, tempSuperblock)

	// Dividir la cadena en líneas
	lines := strings.Split(data, "\n")

	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")

		if len(words) == 5 {
			if (strings.Contains(words[3], user)) && (strings.Contains(words[4], pass)) {
				login = true
				break
			}
		}
	}

	// Imprimir información del Inodo
	fmt.Println("Inode", crrInode.I_block)

	// Si las credenciales son correctas y marcamos como logueado
	if login {
		fmt.Println("Usuario logueado con exito")
		DiskManagement.MarkPartitionAsLoggedIn(id) // Marcar la partición como logueada
	}

	fmt.Println("======End LOGIN======")
}

func InitSearch(path string, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INICIAL ======")
	fmt.Println("path:", path)
	// path = "/ruta/nueva"

	// split the path by /
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 Structs.Inode
	// Read object from bin file
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}

	fmt.Println("======End BUSQUEDA INICIAL======")

	return SarchInodeByPath(StepsPath, Inode0, file, tempSuperblock)
}

// stack
func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func SarchInodeByPath(StepsPath []string, Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INODO POR PATH======")
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	fmt.Println("========== SearchedName:", SearchedName)

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFolderBlock Structs.Folderblock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Folderblock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {
					// fmt.Println("Folder found======")
					fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							fmt.Println("Folder found======")
							return folder.B_inodo
						} else {
							fmt.Println("NextInode======")
							var NextInode Structs.Inode
							// Read object from bin file
							if err := Utilities.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Structs.Inode{})))); err != nil {
								return -1
							}
							return SarchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End BUSQUEDA INODO POR PATH======")
	return 0
}

func GetInodeFileData(Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) string {
	fmt.Println("======Start CONTENIDO DEL BLOQUE======")
	index := int32(0)
	// define content as a string
	var content string

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			//Dentro de los directos
			if index < 13 {
				var crrFileBlock Structs.Fileblock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Fileblock{})))); err != nil {
					return ""
				}

				content += string(crrFileBlock.B_content[:])

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End CONTENIDO DEL BLOQUE======")
	return content
}

// MKUSER
func AppendToFileBlock(inode *Structs.Inode, newData string, file *os.File, superblock Structs.Superblock) error {
	// Leer el contenido existente del archivo utilizando la función GetInodeFileData
	existingData := GetInodeFileData(*inode, file, superblock)

	// Concatenar el nuevo contenido
	fullData := existingData + newData

	// Asegurarse de que el contenido no exceda el tamaño del bloque
	if len(fullData) > len(inode.I_block)*binary.Size(Structs.Fileblock{}) {
		// Si el contenido excede, necesitas manejar bloques adicionales
		return fmt.Errorf("el tamaño del archivo excede la capacidad del bloque actual y no se ha implementado la creación de bloques adicionales")
	}

	// Escribir el contenido actualizado en el bloque existente
	var updatedFileBlock Structs.Fileblock
	copy(updatedFileBlock.B_content[:], fullData)
	if err := Utilities.WriteObject(file, updatedFileBlock, int64(superblock.S_block_start+inode.I_block[0]*int32(binary.Size(Structs.Fileblock{})))); err != nil {
		return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
	}

	// Actualizar el tamaño del inodo
	inode.I_size = int32(len(fullData))
	if err := Utilities.WriteObject(file, *inode, int64(superblock.S_inode_start+inode.I_block[0]*int32(binary.Size(Structs.Inode{})))); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	return nil
}
