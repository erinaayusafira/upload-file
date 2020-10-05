package main
//via ajax, banyak sekaligus, hemat memori dgn multipartreader
import (
	"fmt"
	"net/http"
	"html/template"
	"path/filepath" //menyimpan dan membuat file untuk data
	"io" // untuk io.reader - penyimpanan sementara
	"os"
	"encoding/json"
)
// map
type M map[string] interface{}

//pasrsing ke view.html
func handleIndex(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("fileupload.html"))
	if err := tmpl.Execute(w, nil);
	err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
		http.Error(w, "Only accept POST Request", http.StatusBadRequest)
		return
	}

	basePath, _ := os.Getwd() 
	reader, err := r.MultipartReader() //untuk request milik handler 
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		part, err := reader.NextPart() //mengembalikan 2 info 
		//io.reader, dari file yg diupload dan eror
		//jika menhembalikan eror io.eof maka semua file sudah di proses
		if err == io.EOF{
			break //menghentikan perulangan
		}
		//file destinasi disiapkan 
		fileLocation := filepath.Join(basePath, "files", part.FileName())
		dst, err := os.Create(fileLocation)
		if dst != nil{
			defer dst.Close()
		}
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); // mengisi data dari stream file
		err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte(`all files uploaded`))
}

func handleListFiles(w http.ResponseWriter, r *http.Request){
	files := []M{}
	basePath,_ := os.Getwd()
	filesLocation := filepath.Join(basePath, "files")

	err := filepath.Walk(filesLocation, func(path string, info os.FileInfo, err error) error{
		if err != nil{
			return err
		}

		if info.IsDir(){
			return nil
		}

		files = append(files, M{"filename": info.Name(), "path": path})
		return nil
	})
	
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

		res, err := json.Marshal(files)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    path := r.FormValue("path")
    f, err := os.Open(path)
    if f != nil {
        defer f.Close()
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    contentDisposition := fmt.Sprintf("attachment; filename=%s", f.Name())
    w.Header().Set("Content-Disposition", contentDisposition)

    if _, err := io.Copy(w, f); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}


func main(){
	
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/upload/list-files", handleListFiles)
	http.HandleFunc("/download", handleDownload)
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("assets"))))
	
	fmt.Println("server started at localhost:8080")
	http.ListenAndServe(":8080", nil)
}