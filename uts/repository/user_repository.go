package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
)

// UserRepository adalah kontrak (interface) yang mendefinisikan operasi-operasi database untuk entitas User.
// Catatan: Karena model kita menggunakan Username (NIM/Admin ID) sebagai pengenal, fungsi pencarian
// disesuaikan menjadi FindByUsername (sepadan dengan FindByEmail pada instruksi awal).
type UserRepository interface {
	FindByUsername(username string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	Create(user *domain.User) error
}

// userRepository adalah implementasi nyata dari interface UserRepository menggunakan GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository adalah constructor untuk membuat instance dari userRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	// Mencari satu record pertama yang cocok dengan username
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err // Bisa mengembalikan gorm.ErrRecordNotFound jika tidak ada
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	// Mencari satu record berdasarkan primary key (ID)
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	// Menyimpan data user baru ke database
	return r.db.Create(user).Error
}
