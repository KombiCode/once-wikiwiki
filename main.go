package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var wikiLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

const pageTmpl = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 2em auto; padding: 0 1em; }
        h1 { border-bottom: 2px solid #333; }
        a { color: #0366d6; }
        .missing { color: #d73a49; }
        textarea { width: 100%; height: 400px; }
        input[type="text"] { width: 100%; font-size: 1.2em; }
        .actions { margin-top: 1em; }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <div>{{.Body}}</div>
    <div class="actions">
        <a href="/edit/{{.Title}}">Edit</a> |
        <a href="/">Home</a>
    </div>
</body>
</html>`

const editTmpl = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Edit {{.Title}}</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 2em auto; padding: 0 1em; }
        textarea { width: 100%; height: 400px; }
        input[type="text"] { width: 100%; font-size: 1.2em; }
    </style>
</head>
<body>
    <h1>Edit {{.Title}}</h1>
    <form action="/save/{{.Title}}" method="POST">
        <div><input type="text" name="title" value="{{.Title}}"></div>
        <div><textarea name="body">{{.RawBody}}</textarea></div>
        <div><input type="submit" value="Save"></div>
    </form>
</body>
</html>`

const homeTmpl = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>WikiWiki</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 2em auto; padding: 0 1em; }
    </style>
</head>
<body>
    <h1>WikiWiki</h1>
    <p>Welcome! Go to <a href="/view/Home">Home</a> or create a new page.</p>
</body>
</html>`

type Page struct {
	Title   string
	Body    template.HTML
	RawBody string
}

func initDB() error {
	storagePath := "/storage"
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return err
	}

	var err error
	db, err = sql.Open("sqlite3", storagePath+"/wiki.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pages (
		title TEXT PRIMARY KEY,
		body TEXT
	)`)
	return err
}

func loadPage(title string) (*Page, error) {
	var body string
	err := db.QueryRow("SELECT body FROM pages WHERE title = ?", title).Scan(&body)
	if err == sql.ErrNoRows {
		return &Page{Title: title, RawBody: ""}, nil
	}
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, RawBody: body}, nil
}

func savePage(title, body string) error {
	_, err := db.Exec("INSERT INTO pages (title, body) VALUES (?, ?) ON CONFLICT(title) DO UPDATE SET body=excluded.body", title, body)
	return err
}

func renderLinks(body string) template.HTML {
	body = template.HTMLEscapeString(body)
	body = wikiLinkRe.ReplaceAllStringFunc(body, func(match string) string {
		inner := match[2 : len(match)-2]
		linkText := strings.TrimSpace(inner)
		if linkText == "" {
			return match
		}
		return fmt.Sprintf(`<a href="/view/%s">%s</a>`, linkText, linkText)
	})
	body = strings.ReplaceAll(body, "\n", "<br>\n")
	return template.HTML(body)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/view/")
	if title == "" {
		http.Redirect(w, r, "/view/Home", http.StatusFound)
		return
	}
	page, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page.Body = renderLinks(page.RawBody)
	t, _ := template.New("page").Parse(pageTmpl)
	t.Execute(w, page)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/edit/")
	page, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t, _ := template.New("edit").Parse(editTmpl)
	t.Execute(w, page)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")
	if err := savePage(title, body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func upHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(homeTmpl))
}

func main() {
	if err := initDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	http.HandleFunc("/up", upHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/", homeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Listening on :" + port)
	http.ListenAndServe(":"+port, nil)
}
