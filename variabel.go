package main

import "fmt"

func main() {
	
	//Variabel
	var nama string = "Rozan Aiman"
	var umuur int = 20
	var ipk float64 = 3.99
	var aktif bool = true
	// Slice yang berisi string
	var hobi []string = []string{"Membaca", "Menulis", "Bermain Game"}
	
	fmt.Println("Nama:", nama)
	fmt.Println("Umur:", umuur)
	fmt.Println("IPK:", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Hobi:", hobi)

	// map untuk menyimpan datanya
	mahasiswa := make(map[string]float64)

	// menambahkan data
	mahasiswa["Rozan Aiman"] = 3.99
	mahasiswa["Sumbul"] = 3.75
	mahasiswa["Joseph"] = 3.85

	fmt.Println("Mahasiswa:", mahasiswa)

	// membaca data dan mengecek keberadaan key
	nilai, exist := mahasiswa["Rozan Aiman"]

	if exist {
		fmt.Println("IPK Rozan Aiman:", nilai)
	} else {
		fmt.Println("Data tidak ditemukan")
	}

	// hapus
	delete(mahasiswa, "Sumbul")
	fmt.Println("Setelah Sumbul dihapus:", mahasiswa)

	// menelusuri seluruh isi map
	fmt.Println("\nSeluruh data mahasiswa: ")

	for nama, nilai := range mahasiswa {
		fmt.Printf("%s: %.2f\n", nama, nilai)
	}

}
