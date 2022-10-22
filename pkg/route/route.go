package route

import (
	"net/http"

	"github.com/gorilla/mux"
)

var Router *mux.Router

func Initialize() {
	Router = mux.NewRouter()
}

func NameToURL(routename string, pairs ...string) string {
	url, err := Router.Get(routename).URL(pairs...)
	if err != nil {
		return ""
	}

	return url.String()
}

func GetRouteVariable(paramName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[paramName]
}
