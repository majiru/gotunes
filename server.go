package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	time.Time
	IP       string
	Username string
	Level    UserLevel
}

type Cache map[uuid.UUID]Token

var store = struct {
	*sync.RWMutex
	Cache
}{
	&sync.RWMutex{},
	make(Cache),
}

func genIndexHandle(song chan string, ctl chan string, q chan []string, useAuth bool) http.HandlerFunc {
	basic := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(index))
		case http.MethodPost:
			u := r.FormValue("url")
			if u != "" {
				song <- u
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
	if !useAuth {
		return basic
	}
	admin := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/":
				w.Write([]byte(adminIndex))
			case "/view":
				queue := <-q
				w.Write([]byte("Queue:\n"))
				for _, s := range queue {
					w.Write([]byte(s + "\n"))
				}
			}
		case http.MethodPost:
			switch r.URL.Path {
			case "/skip":
				ctl <- "skip"
			case "/":
				u := r.FormValue("url")
				if u != "" {
					song <- u
				}
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth_token")
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		id, err := uuid.Parse(c.Value)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		store.RLock()
		t, ok := store.Cache[id]
		store.RUnlock()
		if !ok || time.Now().After(t.Add(24*time.Hour)) {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		switch t.Level {
		case Basic:
			basic(w, r)
		case Admin:
			admin(w, r)
		}
	}
}

func handleSignin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(signinPage))
	case http.MethodPost:
		u := r.FormValue("username")
		p := r.FormValue("password")
		if u == "" || p == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		ok, level, err := AuthUser(u, p)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			id := uuid.New()
			t := time.Now()
			store.Lock()
			store.Cache[id] = Token{t, r.RemoteAddr, u, level}
			store.Unlock()
			c := &http.Cookie{
				Name:     "auth_token",
				Value:    id.String(),
				Expires:  t.Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, c)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

const index = `
<!doctype html>
<html lang="en">
	<head>
		<title>GoTunes</title>
	</head>
	<body>
		<form action="/" method="post">
			<input id="url" name="url">
			<button type="submt">Queue Up!</button>
		</form>
	</body>
</html>
`

const adminIndex = `
<!doctype html>
<html lang="en">
	<head>
		<title>GoTunes</title>
	</head>
	<body>
		<form action="/" method="post">
			<input id="url" name="url">
			<button type="submit">Queue Up!</button>
		</form>
		<form action="/skip" method="post">
			<button type="submit">Skip!</button>
		</form>
	</body>
</html>
`

const signinPage = `
<!doctype html>
<html lang="en">
	<head>
		<title>Auth in the Middle</title>
	</head>
	<body>
		<form action="/signin" method="post">
			<input id="username" name="username">
			<input id="password" type="password" name="password">
			<button type="submit">Login</button>
		</form>
	</body>
</html>
`
