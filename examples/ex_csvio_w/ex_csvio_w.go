package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"

	"github.com/l4go/cmdio"
	"github.com/l4go/lineio"
	"github.com/l4go/csvio"
	"github.com/l4go/task"
)

func main() {
	log.Println("START")
	defer log.Println("END")

	m := task.NewMission()

	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)

	std_rw, err := cmdio.StdDup()
	if err != nil {
		defer log.Println("Error:", err)
		return
	}
	go func(cm *task.Mission) {
		defer std_rw.Close()
		echo_worker(cm, std_rw)
	}(m.New())

	select {
	case <-m.Recv():
	case <-signal_ch:
		m.Cancel()
	}
}

func echo_worker(m *task.Mission, rw *cmdio.StdPipe) {
	defer m.Done()
	log.Println("start: echo worker")
	defer log.Println("end: echo worker")

	str_ch := make(chan string)
	defer close(str_ch)

	go func(m *task.Mission) {
		defer m.Done()

		csvio_w, err := csvio.NewWriter(rw)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer csvio_w.Close()
		defer func() {
			if err := csvio_w.Err(); err != nil {
				log.Println("Error:", err)
			}
		}()

		for {
			select {
			case <- m.RecvCancel():
			case str, ok := <- str_ch:
				if !ok {
					return
				}
				clms := strings.SplitN(str, " ", 10)

				select {
				case <- m.RecvCancel():
					return
				case csvio_w.Send() <- clms:
				}
			}
		}
	}(m.New())

	line_r := lineio.NewReader(rw)
	defer func() {
		if err := line_r.Err(); err != nil {
			log.Println("Error:", err)
		}
	}()
	for {
		var ln []byte
		var ok bool
		select {
		case ln, ok = <-line_r.Recv():
		case <-m.RecvCancel():
			return
		}
		if !ok {
			break
		}

		buf_str := string(ln)
		select {
		case <- m.RecvCancel():
			return
		case str_ch <- buf_str:
		}
	}
}
