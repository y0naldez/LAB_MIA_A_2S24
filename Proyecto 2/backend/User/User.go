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

func Login(user string, pass string, id string) (string, error) {
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
				return "", fmt.Errorf("Ya existe un usuario logueado en esta partición")
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
		return "", fmt.Errorf("No se encontró ninguna partición montada con el ID proporcionado")
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return "", fmt.Errorf("Error al abrir el archivo")
	}
	defer file.Close()

	var TempMBR Structs.MRB
	// Leer el MBR desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return "", fmt.Errorf("Error al leer el MBR")
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
					return "", fmt.Errorf("La partición no está montada")
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Println("Partition not found")
		return "", fmt.Errorf("No se encontró la partición con el ID proporcionado")
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock desde el archivo binario
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return "", fmt.Errorf("Error al leer el Superblock")
	}

	// Buscar el archivo de usuarios /users.txt -> retorna índice del Inodo
	indexInode := InitSearch("/users.txt", file, tempSuperblock)

	var crrInode Structs.Inode
	// Leer el Inodo desde el archivo binario
	if err := Utilities.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Structs.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return "", fmt.Errorf("Error al leer el Inodo")
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
		fmt.Println("Usuario logueado con éxito")
		DiskManagement.MarkPartitionAsLoggedIn(id) // Marcar la partición como logueada
		return "Inicio de sesión exitoso", nil
	}

	fmt.Println("Credenciales incorrectas")
	return "", fmt.Errorf("Credenciales incorrectas")
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

func GetCleanedData(data string) string {
	cleanedData := strings.TrimRight(data, "\x00") // Eliminar bytes nulos del final
	return cleanedData
}

func MkusrCommand(path string, newUser string) error {
	// Buscar la partición logueada
	mountedPartitions := DiskManagement.GetMountedPartitions()
	var partitionFound *DiskManagement.MountedPartition
	var filePath string

	// Buscar la partición que esté logueada
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Buscar la partición marcada como logueada
				partitionFound = &partition
				filePath = partition.Path
				break
			}
		}
		if partitionFound != nil {
			break
		}
	}

	if partitionFound == nil {
		fmt.Println("Error: No se encontró ninguna partición logueada.")
		return fmt.Errorf("no se encontró ninguna partición logueada")
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(filePath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo de la partición logueada:", err)
		return err
	}
	defer file.Close()

	// Leer el MBR desde el archivo binario
	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return err
	}

	// Imprimir el MBR
	Structs.PrintMBR(TempMBR)

	var index int = -1
	// Iterar sobre las particiones del MBR para encontrar la correcta
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), partitionFound.ID) {
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' {
					fmt.Println("Partition is mounted")
					index = i
				} else {
					fmt.Println("Partition is not mounted")
					return fmt.Errorf("partición no montada")
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Println("Error: No se encontró la partición en el MBR.")
		return fmt.Errorf("partición no encontrada en el MBR")
	}

	// Aquí se realiza la lectura del Superblock desde la partición correcta
	var tempSuperblock Structs.Superblock
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return err
	}

	// Mostrar información del Superblock
	Structs.PrintSuperblock(tempSuperblock)

	// Buscar el archivo de usuarios /users.txt utilizando el mismo flujo que en login
	fmt.Println("======Start BUSQUEDA INICIAL ======")
	inodeIndex := InitSearch(path, file, tempSuperblock)
	if inodeIndex == -1 {
		return fmt.Errorf("users.txt no encontrado")
	}
	fmt.Println("======End BUSQUEDA INICIAL======")

	// Leer el inodo de users.txt
	var usersInode Structs.Inode
	if err := Utilities.ReadObject(file, &usersInode, int64(tempSuperblock.S_inode_start+inodeIndex*int32(binary.Size(Structs.Inode{})))); err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Mostrar información del inodo para depuración
	Structs.PrintInode(usersInode)

	// Leer datos del archivo users.txt utilizando el inodo, igual que en login
	data := GetInodeFileData(usersInode, file, tempSuperblock)

	// Eliminar bytes nulos del final de los datos actuales
	cleanedData := GetCleanedData(data)

	// Imprimir el tamaño de los datos limpios
	fmt.Printf("Tamaño actual de los datos limpios en users.txt: %d bytes\n", len(cleanedData))

	// Concatenar el nuevo usuario a los datos limpios
	newData := cleanedData + "\n" + newUser

	// Imprimir el contenido y tamaño de cada parte para depuración
	fmt.Printf("Datos actuales: '%s' (%d bytes)\n", cleanedData, len(cleanedData))
	fmt.Printf("Nuevo usuario: '%s' (%d bytes)\n", newUser, len(newUser))
	fmt.Printf("Datos concatenados: '%s' (%d bytes)\n", newData, len(newData))

	// Verificar si el nuevo contenido cabe en el bloque
	if len(newData) > 64 {
		fmt.Println("Advertencia: El contenido excede el tamaño del bloque, se truncará el contenido")
		newData = newData[:64] // Truncar el contenido a 64 bytes
	}

	// Guardar los datos actualizados de nuevo en el bloque de archivo correspondiente
	var fileBlock Structs.Fileblock
	copy(fileBlock.B_content[:], newData)

	// Escribir el bloque actualizado de vuelta en el archivo
	if err := Utilities.WriteObject(file, fileBlock, int64(tempSuperblock.S_block_start+usersInode.I_block[0]*int32(binary.Size(Structs.Fileblock{})))); err != nil {
		return fmt.Errorf("error al escribir en users.txt: %v", err)
	}

	// Verificar el contenido del bloque después de la escritura
	fmt.Println("====== Verificación del bloque de archivo ======")
	var updatedFileBlock Structs.Fileblock
	if err := Utilities.ReadObject(file, &updatedFileBlock, int64(tempSuperblock.S_block_start+usersInode.I_block[0]*int32(binary.Size(Structs.Fileblock{})))); err != nil {
		return fmt.Errorf("error al leer el bloque actualizado: %v", err)
	}
	Structs.PrintFileblock(updatedFileBlock) // Imprimir el bloque de archivo actualizado

	fmt.Println("Usuario creado con éxito:", newUser)

	return nil
}
