package web

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

//go:embed docs/*.md
var docsFS embed.FS

func DocsPageWebHandler(w http.ResponseWriter, r *http.Request) {
	// Default to index.md if no specific page is requested
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "index"
	}

	// Read the markdown file
	md, err := fs.ReadFile(docsFS, filepath.Join("docs", page+".md"))
	if err != nil {
		// If the file doesn't exist, try with .md extension directly
		if _, err := fs.Stat(docsFS, filepath.Join("docs", page)); err == nil {
			md, err = fs.ReadFile(docsFS, filepath.Join("docs", page))
		}
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert(md, &buf); err != nil {
		http.Error(w, "Error rendering markdown", http.StatusInternalServerError)
		return
	}

	// Get all markdown files for the sidebar
	entries, _ := docsFS.ReadDir("docs")
	var pages []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			pages = append(pages, name)
		}
	}

	// Render the page
	component := DocsPage(page, buf.String(), pages)
	component.Render(r.Context(), w)
}
