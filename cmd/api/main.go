package main

import (
	_ "time/tzdata"

	"github.com/Sagarmikeylevi/Pulse-Sever/cmd/api/commands"
)

// @title           Pulse API
// @version         1.0
// @description     Productivity tracking backend API

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer <your_token>"
func main() {
	commands.Execute()
}
