package main

import (
  //_ "github.com/mattn/go-sqlite3"
  //"database/sql"
  "io/ioutil"
  "net/http"
  "html/template"
  "regexp"
  "errors"
  "fmt"
  //"os"
)
const lenPath = len("/view/")
const rootDir = "./"

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")
var jsFile = regexp.MustCompile("\\.js$")
var cssFile = regexp.MustCompile("\\.css$")

type Page struct {
  Title string
  Body []byte
}

var templates *template.Template


func parseTemplates(templateDir string) {
  templates = template.Must(template.ParseGlob(templateDir + "*html"))
}

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err error) {
  title = r.URL.Path[lenPath:]
  if !titleValidator.MatchString(title) {
    http.NotFound(w, r)
    err = errors.New("Invalid Page Title")
  }
  return
}

func (p *Page) save() error {
  const dataDir = "data/"
  filename := rootDir + dataDir + p.Title + ".txt"
  return ioutil.WriteFile(filename,p.Body,0600)
}

func loadPage(title string) (*Page, error) {
  const dataDir = "data/"
  filename := rootDir + dataDir + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if (err != nil) {
    return nil, err
  }
  return &Page{Title:title,Body:body},nil
}

func loadHtml(title string) (*Page, error) {
  const htmlDir = "html/"
  filename := rootDir + htmlDir + title + ".html"
  body, err := ioutil.ReadFile(filename)
  if (err != nil) {
    return nil, err
  }
  return &Page{Title:title,Body:body},nil
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, "renderTemplate:" + err.Error(), http.StatusInternalServerError)
    return
  }
}

func loadPlainText(fullPath string) ([]byte, error) {
  text , err := ioutil.ReadFile(rootDir+fullPath)
  if err != nil {
    fmt.Println(err.Error())
    return nil ,err
  }
  return text, nil
}


func makeHandler (fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Request for "+ r.URL.Path)
    title := r.URL.Path[lenPath:]
    if   !titleValidator.MatchString(title) {
      fmt.Println("Title %s not valid",title)
      http.NotFound(w, r)
      return
    }
  fn (w, r, title)
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}


func sandboxHandler(w http.ResponseWriter, r *http.Request, title string) {
  fmt.Println("Catched by sandboxHandler")
  p, err := loadHtml(title)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  w.Write(p.Body)
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/javascript")
  p, err := loadPlainText(r.URL.Path)
  if err != nil {
    http.NotFound(w, r)
    fmt.Println(err.Error())
    return
  }
  w.Write(p)
}


func cssHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/css")
  p, err := loadPlainText(r.URL.Path)
  if err != nil {
    http.NotFound(w, r)
    fmt.Println(err.Error())
    return
  }
  w.Write(p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string ) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler (w http.ResponseWriter, r *http.Request, title string) { 
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, "saveHandler:" + err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w,r,"/view/"+title,http.StatusFound)
}

/*
func initDb (dbName string) (*sql.DB, error) {
  os.Remove(rootDir + "data/"+ dbName + ".db")
  db, err := sql.Open("sqlite3", rootDir + "data/"+ dbName + ".db")
  if err != nil {
    fmt.Println(err)
    return nil,err
  }
  defer db.Close()
  sqls := []string{
    "create table foo (id integer not null primary key, name text)",
    "delete from foo",
  }

  for _, sql := range sqls {
    _, err = db.Exec(sql)
    if err != nil {
      fmt.Printf("%q: %s\n", err, sql)
      return nil,err
    }
  }
  return db,nil
}
*/


func main () {
  fmt.Println("Staring Webserver at 8080")
  parseTemplates(rootDir + "templates/")

  /*
  _ ,err := initDb("foo")
  if err != nil {
    fmt.Println(err)
    return
  }
  */
  http.HandleFunc("/sand/",makeHandler(sandboxHandler))
  http.HandleFunc("/css/",cssHandler)
  http.HandleFunc("/js/",jsHandler)

  //http.Handle("/js/",http.FileServer(http.Dir("./js/")))
  //http.Handle("/css/",http.FileServer(http.Dir("./css/")))
  http.Handle("/ico/", http.StripPrefix("/ico/", http.FileServer(http.Dir("./ico"))))
  http.ListenAndServe(":8080",nil)
}
