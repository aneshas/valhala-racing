package webshop

import (
	"github.com/labstack/echo/v4"
)

func initService() (*Service, error) {
	s := Service{}

	e := echo.New()

	g := e.Group("backend")

	g.POST("/request-server", s.requestServer)

	s.e = e

	return &s, nil
}
