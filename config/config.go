package config

import (
	"fmt"
	cfg "github.com/mailgun/gotools-config"
	log "github.com/mailgun/gotools-log"
	"github.com/mailgun/vulcanb/model"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Code        int
	Body        string
	ContentType string
	Delay       string
}

type Server struct {
	Addr         string
	Path         string
	Handlers     map[string]map[string]*Response
	ReadTimeout  string
	WriteTimeout string
}

type Config struct {
	Servers []*Server
	Logging []*log.LogConfig
}

type HandlerFn func(http.ResponseWriter, *http.Request)

func ParseConfig(path string) ([]*model.Server, []*log.LogConfig, error) {
	config := Config{}
	if err := cfg.LoadConfig(path, &config); err != nil {
		return nil, nil, fmt.Errorf("Failed to load config file '%s' err:", path, err)
	}
	servers, err := parseServers(config.Servers)
	if err != nil {
		return nil, nil, err
	}
	return servers, config.Logging, nil
}

func parseServers(in []*Server) ([]*model.Server, error) {
	out := make([]*model.Server, len(in))
	for i, cserver := range in {
		serv, err := parseServer(cserver)
		if err != nil {
			return nil, err
		}
		out[i] = serv
	}
	return out, nil
}

func parseServer(in *Server) (*model.Server, error) {
	readTimeout, err := time.ParseDuration(in.ReadTimeout)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := time.ParseDuration(in.WriteTimeout)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	for key, m := range in.Handlers {
		handlerFn, err := buildHandler(m)
		if err != nil {
			return nil, err
		}
		mux.HandleFunc(key, handlerFn)
	}
	return &model.Server{
		Addr:         in.Addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      mux,
	}, nil
}

func buildHandler(distribution map[string]*Response) (HandlerFn, error) {
	responses, err := buildResponses(distribution)
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for {
			for _, r := range responses {
				if r.Delay > 0 {
					time.Sleep(r.Delay)
				}
				w.Header().Set("Content-Type", r.ContentType)
				w.Write([]byte(r.Body))
			}
		}
	}, nil
}

func buildResponses(responses map[string]*Response) ([]*model.Response, error) {
	out := []*model.Response{}
	total := 0
	for percent, re := range responses {
		duration, err := time.ParseDuration(re.Delay)
		if err != nil {
			return nil, fmt.Errorf("Bad delday '%s', should me in form '1s'", re.Delay)
		}
		count, err := parsePercent(percent)
		if err != nil {
			return nil, err
		}
		total += count
		for i := 0; i < count; i += 1 {
			out = append(out, &model.Response{
				Delay:       duration,
				Code:        re.Code,
				Body:        []byte(re.Body),
				ContentType: re.ContentType,
			})
		}
	}
	if total > 100 {
		return nil, fmt.Errorf("Percentages should form 100% in sum")
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
