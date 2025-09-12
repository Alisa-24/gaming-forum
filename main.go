package main

import (
	"fmt"
	db "forum/Backend/DB"
	register "forum/Backend/Register"
	"forum/Backend/home"
	"forum/Backend/login"
	"forum/Backend/posts"
	"forum/Backend/profile"
	"net/http"
)

func main() {
	db.InitDB()
	if db.DB == nil {
		fmt.Println("Failed to connect to the database. Exiting.")
		return
	}
	db.InitCategories()

	mux := http.NewServeMux()

	// Static files without referer check
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.Handle("/static/", staticHandler)

	// Background files without referer check
	bgHandler := http.StripPrefix("/backgrounds/", http.FileServer(http.Dir("backgrounds")))
	mux.Handle("/backgrounds/", bgHandler)

	// App routes
	mux.HandleFunc("/", home.WelcomePage)
	mux.HandleFunc("/homePage", home.AllPosts)
	mux.HandleFunc("/category/general", home.GeneralPosts)
	mux.HandleFunc("/category/minecraft", home.MinecraftPosts)
	mux.HandleFunc("/category/souls", home.SoulsPosts)
	mux.HandleFunc("/category/online", home.OnlinePosts)
	mux.HandleFunc("/category/story", home.StoryPosts)
	mux.HandleFunc("/post", posts.PostShowHandler)
	mux.HandleFunc("/profile", profile.ProfileHandler)
	mux.HandleFunc("/about", home.AboutPage)
	mux.HandleFunc("/register", register.RegisterHandler)
	mux.HandleFunc("/login", login.LoginHandler)
	mux.HandleFunc("/createpost", posts.PostHandler)
	mux.HandleFunc("/logout", login.LogoutHandler)
	mux.HandleFunc("/post/like", posts.LikePostHandler)
	mux.HandleFunc("/post/comment", posts.CommentOnPostHandler)
	mux.HandleFunc("/comment/like", posts.LikePostHandler)

	fmt.Println("Server started on http://localhost:8888")
	if err := http.ListenAndServe(":8888", mux); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
