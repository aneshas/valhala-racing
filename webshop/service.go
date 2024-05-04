package webshop

import (
	"embed"
	"encore.app/checkout"
	"encore.dev/rlog"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/fs"
	"net/http"
	"strconv"
)

var (
	//go:embed pages
	dist embed.FS

	static, _ = fs.Sub(dist, "pages")
	handler   = http.StripPrefix("/shop/", http.FileServer(http.FS(static)))
)

// Service represents web shop fronted service
//
//encore:service
type Service struct {
	e *echo.Echo
}

// ServeShopStatic serves the frontend for web shop.
//
//encore:api public raw path=/shop/*path
func (s *Service) ServeShopStatic(w http.ResponseWriter, req *http.Request) {
	handler.ServeHTTP(w, req)
}

// ServeShopBackend serves web shop backend
//
//encore:api public raw path=/backend/*path
func (s *Service) ServeShopBackend(w http.ResponseWriter, req *http.Request) {
	s.e.ServeHTTP(w, req)
}

type requestServerFormReq struct {
	Email   string `form:"email"`
	Package string `form:"package"`
	Hours   int
}

func (r *requestServerFormReq) Validate() error {
	hours, err := strconv.Atoi(r.Package)
	if err != nil {
		return err
	}

	if hours != 1 &&
		hours != 3 &&
		hours != 5 &&
		hours != 10 {
		return fmt.Errorf("something is fishy here ")
	}

	r.Hours = hours

	return nil
}

func (s *Service) requestServer(c echo.Context) error {
	var req requestServerFormReq

	err := c.Bind(&req)
	if err != nil {
		return err
	}

	err = req.Validate()
	if err != nil {
		return err
	}

	resp, err := checkout.RequestServer(
		c.Request().Context(),
		&checkout.RequestServerReq{
			HoursReserved: req.Hours,
			Email:         req.Email,
		})
	if err != nil {
		rlog.Error(err.Error())

		return c.String(http.StatusInternalServerError, "something went wrong")
	}

	return c.Redirect(http.StatusSeeOther, resp.PaymentRedirect)
}
