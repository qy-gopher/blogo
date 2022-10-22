package route

import "github.com/gorilla/mux"

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
