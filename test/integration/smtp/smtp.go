package smtp

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/rs/zerolog"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

// Borrowed from: https://github.com/emersion/go-smtp/blob/master/example_server_test.go

// The Backend implements SMTP server methods.
type Backend struct {
	log zerolog.Logger
}

// NewSession is called after client greeting (EHLO, HELO).
func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		log: bkd.log,
	}, nil
}

// A Session is returned after successful login.
type Session struct {
	log zerolog.Logger
}

// AuthMechanisms returns a slice of available auth mechanisms; only PLAIN is
// supported in this example.
func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

// Auth is the handler for supported authenticators.
func (s *Session) Auth(mech string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(identity, username, password string) error {
		return nil
	}), nil
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	s.log.Info().Msgf("mail from: %s", from)
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.log.Info().Msgf("rcpt to: %s", to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := io.ReadAll(r); err != nil {
		return err
	} else {
		s.log.Info().Msgf("data: %s", string(b))
	}
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

type Server struct {
	server *smtp.Server
	port   int
	host   string
}

func (s *Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Host() string {
	return s.host
}

func New(host string, port int, log zerolog.Logger) *Server {
	be := &Backend{
		log: log.With().Str("component", "smtp").Logger(),
	}

	s := smtp.NewServer(be)

	s.Addr = fmt.Sprintf("%s:%d", host, port)
	s.Domain = host
	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	return &Server{
		server: s,
		port:   port,
		host:   host,
	}
}
