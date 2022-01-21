package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/l4go/cmdio"
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

func echo_worker(m *task.Mission, rw io.ReadWriter) {
	defer m.Done()
	log.Println("start: echo worker")
	defer log.Println("end: echo worker")

	csvio_r, err := csvio.NewReader(rw)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer csvio_r.Close()

	for {
		select {
		case clms := <-csvio_r.Recv():
			fmt.Fprintln(rw, ">", clms, len(clms))
		case <-m.RecvCancel():
			return
		}
	}
	if err := csvio_r.Err(); err != nil {
		log.Println("Error:", err)
		return
	}
}
