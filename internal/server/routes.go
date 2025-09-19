package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"illuminate/cmd/web"

	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuin/goldmark"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	// Documentation routes
	// e.GET("/docs", echo.WrapHandler(http.HandlerFunc(web.DocsPageWebHandler)))
	// e.GET("/docs/:page", echo.WrapHandler(http.HandlerFunc(web.DocsPageWebHandler)))

	// Converter Pages
	e.GET("/converter", echo.WrapHandler(templ.Handler(web.ConverterPage())))

	e.GET("/", s.HelloWorldHandler)

	e.GET("/health", s.healthHandler)

	e.GET("/websocket", s.websocketHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) websocketHandler(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	socket, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Errorf("could not open websocket: %v", err)
		_, _ = w.Write([]byte("could not open websocket"))
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	defer socket.Close(websocket.StatusGoingAway, "server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		err := socket.Write(socketCtx, websocket.MessageText, []byte(payload))
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	return nil
}

func (s *Server) docsIndexHandler(c echo.Context) error {
	files, err := os.ReadDir("cmd/web/docs")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not read documentation files")
	}

	var pages []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			pages = append(pages, strings.TrimSuffix(file.Name(), ".md"))
		}
	}

	// Redirect to the index page
	return c.Redirect(http.StatusMovedPermanently, "/docs/index")
}

func (s *Server) docsPageHandler(c echo.Context) error {
	pageName := c.Param("page")
	if pageName == "" {
		pageName = "index"
	}

	filePath := filepath.Join("cmd/web/docs", pageName+".md")

	mdContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.String(http.StatusInternalServerError, "Could not read page")
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(mdContent, &buf); err != nil {
		return c.String(http.StatusInternalServerError, "Could not convert markdown")
	}

	// Get list of all documentation pages for the sidebar
	files, err := os.ReadDir("cmd/web/docs")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not read documentation files")
	}

	var pages []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			pages = append(pages, strings.TrimSuffix(file.Name(), ".md"))
		}
	}

	// Render the page using the template
	component := web.DocsPage(pageName, buf.String(), pages)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return err
	})
}
