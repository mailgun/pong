package config

import (
	"fmt"
	"github.com/mailgun/cfg"
	"github.com/mailgun/go-statsd-client/statsd"
	"github.com/mailgun/log"
	"github.com/mailgun/pong/model"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Rate        string
	Code        int
	Body        string
	ContentType string
	Delay       string
	Drop        bool
}

type Handler struct {
	Id        string
	Path      string
	Responses []*Response
}

type Server struct {
	Addr         string
	Path         string
	Handlers     []*Handler
	ReadTimeout  string
	WriteTimeout string
}

type Config struct {
	Statsd struct {
		Url    string
		Prefix string
	}
	Servers []*Server
	Logging []*log.LogConfig
}

type HandlerFn func(http.ResponseWriter, *http.Request)

func ParseConfig(path string) ([]*model.Server, []*log.LogConfig, error) {
	config := Config{}
	if err := cfg.LoadConfig(path, &config); err != nil {
		return nil, nil, fmt.Errorf("Failed to load config file '%s' err:", path, err)
	}

	client, err := statsd.New(config.Statsd.Url, config.Statsd.Prefix)
	if err != nil {
		return nil, nil, err
	}
	builder := &Builder{
		client: client,
	}
	servers, err := builder.parseServers(config.Servers)
	if err != nil {
		return nil, nil, err
	}
	return servers, config.Logging, nil
}

type Builder struct {
	client statsd.Client
}

func (b *Builder) parseServers(in []*Server) ([]*model.Server, error) {
	out := make([]*model.Server, len(in))
	for i, cserver := range in {
		serv, err := b.parseServer(cserver)
		if err != nil {
			return nil, err
		}
		out[i] = serv
	}
	return out, nil
}

func (b *Builder) parseServer(in *Server) (*model.Server, error) {
	readTimeout, err := time.ParseDuration(in.ReadTimeout)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := time.ParseDuration(in.WriteTimeout)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	for _, h := range in.Handlers {
		handler, err := b.buildHandler(h.Id, h.Responses)
		if err != nil {
			return nil, err
		}
		mux.Handle(h.Path, handler)
	}
	return &model.Server{
		Addr:         in.Addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      mux,
	}, nil
}

type Responder struct {
	id        string
	responses []*model.Response
	index     int
	client    statsd.Client
}

func (d *Responder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r := d.responses[d.index]
	d.index = (d.index + 1) % len(d.responses)

	d.client.Inc(metric("requests"), 1, 1)
	d.client.Inc(metric(d.id, "requests"), 1, 1)

	if r.Delay > 0 {
		time.Sleep(r.Delay)
	}
	if r.Drop {
		h := w.(http.Hijacker)
		conn, _, _ := h.Hijack()
		conn.Close()
		return
	}
	w.WriteHeader(r.Code)
	w.Header().Set("Content-Type", r.ContentType)
	w.Write([]byte(r.Body))
}

func (b *Builder) buildHandler(id string, distribution []*Response) (http.Handler, error) {
	responses, err := b.buildResponses(distribution)
	if err != nil {
		return nil, err
	}
	return &Responder{
		id:        id,
		responses: responses,
		client:    b.client,
	}, nil
}

func (b *Builder) buildResponses(responses []*Response) ([]*model.Response, error) {
	out := []*model.Response{}
	total := 0
	for _, re := range responses {
		duration, err := time.ParseDuration(re.Delay)
		if err != nil {
			return nil, fmt.Errorf("Bad delday '%s', should me in form '1s'", re.Delay)
		}
		count, err := parsePercent(re.Rate)
		if err != nil {
			return nil, err
		}
		total += count
		for i := 0; i < count; i += 1 {
			log.Infof("Appending %s", re.Body)
			out = append(out, &model.Response{
				Drop:        re.Drop,
				Delay:       duration,
				Code:        re.Code,
				Body:        []byte(re.Body),
				ContentType: re.ContentType,
			})
		}
	}
	if total != 100 {
		return nil, fmt.Errorf("Percentages should form 100 in sum, got %d", total)
	}
	return shuffle(out), nil
}

func shuffle(in []*model.Response) []*model.Response {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexes := r.Perm(len(in))
	out := make([]*model.Response, len(in))
	for i, v := range indexes {
		out[i] = in[v]
	}
	return out
}

func parsePercent(in string) (int, error) {
	num, err := strconv.Atoi(strings.TrimSuffix(in, "%"))
	if err != nil {
		return -1, fmt.Errorf("Use percentages, e.g 10%")
	}
	if num > 100 || num < 0 {
		return -1, fmt.Errorf("Percentage value should be withing 0% to 100%")
	}
	return num, nil
}

func metric(vals ...string) string {
	return strings.Join(vals, ".")
}
