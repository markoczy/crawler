package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/types"
)

func TestMain(m *testing.M) {
	handler := http.FileServer(http.Dir("./tests"))
	http.Handle("/", handler)
	server := &http.Server{Addr: ":50000", Handler: handler}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic("Server has failed")
		}
	}()

	// Setting up signal capturing
	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, os.Interrupt)

	ret := m.Run()

	// Waiting for SIGINT (pkill -2)
	// <-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic("Server shutdown has failed")
	}
	log.Println("After server shutdown")
	os.Exit(ret)
}

func TestGetLinks0(t *testing.T) {
	log.Println("Start TestGetLinks0")
	expected := []string{
		// Level 0
		"http://localhost:50000/1/index.html",
		"http://localhost:50000/2/index.html",
	}
	depth := 0
	testGetLinks(t, depth, 1*time.Second, expected)
	log.Println("Completed TestGetLinks0")
}

func TestGetLinks1(t *testing.T) {
	log.Println("Start TestGetLinks1")
	expected := []string{
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
	log.Println("Completed TestGetLinks1")
}

func TestGetLinks2(t *testing.T) {
	log.Println("Start TestGetLinks2")
	expected := []string{
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
	log.Println("Completed TestGetLinks2")
}

func TestGetLinks3(t *testing.T) {
	log.Println("Start TestGetLinks3")
	expected := []string{
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
	log.Println("Completed TestGetLinks3")
}

func testGetLinks(t *testing.T, depth int, timeout time.Duration, expected []string) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getLinksRecursive(ctx, "http://localhost:50000", 1*time.Second, 0, depth, types.NewStringSet())
	for _, link := range links.Values() {
		log.Println("Link:", link)
	}

	for _, v := range expected {
		if !links.Exists(v) {
			log.Println("Test Failed - Missing link:", v)
			t.Fail()
		}
		links.Remove(v)
	}

	if links.Len() != 0 {
		log.Println("Test Failed - Unexpected links:")
		for _, link := range links.Values() {
			log.Println(link)
		}
		t.Fail()
	}
}
