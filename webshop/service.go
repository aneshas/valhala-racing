package webshop

import (
	"embed"
	"encoding/json"
	"encore.app/raceroom"
	"encore.dev/rlog"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v76"
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

// Serve serves the frontend for development.
//
//encore:api public raw path=/shop/*path
func (s *Service) Serve(w http.ResponseWriter, req *http.Request) {
	handler.ServeHTTP(w, req)
}

//encore:api public raw path=/backend/*path
func (s *Service) ServeHTML(w http.ResponseWriter, req *http.Request) {
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

	resp, err := raceroom.RequestServer(
		c.Request().Context(),
		&raceroom.RequestServerReq{
			HoursReserved: req.Hours,
			Email:         req.Email,
		})
	if err != nil {
		rlog.Error(err.Error())

		// TODO - Error page
		return fmt.Errorf("could not create server request")
	}

	return c.Redirect(http.StatusSeeOther, resp.PaymentRedirect)
}

func (s *Service) stripeCallback(c echo.Context) error {
	var event stripe.Event

	err := c.Bind(&event)
	if err != nil {
		rlog.Error(err.Error())

		return c.NoContent(http.StatusBadRequest)
	}

	switch event.Type {
	case "checkout.session.completed":
		var s stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &s)
		if err != nil {
			rlog.Error(err.Error())

			return c.NoContent(http.StatusBadRequest)
		}

		err = raceroom.RegisterServerPayment(
			c.Request().Context(),
			&raceroom.RegisterServerPaymentReq{
				PaymentRef: s.ID,
			})
		if err != nil {
			err = errors.Join(err, fmt.Errorf("could not register payment"))

			return c.NoContent(http.StatusServiceUnavailable)
		}

	default:
		rlog.Info("unhandled stripe event", "event", event.Type)
	}

	return nil
}
