package Structs

import (
	"fmt"
)

type MRB struct {
	MbrSize      int32    // 4 bytes //int32 va desde -2,147,483,648 hasta 2,147,483,647.
	CreationDate [10]byte // 10 bytes
	Signature    int32    // 4 bytes
	Fit          [1]byte  // 1 byte
}

func PrintMBR(data MRB) {
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d, Signature: %d",
		string(data.CreationDate[:]),
		string(data.Fit[:]),
		data.MbrSize,
		data.Signature))
}

/*
   MbrSize (4 bytes):
       Hex:

   CreationDate (10 bytes):
       Hex:

	Los siguientes 10 bytes son: 32 30 32 34 2D 30 38 2D 30 39,
	que en texto plano representan la fecha 2024-08-09.


   Signature (4 bytes):
       Hex:

   Fit (1 byte):
       Hex:


*/
