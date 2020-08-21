package proxy_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dtimm/anno/proxy"
	"github.com/gorilla/mux"
)

type testContext struct {
	cleanup func()
}

func TestProxy(t *testing.T) {
	tc := setup(basicConf())
	defer tc.cleanup()

	t.Run("it returns metrics from existing apps", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/metrics/app-id-test")

		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatal(err)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(b), "test-metric") {
			t.Fatal("test-metric not found")
		}
	})

	t.Run("it returns 404 if nothing is found", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/metrics/not-an-app")

		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(err)
		}
	})
}

func setup(c proxy.Config) *testContext {
	p := proxy.NewProxy(c)

	p.Start()

	r := mux.NewRouter()
	r.HandleFunc("/metrics", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(200)
		fmt.Fprint(rw, "test-metric")
	})
	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}
	go srv.ListenAndServe()

	return &testContext{
		cleanup: func() {
			p.Stop()
			srv.Shutdown(context.TODO())
		},
	}
}

func basicConf() proxy.Config {
	return proxy.Config{
		Fetcher: func() (*v1.PodList, error) {
			return &v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "app-id-test",
						Annotations: map[string]string{
							"prometheus.io/scrape": "true",
							"prometheus.io/path":   "/metrics",
							"prometheus.io/porth":  "8081",
						},
					},
					Status: v1.PodStatus{PodIP: "localhost"},
				}},
			}, nil
		},
		Port: 8080,
	}
}
