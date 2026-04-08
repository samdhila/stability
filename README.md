# Stability Test Task API - Laporan Perbaikan Bug

Dokumen ini memuat daftar bug yang ditemukan melalui pengujian menggunakan Postman, beserta detail perbaikannya. Fokus dari perbaikan ini adalah memastikan API berfungsi sesuai dengan standar `RESTful behavior`, me-return HTTP status yang tepat, serta menjaga integritas data tanpa mengubah arsitektur dasar backend. Selain membenahi bug, dokumen ini juga mencakup sebuah improvement, yaitu implementasi penambahan fitur Endpoint Update (`PUT`) untuk melengkapi fungsionalitas CRUD secara utuh.

---

## 1. Incorrect Status Codes

### Bug 1.1: Pencarian Task Gagal yang Me-return Status `200 OK`
*   **Behavior Awal**: Saat men-send request `GET` ke ID task yang tidak terdaftar (contoh: `GET /tasks/999`), aplikasi me-return HTTP status `200 OK` (sukses), meskipun JSON body berisi laporan error `{"error": "task not found"}`.
*   **Behavior yang Diharapkan**: Aplikasi seharusnya me-return status code `404 Not Found` ketika data yang diminta tidak tersedia.
*   **Solusi Perbaikan**: Meng-update logika *response handler* agar me-return `fiber.StatusNotFound` jika hasil pencarian dari database bernilai `nil` (kosong).

**Kode yang Diperbaiki (pada `handlers/task_handler.go`)**:
```go
// Sebelum
if task == nil {
    return c.Status(200).JSON(fiber.Map{
        "error": "task not found",
    })
}

// Sesudah
if task == nil {
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
        "error": "task not found",
    })
}
```

### Bug 1.2: Pembuatan Task Baru yang Me-return Status `200 OK`
*   **Behavior Awal**: Saat sukses meng-create task baru via `POST /tasks`, aplikasi me-return HTTP status `200 OK`.
*   **Behavior yang Diharapkan**: Standar API RESTful mengharuskan proses pembuatan data baru (resource creation) untuk me-return HTTP status `201 Created`.
*   **Solusi Perbaikan**: Meng-update endpoint `POST` agar me-return HTTP status `fiber.StatusCreated` setelah proses pembuatan selesai.

**Kode yang Diperbaiki (pada `handlers/task_handler.go`)**:
```go
// Sebelum
return c.JSON(task)

// Sesudah
return c.Status(fiber.StatusCreated).JSON(createdTask)
```

---

## 2. Missing Validation

### Bug 2.1: Bisa Men-submit POST Body Tanpa Judul Task
*   **Behavior Awal**: User dapat men-send JSON payload kosong `{}` melalui `POST /tasks`, dan API menerimanya sehingga berhasil meng-create baris task tanpa judul di dalam database memori.
*   **Behavior yang Diharapkan**: Sistem perlu mewajibkan User untuk mengisi field penting seperti `Title`. Jika parameter ini kosong, request harus di-reject.
*   **Solusi Perbaikan**: Menambahkan aturan validasi pada isian `task.Title`. Jika bernilai kosong, sistem akan langsung me-reject proses dan me-return pesan error `400 Bad Request`.

**Kode yang Diperbaiki (pada `handlers/task_handler.go`)**:
```go
// Tambahan Kode
if task.Title == "" {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "error": "title is required",
    })
}
```

### Bug 2.2: Penanganan Format ID Parameter yang Invalid
*   **Behavior Awal**: Mengeksekusi pencarian atau proses men-delete task dengan format parameter huruf (contoh: `GET /tasks/abc`) dibiarkan berjalan oleh sistem. Kegagalan konversi tipe data diabaikan (`id, _ := strconv.Atoi(idParam)`), sehingga sistem secara keliru me-return laporan "task not found".
*   **Behavior yang Diharapkan**: Request ID dengan format huruf atau karakter yang tidak valid harus segera di-intercept dan di-reject dengan HTTP status `400 Bad Request`.
*   **Solusi Perbaikan**: Menge-catch potensi error dari proses *parsing* paket `strconv`, lalu menyusun penanganannya agar me-return status `fiber.StatusBadRequest`.

**Kode yang Diperbaiki (pada `handlers/task_handler.go`)**:
```go
// Sebelum
id, _ := strconv.Atoi(idParam)

// Sesudah
id, err := strconv.Atoi(idParam)
if err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "error": "invalid task id format",
    })
}
```

---

## 3. Endpoint Returns Incorrect Data

### Bug 3.1: Task Baru yang Selalu Mendapat Default ID `0`
*   **Behavior Awal**: Sistem mengalokasikan nilai ID `0` secara otomatis untuk setiap payload task baru yang di-submit tanpa field ID. Akibatnya, pembuatan data secara berulang akan menghasilkan identitas duplikat, yaitu nilai `0`.
*   **Behavior yang Diharapkan**: Struktur backend harus memfasilitasi fungsionalitas auto-increment untuk menempatkan angka ID yang spesifik, unik, dan berurutan untuk setiap task resource baru.
*   **Solusi Perbaikan**: Meng-refactor sistem database di `task_store.go` dengan mengimplementasikan variabel counter `nextID`. Fungsi `AddTask` kini langsung meng-overwrite parameter ID sesuai dengan urutan counter internal tersebut.

**Kode yang Diperbaiki (pada `store/task_store.go`)**:
```go
// Sebelum
func AddTask(task models.Task) {
    Tasks = append(Tasks, task)
}

// Sesudah
var nextID = 3

func AddTask(task models.Task) models.Task {
    task.ID = nextID
    nextID++
    Tasks = append(Tasks, task)
    return task
}
```

### Bug 3.2: Proses Men-delete Data yang Tidak Ada, Namun Tetap Dianggap Sukses
*   **Behavior Awal**: Eksekusi perintah `DELETE` untuk menghapus task yang tidak terdaftar di database (contoh: `DELETE /tasks/999`) tetap direspon sukses dan aplikasi me-return `{ "message" : "deleted" }`.
*   **Behavior yang Diharapkan**: Endpoint perlu memberikan feedback operasional yang akurat dengan me-return status `404 Not Found` karena target data dari awal tidak tersedia.
*   **Solusi Perbaikan**: Memodifikasi tipe balasan pada fungsi `store.DeleteTask()` agar men-trigger indikator `boolean`, yang memverifikasi kesuksesan proses pencabutan item. Dari lapisan handler (*controller*), sistem menganalisis indikator ini untuk me-return HTTP status yang akurat.

**Kode yang Diperbaiki (pada `handlers/task_handler.go` & `store/task_store.go`)**:
```go
// Sebelum (pada area Handler)
store.DeleteTask(id)
return c.JSON(fiber.Map{
    "message": "deleted",
})

// Sesudah (pada area Handler)
deleted := store.DeleteTask(id)
if !deleted {
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
        "error": "task not found",
    })
}
return c.JSON(fiber.Map{
    "message": "deleted",
})
```

---

## 4. One Improvement: *NEW:* Update Endpoint

### Penambahan Endpoint `PUT /tasks/:id`
*   **Deskripsi Penggunaan**: Endpoint standar yang didedikasikan untuk meng-update keseluruhan eksistensi data *task* (Fungsi 'U' di dalam operasi CRUD).
*   **Penerapan Validasi Otomatis**:
    *   Memastikan `:id` parameter pada URL memiliki format angka. Sistem me-return `400 Bad Request` jika format tersebut tidak valid.
    *   Menge-catch input format JSON payload yang tidak tepat memanfaatkan *middleware* Fiber `BodyParser`.
    *   Mewajibkan keberadaan teks pada isian `"Title"`. Sistem secara otomatis me-return `400 Bad Request` jika parameter ini dikosongkan.
    *   Mengecek eksistensi ID task di database memori (*in-memory*) sebelum sistem meng-overwrite objek data. Sistem me-return `404 Not Found` jika *target ID* tidak terdeteksi.
*   **Behavior Sistem**: Meng-update isian task (`Title` dan `Done`) berdasarkan JSON payload, tanpa mengizinkan modifikasi terhadap nilai asli `ID` dari entitas tersebut (*enforce immutability*). Proses ini me-return data task yang telah diperbarui beserta HTTP status `200 OK`.
*   **Contoh URL Valid**: `PUT http://localhost:3000/tasks/1` (Instruksi spesifik untuk meng-update entitas task ber-`ID: 1`).
*   **Ekspektasi Input Payload (JSON)**: Input isian `id` di dalam badan JSON bersifat opsional. Regulasi backend secara eksklusif hanya mem-fetch acuan ID yang akan dimutasi berdasarkan validasi target dari parameter URL.
    ```json
    {
      "title": "Complete Go tutorial",
      "done": true
    }
    ```

**Kode Fungsionalitas yang Ditambahkan (pada `main.go`)**:
```go
app.Put("/tasks/:id", handlers.UpdateTask)
```

**Kode Fungsionalitas yang Ditambahkan (pada `handlers/task_handler.go`)**:
```go
func UpdateTask(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id format",
		})
	}

	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if task.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "title is required",
		})
	}

	updatedTask := store.UpdateTask(id, task)
	if updatedTask == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.JSON(updatedTask)
}
```

**Kode Fungsionalitas yang Ditambahkan (pada `store/task_store.go`)**:
```go
func UpdateTask(id int, updatedTask models.Task) *models.Task {
    for i, t := range Tasks {
        if t.ID == id {
            updatedTask.ID = id
            Tasks[i] = updatedTask
            return &Tasks[i]
        }
    }
    return nil
}
```
