package socketmap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
)

// The implementation below was maliciously stolen from https://github.com/d--j/go-socketmap
// I needed to adjust it for graceful shutdown support
// TODO: switch to the original library once it has graceful shutdown support

// MaxSize is the maximum size of a socketmap message in bytes.
const MaxSize = 100000

const (
	colon  = ":"
	comma  = ","
	colonB = ':'
	commaB = ','
)

type Handler func(ctx context.Context, lookup, key string) (result string, found bool, err error)

func write(conn net.Conn, chunks ...[]byte) error {
	rs := make([]io.Reader, 3+len(chunks))
	size := 0
	for i, c := range chunks {
		size += len(c)
		rs[i+2] = bytes.NewReader(c)
	}
	if size > MaxSize {
		return errors.New("data too big")
	}
	rs[0] = strings.NewReader(strconv.FormatInt(int64(size), 10))
	rs[1] = strings.NewReader(colon)
	rs[len(rs)-1] = strings.NewReader(comma)
	_, err := io.Copy(conn, io.MultiReader(rs...))
	return err
}

func read(conn net.Conn) ([]byte, error) {
	sizeBuf := [...]byte{0, 0, 0, 0, 0, 0, 0}
	size := int64(0)
	colonFound := false
	for i := 0; i < len(sizeBuf); i++ {
		_, err := conn.Read(sizeBuf[i : i+1])
		if err != nil {
			return nil, err
		}
		if sizeBuf[i] == colonB {
			size, err = strconv.ParseInt(string(sizeBuf[:i]), 10, 64)
			if err != nil {
				return nil, err
			}
			if size < 0 || size > MaxSize {
				return nil, fmt.Errorf("invalid size %d", size)
			}
			colonFound = true
			break
		}
	}
	if !colonFound {
		return nil, errors.New("colon missing")
	}
	b := make([]byte, size)
	if size > 0 {
		if _, err := io.ReadFull(conn, b); err != nil {
			return nil, err
		}
	}
	lastByte := [1]byte{0}
	if _, err := io.ReadFull(conn, lastByte[:]); err != nil {
		return nil, err
	}
	if lastByte[0] != commaB {
		return nil, fmt.Errorf("expected comma got %c", lastByte[0])
	}
	return b, nil
}

type TempError struct {
	Reason string
}

func (e TempError) Error() string {
	if len(e.Reason) > 0 {
		return fmt.Sprintf("temp error: %s", e.Reason)
	}
	return "temp error"
}

func (TempError) Timeout() bool   { return false }
func (TempError) Temporary() bool { return true }

type TimeoutError struct {
	Reason string
}

func (e TimeoutError) Error() string {
	if len(e.Reason) > 0 {
		return fmt.Sprintf("timeout: %s", e.Reason)
	}
	return "timeout"
}
func (TimeoutError) Timeout() bool   { return true }
func (TimeoutError) Temporary() bool { return true }

type PermanentError struct {
	Reason string
}

func (e PermanentError) Error() string {
	if len(e.Reason) > 0 {
		return fmt.Sprintf("permanent error: %s", e.Reason)
	}
	return "permanent error"
}

func (PermanentError) Timeout() bool   { return false }
func (PermanentError) Temporary() bool { return false }

func handle(ctx context.Context, wg *sync.WaitGroup, conn net.Conn, handler Handler) error {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("closing handler connection", "err", err.Error())
		}
	}()
	defer wg.Done()

	for {
		b, err := read(conn)
		if err != nil {
			return conn.Close()
		}
		parts := strings.SplitN(string(b), " ", 2)
		if len(parts) != 2 {
			return conn.Close()
		}

		result, found, err := handler(ctx, parts[0], parts[1])
		if err := ctx.Err(); err != nil {
			slog.Info("context cancelled, closing connection")
			err = write(conn, []byte("TIMEOUT "), []byte(ctx.Err().Error()))
			if err != nil {
				slog.Error("error closing connection", "msg", err.Error())
				return err
			}
			return nil
		}

		if err != nil {
			switch err.(type) {
			case PermanentError, *PermanentError:
				err = write(conn, []byte("PERM "), []byte(err.Error()))
			case TimeoutError, *TimeoutError:
				err = write(conn, []byte("TIMEOUT "), []byte(err.Error()))
			default:
				err = write(conn, []byte("TEMP "), []byte(err.Error()))
			}
		} else {
			if found {
				err = write(conn, []byte("OK "), []byte(result))
			} else {
				err = write(conn, []byte("NOTFOUND "))
			}
		}
		if err != nil {
			return conn.Close()
		}
	}
}
