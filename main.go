package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func uploadPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 30)

	repoName := r.FormValue("repo")
	repoPath := filepath.Join("uploads", repoName)

	os.MkdirAll(repoPath, os.ModePerm)

	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {

		file, _ := fileHeader.Open()
		defer file.Close()

		dstPath := filepath.Join(repoPath, fileHeader.Filename)

		dst, _ := os.Create(dstPath)
		defer dst.Close()

		io.Copy(dst, file)
	}

	http.Redirect(w, r, "/repos", http.StatusSeeOther)
}

func listRepos(w http.ResponseWriter, r *http.Request) {

	dirs, _ := os.ReadDir("uploads")

	var repos []string

	for _, dir := range dirs {
		if dir.IsDir() {
			repos = append(repos, dir.Name())
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/repos.html"))
	tmpl.Execute(w, repos)
}

func viewRepo(w http.ResponseWriter, r *http.Request) {

	repoName := r.URL.Query().Get("name")
	repoPath := filepath.Join("uploads", repoName)

	files, _ := os.ReadDir(repoPath)

	type File struct {
		Name string
		Path string
	}

	var fileList []File

	for _, f := range files {

		fileList = append(fileList, File{
			Name: f.Name(),
			Path: "/uploads/" + repoName + "/" + f.Name(),
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/files.html"))
	tmpl.Execute(w, fileList)
}

func main() {

	http.HandleFunc("/", uploadPage)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/repos", listRepos)
	http.HandleFunc("/repo", viewRepo)

	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}