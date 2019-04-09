package serve

import (
	"fmt"
	"log"
	"net/http"
)

// Set status to 500, log and show error message
// Calling code should return after calling this to avoid adding further output etc.
// This is not intended for user consumption as a rule. Expected messages should be shown in context and allow retrying
func serverError(resp http.ResponseWriter, message string, err error) {
	// log
	log.Print(fmt.Sprintf("%s: %s", message, err))

	// set http response
	resp.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(resp, fmt.Sprintf("%s:\n\n%s", message, err))
}
