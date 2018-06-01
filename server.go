package main

import (
	"net/http"
	"regexp"
	"fmt"
	"strings"
	/*
	"errors"
	"fmt"
	*/
	"encoding/gob"
    "bytes"
	"time"
	"html/template"
	"os"
  	"github.com/sirupsen/logrus"
  	"github.com/dgraph-io/badger"
)

//template for page
var templates = template.Must(template.ParseFiles("loginregister.html", "register.html", "home.html"))

/*
validPath - MustCompile will parse and compile the regexp
and return a regexp.
Regexep.MustCompile is distinct form Compile in that it will panic if the exp compiliation fails
*/
var validPath = regexp.MustCompile("^/(loginregister|home|register)") ///([a-zA-Z0-9]+)$
var validHomePath = regexp.MustCompile("^/(home)/([a-zA-Z0-9]+)")
type WebPage struct {
	Title string
	User User
}

type User struct {
	UserName string
	Password string
	Email string
}


// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()

/*
loadWebPagePage method --
loads the webpage
*/
func loadWebPage(user User) (*WebPage, error){

    if user.UserName == "" {
		webpage:= &WebPage{
	    Title: "TEST",
	    User: User{}}
	    return webpage,nil
    }else {
    	webpage:= &WebPage{
	    Title: user.UserName,
		User: user}
	    return webpage,nil
    }
    
}

/*
general renderTemplate func
http.Error sends specified internalservice error response code and err msg
*/
func renderTemplate(w http.ResponseWriter, tmpl string, p *WebPage){
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


/*
webPageHandler method --

*/
func loginPageHandler(w http.ResponseWriter, r *http.Request){
	webpage, err := loadWebPage(User{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout
	ip := r.RemoteAddr
	url := r.URL
	requestLogger := log.WithFields(logrus.Fields{"user_ip": ip, "path": url})
	requestLogger.Info(time.Now().Format("2006-01-02 15:04:05")+" - loginPageHandler")
	renderTemplate(w, "loginregister", webpage)
}

/*
registrationHandler method --

*/
func registrationHandler(w http.ResponseWriter, r *http.Request){
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	user := User{
		UserName: username,
		Password: password,
		Email: email,
	}
	if user.UserName == "" {
		http.NotFound(w, r)
		return
	}
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout
	ip := r.RemoteAddr
	url := r.URL
	requestLogger := log.WithFields(logrus.Fields{"user_ip": ip, "path": url})
	requestLogger.Info(time.Now().Format("2006-01-02 15:04:05")+" - registrationHandler")
	_ = writeToDB("users", ip, user)
	webpage, err := loadWebPage(user)
	if err != nil {
		log.Fatal(err)
	}
	renderTemplate(w, "register", webpage)
}


/*
makeHandler method --
wrapper function takes handler functions and returns a function of type http.HandlerFunc
fn is enclosed by closure, fn will be one of the pages available
closure returned by makeHandler is a function that takes http.ResponseWriter and http.Request
then extracts title from request path, validates with TitleValidator regexp.
If title is invalid, error will be written, ResponseWriter, using http.NotFound
If title is valid, enclosed handler function fn will be called with the ResponseWriter, Request and title as args
*/
func makeHandler(fn func (http.ResponseWriter, *http.Request)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        //extract page title from Request
        //call provided handler 'fn'
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil{
            http.NotFound(w, r)
            return
        }
        fn(w, r)
    }
}

/*
makeHomeHandler method --
*/
func makeHomeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        //extract page title from Request
        //call provided handler 'fn'
        m := validHomePath.FindStringSubmatch(r.URL.Path)
        if m == nil{
            http.NotFound(w, r)
            return
        }
        fmt.Printf("M: %s", m)
        fn(w, r, "")
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request, user_name string){
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout
	ip := r.RemoteAddr
	url := r.URL
	requestLogger := log.WithFields(logrus.Fields{"user_ip": ip, "path": url})
	requestLogger.Info(time.Now().Format("2006-01-02 15:04:05")+" - registrationHandler")
	u := strings.Split(url.String(), "/home/")
	fmt.Println(u)
	user := getUserInfo(u[1])
	webpage, err := loadWebPage(user)
	if err != nil {
		log.Fatal(err)
	}
	renderTemplate(w, "home", webpage)
}

func getBytes(key interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(key)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func getInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

/*
writeToDB
*/
func writeToDB(dir string, db_key string, user User) bool {
	opts := badger.DefaultOptions
	opts.Dir = "C:\\Users\\zim\\Documents\\badger\\azorg_logs\\" + dir
	opts.ValueDir = "C:\\Users\\zim\\Documents\\badger\\azorg_logs\\" + dir
	db, err := badger.Open(opts)
	if err != nil {
	  log.Fatal(err)
	  return false
	}
	defer db.Close()
	log.Out = os.Stdout
	fmt.Printf("U: %s\nP: %s\nE: %s\n", user.UserName, user.Password, user.Email)
	switch dir {
	case "users":
		// gob encoding
		key := []byte(user.UserName)
		encBuf := new(bytes.Buffer)
		err := gob.NewEncoder(encBuf).Encode(user)
		if err != nil {
			log.Fatal(err)
			return false
		}
		value := encBuf.Bytes()
		err1 := db.Update(func(txn *badger.Txn) error {
			err1 := txn.Set(key, value)
			return err1
		})
	  	if err1 != nil {
		  log.Fatal(err)
		  return false
		}
		return true
	default:
		requestLogger := log.WithFields(logrus.Fields{"directory": dir, "ip": db_key})
		requestLogger.Panic(time.Now().Format("2006-01-02 15:04:05 -- "))
		panic("unrecognized badger directory")
		return false
	}
	
}

func getUserInfo(username string) User {
	opts := badger.DefaultOptions
	opts.Dir = "/Documents/badger/azorg_logs/users"
	opts.ValueDir = "C:\\Users\\zim\\Documents\\badger\\azorg_logs\\users"
	db, err := badger.Open(opts)
	if err != nil {
	  log.Fatal(err)
	}
	user := User{}
	defer db.Close()
	err1 := db.View(func(txn *badger.Txn) error {
	    item, err := txn.Get([]byte(username))
	    if err != nil {
	      return err
	    }
	    val, err := item.Value()
	    if err != nil {
	      return err
	    }
	    decBuf := bytes.NewBuffer(val)
	    // user := User{}
	    err = gob.NewDecoder(decBuf).Decode(&user)
	    // fmt.Printf("key=%s\nvalue=%s\n", k, v)
	    // user.UserName = UserName
	    fmt.Printf("Key: %s\nValue: %s\n", username, user)
	    // u := User{
	    // 	UserName: user.UserName,
	    // 	Password: user.Password,
	    // 	Email: user.Email,
	    // }
	    return nil
  	})
  	if err1 != nil {
    log.Fatal(err1)
  	}
  	return user
}


func main(){
	http.HandleFunc("/loginregister/", makeHandler(loginPageHandler))
	http.HandleFunc("/register/", makeHandler(registrationHandler))
	http.HandleFunc("/home/", makeHomeHandler(homeHandler))
	http.ListenAndServe(":8000", nil)
}