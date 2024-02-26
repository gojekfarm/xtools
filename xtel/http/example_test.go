package http_test

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	httptrace "github.com/gojekfarm/xtools/xtel/http"
)

func ExampleMuxMiddleware() {
	mwf := httptrace.MuxMiddleware("service-a", ignoreRoutes([]string{
		"/login",
	}))

	r := mux.NewRouter()
	r.Use(mwf)
}

func ignoreRoutes(in []string) httptrace.PathWhitelistFunc {
	return func(r *http.Request) bool {
		spanName := ""

		route := mux.CurrentRoute(r)
		if route != nil {
			var err error

			spanName, err = route.GetPathTemplate()
			if err != nil {
				spanName, err = route.GetPathRegexp()
				if err != nil {
					spanName = ""
				}
			}
		}

		for _, s := range in {
			if strings.EqualFold(spanName, s) {
				return true
			}
		}

		return false
	}
}
