package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const filePath = "todo.html"
const placeholder = "<!-- Existing todos go here -->"

func readFile() (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

func getNewID(content string) (int, error) {
	r := regexp.MustCompile(`id="item-(\d+)"`)
	matches := r.FindAllStringSubmatch(content, -1)

	highestID := 0
	for _, match := range matches {
		if id, err := strconv.Atoi(match[1]); err == nil && id > highestID {
			highestID = id
		}
	}

	return highestID + 1, nil
}

func deleteTodoInHTMLFile(content string, re *regexp.Regexp) error {
	newContent := re.ReplaceAllString(content, "")
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

func appendTodoInHTMLFile(content, newLiContent string) (string, error) {
	const spacing = "\n\t\t"
	updatedContent := strings.Replace(content, placeholder, newLiContent+spacing+placeholder, 1)
	err := os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return "", err
	}
	return newLiContent, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	content, err := readFile()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {

	// GET - Show todos
	case http.MethodGet:
		http.ServeFile(w, r, "todo.html")

	// POST - Add todo
	case http.MethodPost:
		newTodo := r.FormValue("todo")
		newID, err := getNewID(content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		deleteButton := fmt.Sprintf(`<button hx-delete="/todos?id=item-%d" hx-swap="outerHTML" hx-target="closest li">grug did it</button>`, newID)
		todoHeadline := fmt.Sprintf(`<span">%s</span>`, newTodo)
		liContent := fmt.Sprintf(`<li id="item-%d">%s %s</li>`, newID, todoHeadline, deleteButton)

		newLi, err := appendTodoInHTMLFile(content, liContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write([]byte(newLi))

	// DELETE - Delete todo
	case http.MethodDelete:
		todoID := r.URL.Query().Get("id")
		if todoID == "" {
			http.Error(w, "ID missing", http.StatusBadRequest)
			return
		}

		// Create a regular expression pattern based on the ID
		pattern := fmt.Sprintf(`<li id="%s">.*?</li>\n\t\t`, todoID)
		re := regexp.MustCompile(pattern)

		err := deleteTodoInHTMLFile(content, re)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/todos", handler)
	http.ListenAndServe(":8080", nil)
}
