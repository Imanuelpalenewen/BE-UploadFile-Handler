# UploadFile API (Go)

Project ini mengikuti pola pembelajaran di kelas: main -> routes -> handlers dengan middleware sebagai jembatan untuk memvalidasi API key.

## Struktur Folder

```
UploadFile/
├── handlers/
│   └── UploadHandler.go
├── middlewares/
│   └── CheckAPI.go
├── routes/
│   └── web.go
├── uploads/
│   └── .gitkeep
├── .env
├── .gitignore
├── go.mod
├── main.go
└── uploadfile.sql
```

## Setup

1. Jalankan database yang ada di phpMyAdmin
    
2. Konfigurasi environment di `.env`:

```env
API_KEY=isi sendiri
DSN=root:@tcp(127.0.0.1:3306)/upload_file_db?parseTime=true
```

3. Jalankan project:

```bash
go mod tidy
go run main.go
```

Server aktif di: `http://localhost:8080`

## Endpoint

### 1) Upload File
- Method: `POST`
- URL: `/upload_file`
- Headers:
  - `X-API-Key: <API_KEY>`
  - `Content-Type: multipart/form-data`
- Body (form-data):
  - key: `file` (type: File)

Response sukses berisi metadata file dan status insert metadata ke DB.

### 2) Ambil List File (dari database)
- Method: `GET`
- URL: `/my_files`
- Headers:
  - `X-API-Key: <API_KEY>`
