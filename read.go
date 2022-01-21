package csvio

import (
	"github.com/l4go/task"
	"io"
)

const defaultBufSize = 4096

const (
	mStart = iota
	mNeutral
	mEscape
	mChange
)

type Reader struct {
	FieldsPerRecord int

	io   io.Reader
	send chan []string
	err  error

	buf  []byte
	cur  int
	esc  uint8
	cols []string

	cancel task.Canceller

	cnf *Config
}

func NewReader(rio io.Reader) (*Reader, error) {
	return NewReaderWithConfig(rio, &DefaultConfig)
}

func NewReaderWithConfig(rio io.Reader, conf *Config) (*Reader, error) {
	if err := conf.Check(); err != nil {
		return nil, err
	}

	dup_conf := &Config{}
	*dup_conf = *conf
	r := &Reader{
		FieldsPerRecord: 0,

		io:     rio,
		send:   make(chan []string),
		err:    nil,
		buf:    nil,
		cur:    0,
		cancel: task.NewCancel(),
		esc:    mStart,
		cnf:    dup_conf,
	}

	go r.start()
	return r, nil
}

func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) Recv() <-chan []string {
	return r.send
}

func (r *Reader) Close() {
	r.cancel.Cancel()
}

func (r *Reader) init_cols() {
	r.cols = []string{}
}

func (r *Reader) init_buf() {
	r.buf = make([]byte, defaultBufSize)
}

func (r *Reader) resize_buf() error {
	newsize := len(r.buf) << 1
	if newsize <= len(r.buf) {
		return ErrFieldSizeLimit
	}
	newbuf := make([]byte, newsize)
	copy(newbuf, r.buf)
	r.buf = newbuf

	return nil
}

func (r *Reader) push() {
	b := r.buf[0:r.cur]
	r.cols = append(r.cols, string(b))
	r.cur = 0
}

func (r *Reader) flush() bool {
	cols := r.cols
	r.init_cols()

	switch {
	case r.FieldsPerRecord < 0:
	case r.FieldsPerRecord == 0:
		r.FieldsPerRecord = len(cols)
	default:
		for len(cols) < r.FieldsPerRecord {
			cols = append(cols, "")
		}
		if len(cols) > r.FieldsPerRecord {
			r.err = ErrFieldCount
			return false
		}
	}

	if len(cols) > 0 {
		select {
		case <-r.cancel.RecvCancel():
			r.err = ErrCancel
			return false
		case r.send <- cols:
		}
	}

	return true
}

func (r *Reader) start() {
	defer close(r.send)

	r.init_cols()
	r.init_buf()
	r.cur = 0
read_loop:
	for r.err == nil {
		if r.cur >= len(r.buf) {
			if err := r.resize_buf(); err != nil {
				r.err = err
				break read_loop
			}
		}
		b := r.buf[r.cur : r.cur+1]
		_, e := io.ReadFull(r.io, b)

		if task.IsCanceled(r.cancel) {
			r.err = ErrCancel
			break read_loop
		}

		switch e {
		default:
			r.err = e
			break read_loop
		case io.EOF:
			switch r.esc {
			case mStart:
			case mNeutral:
				r.push()
				r.flush()
			case mEscape:
				r.err = ErrSyntax
			case mChange:
				r.push()
				r.flush()
			}
			break read_loop
		case nil:
			switch r.esc {
			case mStart:
				r.esc = mNeutral
				fallthrough
			case mNeutral:
				def := func() {
					r.cur++
				}
				switch b[0] {
				case r.cnf.Quote:
					if r.cnf.UseQuote {
						r.esc = mEscape
					} else {
						def()
					}
				case '\r':
					if !r.cnf.UseCRLF {
						def()
					}
				case '\n':
					r.push()
					if !r.flush() {
						break read_loop
					}
					r.esc = mStart
				case r.cnf.Comma:
					r.push()
				default:
					def()
				}
			case mEscape:
				def := func() {
					r.cur++
				}
				switch b[0] {
				case r.cnf.Quote:
					if r.cnf.UseQuote {
						r.esc = mChange
					} else {
						def()
					}
				default:
					def()
				}
			case mChange:
				def := func() {
					r.cur++
					r.esc = mEscape
				}
				switch b[0] {
				case r.cnf.Comma:
					r.push()
					r.esc = mNeutral
				case '\r':
					if !r.cnf.UseCRLF {
						def()
					} else {
						r.esc = mNeutral
					}
				case '\n':
					r.push()
					if !r.flush() {
						break read_loop
					}
					r.esc = mStart
				case r.cnf.Quote:
					def()
				default:
					def()
				}
			}
		}
	}
}
