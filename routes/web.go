package routes

import (
	"UploadFile/handlers"
	"UploadFile/middlewares"
	"net/http"
)

func RegisterRoutes() {
	http.HandleFunc("/upload_file", middlewares.CheckAPI(handlers.UploadFile))
	http.HandleFunc("/my_files", middlewares.CheckAPI(handlers.GetUploadedFiles))

	http.ListenAndServe(":8080", nil)
}
