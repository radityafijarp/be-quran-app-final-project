# Hacktogram - Backend

## Course Project

### Description

Pada project ini, Anda diminta untuk membangun backend menggunakan **Gin** dan **Gorm** untuk aplikasi _Hacktogram_ yang telah kamu bangun sebelumnya sebagai pengganti dari json-server yang ada. Fitur-fitur yang perlu Anda implementasikan adalah:

- Membuat API untuk _Register_.
- Membuat API untuk _Sign In_.
- Membuat middleware untuk _protected route_.
- Membuat API untuk mendapatkan seluruh data _photo_ yang dimiliki oleh _user_ yang berhasil _Sign In_.
- Membuat API untuk mendapatkan detail _photo_.
- Membuat API untuk menambahkan _photo_ _user_.
- Membuat handler untuk menangani **Not found**.

API ini memiliki dua endpoint yaitu: `/users` dan `/photos`.

- Endpoint `/users` digunakan management _user authentication_.
- Endpoint `/photos` digunakan untuk mengakses data photo.

### Struktur file dan folder

API ini memiliki tiga folder yaitu: `/model`, `/repository` dan `/service`.
- Folder `/` (root) berisi file utama `main.go` yang kita gunakan untuk inisialisasi aplikasi dan routing serta file pendukung lainnya.
- Folder `/model` berisi struktur data yang digunakan termasuk struktur data yang terdapat di database.
- Folder `/repository` berisi 2 sub-folder yaitu:
  - Subfolder `authRepository` berisi kode untuk melakukan otentikasi user.
  - Subfolder `dbRepository` berisi kode untuk berinteraksi dengan database.
- Folder `/service` berisi kode untuk melakukan interaksi dengan data, dalam hal ini melakukan registrasi dan login.

### Database Model and Schema

Aplikasi ini memiliki 2 tabel utama, yaitu `users`, `photos`. Tabel `users` digunakan untuk menyimpan data-data user, tabel `photos` digunakan untuk menyimpan data photo.

Tabel `users` hanya dapat memiliki satu sessions, dan tabel `sessions` dapat memiliki banyak users. Tabel `users` dan `sessions` memiliki relasi one-to-many.

Tabel `users` memiliki relasi one-to-many dengan tabel `photos`, dimana banyak photos dapat terdaftar pada satu user. Kolom `user_id` pada tabel `photos` merupakan foreign key yang mengacu pada primary key `id` pada tabel `users`.

> **Note**: aplikasi ini menggunakan GORM untuk management data repository ke database postgresql

### Middleware

Pada sebagian endpoint kamu perlu menaruh middleware untuk memastikan bahwa endpoint tersebut hanya bisa diakses oleh _user_ yang sudah _Sign In_.

### Endpoint Specifications

Berikut adalah spesifikasi dari masing-masing endpoint yang harus Anda buat:

#### 1. **Register User**

- **Method**: `POST`
- **Endpoint**: `/users`
- **Description**: Membuat pengguna baru.
- **Request Body**:
  ```json
  {
    "username": "djarotpurnomo",
    "password": "admin#1234",
    "fullname": "Djarot Purnomo",
    "desc": "Admin User",
    "profilePic": "https://images.unsplash.com/photo-1544642058-1f01423e7a16"
  }
  ```
- **Response**:
  - Success: Status `201 Created` dengan data pengguna yang baru dibuat.
  - Failure: Status `400 Bad Request` jika data tidak valid.
  - Failure: Status `409 Conflict` jika username sudah digunakan oleh pengguna lain.

#### 2. **Sign In User**

- **Method**: `POST`
- **Endpoint**: `/signin`
- **Description**: Autentikasi pengguna dan menyimpan informasi pengguna yang berhasil login di memori.
- **Request Body**:
  ```json
  {
    "username": "djarotpurnomo",
    "password": "admin#1234"
  }
  ```
- **Response**:
  - Success: Status `200 OK` dengan data pengguna yang berhasil login. Informasi pengguna yang login akan disimpan di in-memory.
  - Failure: Status `401 Unauthorized` jika username atau password salah.

#### 3. **Get All Photos by User**

- **Method**: `GET`
- **Endpoint**: `/photos`
- **Description**: Mendapatkan semua foto yang dimiliki oleh pengguna yang sedang login.
- **Response**:
  - Success: Status `200 OK` dengan daftar foto.
  - Failure: Status `401 Unauthorized` jika pengguna tidak terautentikasi.

#### 4. **Get Photo Detail by ID**

- **Method**: `GET`
- **Endpoint**: `/photos/:id`
- **Description**: Mendapatkan detail foto berdasarkan ID.
- **Response**:
  - Success: Status `200 OK` dengan detail foto.
  - Failure: Status `404 Not Found` jika foto tidak ditemukan.
  - Failure: Status `401 Unauthorized` jika pengguna tidak terautentikasi.

#### 5. **Create New Photo**

- **Method**: `POST`
- **Endpoint**: `/photos`
- **Description**: Menambahkan foto baru untuk pengguna yang sedang login.
- **Request Body**:
  ```json
  {
    "url": "https://images.unsplash.com/photo-1544642058-1f01423e7a16",
    "caption": "A beautiful sunset"
  }
  ```
- **Response**:
  - Success: Status `201 Created` dengan data foto yang baru ditambahkan.
  - Failure: Status `401 Unauthorized` jika pengguna tidak terautentikasi.

#### 6. **Delete Photo**

- **Method**: `DELETE`
- **Endpoint**: `/photos/:id`
- **Description**: Menghapus foto berdasarkan ID.
- **Response**:
  - Success: Status `200 OK` jika foto berhasil dihapus.
  - Failure: Status `404 Not Found` jika foto tidak ditemukan.
  - Failure: Status `401 Unauthorized` jika pengguna tidak terautentikasi.

#### 7. **Health Check**

- **Method**: `GET`
- **Endpoint**: `/health`
- **Description**: Menampilkan status kesehatan server.
- **Response**: Status `200 OK` dengan pesan `OK`.

#### 8. **Page Not Found Handler**

- **Method**: `ANY`
- **Endpoint**: `*`
- **Description**: Handler untuk menangani route yang tidak ditemukan.
- **Response**: Status `404 Not Found` dengan pesan error.

## In-Memory Session Management

Untuk menyimpan informasi pengguna yang telah login, kamu bisa mempelajarinya di `repository/authRepository`. Untuk service ini kita hanya menyimpan satu user yang bisa login di API kita.

### **Perhatian**

Sebelum kalian menjalankan `grader-cli test`, pastikan kalian sudah mengubah database credentials pada file **`main.go`** (line 208) dan **`main_test.go`** (line 29) sesuai dengan database kalian. Kalian cukup mengubah nilai dari `"username"`, `"password"` dan `"database_name"`saja.

Contoh:

```go
dbCredentials = Credential{
    Host:         "localhost",
    Username:     "postgres", // <- ubah ini
    Password:     "postgres", // <- ubah ini
    DatabaseName: "kampusmerdeka", // <- ubah ini
    Port:         5432,
}
```

### Test Case Examples

#### Test Case 1

**Input**:

```http
GET /health
```

**Expected Output / Behavior**:

```http
HTTP status code: 200 OK
Response body: "OK"
```

**Explanation**:

```txt
Ketika melakukan request GET /health, server akan merespons dengan kode status HTTP 200 OK dan body "OK", yang menunjukkan bahwa server berjalan dengan normal.
```

#### Test Case 2

**Input**:

```http
POST /users
Content-Type: application/json
{
    "username": "djarotpurnomo",
    "password": "admin#1234",
    "fullname": "Djarot Purnomo",
    "desc": "Admin User",
    "profilePic": "https://images.unsplash.com/photo-1544642058-1f01423e7a16"
}
```

**Expected Output / Behavior**:

```http
HTTP status code: 201 Created
Response body: {
  "status": "Created",
  "User": {
    "username": "djarotpurnomo",
    "password": "admin#1234",
    "fullname": "Djarot Purnomo",
    "desc": "Admin User",
    "profile_pic": "https://images.unsplash.com/photo-1544642058-1f01423e7a16"
  }
}
```

**Explanation**:

```txt
Ketika melakukan request POST /users dengan data pengguna baru yang valid, server akan mendaftarkan pengguna baru dan merespons dengan kode status HTTP 201 Created serta informasi pengguna yang baru terdaftar.
```

#### Test Case 3

**Input**:

```http
POST /signin
Content-Type: application/json

{
  "username": "djarotpurnomo",
  "password": "admin#1234"
}
```

**Expected Output / Behavior**:

```http
HTTP status code: 200 OK
Response body: {
  "status": "Logged in",
  "User": {
    "username": "djarotpurnomo",
    "fullname": "Djarot Purnomo",
    "desc": "Admin User",
    "profile_pic": "https://images.unsplash.com/photo-1544642058-1f01423e7a16"
  }
}
```

**Explanation**:

```txt
Ketika melakukan request POST /signin dengan kredensial yang valid, server akan melakukan proses login dan merespons dengan kode status HTTP 200 OK serta informasi pengguna yang berhasil login.
```

#### Test Case 4

**Input**:

```http
GET /photos
```

**Expected Output / Behavior**:

```http
HTTP status code: 401 Unauthorized
Response body: {
  "error": "Unauthorized"
}
```

**Explanation**:

```txt
Ketika melakukan request GET /photos tanpa login, server akan merespons dengan kode status HTTP 401 Unauthorized dan pesan kesalahan "Unauthorized".
```

#### Test Case 5

**Input**:

```http
POST /photos
Content-Type: application/json

{
  "url": "https://images.unsplash.com/photo-1544642058-1f01423e7a16",
  "caption": "A beautiful sunset"
}
```

**Expected Output / Behavior**:

```http
HTTP status code: 201 Created
Response body: {
  "photo_id": 1
}
```

**Explanation**:

```txt
Ketika melakukan request POST /photos dengan data foto yang valid setelah login, server akan menambahkan foto baru yang terkait dengan pengguna yang login dan merespons dengan kode status HTTP 201 Created serta ID dari foto yang baru ditambahkan.
```

#### Test Case 6

**Input**:

```http
GET /photos/1
```

**Expected Output / Behavior**:

```http
HTTP status code: 200 OK
Response body: {
  "id": 1,
  "user_id": 1,
  "url": "https://images.unsplash.com/photo-1544642058-1f01423e7a16",
  "caption": "A beautiful sunset",
  "created_at": "2023-04-05T12:00:00Z",
  "updated_at": "2023-04-05T12:00:00Z"
}
```

**Explanation**:

```txt
Ketika melakukan request GET /photos/1 setelah login, server akan merespons dengan kode status HTTP 200 OK serta informasi detail foto dengan ID 1 yang ditemukan dalam database.
```

#### Test Case 7

**Input**:

```http
DELETE /photos/1
```

**Expected Output / Behavior**:

```http
HTTP status code: 200 OK
Response body: {
  "status": "Photo deleted"
}
```

**Explanation**:

```txt
Ketika melakukan request DELETE /photos/1 setelah login, server akan menghapus foto dengan ID 1 dan merespons dengan kode status HTTP 200 OK serta pesan konfirmasi bahwa foto telah dihapus.
```
