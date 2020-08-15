package httpsd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/boynton/conf"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	http.Server
	Conf           *conf.Data
	Hostnames      []string
	Certs          string
	Mux            *http.ServeMux
	HandlerFunc    func(http.ResponseWriter, *http.Request)
	logFile        *os.File
	logger         *log.Logger
	accessLogFile  *os.File
	RedirectServer *http.Server
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wp := NewHttpsdResponseWriter(w)
	server.HandlerFunc(wp, r)
	rhost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		rhost = r.RemoteAddr
	}
	method := r.Method
	path := r.URL.String()
	status := wp.Status
	size := wp.Size
	started := wp.Started
	elapsed := wp.Elapsed
	agent := r.UserAgent()
	referer := r.Referer()
	proto := r.Proto
	user := "nobody" //fix when I add client certs
	server.LogAccess(rhost, user, method, path, proto, agent, referer, status, size, started, elapsed)
}

func (server *Server) ConfigureFromFile(confPath string, handler func(http.ResponseWriter, *http.Request)) error {
	conf, err := conf.FromFile(confPath)
	if err != nil {
		return fmt.Errorf("Cannot read YAML config file (%s): %v\n", confPath, err)
	}
	return server.Configure(conf, handler)
}

func (server *Server) Configure(conf *conf.Data, handler func(http.ResponseWriter, *http.Request)) error {
	var err error
	server.Conf = conf
	server.Hostnames = conf.GetStrings("hostnames", []string{"localhost"})
	server.Certs = conf.GetString("certs", "certs")
	server.HandlerFunc = handler
	accessLogPath := conf.GetString("access_log", "access_log")
	if accessLogPath != "" {
		server.accessLogFile, err = os.OpenFile(accessLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	logPath := conf.GetString("log", "")
	if logPath != "" {
		server.logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		server.logger = log.New(server.logFile, "httpsd: ", log.Lshortfile)
	}
	server.Log("Using %q to cache SSL certs from Let's Encrypt\n", server.Certs)
	cm := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(server.Hostnames...),
		Cache:      autocert.DirCache(server.Certs),
	}
	server.Mux = http.NewServeMux()
	server.Mux.Handle("/", server)
	server.TLSConfig = cm.TLSConfig()
	server.Addr = fmt.Sprintf(":https")
	server.RedirectServer = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      cm.HTTPHandler(nil),
		Addr:         ":http",
	}
	return nil
}

func (server *Server) Serve() error {
	go func() {
		log.Fatal(server.RedirectServer.ListenAndServe())
	}()
	return server.ListenAndServeTLS("", "") //we already have everything set up.
}

func (server *Server) LogAccess(rhost, user, method, path, proto, agent, referer string, status int, size int64, started time.Time, elapsed int64) {
	date := started.Format("02/Jan/2006:15:04:05 -0700")
	line := fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d %q %q [%d ms]\n", rhost, user, date, method, path, proto, status, size, referer, agent, elapsed)
	if server.accessLogFile != nil {
		server.accessLogFile.Write([]byte(line))
	}
}

func (server *Server) Log(format string, args ...interface{}) {
	if server.logger == nil {
		log.Printf(format, args...)
	} else {
		server.logger.Printf(format, args...)
	}
}

type HttpsdResponseWriter struct {
	http.ResponseWriter
	Size    int64
	Started time.Time
	Status  int
	Elapsed int64
}

func NewHttpsdResponseWriter(rw http.ResponseWriter) *HttpsdResponseWriter {
	return &HttpsdResponseWriter{
		ResponseWriter: rw,
		Started:        time.Now(),
	}
}

func (w *HttpsdResponseWriter) Write(buf []byte) (int, error) {
	n, err := w.ResponseWriter.Write(buf)
	w.Size = w.Size + int64(n)
	return n, err
}

func (w *HttpsdResponseWriter) WriteHeader(status int) {
	w.Elapsed = time.Since(w.Started).Milliseconds()
	w.Status = int(status)
	w.ResponseWriter.WriteHeader(status)
}
