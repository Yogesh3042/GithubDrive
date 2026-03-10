package main

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	// "time"
)

type File struct {
	Name    string
	Path    string
	Size    string
	Time    string
	IsImage bool
	IsVideo bool
	Repo    string
}

func formatSize(size int64) string {
	return fmt.Sprintf("%.2f MB", float64(size)/1024.0/1024.0)
}

func uploadPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(20 << 30)

	repo := r.FormValue("repo")
	repoPath := filepath.Join("uploads", repo)

	os.MkdirAll(repoPath, os.ModePerm)

	files := r.MultipartForm.File["files"]

	for _, fh := range files {

		file, _ := fh.Open()
		defer file.Close()

		dstPath := filepath.Join(repoPath, fh.Filename)

		dst, _ := os.Create(dstPath)
		defer dst.Close()

		io.Copy(dst, file)
	}

	http.Redirect(w, r, "/repo?name="+repo, http.StatusSeeOther)
}

func listRepos(w http.ResponseWriter, r *http.Request) {

	dirs, _ := os.ReadDir("uploads")

	var repos []string

	for _, d := range dirs {
		if d.IsDir() {
			repos = append(repos, d.Name())
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/repos.html"))
	tmpl.Execute(w, repos)
}

func viewRepo(w http.ResponseWriter, r *http.Request) {

	repo := r.URL.Query().Get("name")
	search := strings.ToLower(r.URL.Query().Get("search"))

	repoPath := filepath.Join("uploads", repo)

	files, _ := os.ReadDir(repoPath)

	var fileList []File

	for _, f := range files {

		if search != "" && !strings.Contains(strings.ToLower(f.Name()), search) {
			continue
		}

		info, _ := f.Info()

		ext := strings.ToLower(filepath.Ext(f.Name()))

		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
		isVideo := ext == ".mp4" || ext == ".webm" || ext == ".mov"

		fileList = append(fileList, File{
			Name:    f.Name(),
			Path:    "/uploads/" + repo + "/" + f.Name(),
			Size:    formatSize(info.Size()),
			Time:    info.ModTime().Format("2006-01-02 15:04"),
			IsImage: isImage,
			IsVideo: isVideo,
			Repo:    repo,
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/files.html"))
	tmpl.Execute(w, struct {
		Files []File
		Repo  string
	}{
		fileList,
		repo,
	})
}

func deleteFile(w http.ResponseWriter, r *http.Request) {

	repo := r.URL.Query().Get("repo")
	file := r.URL.Query().Get("file")

	path := filepath.Join("uploads", repo, file)

	os.Remove(path)

	http.Redirect(w, r, "/repo?name="+repo, http.StatusSeeOther)
}

func zipRepo(w http.ResponseWriter, r *http.Request) {

	repo := r.URL.Query().Get("name")
	repoPath := filepath.Join("uploads", repo)

	w.Header().Set("Content-Disposition", "attachment; filename="+repo+".zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		file, _ := os.Open(path)
		defer file.Close()

		f, _ := zipWriter.Create(info.Name())

		io.Copy(f, file)

		return nil
	})
}

func main() {

	http.HandleFunc("/", uploadPage)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/repos", listRepos)
	http.HandleFunc("/repo", viewRepo)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/download", zipRepo)

	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	fmt.Println("Server running :8080")
	http.ListenAndServe(":8080", nil)
}