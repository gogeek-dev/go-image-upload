package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "Welcome123@$"
	dbName := "goblog"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Compile templates on start of the application
var tmp = template.Must(template.ParseFiles("form/index.html"))

// Display the named template
func index(w http.ResponseWriter, r *http.Request) {
	tmp.ExecuteTemplate(w, "index.html", nil)
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(200000)

	// ok, no problem so far, read the Form data
	formdata := r.MultipartForm

	//get the *fileheaders
	fil := formdata.File["files"] // grab the filenames
	for i := range fil {          // loop through the files one by one

		//file save to open
		file, err := fil[i].Open()

		// fmt.Println("File Name : ", fil[i].Filename)
		// // fmt.Println("File Name : ", fil[i].Filetype)
		// fmt.Println("File Size : ", fil[i].Size)
		// fmt.Println("File Type : ", fil[i].Header.Get("Content-type"))

		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// Create file
		dst, err := os.Create("images/" + fil[i].Filename)
		defer dst.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy the uploaded file to the created file on the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// if _, err := io.Copy(w, file); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		fname := fil[i].Filename
		fsize := fil[i].Size
		ftype := fil[i].Header.Get("Content-type")
		insForm, err := db.Prepare("INSERT INTO upload(fname, fsize ,ftype) VALUES(?,?,?)")
		if err != nil {
			panic(err.Error())
		} else {
			log.Println("data insert successfully . . .")
		}
		insForm.Exec(fname, fsize, ftype)
		log.Println("File name", fname)
		log.Println("File size", fsize)
		log.Println("File type", ftype)

		fmt.Fprintf(w, "Successfully Uploaded File\n")
		fmt.Fprintf(w, fil[i].Filename+"\n")

		defer file.Close()
		defer db.Close()
		http.Redirect(w, r, "view.html", 301)
		// f, err := os.Open(os.Getenv("images/") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
		// if err != nil {
		// 	log.Println("file does not access. . .")
		// 	log.Println(err)
		// 	w.WriteHeader(http.StatusNotFound)
		// 	return
		// }
		// defer f.Close()
		// io.Copy(w, f)

	}

}

// func showPicHandle(w http.ResponseWriter, req *http.Request) {
// 	file, err := os.Open("." + req.URL.Path)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	defer file.Close()
// 	buff, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	w.Write(buff)
// }

// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case "GET":
// 		display(w, "upload", nil)
// 	case "POST":
// 		uploadFiles(w, r)
// 	}
// }

func view(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM upload ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}

	for selDB.Next() {
		var id int
		var name, size, typ string
		err = selDB.Scan(&id, &name, &size, &typ)
		if err != nil {
			panic(err.Error())
		}
	}
	// 	emp.ID = id
	// 	emp.Name = name
	// 	emp.Size = size
	// 	emp.Type = typ
	// 	res = append(res, emp)
	// }
	// tmpl.ExecuteTemplate(w, "view.html", res)
	defer db.Close()

}

func main() {
	// Upload route
	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", uploadFiles)
	http.HandleFunc("/view", view)

	//Listen on port 8080
	http.ListenAndServe(":8080", nil)
}
