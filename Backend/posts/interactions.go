package posts

import (
	"forum/Backend/DB"
	"forum/Backend/errors"
	"forum/Backend/login"
	"net/http"
	"strconv"
	"strings"
)

// LikePostHandler handles likes/dislikes for posts and comments.
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.InternalServerError(w, r, "Method not allowed")
		return
	}

	session, err := login.GetSessionFromRequest(r)
	if err != nil || session.IsGuest || session.UserID == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id")
	isLikeStr := r.FormValue("is_like")

	if (postIDStr == "" && commentIDStr == "") || (postIDStr != "" && commentIDStr != "") {
		errors.BadRequest(w, r, "Must provide either post_id or comment_id")
		return
	}

	isLike, err := strconv.Atoi(isLikeStr)
	if err != nil || (isLike != 0 && isLike != 1) {
		errors.BadRequest(w, r, "is_like must be 0 or 1")
		return
	}

	if postIDStr != "" {
		postID, _ := strconv.Atoi(postIDStr)
		ok, err := db.CheckPostExists(db.DB, postID)
		if err != nil || !ok {
			errors.BadRequest(w, r, "Post does not exist")
			return
		}
		err = db.ToggleLike(db.DB, db.LikeTarget{ID: postID, UserID: *session.UserID, IsPost: true, IsLike: isLike == 1})
		if err != nil {
			errors.InternalServerError(w, r, "DB error: "+err.Error())
			return
		}
	}

	if commentIDStr != "" {
		commentID, _ := strconv.Atoi(commentIDStr)
		ok, err := db.CheckCommentExists(db.DB, commentID)
		if err != nil || !ok {
			errors.BadRequest(w, r, "Comment does not exist")
			return
		}
		err = db.ToggleLike(db.DB, db.LikeTarget{ID: commentID, UserID: *session.UserID, IsPost: false, IsLike: isLike == 1})
		if err != nil {
			errors.InternalServerError(w, r, "DB error: "+err.Error())
			return
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// CommentOnPostHandler handles comments on posts.
func CommentOnPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	session, err := login.GetSessionFromRequest(r)
	if err != nil || session == nil || session.IsGuest || session.UserID == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.FormValue("post_id")
	content := strings.TrimSpace(r.FormValue("content"))

	if postIDStr == "" || content == "" {
		errors.BadRequest(w, r, "Post ID and content are required")
		return
	}

	if len(content) > 100 {
		errors.BadRequest(w, r, "Comment too long (max 100 characters)")
		return
	}

	postID, _ := strconv.Atoi(postIDStr)
	ok, err := db.CheckPostExists(db.DB, postID)
	if err != nil || !ok {
		errors.BadRequest(w, r, "Post does not exist")
		return
	}

	if err := db.AddComment(db.DB, postID, *session.UserID, content); err != nil {
		errors.InternalServerError(w, r, "DB error adding comment: "+err.Error())
		return
	}

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}
