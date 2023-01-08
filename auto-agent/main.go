package main

import (
	"github.com/toisin/astro-world/auto-agent/server"
	"github.com/toisin/astro-world/auto-agent/workflow"
	"google.golang.org/appengine"
)

func main() {
	server.InitHandlers()
	workflow.InitWorkflow()
	appengine.Main()
}
