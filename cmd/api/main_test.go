package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	main "github.com/scottyfionnghall/ozonurlshortener/cmd/api"
	"go.uber.org/zap"
)

type Test struct {
	Name     string
	URL      string
	WantCode int
	WantBody string
}

var (
	testPost = []Test{
		{
			Name:     "Valid URL",
			URL:      "https://pkg.go.dev/testing",
			WantCode: http.StatusOK,
		},
		{
			Name:     "Empty URL",
			URL:      "",
			WantCode: http.StatusBadRequest,
		},
		{
			Name:     "Non Valid URL",
			URL:      "testeteststsdtsdtstsdt",
			WantCode: http.StatusBadRequest,
		},
		{
			Name:     "Long URL",
			URL:      "https://www.google.com/search?q=bad+request&oq=bad+req&gs_lcrp=EgZjaHJvbWUqBwgAEAAYgAQyBwgAEAAYgAQyBggBEEUYOTIHCAIQABiABDIHCAMQABiABDIHCAQQABiABDIHCAUQABiABDIHCAYQABiABDIHCAcQABiABDIHCAgQABiABDIHCAkQABiABNIBCDIyNTRqMGo3qAIAsAIA&sourceid=chrome&ie=UTF-8",
			WantCode: http.StatusOK,
		},
	}

	testGet = []Test{
		{
			Name:     "Invalid ID/Non-Existant ID",
			URL:      "ast4o534o",
			WantCode: http.StatusBadRequest,
		},
		{
			Name:     "Empty ID",
			URL:      "",
			WantCode: http.StatusBadRequest,
		},
	}
)

func TestAPI(t *testing.T) {
	t.Run("With DB", func(t *testing.T) {
		runAPI(t, false)
	})

	t.Run("With InMemmory Cache", func(t *testing.T) {
		runAPI(t, true)
	})
}

func runAPI(t *testing.T, inMemory bool) {
	app := newTestApplication(t, inMemory)
	sentPostRequest(t, app)
	sentGetRequest(t, app)
	err := app.Store.ClearTable()
	if err != nil {
		t.Fatal(err)
	}
}

func sentPostRequest(t *testing.T, app *main.APIServer) {
	for _, tt := range testPost {
		t.Run("PostTest", func(t *testing.T) {
			type testURL struct {
				URL string
			}

			url := testURL{URL: tt.URL}

			var buf bytes.Buffer

			err := json.NewEncoder(&buf).Encode(url)

			if err != nil {
				t.Fatal(err)
			}

			request, _ := http.NewRequest(http.MethodPost, "/", &buf)
			response := httptest.NewRecorder()
			(*main.APIServer).PostURL(app, response, request, nil)

			got := response.Code
			want := tt.WantCode
			// Удачные запросы добавляем к будущим
			if got == http.StatusOK {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal("Could not read response body")
				}
				strBody := string(body)
				testGet = append(testGet, Test{
					Name:     tt.Name,
					URL:      strBody,
					WantCode: http.StatusOK,
				})
			}

			if got != want {
				t.Errorf("got %d; want %d;", got, want)
			}
		})
	}
}

func sentGetRequest(t *testing.T, app *main.APIServer) {

	for _, tt := range testGet {
		t.Run("GetTest", func(t *testing.T) {
			type testURL struct {
				URL string
			}

			url := testURL{URL: tt.URL}

			var buf bytes.Buffer

			err := json.NewEncoder(&buf).Encode(url)
			if err != nil {
				t.Fatal(err)
			}

			request, _ := http.NewRequest(http.MethodPost, "/", &buf)
			response := httptest.NewRecorder()

			(*main.APIServer).PostURL(app, response, request, nil)

			got := response.Code
			want := tt.WantCode

			if got != want {
				t.Errorf("got %d; want %d;", got, want)
			}
		})
	}
}

// Протестировать с базой данных в памяти
func newTestApplication(t *testing.T, inMemory bool) *main.APIServer {
	sqlvars := &main.SQLDBVars{
		User:     os.Getenv("TEST_POSTGRES_USER"),
		DBName:   os.Getenv("TEST_POSTGRES_DB"),
		SSLMode:  os.Getenv("TEST_POSTGRES_SSLMODE"),
		Host:     os.Getenv("TEST_POSTGRES_HOST"),
		Port:     os.Getenv("TEST_POSTGRES_PORT"),
		Password: os.Getenv("TEST_POSTGRES_PASSWORD"),
	}
	main.InMemoryDB = inMemory
	logger := zap.Must(zap.NewProduction())
	store, err := main.GetDB(os.Getenv("TEST_API_ADDR"), logger, *sqlvars)
	if err != nil {
		t.Fatal(err)
	}
	return main.NewAPISever(os.Getenv("TEST_API_ADDR"), store, *logger)
}
