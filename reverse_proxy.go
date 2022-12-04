package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func redirectToApi(w http.ResponseWriter, r *http.Request) {

}

func redirectToWebsite(w http.ResponseWriter, r *http.Request) {
	URL, _ := url.Parse("http://localhost:" + os.Getenv("WEBSITE_PORT"))
	r.URL.Scheme = URL.Scheme
	r.URL.Host = URL.Host
	r.URL.Path = singleJoiningSlash(URL.Path, r.URL.Path)
	r.RequestURI = ""

	r.Header.Set("X-Forwarded-For", r.RemoteAddr)

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":" + os.Getenv("INCOMING_PORT"),
		Handler: mux,
	}

	mux.HandleFunc("/", redirectToWebsite)
	mux.HandleFunc("/api", redirectToApi)

	err = s.ListenAndServeTLS(os.Getenv("CERTIFICATE_LOCATION"), os.Getenv("KEY_LOCATION"))
	if err != nil {
		log.Fatal(err)
	}
}
