package main

import (
	"log"
	"net/http"

	"donaldle.com/m/handler"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

func main() {
	router := httprouter.New()
	router.GET("/", handler.AllBlogs)
	router.POST("/blog", handler.CreateBlog)
	router.GET("/blog/:id", handler.OneBlog)
	router.PUT("/blog/:id", handler.UpdateBlog)
	router.DELETE("/blog/:id", handler.DeleteBlog)

	log.Fatal(http.ListenAndServe(":8081", router))
}
