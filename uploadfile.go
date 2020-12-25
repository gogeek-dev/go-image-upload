package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func dbConn() (db *sql.DB) {

	er := godotenv.Load(".env")
	if er != nil {
		panic(er.Error())
	}
	dbDriver := os.Getenv("DB_Driver")
	dbUser := os.Getenv("DB_User")
	dbPass := os.Getenv("DB_Password")
	dbName := os.Getenv("DB_Name")
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type upfile struct {
	ID    int
	Fname string
	Fsize string
	Ftype string
	Path  string
	Count int
}

var tmpl = template.Must(template.ParseGlob("templates/*"))

func upload(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM upload ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	upld := upfile{}
	res := []upfile{}
	for selDB.Next() {
		var id int
		var fname, fsize, ftype, path string
		err = selDB.Scan(&id, &fname, &fsize, &ftype, &path)
		if err != nil {
			panic(err.Error())
		}
		upld.ID = id
		upld.Fname = fname
		upld.Fsize = fsize
		upld.Ftype = ftype
		upld.Path = path
		res = append(res, upld)

	}

	upld.Count = len(res)

	if upld.Count > 0 {
		tmpl.ExecuteTemplate(w, "uploadfile.html", res)
	} else {
		tmpl.ExecuteTemplate(w, "uploadfile.html", nil)
	}

	db.Close()

}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	// tmpl.ExecuteTemplate(w, "uploadfile.html", r)
	db := dbConn()
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(200000)
	if r == nil {
		fmt.Fprintf(w, "No files can be selected\n")
	}
	// ok, no problem so far, read the Form data
	formdata := r.MultipartForm

	//get the *fileheaders
	fil := formdata.File["files"] // grab the filenames

	for i := range fil { // loop through the files one by one

		//file save to open
		file, err := fil[i].Open()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer file.Close()

		fname := fil[i].Filename
		fsize := fil[i].Size
		kilobytes := fsize / 1024
		// megabytes := (float64)(kilobytes / 1024) // cast to type float64

		ftype := fil[i].Header.Get("Content-type")

		// Create file

		tempFile, err := ioutil.TempFile("assets/uploadimage", "upload-*.jpg")
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()
		filepath := tempFile.Name()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)

		// return that we have successfully uploaded our file!

		insForm, err := db.Prepare("INSERT INTO upload(fname, fsize, ftype, path) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		} else {
			log.Println("data insert successfully . . .")
		}
		insForm.Exec(fname, kilobytes, ftype, filepath)

		log.Printf("Successfully Uploaded File\n")
		defer db.Close()

		http.Redirect(w, r, "/", 301)
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	emp := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM upload WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(emp)
	log.Println("deleted successfully")
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func main() {
	log.Println("Server started on: http://localhost:9000")
	http.HandleFunc("/dele", delete)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/", upload)
	http.HandleFunc("/uploadfiles", uploadFiles)

	http.ListenAndServe(":9000", nil)

}
