package domain

type AuthMeResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type AuthMeStudentResponse struct {
	AuthMeResponse
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}
