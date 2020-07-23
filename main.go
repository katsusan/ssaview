package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const SSAHTML = "ssa.html"

var (
	host    string
	port    string
	ssafunc string
	gcflags string
	files   []string

	tmpobj string
)

func init() {
	flag.StringVar(&ssafunc, "f", "main", "functions to show")
	flag.StringVar(&gcflags, "args", "", "arguments passed to go compiler")
	flag.StringVar(&host, "h", "127.0.0.1", "host")
	flag.StringVar(&port, "p", "9000", "port")
}

const usage = `SSA viewer.
Usage: ssaview [-f=func] [-args="..."] [-h=host] [-p=port] file

Examples:
$ ssaview -f=main -args="-l" -h=127.0.0.1 -p=9000 hello.go
$ ssaview -f=main hello.go
$ ssaview hello.go 
`

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	//set files to be compiles and generate temp object filename
	files = flag.Args()
	tmpobj = "tmp-" + randFile()

	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "ssaview: %v\n", err)
		os.Exit(2)
	}
}

func doMain() error {
	clean := func() {
		if err := os.Remove(tmpobj); err != nil && checkExist(tmpobj) {
			fmt.Fprintf(os.Stderr, "failed to remove temporary object file,%v\n", err)
		}
		if err := os.Remove(SSAHTML); err != nil && checkExist(SSAHTML) {
			fmt.Fprintf(os.Stderr, "failed to remove ssa.html,%v\n", err)
		}
	}
	//1. use GOSSAFUNC=f go tool compile file to generate ssa.html
	cmd := exec.Command("go", "tool", "compile")
	cmd.Env = append(os.Environ(),
		"GOSSAFUNC="+ssafunc,
	)

	if len(gcflags) != 0 {
		var flg []string //turn space-separated arguments into slice
		flg = strings.Split(gcflags, " ")
		cmd.Args = append(cmd.Args, flg...)
	}
	cmd.Args = append(cmd.Args, "-o", tmpobj)
	cmd.Args = append(cmd.Args, files...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	defer clean()
	if err != nil {
		return fmt.Errorf("go compile failed, %v\n%v", err, out.String())
	}

	if _, err := os.Stat(SSAHTML); err != nil {
		return fmt.Errorf("%s not found, %v", SSAHTML, err)
	}

	//2. serve ssa.html at host:port
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveFile)
	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: mux,
	}

	fmt.Fprintln(os.Stdout, "INFO start processing")
	fmt.Fprintf(os.Stdout, "INFO ssaview is running at http://%s:%s . Press Ctrl-C to stop\n", host, port)

	sigch := make(chan os.Signal, 1)
	srvch := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//fmt.Fprintf(os.Stderr, "http listen failed, %v\n", err)
			srvch <- err
		}
	}()

	signal.Notify(sigch, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigch:
			fmt.Fprintf(os.Stdout, "INFO [%v] received..will exit soon\n", sig)
			if err := srv.Shutdown(context.Background()); err != nil {
				return fmt.Errorf("shutdown server,%v", err)
			}
			return nil
		case err := <-srvch:
			return fmt.Errorf("http listen failed, %v", err)
		}
	}
}

func serveFile(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, SSAHTML)
}

// return true if file is accessible
func checkExist(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

// generate a temporary filename for go tool compile -o,
// inspired by ioutil.TempFile
func randFile() string {
	r := uint32(seed())
	r = r*1664525 + 1013904223
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func seed() int64 {
	return time.Now().Unix() + int64(os.Getpid())
}
