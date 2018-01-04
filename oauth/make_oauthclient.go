package oauth

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
	"runtime"
	"path/filepath"
	"hash/fnv"
	"strings"
	"os/exec"

	// Can't I use the non-experimental one?
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

)

// GetConfig retrieves configuration from the specified configfile path. 
func GetConfigMap(configfile string)  (map[string]string, error) {

	f, err := os.Open(configfile)
	if err != nil {
		return map[string]string{}, fmt.Errorf("GetConfigMap failed to open %s because %v", configfile, err)
	}
	defer f.Close()

	// I can configure this for more as I discover needing them.
	configmap := map[string]string{
		"client_secret": "",
		"client_id":     "",
	}
	if err := FindProperties(f, configmap); err != nil {
		return map[string]string{}, fmt.Errorf("GetConfigMap failed because %v", err)
	}
	return configmap, nil
}


func FriendlyNewOauthClient(client_id, client_secret, scope string, cacheToken bool, trans http.RoundTripper) (*http.Client, context.Context) {
	ctx := context.Background()
	if trans != http.DefaultTransport {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: trans,
		})
	}

	config := &oauth2.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{scope},
	}
	return newOAuthClient(ctx, config, cacheToken), ctx
}

// newOAuthClient creates a new oauth2 authenticated http.Client given a
// context and configuration.
func newOAuthClient(ctx context.Context, config *oauth2.Config, cacheToken bool) *http.Client {
	cacheFile := tokenCacheFile(config)
	token, err := tokenFromFile(cacheFile, cacheToken)
	if err != nil {
		token = tokenFromWeb(ctx, config)
		saveToken(cacheFile, token)
	} else {
		log.Printf("Using cached token %#v from %q", token, cacheFile)
	}

	return config.Client(ctx, token)
}

// tokenFromWeb gets an OAuth token from the web.
// really? What does this do?
func tokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	log.Printf("Authorize this app at: %s", authURL)
	code := <-ch
	log.Printf("Got code: %s", code)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return token
}


// saveToken saves an OAuth token to the specified cache file.
func saveToken(file string, token *oauth2.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

func osUserCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	log.Printf("TODO: osUserCacheDir on GOOS %q", runtime.GOOS)
	return "."
}

// tokenCacheFile builds a unique file name based on a hash of a
// client id and secret.
func tokenCacheFile(config *oauth2.Config) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientID))
	hash.Write([]byte(config.ClientSecret))
	hash.Write([]byte(strings.Join(config.Scopes, " ")))
	fn := fmt.Sprintf("go-api-demo-tok%v", hash.Sum32())
	return filepath.Join(osUserCacheDir(), url.QueryEscape(fn))
}

// tokenFromFile reads an OAuth token from the specified cache file
// name.
func tokenFromFile(file string, cacheToken bool) (*oauth2.Token, error) {
	if !cacheToken {
		return nil, errors.New("cachetoken is false")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func openURL(url string) {
	try := []string{"xdg-open", "open", "google-chrome"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}

