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

func PrintSuperblock(sb Superblock) {
	fmt.Println("====== Superblock ======")
	fmt.Printf("S_filesystem_type: %d\n", sb.S_filesystem_type)
	fmt.Printf("S_inodes_count: %d\n", sb.S_inodes_count)
	fmt.Printf("S_blocks_count: %d\n", sb.S_blocks_count)
	fmt.Printf("S_free_blocks_count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("S_free_inodes_count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("S_mtime: %s\n", string(sb.S_mtime[:]))
	fmt.Printf("S_umtime: %s\n", string(sb.S_umtime[:]))
	fmt.Printf("S_mnt_count: %d\n", sb.S_mnt_count)
	fmt.Printf("S_magic: 0x%X\n", sb.S_magic) // Usamos 0x%X para mostrarlo en formato hexadecimal
	fmt.Printf("S_inode_size: %d\n", sb.S_inode_size)
	fmt.Printf("S_block_size: %d\n", sb.S_block_size)
	fmt.Printf("S_fist_ino: %d\n", sb.S_fist_ino)
	fmt.Printf("S_first_blo: %d\n", sb.S_first_blo)
	fmt.Printf("S_bm_inode_start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("S_bm_block_start: %d\n", sb.S_bm_block_start)
	fmt.Printf("S_inode_start: %d\n", sb.S_inode_start)
	fmt.Printf("S_block_start: %d\n", sb.S_block_start)
	fmt.Println("========================")
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

func PrintInode(inode Inode) {
	fmt.Println("====== Inode ======")
	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", string(inode.I_atime[:]))
	fmt.Printf("I_ctime: %s\n", string(inode.I_ctime[:]))
	fmt.Printf("I_mtime: %s\n", string(inode.I_mtime[:]))
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Println("===================")
}

type Folderblock struct {
	B_content [4]Content
}

func PrintFolderblock(folderblock Folderblock) {
	fmt.Println("====== Folderblock ======")
	for i, content := range folderblock.B_content {
		fmt.Printf("Content %d: Name: %s, Inodo: %d\n", i, string(content.B_name[:]), content.B_inodo)
	}
	fmt.Println("=========================")
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type Fileblock struct {
	B_content [64]byte
}

func PrintFileblock(fileblock Fileblock) {
	fmt.Println("====== Fileblock ======")
	fmt.Printf("B_content: %s\n", string(fileblock.B_content[:]))
	fmt.Println("=======================")
}

type Pointerblock struct {
	B_pointers [16]int32
}

func PrintPointerblock(pointerblock Pointerblock) {
	fmt.Println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_pointers {
		fmt.Printf("Pointer %d: %d\n", i, pointer)
	}
	fmt.Println("=========================")
}

//EXT3

type Content_J struct {
	Operation [10]byte
	Path      [100]byte
	Content   [100]byte
	Date      [17]byte
}

type Journaling struct {
	Size      int32
	Ultimo    int32
	Contenido [50]Content_J
}
