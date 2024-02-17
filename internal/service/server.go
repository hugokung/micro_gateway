package service

import (
	"fmt"
	"log"

	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
	"github.com/sourcegraph/conc"
)

var (
	httpServers    = newServerPool[*httpServer]()
)

const (
	actOnStart byte = iota
	actOnStop
	actStart
	actStop
)

type server interface {
	start() error
	stop() error
	services() []Service
}

type serverPool[T server] struct {
	servers map[string]T
}

type baseServer struct {
	ss map[string]Service
}

func (p *serverPool[T]) from(addr string, newServer func() T) T {
	s, exist := p.servers[addr]
	if exist {
		return s
	}
	s = newServer()
	p.servers[addr] = s
	return s
}

func (p *serverPool[T]) startServer(wg *conc.WaitGroup, maxSidSize int) {
	for _, srv := range p.servers {
		ss := srv.services()
		if len(ss) == 0 {
			continue
		}
		startSrv := srv.start
		wg.Go(func() {
			for _, s := range ss {
				colorPrint(actOnStart, s.OnStart(), maxSidSize, s)
			}
			colorPrint(actStart, startSrv(), maxSidSize, ss...)
		})
	}
}

func (p *serverPool[T]) stopServer(maxSidSize int) {
	for _, srv := range p.servers {
		ss := srv.services()
		if len(ss) < 1 {
			return
		}
		for _, s := range ss {
			colorPrint(actOnStop, s.OnStop(), maxSidSize, s)
		}
		colorPrint(actStop, srv.stop(), maxSidSize, ss...)
	}
}

func (p *serverPool[T]) allServices() (ss []Service) {
	for _, srv := range p.servers {
		ss = append(ss, srv.services()...)
	}
	return
}

func (s *baseServer) addService(srv Service) {
	if srv != nil {
		sid := srv.Name() + "@" + srv.Version().String()
		s.ss[sid] = srv
	}
}

func (s *baseServer) services() (ss []Service) {
	for _, s := range s.ss {
		ss = append(ss, s)
	}
	return
}

func newServerPool[T server]() *serverPool[T] {
	return &serverPool[T]{
		servers: make(map[string]T),
	}
}

func newBaseServe() *baseServer {
	return &baseServer{
		ss: make(map[string]Service),
	}
}

func checkServices() (int, int) {
	var ss []Service
	ss = append(ss, httpServers.allServices()...)
	return len(ss), maxSidSize(ss)
}

// maxSidSize max service id string length
func maxSidSize(ss []Service) int {
	length := 0
	for _, s := range ss {
		size := len(s.Name() + "@" + s.Version().String())
		if size > length {
			length = size
		}
	}
	return length
}

func SidStr(name string, version *semver.Version, size int) string {
	return fmt.Sprintf(fmt.Sprintf("%%s@%%-%ds", size-len(name+version.String())+4), name, version)
}

func colorPrint(act byte, err error, l int, ss ...Service) {
	s := ss[0]
	switch act {
	case actOnStart:
		if err == nil {
			// fmt.Fprintf(color.Output, "%s [start] - %s", SidStr(s.Name(), s.Version(), l), s)
			log.Printf("%s [start] - %s", SidStr(s.Name(), s.Version(), l), s)
		} else {
			// fmt.Fprintf(color.Output, "%s [start] - run OnStart error: %s\n", SidStr(s.Name(), s.Version(), l), err)
			log.Printf("%s [start] - run OnStart error: %s\n", SidStr(s.Name(), s.Version(), l), err)
		}
	case actOnStop:
		if err == nil {
			// fmt.Fprintf(color.Output, "%s [stop]  - finish...\n", SidStr(s.Name(), s.Version(), l))
			log.Printf("%s [stop]  - finish...\n", SidStr(s.Name(), s.Version(), l))
		} else {
			// fmt.Fprintf(color.Output, "%s [stop]  - run OnStop error: %s\n", SidStr(s.Name(), s.Version(), l), err)
			log.Printf("%s [stop]  - run OnStop error: %s\n", SidStr(s.Name(), s.Version(), l), err)
		}
	case actStart:
		if err != nil {
			for _, s = range ss {
				// fmt.Fprintf(color.Output, "%s [start] - starting server occurs error: %s\n", SidStr(s.Name(), s.Version(), l), err)
				log.Printf("%s [start] - starting server occurs error: %s\n", SidStr(s.Name(), s.Version(), l), err)	
			}
		}
	case actStop:
		if err != nil {
			for _, s = range ss {
				// fmt.Fprintf(color.Output, "%s [stop] - stopping server occurs error: %s\n", SidStr(s.Name(), s.Version(), l), err)
				log.Printf("%s [stop] - stopping server occurs error: %s\n", SidStr(s.Name(), s.Version(), l), err)
			}
		}
	}
}

// Start start all servers
func Start(wg *conc.WaitGroup) {
	srvSize, maxSidSize := checkServices()
	if srvSize < 1 {
		return
	}

	// some initialize for server engine
	gin.SetMode(lib.GetStringConf("base.debug_mode"))

	// start servers
	httpServers.startServer(wg, maxSidSize)
}

// Stop stop all servers
func Stop() {
	srvSize, maxSidSize := checkServices()
	if srvSize < 1 {
		return
	}
	// stop servers
	httpServers.stopServer(maxSidSize)
}