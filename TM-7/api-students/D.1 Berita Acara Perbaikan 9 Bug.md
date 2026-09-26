Bug #3: DecodeCursor menggunakan ErrInvalidCursor yang belum dideklarasikan sehingga menyebabkan compile error undefined: ErrInvalidCursor. Perbaikan dilakukan dengan mendeklarasikan ErrInvalidCursor sebagai error package-level menggunakan errors.New().

Bug #5: Kondisi penentuan level logging pada newErrorHandler menggunakan operator < fiber.StatusInternalServerError, sehingga error 4xx masuk ke level ERROR. Sesuai spesifikasi, error 4xx harus dicatat sebagai WARN, sedangkan error 5xx sebagai ERROR. Perbaikan dilakukan dengan mengubah operator menjadi >=.

Bug #6: RequestLogger telah menghitung status response yang benar pada variabel lokal status, tetapi atribut access log masih mengambil c.Response().StatusCode(). Akibatnya status pada access log dapat berbeda dari status yang diterima client, khususnya ketika handler mengembalikan error. Perbaikan dilakukan dengan menggunakan variabel status pada slog.Int("status", status).

Bug #7: translateError() mengembalikan nil pada error yang tidak dikenali. Hal tersebut menyebabkan error teknis yang seharusnya diteruskan sebagai kegagalan internal dapat tertelan. Sesuai prinsip fail closed, error yang tidak dikenali harus diterjemahkan menjadi helper.Internal(err) sehingga menghasilkan HTTP 500 INTERNAL_ERROR.

Bug #8: Custom validation strongpassword menggunakan kondisi != "", sehingga hasil validasi terbalik. Password yang menghasilkan string error justru dianggap valid, sedangkan password yang memenuhi persyaratan dianggap tidak valid. Perbaikan dilakukan dengan mengubah kondisi menjadi passwordStrength(fl.Field().String()) == "".

