package types

type ErrorResponse struct {
	Error string `json:"error"`
}

type FileResponse struct {
	Name string `json:"name"`
	Data string `json:"data"` // base64 encoded file content
}

type FilesListResponse struct {
	Files []FileResponse `json:"files"`
}
