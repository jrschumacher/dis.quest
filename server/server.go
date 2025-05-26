package server

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
)

func Start() {
	component := components.Page(false)

	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
