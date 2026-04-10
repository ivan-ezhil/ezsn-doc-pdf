package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type RequestData struct {
	Template string                 `json:"template"`
	Pages    []string               `json:"pages"`
	Data     map[string]interface{} `json:"data"`
}

// 🔹 Resolve template paths
func getTemplatePaths(templateName string) (string, string, string) {
	base := "templates"

	body := filepath.Join(base, templateName+".html")
	header := filepath.Join(base, templateName+"_header.html")
	footer := filepath.Join(base, templateName+"_footer.html")

	return body, header, footer
}

func renderTemplate(path string, data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("template error: %w", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("render error: %w", err)
	}

	return buf.String(), nil
}

// 🔹 Write temp file (for header/footer)
func writeTempFile(content, name string) (string, error) {
	filePath := filepath.Join(os.TempDir(), name+".html")
	err := os.WriteFile(filePath, []byte(content), 0644)
	return filePath, err
}

// 🔹 Check file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 🔹 Build final HTML with dynamic pages
func buildFinalHTML(templateName string, pages []string, data map[string]interface{}) (string, error) {

	// main body
	bodyPath := filepath.Join("templates", templateName+".html")

	bodyHTML, err := renderTemplate(bodyPath, data)
	if err != nil {
		return "", err
	}

	finalHTML := bodyHTML

	// dynamic pages
	for _, page := range pages {

		path := filepath.Join("templates", templateName+"_"+page+".html")

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		extraHTML, err := renderTemplate(path, data)
		if err != nil {
			continue
		}

		finalHTML += `<div style="margin-top:40px;"></div>` + extraHTML
		// finalHTML += `<div style="page-break-before: always;"></div>` + extraHTML
	}

	return finalHTML, nil
}

// HTTP Handler
func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestData

	// Decode request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Template == "" {
		http.Error(w, "template is required", 400)
		return
	}

	// // Validate minimal fields
	// if req.Title == "" || req.Name == "" {
	// 	http.Error(w, "Missing required fields", http.StatusBadRequest)
	// 	return
	// }

	// 🔹 Get template paths
	bodyPath, headerPath, footerPath := getTemplatePaths(req.Template)

	// 🔹 Validate templates
	if !fileExists(bodyPath) {
		http.Error(w, "body template not found", 400)
		return
	}

	// 🔹 Build full HTML (body + dynamic pages)
	finalHTML, err := buildFinalHTML(req.Template, req.Pages, req.Data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var headerHTML, footerHTML string
	hasHeader := false
	hasFooter := false

	if _, err := os.Stat(headerPath); err == nil {
		headerHTML, _ = renderTemplate(headerPath, req.Data)
		hasHeader = true
	}

	if _, err := os.Stat(footerPath); err == nil {
		footerHTML, _ = renderTemplate(footerPath, req.Data)
		hasFooter = true
	}

	// 🔹 Create PDF generator
	pdfg, err := wkhtml.NewPDFGenerator()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	page := wkhtml.NewPageReader(strings.NewReader(finalHTML))

	// 🔥 Enable background colors
	page.PrintMediaType.Set(true)

	// 🔹 Temp files for header/footer
	var headerFile, footerFile string

	uniqueName := fmt.Sprintf("%s_%d", req.Template, time.Now().UnixNano())

	if hasHeader {
		headerFile, _ = writeTempFile(headerHTML, uniqueName+"_header")
		page.HeaderHTML.Set(headerFile)
		// page.HeaderSpacing.Set(5)
		defer os.Remove(headerFile)
	}

	if hasFooter {
		footerFile, _ = writeTempFile(footerHTML, uniqueName+"_footer")
		page.FooterHTML.Set(footerFile)
		// page.FooterSpacing.Set(5)
		page.FooterFontSize.Set(10)
		defer os.Remove(footerFile)
	}

	// 🔹 Margins (important)
	// pdfg.MarginTop.Set(20)
	// pdfg.MarginBottom.Set(20)

	// pdfg.MarginTop.Set(0)
	// pdfg.MarginBottom.Set(0)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)

	// pdfg.PageSize.Set(wkhtml.PageSizeA4)
	// pdfg.Dpi.Set(300)
	// pdfg.Zoom.Set(1.0)

	pdfg.AddPage(page)

	// 🔹 Generate PDF
	if err := pdfg.Create(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 🔹 Response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+req.Template+".pdf")

	w.Write(pdfg.Bytes())

	log.Printf("PDF generated in %s\n", time.Since(start))
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/generate-pdf", generatePDFHandler)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Println("🚀 Go PDF Service running on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
