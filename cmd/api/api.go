package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/scottyfionnghall/ozonurlshortener/internal/linkgen"
	"github.com/scottyfionnghall/ozonurlshortener/internal/storage"
	"go.uber.org/zap"
)

// Функция запускает API сервер и инициализирует в mux необходимые эндпойнты.
func (s *APIServer) Run() error {

	s.Router.GET("/", s.GetURL)
	s.Router.POST("/", s.PostURL)

	return http.ListenAndServe(s.ListenAddr, s.Router)
}

// Хэндлер для GET запросов. Он направляет запрос в базу данных используя сокращённую
// ссылку как параметр, если база данных ничего не находит то возвращает Not Found.
func (s *APIServer) GetURL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shortenUrl, err := s.parseBody(w, r)
	if err != nil {
		s.badRequest(w, err, "error while parsing url", r)
		return
	}
	_, path, err := splitURL(shortenUrl)
	if err != nil {
		s.badRequest(w, err, "error while splitting url", r)
		return
	}
	url, err := s.Store.ReturnURL(r.Context(), path[1:])
	if err != nil {
		s.notFound(w, err, "short not found url:"+path[1:], r)
		return
	}

	s.logRequest(r)
	fmt.Fprint(w, url.Domain+url.OriginalPath)
}

// Хэндлер для POST запросов. Он проверяет тело запроса, если тело не отвечает
// требованиям структуры requsetURL то он возвращает ошибку Bad Request. После
// чего функция проверяет не является ли тело пустым, проверяет переданный URL
// на валидность используя Regex. Если все проверки пройдены, функция генериурет
// рандомный ID, создаёт объект типа URL, записывает его в базы данных и возвращает
// сокращённую ссылку.

func (s *APIServer) PostURL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	url, err := s.parseBody(w, r)
	if err != nil {
		s.badRequest(w, err, "error while parsing url", r)
		return
	}

	domain, path, err := splitURL(url)
	if err != nil {
		s.badRequest(w, err, "error while splitting url", r)
		return
	}

	if path[1:] == "" {
		s.badRequest(w, err, "empty path; nothing to shorten:"+url, r)
		return
	}

	exists, err := s.Store.CheckExists(domain, path)
	if err != nil {
		if err.Error() != "not found" {
			s.serverError(w, err, r)
			return
		}
	}

	if exists != nil {
		fmt.Fprintf(w, exists.Domain+"/"+exists.ShortenPath)
		return
	}

	shortUrl := linkgen.GenerateShortURL()

	err = s.Store.AddURL(r.Context(), &storage.URL{
		ShortenPath:  shortUrl,
		Domain:       domain,
		OriginalPath: path,
	})

	if err != nil {
		s.serverError(w, err, r)
		return
	}

	s.logRequest(r)
	fmt.Fprintf(w, domain+"/"+shortUrl)
}

// Паттерн regex чтобы проверить получаемый URL на то является ли он URL.
var pattern = `(http(s)?:\/\/.)?(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&\/=]*)`

type APIServer struct {
	Router     *httprouter.Router
	ListenAddr string
	Store      storage.Storage
	Logger     zap.Logger
}

// Функция создаёт новый API сервер, возвращая ссылку на объект APIServer.
// В качестве параметров используются listenAddr (string) который определяет hostname и
// порт на котором будет работать сервер, store(storage.Storage) ссылка на объект
// удовлетворяющий  интерфейсу storage.Storage который определяет все нужные функции
// для работы с СУБД и logger(zap.Logger) ссылка на zap логгер.

func NewAPISever(listenAddr string, store storage.Storage, logger zap.Logger) *APIServer {
	router := httprouter.New()

	return &APIServer{Router: router, ListenAddr: listenAddr, Store: store, Logger: logger}
}

func splitURL(url string) (string, string, error) {

	domain := ""
	parts := strings.Split(url, "/")

	if len(parts) == 1 {
		return "", "", fmt.Errorf("error while converting url")
	}

	switch parts[0] {
	case "http:":
		domain = "http://" + parts[2]
	case "https:":
		domain = "https://" + parts[2]
	}

	path, ok := strings.CutPrefix(url, domain)
	if !ok {
		return "", "", fmt.Errorf("error while converting url")
	}

	return domain, path, nil
}

func (s *APIServer) parseBody(w http.ResponseWriter, r *http.Request) (string, error) {
	requestUrl := new(storage.Requset)

	err := json.NewDecoder(r.Body).Decode(requestUrl)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if requestUrl.URL == "" {
		return "", fmt.Errorf("empty url")
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	if !reg.MatchString(requestUrl.URL) {
		return "", fmt.Errorf("non valid url" + requestUrl.URL)
	}

	return requestUrl.URL, nil
}
