package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	Start()
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

func (p *proxy) Start() {
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
			u := getPath(pod.Status.PodIP, pod.GetAnnotations())
			resp, err := http.Get(u.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			io.Copy(w, resp.Body)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func getPath(ip string, annotations map[string]string) url.URL {
	u := url.URL{
		Scheme: "http",
		Path:   ip,
	}
	for k, v := range annotations {
		if strings.Contains(k, "prometheus.io/port") {
			u.Host = fmt.Sprintf("%s:%s", ip, v)
		}
		if strings.Contains(k, "prometheus.io/path") {
			u.Path = v
		}
	}

	return u
}
