package serve

import (
	"github.com/timabell/schema-explorer/trail"
	"net/http"
	"strings"
	"time"
)

const trailCookieName = "table-trail-"

func ReadTrail(databaseName string, req *http.Request) *trail.TrailLog {
	trailCookie, _ := req.Cookie(trailCookieName + databaseName)
	if trailCookie != nil && trailCookie.Value != "" {
		return trailFromCsv(trailCookie.Value)
	}
	return &trail.TrailLog{}
}

func SetTrailCookie(databaseName string, trail *trail.TrailLog, resp http.ResponseWriter) {
	trailCookie := &http.Cookie{Name: trailCookieName + databaseName, Value: trailString(trail), Path: "/"}
	http.SetCookie(resp, trailCookie)
}
func ClearTrailCookie(databaseName string, resp http.ResponseWriter) {
	trailCookie := &http.Cookie{Name: trailCookieName + databaseName, Value: "", Path: "/", Expires: time.Now().Add(-10000)}
	http.SetCookie(resp, trailCookie)
}

func trailString(trail *trail.TrailLog) string {
	return strings.Join(trail.Tables, ",")
}

func trailFromCsv(values string) *trail.TrailLog {
	return &trail.TrailLog{Tables: strings.Split(values, ",")}
}
