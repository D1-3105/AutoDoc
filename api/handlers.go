package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BaseCDNUrl is the base CDN URL used to construct export links.
var (
	BaseCDNUrl = os.Getenv("CDN_URL")
)

// FullOpenAPI represents the OpenAPI schema payload submitted to the exporter.
// @description Full OpenAPI schema body used in export requests.
type FullOpenAPI struct {
	Openapi string `json:"openapi" example:"3.0.0"`
	Info    struct {
		Title          string `json:"title" example:"example-service"`
		Version        string `json:"version" example:"1.0.0"`
		Description    string `json:"description,omitempty"`
		TermsOfService string `json:"termsOfService,omitempty"`
	} `json:"info"`
	Paths      map[string]interface{}   `json:"paths,omitempty"`
	Components map[string]interface{}   `json:"components,omitempty"`
	Servers    []map[string]interface{} `json:"servers,omitempty"`
	Tags       []map[string]interface{} `json:"tags,omitempty"`
	Security   []map[string]interface{} `json:"security,omitempty"`
}

// ErrorResponse represents a generic error response.
// @description Returned when an error occurs during request processing.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid JSON body"`
}

// ExportSuccess represents a successful export response.
// @description Returned when the OpenAPI schema has been successfully exported.
type ExportSuccess struct {
	SwaggerUrl string `json:"url" example:"https://cdn.example.com/example-service/index.html"`
	RedocUrl   string `json:"redocUrl" example:"https://cdn.example.com/example-service/redoc.html"`
}

// returnError sends an error response to the client.
func returnError(w http.ResponseWriter, err error) {
	slog.Error("Error handling request: %v", err)
	w.WriteHeader(http.StatusBadRequest)
	newError := ErrorResponse{Error: err.Error()}
	_ = json.NewEncoder(w).Encode(newError)
}

// openapiExport handles OpenAPI export requests.
//
// @Summary Export OpenAPI schema to Redoc
// @Description Accepts a full OpenAPI JSON schema, builds static Redoc documentation, and returns a CDN URL.
// @Tags openapi
// @Accept application/json
// @Produce application/json
// @Param schema body FullOpenAPI true "Full OpenAPI Schema"
// @Success 200 {object} ExportSuccess
// @Failure 400 {object} ErrorResponse
// @Router /openapi-export [post]
func openapiExport(w http.ResponseWriter, r *http.Request) {
	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading body: %v", err)
		returnError(w, err)
		return
	}

	fullSchema := FullOpenAPI{}
	if err := json.NewDecoder(bytes.NewReader(jsonData)).Decode(&fullSchema); err != nil {
		slog.Error("Error decoding JSON: %v", err)
		returnError(w, err)
		return
	}

	slog.Info("Accepted a new schema: %s", fullSchema.Info.Title)
	fullPth := "./schemas/" + fullSchema.Info.Title + ".json"
	if err = os.MkdirAll(filepath.Dir(fullPth), 0777); err != nil {
		slog.Error("Error creating directory %s: %v", filepath.Dir(fullPth), err)
		returnError(w, err)
		return
	}

	err = os.WriteFile(fullPth, jsonData, 0644)
	if err != nil {
		slog.Error("Error writing file: %v", err)
		returnError(w, err)
		return
	}

	slog.Info("Wrote file: %s", fullSchema.Info.Title+".json")
	slog.Info("File path: %s", fullPth)

	redocShortPath := fmt.Sprintf("./exported/%s", fullSchema.Info.Title)
	fullPth, _ = filepath.Abs(fullPth)
	redocPath, _ := filepath.Abs(redocShortPath)
	makeUI := func(cmdCommand []string, cmdDir string) error {
		slog.Info("Running command: %v", cmdCommand)
		_ = os.MkdirAll(redocPath, 0755)
		var stderr bytes.Buffer
		cmd := exec.Command(cmdCommand[0], cmdCommand[1:]...)
		cmd.Dir = cmdDir
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			err := fmt.Errorf("error running command: %w, stderr: %s", err, stderr.String())
			return err
		}
		return nil
	}
	err = makeUI([]string{"/bin/bash", "html-swagger.sh", fullPth, filepath.Join(redocPath, "swagger.html")}, "")
	if err != nil {
		returnError(w, err)
		return
	}
	err = makeUI(
		[]string{
			"npx", "--no-install", "@redocly/cli", "build-docs", fullPth, "--output=" + filepath.Join(
				redocPath,
				"redoc.html",
			),
		},
		"./node_js",
	)
	if err != nil {
		returnError(w, err)
		return
	}
	success := ExportSuccess{
		SwaggerUrl: fmt.Sprintf("%s%s/swagger.html", BaseCDNUrl, fullSchema.Info.Title),
		RedocUrl:   fmt.Sprintf("%s%s/redoc.html", BaseCDNUrl, fullSchema.Info.Title),
	}
	_ = json.NewEncoder(w).Encode(success)

	slog.Info("Exported: %s", fullSchema.Info.Title)
	slog.Info("URL: %s", success.SwaggerUrl)
}

// expandedOpenapi handles OpenAPI export requests.
//
// @Summary Export OpenAPI schema to Redoc
// @Description Returns the expanded openapi schema
// @Tags openapi
// @Param        name   path      string  true  "JSON name"
// @Accept application/json
// @Produce application/json
// @Success 200 {object} FullOpenAPI
// @Failure 400 {object} ErrorResponse
// @Router /expand/{name} [get]
func expandedOpenapi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemaName := vars["name"]
	slog.Info("Received request for schema: %s", schemaName)
	schemaPath, err := filepath.Abs(
		fmt.Sprintf("./schemas/%s.json", strings.TrimSuffix(schemaName, ".json")),
	)
	if err != nil {
		returnError(w, err)
		return
	}
	cmd := exec.CommandContext(context.Background(), "node", "node_js/deref.js", schemaPath)
	cmd.Dir = ""
	output, err := cmd.CombinedOutput()
	if err != nil {
		returnError(w, fmt.Errorf("error running command: %w; output: %s", err, output))
		return
	}
	if !json.Valid(output) {
		slog.Error("Failed to parse JSON: %s", output)
		returnError(w, fmt.Errorf("failed to parse JSON: %s", output))
		return
	}
	_, _ = w.Write(output)
}

type AllResponse struct {
	AllFiles []string `json:"allFiles"`
}

// listAll handles a GET request to list all `index.html` files in the "exported" directory
// and returns their full CDN URLs.
//
// @Summary      Get all index.html files in the "exported" directory
// @Description  Recursively scans the ./exported directory and returns a list of CDN URLs pointing to each `index.
// html` found in subdirectories.
//
// @Tags         Files
// @Produce      application/json
// @Success      200  {object}  AllResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /all [get]
func listAll(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir("./exported")
	if err != nil {
		returnError(w, err)
		return
	}
	allFiles := AllResponse{AllFiles: []string{}}
	for _, f := range files {
		allFiles.AllFiles = append(allFiles.AllFiles, fmt.Sprintf("%s%s/swagger.html", BaseCDNUrl, f.Name()))
	}
	parser := json.NewEncoder(w)
	_ = parser.Encode(allFiles)
	return
}

// Router returns a new Gorilla mux router with the OpenAPI export endpoint registered.
func Router() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/openapi-export", openapiExport).Methods(http.MethodPost)
	r.HandleFunc("/all", listAll).Methods(http.MethodGet)
	r.HandleFunc("/expand/{name}", expandedOpenapi).Methods(http.MethodGet)
	return r
}
