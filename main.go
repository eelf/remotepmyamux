package main

import (
	"github.com/hashicorp/yamux"
	"net"
	"flag"
	"io"
	"log"
	"time"
)

func pipe(stream, service net.Conn, label string) {
	go func() {
		written, err := io.Copy(stream, service)
		if err != nil {
			log.Println(label, "to stream", err)
			return
		}
		log.Println(label, "to stream", written)
		stream.Close()
		service.Close()
	}()
	written, err := io.Copy(service, stream)
	if err != nil {
		log.Println("stream to", label, err)
		return
	}
	log.Println("stream to", label, written)
	stream.Close()
	service.Close()
}

func main() {
	var controlNet, controlAddr, localNet, localAddr, serviceNet, serviceAddr string
	var tries, triesSleep uint
	flag.StringVar(&controlNet, "control-net", "tcp", "control network")
	flag.StringVar(&controlAddr, "control-addr", ":2023", "control address")
	flag.StringVar(&localNet, "local-net", "tcp", "local network")
	flag.StringVar(&localAddr, "local-addr", "", "local address")
	flag.StringVar(&serviceNet, "service-net", "tcp", "service network")
	flag.StringVar(&serviceAddr, "service-addr", "localhost:22", "service address")
	flag.UintVar(&tries, "tries", 0, "connect to control tries")
	flag.UintVar(&triesSleep, "tries-sleep", 10, "connect to control sleep between tries (sec)")
	flag.Parse()

	if localAddr != "" {
		controlLn, err := net.Listen(controlNet, controlAddr)
		if err != nil {
			panic(err)
		}
		log.Println("control connection listening", controlLn.Addr().String())
		localLn, err := net.Listen(localNet, localAddr)
		if err != nil {
			panic(err)
		}
		log.Println("exposed remote service at", localLn.Addr().String())
		for {
			controlConn, err := controlLn.Accept()
			if err != nil {
				panic(err)
			}
			log.Println("control connected from", controlConn.RemoteAddr().String())
			session, err := yamux.Server(controlConn, nil)
			if err != nil {
				panic(err)
			}
			go func(session *yamux.Session) {
				for {
					clientConn, err := localLn.Accept()
					if err != nil {
						panic(err)
					}
					log.Println("client connected from", clientConn.RemoteAddr().String())
					stream, err := session.Open()
					if err != nil {
						panic(err)
					}
					go pipe(stream, clientConn, "client")
				}
			}(session)
		}
	} else {
		for tries >= 0 {
			controlConn, err := net.Dial(controlNet, controlAddr)
			if err != nil {
				log.Println("error connecting to control", err)
				tries--
				if tries == 0 {
					break
				}
				time.Sleep(time.Duration(triesSleep) * time.Second)
				continue
			}
			log.Println("connected to control", controlConn.RemoteAddr().String())
			session, err := yamux.Client(controlConn, nil)
			if err != nil {
				panic(err)
			}
			for {
				stream, err := session.Accept()
				if err != nil {
					log.Println("control session accept err:", err)
					break
				}
				serviceConn, err := net.Dial(serviceNet, serviceAddr)
				if err != nil {
					panic(err)
				}
				log.Println("connected to service", serviceConn.RemoteAddr().String())
				go pipe(stream, serviceConn, "service")
			}
		}
	}
}
