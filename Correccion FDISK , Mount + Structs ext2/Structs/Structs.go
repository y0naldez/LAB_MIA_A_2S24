package Structs

import (
	"fmt"
)

type MRB struct {
	MbrSize      int32    // 4 bytes //int32 va desde -2,147,483,648 hasta 2,147,483,647.
	CreationDate [10]byte // 10 bytes
	Signature    int32    // 4 bytes
	Fit          [1]byte  // 1 byte
	Partitions   [4]Partition
}

func PrintMBR(data MRB) {
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d", string(data.CreationDate[:]), string(data.Fit[:]), data.MbrSize))
	for i := 0; i < 4; i++ {
		PrintPartition(data.Partitions[i])
	}
}

type Partition struct {
	Status      [1]byte
	Type        [1]byte
	Fit         [1]byte
	Start       int32
	Size        int32
	Name        [16]byte
	Correlative int32
	Id          [4]byte
}

func PrintPartition(data Partition) {
	fmt.Println(fmt.Sprintf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s", string(data.Name[:]), string(data.Type[:]), data.Start, data.Size, string(data.Status[:]), string(data.Id[:])))
}

type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
}

func PrintEBR(data EBR) {
	fmt.Println(fmt.Sprintf("Name: %s, fit: %c, start: %d, size: %d, next: %d, mount: %c",
		string(data.PartName[:]),
		data.PartFit,
		data.PartStart,
		data.PartSize,
		data.PartNext,
		data.PartMount))
}

//Estructuras relacionadas a EXT2

type Superblock struct {
	S_filesystem_type   int32    // Guarda el número que identifica el sistema de archivos utilizado
	S_inodes_count      int32    // Guarda el número total de inodos
	S_blocks_count      int32    // Guarda el número total de bloques
	S_free_blocks_count int32    // Contiene el número de bloques libres
	S_free_inodes_count int32    // Contiene el número de inodos libres
	S_mtime             [17]byte // Última fecha en el que el sistema fue montado
	S_umtime            [17]byte // Última fecha en que el sistema fue desmontado
	S_mnt_count         int32    // Indica cuantas veces se ha montado el sistema
	S_magic             int32    // Valor que identifica al sistema de archivos, tendrá el valor 0xEF53
	S_inode_size        int32    // Tamaño del inodo
	S_block_size        int32    // Tamaño del bloque
	S_fist_ino          int32    // Primer inodo libre (dirección del inodo)
	S_first_blo         int32    // Primer bloque libre (dirección del inodo)
	S_bm_inode_start    int32    // Guardará el inicio del bitmap de inodos
	S_bm_block_start    int32    // Guardará el inicio del bitmap de bloques
	S_inode_start       int32    // Guardará el inicio de la tabla de inodos
	S_block_start       int32    // Guardará el inicio de la tabla de bloques
}

type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime [17]byte
	I_ctime [17]byte
	I_mtime [17]byte
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

type Folderblock struct {
	B_content [4]Content
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type Fileblock struct {
	B_content [64]byte
}

type Pointerblock struct {
	B_pointers [16]int32
}
