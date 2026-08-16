package main

import "fmt"

func swap(a, b *int) {
	// Simpan nilai a sementara
	temp := *a
	// masukkan nilai b ke a
	*a = *b
	// masukkan nilai sementara ke b
	*b = temp
}

// s adalah pointer ke slice string
// newItem adalah item baru yang akan ditambahkan
func updateSlice(s *[]string, newItem string) {
	// Ambil slice asli, lalu tambahkan item baru ke dalamnya
	*s = append(*s, newItem)
}

func main() {
	// swap
	a := 10
	b := 20

	fmt.Println("Sebelum swap:")
	fmt.Println("a =", a)
	fmt.Println("b =", b)

	// Memanggil fungsi swap dengan passing address dari variabel a dan b
	swap(&a, &b)

	fmt.Println("\nSetelah swap:")
	fmt.Println("a =", a)
	fmt.Println("b =", b)

	// updateSlice
	hobi := []string{"Membaca", "Menulis"}

	fmt.Println("\nSlice sebelum update:")
	fmt.Println(hobi)

	// Memanggil fungsi updateSlice dengan passing address dari slice hobi dan item baru "Gaming"
	updateSlice(&hobi, "Gaming")

	fmt.Println("Slice setelah update:")
	fmt.Println(hobi)
}