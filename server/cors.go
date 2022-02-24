package server

import (
	"github.com/rs/cors"
)

var CORS = cors.AllowAll().Handler
