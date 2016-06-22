package main

type addReceiver struct {
	receiver *Client
}

type removeReceiver struct {
	receiver *Client
}

type sendData struct {
	data interface{}
}

type Producer struct {
	receivers map[*Client]struct{}
	events    chan interface{}
}

func NewProducer() *Producer {
	p := &Producer{
		receivers: make(map[*Client]struct{}),
		events:    make(chan interface{}),
	}
	go p.run()
	return p
}

func (p *Producer) AddReceiver(e *Client) {
	p.events <- &addReceiver{
		receiver: e,
	}
}

func (p *Producer) RemoveReceiver(e *Client) {
	//debug.PrintStack()

	p.events <- &removeReceiver{
		receiver: e,
	}
}

func (p *Producer) Send(data interface{}) {
	p.events <- &sendData{
		data: data,
	}
}

func (p *Producer) run() {
	for e := range p.events {
		switch e := e.(type) {
		case *addReceiver:
			p.receivers[e.receiver] = struct{}{}
		case *removeReceiver:
			delete(p.receivers, e.receiver)
		case *sendData:
			for r := range p.receivers {
				//r <- e.data
				select {
				case r.eventChan <- e.data:
				default:
					delete(p.receivers, r)
					r.Overflow()
				}
			}
		}
	}
}

func (p *Producer) Close() {
	close(p.events)
}
