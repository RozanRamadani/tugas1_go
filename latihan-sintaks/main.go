package main 
  
import "fmt" 
  
type User struct { 
    ID       int    `json:"id"` 
    Username string `json:"username"` 
    IsActive bool   `json:"is_active"` 
} 
  
func (u *User) Activate()   { u.IsActive = true } 
func (u User) Info() string { return fmt.Sprintf("%s aktif: %v", u.Username, u.IsActive) } 
  
func tukarNilai(a, b *int) { *a, *b = *b, *a } 

func main() { 
    // Variabel 
    var nama string = "Sari" 
    umur := 20 
    var kosong string 

  
    // Slice 
    nilai := []int{10, 20, 30} 
    nilai = append(nilai, 40) 
  
  
    // Map 
    skor := make(map[string]int) 
    skor["Sari"] = 90 
    if n, ada := skor["Budi"]; ada { 

    } else { 

    } 
  
    // Channel 
    ch := make(chan string) 
    go func() { ch <- "halo dari goroutine" }() 

  
    // Pointer 
    a, b := 1, 2 
    tukarNilai(&a, &b) 
    
  
    // Struct 
    u := User{ID: 1, Username: "sari"} 
    u.Activate() 
    
} 