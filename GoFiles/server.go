package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

var store = sessions.NewCookieStore([]byte("69-420"))

type User struct {
	Name  string
	Email string
	Id    int
}

type NewsApi struct {
	Status       string
	TotalResults int
	Articles     []Articles
}

type Articles struct {
	Title       string
	Author      string
	Source      ArticleSource
	PublishedAt string
	Url         string
}

type ArticleSource struct {
	Id   string
	Name string
}

var currentUser User

type DashboardData struct {
	CurrentUser User
	Articles    []Articles
}

func seeIfLoggedIn(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth")
		if err != nil {
			panic(err)
		}
		_, ok := session.Values["Beep"]
		if ok {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func seeIfNotLoggedIn(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth")
		if err != nil {
			panic(err)
		}
		_, ok := session.Values["Beep"]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func main() {
	http.HandleFunc("/", seeIfLoggedIn(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Fprintf(w, "<h1>404, page not found</h1> <br/> <a href=\"/\">Home</a>")
			return
		}
		switch r.Method {
		case "GET":
			page, err := template.ParseFiles("../HtmlPages/index.html")
			if err != nil {
				panic(err)
			}
			page.Execute(w, nil)

		}
	}))

	http.HandleFunc("/login", seeIfLoggedIn(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			page, err := template.ParseFiles("../HtmlPages/login.html")
			if err != nil {
				panic(err)
			}
			page.Execute(w, nil)
		case "POST":
			r.ParseForm()
			email := r.FormValue("email")
			password := r.FormValue("password")
			conn, err := pgx.Connect(context.Background(), "postgres://xzvhavyz:VRXYX321NWGOPvMgBwJ477rmBc5jQJG9@surus.db.elephantsql.com/xzvhavyz")
			defer conn.Close(context.Background())
			if err != nil {
				panic(err)
			}
			var dbEmail string
			var dbPassword string
			var dbName string
			var dbId int
			err = conn.QueryRow(context.Background(), "SELECT name, email, password, id FROM users WHERE email=$1", email).Scan(&dbName, &dbEmail, &dbPassword, &dbId)

			if err != nil {
				if err == pgx.ErrNoRows {
					page, err1 := template.ParseFiles("../HtmlPages/login.html")
					if err1 != nil {
						panic(err)
					}
					page.Execute(w, "ERROR - Credentials not found")
					return
				}
				panic(err)
			}
			if password == dbPassword {
				session, err := store.Get(r, "auth")
				if err != nil {
					panic(err)
				}
				session.Values["Beep"] = "Boop"
				session.Save(r, w)
				currentUser = User{Name: dbName, Email: dbEmail, Id: dbId}
				http.Redirect(w, r, "/dashboard", http.StatusFound)
			} else {
				if err == pgx.ErrNoRows {
					page, err := template.ParseFiles("../HtmlPages/login.html")
					if err != nil {
						panic(err)
					}
					page.Execute(w, "ERROR - Invalid credentials")
					return
				}
			}

		}
	}))

	http.HandleFunc("/signup", seeIfLoggedIn(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			pages, err := template.ParseFiles("../HtmlPages/signup.html")
			if err != nil {
				panic(err)
			}
			pages.Execute(w, nil)

		case "POST":
			conn, err := pgx.Connect(context.Background(), "postgres://xzvhavyz:VRXYX321NWGOPvMgBwJ477rmBc5jQJG9@surus.db.elephantsql.com/xzvhavyz")
			if err != nil {
				panic(err)
			}
			defer conn.Close(context.Background())
			r.ParseForm()
			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")
			var result string
			err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE email=$1", email).Scan(&result)

			

			if err != nil {
				panic(err)

			}

			if result == "0" {
				conn.QueryRow(context.Background(), "INSERT INTO users (name, email, password) VALUES($1,$2,$3)", name, email, password)

				http.Redirect(w, r, "/login", http.StatusFound)
				return
			} else if len(result) > 0 {
				pages, err := template.ParseFiles("../HtmlPages/signup.html")
				if err != nil {
					fmt.Println("line 66")
					panic(err)
				}
				pages.Execute(w, "ERROR - Credentials already in use")
				return
			}

		}
	}))

	http.HandleFunc("/dashboard", seeIfNotLoggedIn(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":

			resp, err := http.Get("https://newsapi.org/v2/top-headlines?category=general&language=en&apiKey=1272c48f3ed445d7819ac65b3548b2ac")
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			var response NewsApi
			json.Unmarshal(body, &response)

			page, err := template.ParseFiles("../HtmlPages/dashboard.html")
			if err != nil {
				panic(err)
			}
			Payload := DashboardData{currentUser, response.Articles}
			page.Execute(w, Payload)

		}
	}))

	type collection1 struct {
		Title string
		Url   string
		Name  string
	}

	type Blogs struct {
		Blog []collection1
	}

	type BlogData struct {
		Title string
		Name  string
		Email string
		Body  string
	}
	http.HandleFunc("/blogs/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/blogs/")
		if len(id) == 0 {
			switch r.Method {
			case "GET":
				conn, err := pgx.Connect(context.Background(), "postgres://xzvhavyz:VRXYX321NWGOPvMgBwJ477rmBc5jQJG9@surus.db.elephantsql.com/xzvhavyz")
				if err != nil {
					panic(err)
				}
				defer conn.Close(context.Background())
				Rows, err := conn.Query(context.Background(), "SELECT blogs.title, blogs.url, users.name FROM blogs INNER JOIN users ON blogs.author_id = users.id")
				if err != nil {
					panic(err)
				}
				var BlogData []collection1
				for Rows.Next() {
					var title string
					var url string
					var name string
					Rows.Scan(&title, &url, &name)
					BlogData = append(BlogData, collection1{title, url, name})
				}
				BlogPayload := Blogs{BlogData}

				page, err := template.ParseFiles("../HtmlPages/blog.html")
				if err != nil {
					panic(err)
				}
				page.Execute(w, BlogPayload)

			default:
				fmt.Fprintf(w, "ERROR - Invalid method")
			}

		} else if id == "publish" {
			switch r.Method {
			case "GET":
				session, err := store.Get(r, "auth")
				if err != nil {
					panic(err)
				}
				_, ok := session.Values["Beep"]
				if !ok {
					http.Redirect(w, r, "/login", http.StatusFound)
					return
				}
				page, err := template.ParseFiles("../HtmlPages/publishBlog.html")
				if err != nil {
					panic(err)
				}
				page.Execute(w, currentUser)
			case "POST":
				r.ParseForm()
				title := r.FormValue("title")
				author_id := r.FormValue("user_id")
			
				body := r.FormValue("body")
				conn, err := pgx.Connect(context.Background(), "postgres://xzvhavyz:VRXYX321NWGOPvMgBwJ477rmBc5jQJG9@surus.db.elephantsql.com/xzvhavyz")
				if err != nil {
					panic(err)
				}
				defer conn.Close(context.Background())
				var test string
				err = conn.QueryRow(context.Background(), "INSERT INTO blogs (author_id, title, body) VALUES($1,$2,$3)", author_id, title, body).Scan(&test)
				if err != nil && err != pgx.ErrNoRows {
					fmt.Println(306)
					panic(err)
				}
				http.Redirect(w, r, "/blogs", http.StatusFound)
			}

		} else {
			conn, err := pgx.Connect(context.Background(), "postgres://xzvhavyz:VRXYX321NWGOPvMgBwJ477rmBc5jQJG9@surus.db.elephantsql.com/xzvhavyz")
			if err != nil {
				fmt.Println("310")
				panic(err)
			}

			var title, name, email, body string
			err = conn.QueryRow(context.Background(), "SELECT blogs.title, users.name, users.email, blogs.body FROM blogs INNER JOIN users ON blogs.author_id = users.id WHERE blogs.url=$1", id).Scan(&title, &name, &email, &body)
			if err == pgx.ErrNoRows {
				fmt.Fprintf(w, "<h1>404, page not found</h1> <br/> <a href=\"/\">Home</a>")
				return
			}
			conn.Close(context.Background())
			Payload := BlogData{title, name, email, body}
			page, _ := template.ParseFiles("../HtmlPages/blogPage.html")

			page.Execute(w, Payload)
		}
	})

	http.ListenAndServe(":8000", nil)
}
