package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent menentukan apakah user boleh mengakses
// data student tertentu.
//
// User boleh mengakses jika:
// 1. User adalah owner dari student tersebut.
// 2. User memiliki permission :any yang sesuai.
//
// Ownership diperiksa di service layer karena membutuhkan
// data student dari database.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {

	// Owner selalu boleh mengakses datanya sendiri.
	if current.UserID == ownerID {
		return true
	}

	// Selain owner, harus mempunyai permission :any.
	return perms.Can(current.Role, anyPermission)
}
