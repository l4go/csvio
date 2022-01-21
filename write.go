package csvio

import (
	"bytes"
	"io"
)

import (
	"github.com/l4go/task"
)

type Writer struct {
	io     io.WriteCloser
	err    error
	recv   chan []string
	cancel task.Canceller
	done   chan bool

	commaBytes  []byte
	dquoteBytes []byte
	lineBytes   []byte
}

func NewWriter(wio io.WriteCloser) (*Writer, error) {
	return NewWriterWithConfig(wio, &DefaultConfig)
}

func NewWriterWithConfig(wio io.WriteCloser, cnf *Config) (*Writer, error) {
	if err := cnf.Check(); err != nil {
		return nil, err
	}

	commaBytes := []byte{cnf.Comma}
	dquoteBytes := []byte{cnf.Quote}
	if !cnf.UseQuote {
		dquoteBytes = []byte{}
	}
	lineBytes := []byte{'\r', '\n'}
	if !cnf.UseCRLF {
		lineBytes = []byte{'\n'}
	}

	w := &Writer{
		io:     wio,
		recv:   make(chan []string),
		err:    nil,
		cancel: task.NewCancel(),
		done:   make(chan bool),

		commaBytes:  commaBytes,
		dquoteBytes: dquoteBytes,
		lineBytes:   lineBytes,
	}

	go w.start()
	return w, nil
}

func (w *Writer) Err() error {
	return w.err
}

func (w *Writer) Send() chan<- []string {
	return w.recv
}

func (w *Writer) Cancel() {
	close(w.recv)
	w.cancel.Cancel()
}

func (w *Writer) Close() {
	close(w.recv)
	select {
	case <-w.cancel.RecvCancel():
	}
}

func (w *Writer) write_raw_column(b []byte) bool {
	for len(b) > 0 {
		n, e := w.io.Write(b)
		if e != nil {
			w.err = e
			return false
		}
		if task.IsCanceled(w.cancel) {
			w.err = ErrCancel
			return false
		}
		b = b[n:]
	}

	return true
}

func (w *Writer) write_escape_column(b []byte) bool {
	if _, e := w.io.Write(w.dquoteBytes); e != nil {
		w.err = e
		return false
	}
	if task.IsCanceled(w.cancel) {
		w.err = ErrCancel
		return false
	}

	for i, c := range b {
		if c == '"' {
			if _, e := w.io.Write(w.dquoteBytes); e != nil {
				w.err = e
				return false
			}
			if task.IsCanceled(w.cancel) {
				w.err = ErrCancel
				return false
			}
		}

		if _, e := w.io.Write(b[i : i+1]); e != nil {
			w.err = e
			return false
		}
		if task.IsCanceled(w.cancel) {
			w.err = ErrCancel
			return false
		}
	}

	if _, e := w.io.Write(w.dquoteBytes); e != nil {
		w.err = e
		return false
	}
	if task.IsCanceled(w.cancel) {
		w.err = ErrCancel
		return false
	}

	return true
}

func (w *Writer) write_line() bool {
	if _, e := w.io.Write(w.lineBytes); e != nil {
		w.err = e
		return false
	}
	if task.IsCanceled(w.cancel) {
		w.err = ErrCancel
		return false
	}

	return true
}

func (w *Writer) write_delim() bool {
	if _, e := w.io.Write(w.commaBytes); e != nil {
		w.err = e
		return false
	}
	if task.IsCanceled(w.cancel) {
		w.err = ErrCancel
		return false
	}

	return true
}

func (w *Writer) write_column(col []byte) bool {
	if bytes.IndexAny(col, ",\"\r\n") >= 0 {
		if !w.write_escape_column(col) {
			return false
		}
	} else {
		if !w.write_raw_column(col) {
			return false
		}
	}

	return true
}

func (w *Writer) start() {
	defer w.cancel.Cancel()
	defer w.io.Close()

	for cols := range w.recv {
		if len(cols) == 0 {
			continue
		}

		for _, col := range cols[0 : len(cols)-1] {
			if !w.write_column([]byte(col)) {
				return
			}
			if !w.write_delim() {
				return
			}
		}
		if !w.write_column([]byte(cols[len(cols)-1])) {
			return
		}
		if !w.write_line() {
			return
		}
	}
}
