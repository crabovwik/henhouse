/**
 * @file auth.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU AGPLv3
 * @date December, 2015
 * @brief auth helpers and middleware
 */

package scoreboard

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jollheef/henhouse/db"
)

const (
	sessionCookieName = "session"
	contextTeamIDName = "teamID"
)

var authEnabled = true

var underProxy bool

func getClientAddr(r *http.Request) (clientAddr string) {
    
    if underProxy {
        clientAddr = r.Header.Get("x-real-ip")
    } else {
        clientAddr = r.RemoteAddr
    }
              
    return
}

func genSession() (s string, err error) {

	sessionLen := 256

	randBuf := make([]byte, sessionLen)

	_, err = rand.Read(randBuf)

	if err != nil {
		return
	}

	s = fmt.Sprintf("%x", randBuf)

	return
}

func getSessionTeamID(database *sql.DB, r *http.Request) (teamID int,
	err error) {

	session, err := r.Cookie(sessionCookieName)
	if err != nil {
		return
	}

	teamID, err = db.GetSessionTeam(database, session.Value)
	if err != nil {
		return
	}

	return
}

func setSessionTeamID(database *sql.DB, w http.ResponseWriter,
	teamID int) (err error) {

	session, err := genSession()
	if err != nil {
		return
	}

	cookie := http.Cookie{Name: sessionCookieName, Value: session}

	err = db.AddSession(database, &db.Session{
		TeamID:  teamID,
		Session: session,
	})
	if err != nil {
		return
	}

	http.SetCookie(w, &cookie)

	return
}

func getTeamID(r *http.Request) int {
	if rv := context.Get(r, contextTeamIDName); rv != nil {
		return rv.(int)
	}
	return 0
}

func authorized(database *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		teamID, err := getSessionTeamID(database, r)
		if err != nil && authEnabled {
			http.Redirect(w, r, "/auth.html", 307)
		} else {
			context.Set(r, contextTeamIDName, teamID)
			context.ClearHandler(next).ServeHTTP(w, r)
		}
	})
}

func authHandler(database *sql.DB, w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Redirect(w, r, "/", 307)
		return
	}

	token := r.FormValue("token")

	if token == "" {
		http.Redirect(w, r, "/", 307)
		return
	}

	log.Printf("auth ip: %s, access_token: %s", getClientAddr(r), token)

	teamID, err := db.GetTeamIDByToken(database, token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		tmpl, err := getTmpl("auth_error")
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprint(w, tmpl)
		return
	}

	err = setSessionTeamID(database, w, teamID)
	if err != nil {
		log.Println("Set session id fail:", err)
		return
	}

	log.Printf("Success auth ip: %s team ID: %d", getClientAddr(r), teamID)

	// Success auth
	http.Redirect(w, r, "/", 303)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName})
	http.Redirect(w, r, "/", 307)
	return
}
