package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/logger"
)

func TestMain(m *testing.M) {
	log = logger.New(true, true, true)
	handler := http.FileServer(http.Dir("./tests"))
	http.Handle("/", handler)
	server := &http.Server{Addr: ":50000", Handler: handler}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic("Server has failed")
		}
	}()

	ret := m.Run()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic("Server shutdown has failed")
	}
	log.Info("After server shutdown")
	os.Exit(ret)
}

func TestGetLinks0(t *testing.T) {
	log.Info("Start TestGetLinks0")
	expected := []string{
		"http://localhost:50000/",
		// Level 0
		"http://localhost:50000/1/index.html",
		"http://localhost:50000/2/index.html",
	}
	depth := 0
	testGetLinks(t, depth, 1*time.Second, expected)
	log.Info("Completed TestGetLinks0")
}

func TestGetLinks1(t *testing.T) {
	log.Info("Start TestGetLinks1")
	expected := []string{
		"http://localhost:50000/",
		// Level 0
		"http://localhost:50000/1/index.html",
		"http://localhost:50000/2/index.html",
		// Level 1
		"http://localhost:50000/1/1/index.html",
		"http://localhost:50000/1/2/index.html",
		"http://localhost:50000/2/1/index.html",
		"http://localhost:50000/2/2/index.html",
	}
	depth := 1
	testGetLinks(t, depth, 1*time.Second, expected)
	log.Info("Completed TestGetLinks1")
}

func TestGetLinks2(t *testing.T) {
	log.Info("Start TestGetLinks2")
	expected := []string{
		"http://localhost:50000/",
		// Level 0
		"http://localhost:50000/1/index.html",
		"http://localhost:50000/2/index.html",
		// Level 1
		"http://localhost:50000/1/1/index.html",
		"http://localhost:50000/1/2/index.html",
		"http://localhost:50000/2/1/index.html",
		"http://localhost:50000/2/2/index.html",
		// Level 2
		"http://localhost:50000/1/1/1/index.html",
		"http://localhost:50000/1/1/2/index.html",
		"http://localhost:50000/1/2/1/index.html",
		"http://localhost:50000/1/2/2/index.html",
		"http://localhost:50000/2/1/1/index.html",
		"http://localhost:50000/2/1/2/index.html",
		"http://localhost:50000/2/2/1/index.html",
		"http://localhost:50000/2/2/2/index.html",
	}
	depth := 2
	testGetLinks(t, depth, 1*time.Second, expected)
	log.Info("Completed TestGetLinks2")
}

func TestGetLinks3(t *testing.T) {
	log.Info("Start TestGetLinks3")
	expected := []string{
		"http://localhost:50000/",
		// Level 0
		"http://localhost:50000/1/index.html",
		"http://localhost:50000/2/index.html",
		// Level 1
		"http://localhost:50000/1/1/index.html",
		"http://localhost:50000/1/2/index.html",
		"http://localhost:50000/2/1/index.html",
		"http://localhost:50000/2/2/index.html",
		// Level 2
		"http://localhost:50000/1/1/1/index.html",
		"http://localhost:50000/1/1/2/index.html",
		"http://localhost:50000/1/2/1/index.html",
		"http://localhost:50000/1/2/2/index.html",
		"http://localhost:50000/2/1/1/index.html",
		"http://localhost:50000/2/1/2/index.html",
		"http://localhost:50000/2/2/1/index.html",
		"http://localhost:50000/2/2/2/index.html",
		// Level 3
		"http://localhost:50000/1/1/1/1/index.html",
		"http://localhost:50000/1/1/1/2/index.html",
		"http://localhost:50000/1/1/2/1/index.html",
		"http://localhost:50000/1/1/2/2/index.html",
		"http://localhost:50000/1/2/1/1/index.html",
		"http://localhost:50000/1/2/1/2/index.html",
		"http://localhost:50000/1/2/2/1/index.html",
		"http://localhost:50000/1/2/2/2/index.html",
		"http://localhost:50000/2/1/1/1/index.html",
		"http://localhost:50000/2/1/1/2/index.html",
		"http://localhost:50000/2/1/2/1/index.html",
		"http://localhost:50000/2/1/2/2/index.html",
		"http://localhost:50000/2/2/1/1/index.html",
		"http://localhost:50000/2/2/1/2/index.html",
		"http://localhost:50000/2/2/2/1/index.html",
		"http://localhost:50000/2/2/2/2/index.html",
	}
	depth := 3
	testGetLinks(t, depth, 1*time.Second, expected)
	log.Info("Completed TestGetLinks3")
}

func testGetLinks(t *testing.T, depth int, timeout time.Duration, expected []string) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	os.Args = []string{"cmd",
		"-url=" + "http://localhost:50000/",
		"-depth=" + strconv.Itoa(depth),
		"-timeout=" + strconv.Itoa(int(timeout)),
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg := cli.ParseFlags()

	links := getAllLinks(cfg)
	for _, link := range links.Values() {
		log.Info("Link:", link)
	}

	for _, v := range expected {
		if !links.Exists(v) {
			log.Info("Test Failed - Missing link:", v)
			t.Fail()
		}
		links.Remove(v)
	}

	if links.Len() != 0 {
		log.Info("Test Failed - Unexpected links:")
		for _, link := range links.Values() {
			log.Info(link)
		}
		t.Fail()
	}
}
