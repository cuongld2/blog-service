package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"

	"donaldle.com/m/config"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

type Blogs struct {
	BlogID      int       `json:"blog_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	CreatedOn   time.Time `json:"created_on"`
	LastUpdated time.Time `json:"last_updated"`
}

type CreatingBlog struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func AllBlogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// We only accept 'GET' method here
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	allBlogsCached := getAllBlogsRedis(getEnv("RedisHost", "tcps://"), getEnv("RedisPassword", "password"))

	// Get all blogs from DB
	// rows, err := config.DB.Query("SELECT * FROM blogs")
	// if err != nil {
	// 	http.Error(w, http.StatusText(500), 500)
	// 	return
	// }
	// // Close the db connection at the end
	// defer rows.Close()

	// // Create blog object list
	// blogs := make([]Blogs, 0)
	// for _, element := range allBlogsCached {
	// 	blog := Blogs{}
	// 	err := rows.Scan(&blog.BlogID, &blog.Title, &blog.Content, &blog.Author, &blog.CreatedOn, &blog.LastUpdated) // order matters
	// 	if err != nil {
	// 		http.Error(w, http.StatusText(500), 500)
	// 		return
	// 	}
	// 	blogs = append(blogs, blog)
	// }
	// if err = rows.Err(); err != nil {
	// 	http.Error(w, http.StatusText(500), 500)
	// 	return
	// }

	// Returns as JSON (List of Blog objects)

	blogs := []Blogs{}

	// fmt.Println("11111")
	// fmt.Println(allBlogsCached)

	// fmt.Println("2222")

	for _, element := range allBlogsCached {

		blog := Blogs{}

		err := json.Unmarshal([]byte(element), &blog)
		if err != nil {
			fmt.Println(err)
			return
		}

		blogs = append(blogs, blog)

	}

	// fmt.Println("aaaaaaaaaaa")
	// fmt.Println(blogs)
	// fmt.Println("bbbbbbb")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(blogs); err != nil {
		panic(err)
	}
}

func CreateBlog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var blog CreatingBlog

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("INSERT INTO blogs (CONTENT,TITLE,AUTHOR,CREATED_ON,LAST_UPDATED) VALUES ($1,$2,$3,$4,$5)", blog.Content, blog.Title, blog.Author, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
}

func getOneBlogRedis(host string, password string, key string) string {

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})

	defer client.Close()

	val, err := client.Get(key).Result()
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val)

	return val

}

func getAllBlogsRedis(host string, password string) []string {

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})

	defer client.Close()

	var cursor uint64

	var blogsCached []string
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(cursor, "blogId*", 0).Result()
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			fmt.Println("key", key)
			val, err := client.Get(key).Result()
			if err != nil {
				fmt.Println(err)
			}

			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(val)
			blogsCached = append(blogsCached, val)
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	// strings.Join(blogsCached, "\n")

	return blogsCached

}

func OneBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// We only accept 'GET' method here
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	blogID := ps.ByName("id")

	blogCached := getOneBlogRedis(getEnv("RedisHost", "tcps://"), getEnv("RedisPassword", "password"), "blogId-"+blogID)

	// Get the specific blog from DB
	// row := config.DB.QueryRow("SELECT * FROM blogs WHERE blog_id = $1", blogID)

	// Create blog object
	blog := Blogs{}

	err := json.Unmarshal([]byte(blogCached), &blog)
	if err != nil {
		fmt.Println(err)
		return
	}
	// err := row.Scan(&blog.BlogID, &blog.Title, &blog.Content, &blog.Author, &blog.CreatedOn, &blog.LastUpdated)
	// switch {
	// case err == sql.ErrNoRows:
	// 	http.NotFound(w, r)
	// 	return
	// case err != nil:
	// 	fmt.Println(err)
	// 	http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	// 	return
	// }

	// Returns as JSON (single Blog object)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(blog); err != nil {
		panic(err)
	}
}

func UpdateBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Needs to convert float64 to int for the value from context

	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM blogs WHERE blog_id = $1", blogID)
	// Create blog object
	updatingBlog := Blogs{}
	er := row.Scan(&updatingBlog.BlogID, &updatingBlog.Title,
		&updatingBlog.Content, &updatingBlog.Author, &updatingBlog.CreatedOn, &updatingBlog.LastUpdated)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		fmt.Println(er)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	var blog CreatingBlog

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("UPDATE blogs SET content = $1, title = $2, author = $3, last_updated = $4 WHERE blog_id = $5", blog.Content, blog.Title, blog.Author, time.Now().UTC(), updatingBlog.BlogID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func DeleteBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM blogs WHERE blog_id = $1", blogID)
	// Create blog object
	deletingBlog := Blogs{}
	er := row.Scan(&deletingBlog.BlogID, &deletingBlog.Title,
		&deletingBlog.Content, &deletingBlog.Author, &deletingBlog.CreatedOn, &deletingBlog.LastUpdated)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	_, err := config.DB.Exec("DELETE FROM blogs WHERE blog_id = $1", deletingBlog.BlogID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}
