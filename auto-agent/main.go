package main

import (
	"github.com/toisin/astro-world/auto-agent/server"
	"github.com/toisin/astro-world/auto-agent/workflow"
	"google.golang.org/appengine"
)

// // indexHandler responds to requests with our greeting.
// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	ctx := appengine.NewContext(r)
// 	key := datastore.NewKey(ctx, "User", "user1", 0, nil)
// 	fmt.Fprintf(w, "Hello, World! %s", key)
// 	// fmt.Fprintf(w, "Hello, World!")
// }

// func handle(w http.ResponseWriter, r *http.Request) {
// 	ctx := appengine.NewContext(r)
// 	key := datastore.NewKey(ctx, "User", "user1", 0, nil)

// }

func main() {
	// http.HandleFunc("/", indexHandler)

	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// }

	server.InitHandlers()
	workflow.InitWorkflow()
	appengine.Main()

	// log.Printf("Listening on port %s", port)
	// if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 	log.Fatal(err)
	// }

}
