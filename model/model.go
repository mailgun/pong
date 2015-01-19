package model

import (
	"github.com/mailgun/log"
	"net/http"
	"time"
)

type Response struct {
	Code        int
	Body        []byte
	ContentType string
	Delay       time.Duration
	Drop        bool
}

type Handler struct {
	Id        string
	Responses map[string]*Response
}

type Server struct {
	Addr         string
	Path         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Handler      http.Handler
}

func StartServer(server *Server) {
	log.Infof("Starting server %s", server.Addr)
	s := &http.Server{
		Addr:           server.Addr,
		Handler:        server.Handler,
		ReadTimeout:    server.ReadTimeout,
		WriteTimeout:   server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Errorf("Server %s exited with error: %s", server.Addr, err)
	}
}
