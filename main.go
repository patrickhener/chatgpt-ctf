package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	_ "github.com/lib/pq"
)

type User struct {
	Username string
	Password string
}

func (u *User) Exec(cmd string) string {
	out, _ := exec.Command(cmd).CombinedOutput()
	return string(out)
}

var (
	dbUsername string = os.Getenv("DBUSER")
	dbPassword string = os.Getenv("DBPASSWORD")
	dbName     string = os.Getenv("DBNAME")
	dbHost     string = os.Getenv("DBHOST")
	dbPort     string = os.Getenv("DBPORT")
)

//go:embed static
var static embed.FS

func main() {
	http.HandleFunc("/getUser", handleGetUser)
	http.HandleFunc("/fetchCreds", handleRequest)
	http.HandleFunc("/", handleIndex)
	if err := http.ListenAndServe(":5000", nil); err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	staticIndex, err := static.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}

	if _, err := w.Write(staticIndex); err != nil {
		fmt.Println(err)
	}
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("test-user")); err != nil {
		fmt.Println(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(static, "static/display.html")
	if err != nil {
		panic(err)
	}

	users := []User{}
	query := r.URL.Query()
	urlString := query.Get("url")

	// check for ssrf vulnerability
	if urlString == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	parsedURL, err := url.Parse(urlString)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		http.Error(w, "invalid url parameter", http.StatusBadRequest)
		return
	}

	// make request to url provided in url parameter
	resp, err := http.Get(urlString)
	if err != nil {
		http.Error(w, "error making request to provided url", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// extract and sanitize user input from response body
	userInput := sanitizeInput(resp.Body)

	// perform sql injection with user input
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, dbHost, dbPort, dbName))
	if err != nil {
		http.Error(w, "error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	sqlQuery := fmt.Sprintf("SELECT * FROM users WHERE username='%s'", userInput)
	if os.Getenv("DEBUG") == "TRUE" {
		fmt.Printf("Query is %s\n", sqlQuery)
	}
	rows, err := db.Query(sqlQuery)
	if err != nil {
		http.Error(w, "error executing sql query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// process and display results
	var username string
	var password string
	for rows.Next() {
		err := rows.Scan(&username, &password)
		if err != nil {
			http.Error(w, "error processing sql query results", http.StatusInternalServerError)
			return
		}
		if os.Getenv("DEBUG") == "TRUE" {
			fmt.Printf("username: %s, password: %s\n", username, password)
		}
		users = append(users, User{
			Username: username,
			Password: password,
		})
	}

	if err := t.Execute(w, users); err != nil {
		fmt.Println(err)
	}
}

func sanitizeInput(input io.Reader) string {
	userInput, err := ioutil.ReadAll(input)
	if err != nil {
		return ""
	}
	// perform input sanitization
	// TODO none until now
	// Can use to make things more difficult
	stringUserInput := string(strings.TrimSuffix(string(userInput), "\n"))

	return stringUserInput
}
