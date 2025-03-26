package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Todo represents a todo item
type Todo struct {
	ID        int    `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

// Template renderer
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.Static("/static", "static")

	// Templates
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = t

	// In-memory todo storage
	var (
		todos = []Todo{
			{ID: 1, Task: "Learn HTMX", Completed: false},
			{ID: 2, Task: "Build with Echo", Completed: false},
			{ID: 3, Task: "Style with Tailwind", Completed: false},
		}
		mu       sync.Mutex
		nextID   = 4
	)

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"todos": todos,
		})
	})

	// Get all todos (for htmx partial refresh)
	e.GET("/todos", func(c echo.Context) error {
		mu.Lock()
		defer mu.Unlock()
		return c.Render(http.StatusOK, "todos.html", map[string]interface{}{
			"todos": todos,
		})
	})

	// Add new todo
	e.POST("/todos", func(c echo.Context) error {
		task := c.FormValue("task")
		if task == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		mu.Lock()
		todos = append(todos, Todo{ID: nextID, Task: task, Completed: false})
		nextID++
		mu.Unlock()

		return c.Render(http.StatusOK, "todos.html", map[string]interface{}{
			"todos": todos,
		})
	})

	// Toggle todo completion
	e.PUT("/todos/:id/toggle", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		mu.Lock()
		for i := range todos {
			if todos[i].ID == id {
				todos[i].Completed = !todos[i].Completed
				break
			}
		}
		mu.Unlock()

		return c.Render(http.StatusOK, "todos.html", map[string]interface{}{
			"todos": todos,
		})
	})

	// Add new todo (using PUT instead of POST)
	e.PUT("/todos", func(c echo.Context) error {
		task := c.FormValue("task")
		if task == "" {
			return c.NoContent(http.StatusBadRequest)
		}
	
		mu.Lock()
		todos = append(todos, Todo{ID: nextID, Task: task, Completed: false})
		nextID++
		mu.Unlock()
	
		return c.Render(http.StatusOK, "todos.html", map[string]interface{}{
			"todos": todos,
		})
	})
	
	// Delete todo
	e.DELETE("/todos/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		mu.Lock()
		for i := range todos {
			if todos[i].ID == id {
				todos = append(todos[:i], todos[i+1:]...)
				break
			}
		}
		mu.Unlock()

		return c.Render(http.StatusOK, "todos.html", map[string]interface{}{
			"todos": todos,
		})
	})

	// Start server
	e.Logger.Fatal(e.Start(":3002"))
}
