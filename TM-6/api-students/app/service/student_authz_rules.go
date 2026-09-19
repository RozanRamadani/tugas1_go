package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh data mahasiswa.
//
// Dua jalur yang diizinkan:
// 1. Kepemilikan (ownership) — ownerID mahasiswa sama dengan UserID pemanggil.
// 2. Permission — role-nya memiliki permission :any yang sesuai.
//
// Urutannya disengaja: pemeriksaan kepemilikan didahulukan karena paling
// murah dan paling sering benar. Bila keduanya gagal, jawabannya false.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// Jalur 1: Kepemilikan (ownership)
	if current.UserID == ownerID {
		return true
	}

	// Jalur 2: Permission :any
	return perms.Can(current.Role, anyPermission)
}
