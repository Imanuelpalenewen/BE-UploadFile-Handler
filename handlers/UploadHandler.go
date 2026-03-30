package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UploadedFile struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_name"`
	SavedName    string `json:"saved_name"`
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	UploadedAt   string `json:"uploaded_at"`
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"code":    405,
			"message": "Method not allowed. Use POST",
		})
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "Invalid multipart form",
		})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "File field is required (form-data key: file)",
		})
		return
	}

	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "Failed to prepare upload directory",
		})
		return
	}

	safeOriginal := filepath.Base(fileHeader.Filename)
	safeOriginal = strings.ReplaceAll(safeOriginal, " ", "_")
	savedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeOriginal)
	savedPath := filepath.Join("uploads", savedName)

	dst, err := os.Create(savedPath)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "Failed to create destination file",
		})
		return
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "Failed to save uploaded file",
		})
		return
	}

	dsn := os.Getenv("DSN")
	dbMessage := "Skipped metadata insert because DSN is empty"
	if dsn != "" {
		err = insertUploadMetadata(dsn, safeOriginal, savedName, savedPath, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			fmt.Println(err)
			dbMessage = "File uploaded but failed to insert metadata to DB"
		} else {
			dbMessage = "Metadata inserted to DB"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":    200,
		"message": "File uploaded",
		"db":      dbMessage,
		"data": map[string]any{
			"original_name": safeOriginal,
			"saved_name":    savedName,
			"file_path":     savedPath,
			"file_size":     fileHeader.Size,
			"mime_type":     fileHeader.Header.Get("Content-Type"),
		},
	})
}

func GetUploadedFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"code":    405,
			"message": "Method not allowed. Use GET",
		})
		return
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "DSN is empty. Configure database connection first",
		})
		return
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "Failed to connect database",
		})
		return
	}

	rows, err := db.Query(`SELECT id, original_name, saved_name, file_path, file_size, mime_type, uploaded_at FROM uploaded_files ORDER BY id DESC`)
	if err != nil {
		fmt.Println(err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "Failed to get uploaded files",
		})
		return
	}

	files := []UploadedFile{}
	for rows.Next() {
		var item UploadedFile
		err = rows.Scan(&item.ID, &item.OriginalName, &item.SavedName, &item.FilePath, &item.FileSize, &item.MimeType, &item.UploadedAt)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"code":    500,
				"message": "Failed to read uploaded file rows",
			})
			return
		}
		files = append(files, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":    200,
		"message": "Uploaded files fetched",
		"data":    files,
	})
}

func insertUploadMetadata(dsn, originalName, savedName, filePath string, fileSize int64, mimeType string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`INSERT INTO uploaded_files (original_name, saved_name, file_path, file_size, mime_type) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(originalName, savedName, filePath, fileSize, mimeType)
	if err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}
