package main

import (
	
	"github.com/nishanths/go-xkcd"
	"net/http"
	"log"
	"regexp"
	"strconv"
	"io/ioutil"
	/*
	"errors"
	"fmt"
	*/
	"html/template"
	"time"
	"os"
	"strings"
	
)

//template for page
var templates = template.Must(template.ParseFiles("twitchindex.html"))

/*
validPath - MustCompile will parse and compile the regexp
and return a regexp.
Regexep.MustCompile is distinct form Compile in that it will panic if the exp compiliation fails
*/
var validPath = regexp.MustCompile("^/(twitch)/([a-zA-Z0-9]+)") ///([a-zA-Z0-9]+)$

//struct for page
//xkcd
type Page struct {
	Title string
	Comic
	Comments []string
}
type Comic struct {
	Alt string
	ImgUrl string
	Number string
	cTitle string
	Transcript string
}

/*
save method new --
*/
func (p *Page) save() error {
	log.Println("Comic Number: " + p.Comic.Number)
    filename := p.Comic.Number + ".txt"
    log.Println("Filename: " + filename)
    f, err := os.OpenFile (filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("save err")
		panic(err)
	}
	defer f.Close()
    dlog := "[" + time.Now().String() +"]:\n" + strings.Join(p.Comments, ", ") + "\n"
	if _, err = f.WriteString(dlog); err != nil {
    	panic(err)
	}
	f.Close()
	return err
}

/*
loadPage method --
loads the comic page
*/
func loadPage(number string) (*Page, error){
	filename := number + ".txt"
   	client := xkcd.NewClient()
    i, err := strconv.Atoi(number)
    comic, err := client.Get(int(i))
    if err != nil {
        log.Fatal(err)
    }
   	data, err := ioutil.ReadFile(filename)
	if err != nil {
	  	p := &Page{
	  		Title: comic.SafeTitle,
			Comic: Comic {
				Alt: comic.Alt,
				ImgUrl: comic.ImageURL,
				Number: number,
				cTitle: comic.Title,
				Transcript: comic.Transcript}}
		return p, nil
	}else{
		s := string(data)
		sarr := strings.Split(s, "\n")
		var sarr2 []string
		for i := 0; i < len(sarr) -1; i++ {
			if i % 2 != 0 {
				log.Println(sarr[i])
				sarr2 = append(sarr2, sarr[i])
			}
		}
		p := &Page{
	  		Title: comic.SafeTitle,
			Comic: Comic {
				Alt: comic.Alt,
				ImgUrl: comic.ImageURL,
				Number: number,
				cTitle: comic.Title,
				Transcript: comic.Transcript}, Comments: sarr2}
		return p, nil
	}
    
}


/*
loadRandom method --
*/
func loadRandom() (*Page, error){
	client := xkcd.NewClient()
	comic, err := client.Random()
	if err != nil { log.Fatal(err) }
	return &Page{Title: comic.SafeTitle,
		Comic: Comic{ Alt: comic.Alt,
				ImgUrl: comic.ImageURL,
				Number: string(comic.Number),
				cTitle: comic.Title,
				Transcript: comic.Transcript}}, nil
}

/*
general renderTemplate func
http.Eerror sends specified internalservice error response code and err msg
*/
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

/*
comicHandler method --

*/
func comicHandler(w http.ResponseWriter, r *http.Request, title string){
	p, err := loadPage(title)
	if err != nil {
		p, err := loadRandom()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "xkcdindex", p)
	}
	renderTemplate(w, "xkcdindex", p)
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
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        //extract page title from Request
        //call provided handler 'fn'
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil{
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

/*
saveHandler method --
handle submission of forms located on edit pages
page title and form's only field, Body, are stored in new Page
save() method called to write data to file and client redirected to /view/ page
FormValue - returns type string, must convert to []byte slice before it will fit into Page struct
accepts a title (string)
*/
func saveHandler(w http.ResponseWriter, r *http.Request, number string){
	log.Println("Number: " + number)
	body := r.FormValue("comment")
	log.Println("Comment: " + body)
	body_arr := []string{body}
    client := xkcd.NewClient()
    i, err := strconv.ParseInt(number, 10, 0)
    comic, err2 := client.Get(int(i))
    if err2 != nil { log.Fatal(err) }
    p := &Page{Title: comic.SafeTitle,
		Comic: Comic{ Alt: comic.Alt,
				ImgUrl: comic.ImageURL,
				Number: number,
				cTitle: comic.Title,
				Transcript: comic.Transcript}, Comments: body_arr}
	log.Println(number)
	p.save()
    http.Redirect(w, r, "/xkcd/"+number, http.StatusFound)
}


func main(){
	http.HandleFunc("/xkcd/", makeHandler(comicHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8000", nil)
}
