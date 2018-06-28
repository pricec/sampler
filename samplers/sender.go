package samplers

import (
	"fmt"
	"net"

	"golang.org/x/net/context"
)

type Sender struct {
	sampleChan chan Sample  // Internal communication
	conn       *net.UDPConn // Destination for stats
	prefix     string       // Prefix all stats with this string
}

func NewSender(
	ctx context.Context,
	prefix string,
	host string,
	port string,
) (*Sender, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		return nil, err
	}

	udpconn, ok := conn.(*net.UDPConn)
	if !ok {
		return nil, err
	}

	sender := &Sender{
		sampleChan: make(chan Sample),
		conn:       udpconn,
		prefix:     prefix,
	}

	return sender, sender.start(ctx)
}

func (s *Sender) Send(sample Sample) {
	s.sampleChan <- sample
}

func (s *Sender) start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <- ctx.Done():
				close(s.sampleChan)
				if err := s.conn.Close(); err != nil {
					fmt.Printf("Error closing UDP connection: %v\n", err)
				}
				return
			case sample := <- s.sampleChan:
				s.sendSample(sample)
			}
		}
	}()
	return nil
}

func (s *Sender) sendSample(sample Sample) {
	var extension string
	switch(sample.metric) {
	case METRIC_TYPE_COUNTER:
		extension = "c"
	case METRIC_TYPE_SET:
		extension = "s"
	case METRIC_TYPE_GAUGE:
		extension = "g"
	default:
		fmt.Printf("Unrecognized metric type '%v'\n", sample.metric)
		return
	}

	stat := sample.name

	if sample.suffix != "" {
		stat = fmt.Sprintf("%v.%v", stat, sample.suffix)
	}

	stat = fmt.Sprintf("%v:%v|%v\n", stat, sample.value, extension)

	if s.prefix != "" {
		stat = fmt.Sprintf("%v.%v", s.prefix, stat)
	}

	// TODO: debug output fmt.Printf("Sending '%v' to statsd\n", stat)
	if _, err := fmt.Fprintf(s.conn, stat); err != nil {
		fmt.Printf("Error sending '%v' to statsd: %v\n", stat, err)
	}
}
