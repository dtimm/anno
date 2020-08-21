package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/core/v1"
)

type Fetcher func() (*v1.PodList, error)

type Config struct {
	Fetcher
	Port int
}

type Proxy interface {
	Start() error
	Stop()
}

func NewProxy(c Config) Proxy {
	return &proxy{
		fetch: c.Fetcher,
		port:  c.Port,
	}
}

type proxy struct {
	fetch Fetcher
	port  int
	srv   *http.Server
}

func (p *proxy) Start() error {
	log.Printf("serving anno-proxy on %d...\n", p.port)

	r := mux.NewRouter()
	r.HandleFunc("/metrics/{appID}", p.serveMetrics)

	p.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", p.port),
		Handler:      r,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	go func() {
		if err := p.srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

func (p *proxy) Stop() {
	log.Println("shutting down...")
	p.srv.Shutdown(context.TODO())
}

func (p *proxy) serveMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pods, err := p.fetch()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, pod := range pods.Items {
		if pod.GetName() == vars["appID"] {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
