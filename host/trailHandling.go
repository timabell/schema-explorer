package host

import (
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"net/http"
	"strings"
	"time"
)

const trailCookieName = "table-trail"

func ReadTrail(req *http.Request) *trail.TrailLog {
	trailCookie, _ := req.Cookie(trailCookieName)
	if trailCookie != nil {
		return trailFromCsv(trailCookie.Value)
	}
	return &trail.TrailLog{}
}

func trailFromCsv(values string) *trail.TrailLog {
	return &trail.TrailLog{Tables: strings.Split(values, ",")}
}

func SetCookie(trail *trail.TrailLog, resp http.ResponseWriter) {
	trailCookie := &http.Cookie{Name: trailCookieName, Value: TrailString(trail), Path: "/"}
	http.SetCookie(resp, trailCookie)
}
func ClearTrailCookie(resp http.ResponseWriter) {
	trailCookie := &http.Cookie{Name: trailCookieName, Value: "", Path: "/", Expires: time.Now().Add(-10000)}
	http.SetCookie(resp, trailCookie)
}

func TrailString(trail *trail.TrailLog) string {
	return strings.Join(trail.Tables, ",")
}
