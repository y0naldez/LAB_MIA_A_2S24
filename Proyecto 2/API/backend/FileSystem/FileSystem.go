package FileSystem

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/DiskManagement"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
)

func Mkfs(id string, type_ string, fs_ string) {
	fmt.Println("======INICIO MKFS======")
	fmt.Println("Id:", id)
	fmt.Println("Type:", type_)
	fmt.Println("Fs:", fs_)

	// Buscar la partición montada por ID
	var mountedPartition DiskManagement.MountedPartition
	var partitionFound bool

	for _, partitions := range DiskManagement.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Particion no encontrada")
		return
	}

	if mountedPartition.Status != '1' { // Verifica si la partición está montada
		fmt.Println("La particion aun no esta montada")
		return
	}

	// Abrir archivo binario
	file, err := Utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return
	}

	var TempMBR Structs.MRB
	// Leer objeto desde archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	// Imprimir objeto
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	var index int = -1
	// Iterar sobre las particiones para encontrar la que tiene el nombre correspondiente
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				index = i
				break
			}
		}
	}

	if index != -1 {
		Structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Particion no encontrada (2)")
		return
	}

	// Calcular el número de inodos y bloques
	numerador := int32(TempMBR.Partitions[index].Size - int32(binary.Size(Structs.Superblock{})))
	denominador_base := int32(4 + int32(binary.Size(Structs.Inode{})) + 3*int32(binary.Size(Structs.Fileblock{})))
	var temp int32 = 0

	// Agregar journaling si es EXT3
	if fs_ == "3fs" {
		temp = int32(binary.Size(Structs.Journaling{})) // Añadir journaling para EXT3
	} else if fs_ == "2fs" {
		temp = 0 // Sin journaling para EXT2
	} else {
		fmt.Println("Error: Sólo están disponibles los sistemas de archivos 2FS y 3FS.")
		return
	}

	denominador := denominador_base + temp
	n := int32(numerador / denominador)

	fmt.Println("INODOS:", n)

	// Crear el Superblock con todos los campos calculados
	var newSuperblock Structs.Superblock
	if fs_ == "2fs" {
		newSuperblock.S_filesystem_type = 2 // EXT2
	} else if fs_ == "3fs" {
		newSuperblock.S_filesystem_type = 3 // EXT3
	}
	newSuperblock.S_inodes_count = n
	newSuperblock.S_blocks_count = 3 * n
	newSuperblock.S_free_blocks_count = 3*n - 2
	newSuperblock.S_free_inodes_count = n - 2
	copy(newSuperblock.S_mtime[:], "23/08/2024")
	copy(newSuperblock.S_umtime[:], "23/08/2024")
	newSuperblock.S_mnt_count = 1
	newSuperblock.S_magic = 0xEF53
	newSuperblock.S_inode_size = int32(binary.Size(Structs.Inode{}))
	newSuperblock.S_block_size = int32(binary.Size(Structs.Fileblock{}))

	// Calcula las posiciones de inicio
	newSuperblock.S_bm_inode_start = TempMBR.Partitions[index].Start + int32(binary.Size(Structs.Superblock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + n
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*n
	newSuperblock.S_block_start = newSuperblock.S_inode_start + n*newSuperblock.S_inode_size

	// Llamar a la función correspondiente para crear el sistema de archivos
	if fs_ == "2fs" {
		create_ext2(n, TempMBR.Partitions[index], newSuperblock, "23/08/2024", file)
	} else if fs_ == "3fs" {
		create_ext3(n, TempMBR.Partitions[index], newSuperblock, "23/08/2024", file)
	}

	// Cerrar archivo binario
	defer file.Close()

	fmt.Println("======FIN MKFS======")
}

func create_ext2(n int32, partition Structs.Partition, newSuperblock Structs.Superblock, date string, file *os.File) {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("INODOS:", n)

	// Imprimir Superblock inicial
	Structs.PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	// Escribe los bitmaps de inodos y bloques en el archivo
	for i := int32(0); i < n; i++ {
		if err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	for i := int32(0); i < 3*n; i++ {
		if err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	// Inicializa inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Crea la carpeta raíz y el archivo users.txt
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Escribe el superbloque actualizado al archivo
	if err := Utilities.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Marca los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Imprimir Fileblocks
	for i := int32(0); i < 1; i++ {
		var fileblock Structs.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Structs.Folderblock{})) + i*int32(binary.Size(Structs.Fileblock{})))
		if err := Utilities.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return
		}
		Structs.PrintFileblock(fileblock)
	}

	/*
		// Imprimir el Superblock final
		Structs.PrintSuperblock(newSuperblock)
	*/

	fmt.Println("======End CREATE EXT2======")
}

func create_ext3(n int32, partition Structs.Partition, newSuperblock Structs.Superblock, date string, file *os.File) {
	fmt.Println("======Start CREATE EXT3======")
	fmt.Println("INODOS:", n)

	// Imprimir Superblock inicial
	Structs.PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	// Inicializa el journaling
	if err := initJournaling(newSuperblock, file); err != nil {
		fmt.Println("Error al inicializar el Journaling: ", err)
		return
	}
	fmt.Println("Journaling inicializado correctamente.")

	// Escribe los bitmaps de inodos y bloques en el archivo
	for i := int32(0); i < n; i++ {
		if err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}
	fmt.Println("Bitmap de inodos escrito correctamente.")

	for i := int32(0); i < 3*n; i++ {
		if err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}
	fmt.Println("Bitmap de bloques escrito correctamente.")

	// Inicializa inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Inodos y bloques inicializados correctamente.")

	// Crea la carpeta raíz y el archivo users.txt
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Carpeta raíz y archivo users.txt creados correctamente.")

	// Escribe el superbloque actualizado al archivo
	if err := Utilities.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Superbloque escrito correctamente.")

	// Marca los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Inodos y bloques iniciales marcados como usados correctamente.")

	// Imprimir Fileblocks (similar a EXT2 para verificación)
	for i := int32(0); i < 1; i++ {
		var fileblock Structs.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Structs.Folderblock{})) + i*int32(binary.Size(Structs.Fileblock{})))
		if err := Utilities.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return
		}
		Structs.PrintFileblock(fileblock)
	}
	fmt.Println("Fileblocks impresos correctamente.")

	fmt.Println("======End CREATE EXT3======")
}

// Función para inicializar el journaling
func initJournaling(newSuperblock Structs.Superblock, file *os.File) error {
	var journaling Structs.Journaling
	journaling.Size = 50
	journaling.Ultimo = 0

	// Calcula la posición de inicio del journaling (después del superbloque y bitmaps)
	journalingStart := newSuperblock.S_inode_start - int32(binary.Size(Structs.Journaling{}))*journaling.Size

	// Escribir el journaling en el archivo en la ubicación adecuada
	for i := 0; i < 50; i++ {
		if err := Utilities.WriteObject(file, journaling, int64(journalingStart+int32(i*binary.Size(journaling)))); err != nil {
			return fmt.Errorf("error al inicializar el journaling: %v", err)
		}
	}

	fmt.Println("Journaling inicializado correctamente.")
	return nil
}

// Función auxiliar para inicializar inodos y bloques
func initInodesAndBlocks(n int32, newSuperblock Structs.Superblock, file *os.File) error {
	var newInode Structs.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := Utilities.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Structs.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock Structs.Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := Utilities.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Structs.Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock Structs.Superblock, date string, file *os.File) error {
	var Inode0, Inode1 Structs.Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	// Asignar el tamaño real del contenido
	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize // Esto ahora refleja el tamaño real del contenido

	// Inicializar el bloque de archivo, pero solo copiar el contenido real
	var Fileblock1 Structs.Fileblock
	copy(Fileblock1.B_content[:actualSize], data) // Aquí aseguramos que solo copiamos los bytes necesarios

	// Depuración: Mostrar los datos byte por byte en el bloque
	fmt.Println("Contenido del bloque antes de escribir (byte por byte):")
	for i := 0; i < len(Fileblock1.B_content); i++ {
		fmt.Printf("Byte %d: %v (%c)\n", i, Fileblock1.B_content[i], Fileblock1.B_content[i])
	}

	// Comprobar si el bloque tiene espacios no utilizados (bytes nulos o extra)
	var nullBytes int
	for i := 0; i < len(Fileblock1.B_content); i++ {
		if Fileblock1.B_content[i] == 0 {
			nullBytes++
		}
	}
	fmt.Printf("Cantidad de bytes nulos en el bloque: %d\n", nullBytes)

	var Folderblock0 Structs.Folderblock
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	// Escribir los inodos y bloques en las posiciones correctas
	if err := Utilities.WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(Structs.Inode{})))); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(Structs.Folderblock{})))); err != nil {
		return err
	}

	return nil
}

// Función auxiliar para inicializar un inodo
func initInode(inode *Structs.Inode, date string) {
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	copy(inode.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_block[i] = -1
	}
}

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock Structs.Superblock, file *os.File) error {
	if err := Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}
