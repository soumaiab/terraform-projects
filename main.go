package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

var page = template.Must(template.New("index").Parse(`
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Hello Go</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>body{font-family:system-ui,Segoe UI,Roboto,Arial,sans-serif;margin:2rem} code{background:#f6f8fa;padding:.2rem .4rem;border-radius:4px}</style>
</head>
<body>
  <h1>Hello, world 👋</h1>
  <p>Message from <code>APP_MESSAGE</code> (set in Terraform):</p>
  <h2>{{.Message}}</h2>
</body>
</html>
`))

func handler(w http.ResponseWriter, r *http.Request) {
	msg := os.Getenv("APP_MESSAGE")
	if msg == "" {
		msg = "(APP_MESSAGE not set)"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = page.Execute(w, struct{ Message string }{Message: msg})
}

func main() {
	http.HandleFunc("/", handler)
	addr := ":8888"
	fmt.Println("Listening on", addr)
	_ = http.ListenAndServe(addr, nil)
}
