package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox/internal/models"
	"strconv"
	"strings"
	"unicode/utf8"
)

type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData()
	data.Snippets = snippets
	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	data := app.newTemplateData()
	data.Snippet = snippet
	app.render(w, r, http.StatusOK, "view.html", data)
}
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData()
	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// First we call r.ParseForm() which adds any data in POST request bodies
	// to the r.PostForm map. This also works in the same way for PUT and PATCH
	// requests. If there are any errors, we use our app.ClientError() helper to
	// send a 400 Bad Request response to the user.
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// The r.PostForm.Get() method always returns the form data as a *string*.
	// However, we're expecting our expires value to be a number, and want to
	// represent it in our Go code as an integer. So we need to manually convert
	// the form data to an integer using strconv.Atoi(), and send a 400 Bad
	// Request response if the conversion fails.
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := snippetCreateForm{
		Title:       r.PostForm.Get("title"),
		Content:     r.PostForm.Get("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}
	// Initialize a map to hold any validation errors for the form fields.
	fieldErrors := make(map[string]string)
	// Check that the title value is not blank and is not more than 100
	// characters long. If it fails either of those checks, add a message to the
	// errors map using the field name as the key.
	if strings.TrimSpace(form.Title) == "" {
		fieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		// len() counts bytes. If you input "世界", it contains 3 bytes.
		fieldErrors["title"] = "This field cannot be more than 100 characters long"
	}
	// Check that the content value isn't blank.
	if strings.TrimSpace(form.Content) == "" {
		fieldErrors["content"] = "This field cannot be blank"
	}
	// Check the expires value matches one of the permitted values (1, 7 or
	// 365).
	if expires != 1 && expires != 7 && expires != 365 {
		fieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}
	// If there are any errors, dump them in a plain-text HTTP response and
	// return from the handler.
	if len(fieldErrors) > 0 {
		data := app.newTemplateData()
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}
	id, err := app.snippets.Insert(form.Title, form.Content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
