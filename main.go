package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gbin/goncurses"
)

const (
	NORMAL = iota
	INSERT
	VISUAL
)

type State struct {
	key    goncurses.Key
	buf    TextBuffer
	status int
	window *goncurses.Window
	y      int
	x      int
}

const pad = 2

// gap buffer
type TextBuffer interface {
	ReadAll() string
	Write(string) error
	WriteChar(rune) error
	Delete() error
	ChangeCursorPosition(int, int) error
	LineLength() int
	Select(int, int) error
}

type TextGapBuffer struct {
	left  strings.Builder
	right strings.Builder
}

func NewTextGapBuffer(text string) (*TextGapBuffer, error) {
	tgb := &TextGapBuffer{}
	_, err := tgb.left.WriteString(text)
	return tgb, err
}

func (tgb *TextGapBuffer) ReadAll() string {
	return tgb.left.String() + tgb.right.String()
}

func (tgb *TextGapBuffer) Write(text string) error {
	_, err := tgb.left.WriteString(text)
	return err
}

func (tgb *TextGapBuffer) WriteChar(char rune) error {
	_, err := tgb.left.WriteString(string(char))
	return err
}

func (tgb *TextGapBuffer) Delete() error {
	s := tgb.left.String()
	tgb.left.Reset()
	if len(s) > 0 {
		_, err := tgb.left.WriteString(s[:len(s)-1])
		return err
	}
	return nil
}

func (tgb *TextGapBuffer) ChangeCursorPosition(y int, x int) error {
	return errors.New("not implemented")
}

func (tgb *TextGapBuffer) LineLength() int {
	s := tgb.left.String()
	i := strings.LastIndex(s, "\n")
	if i == -1 { // on first line
		return len(s)
	} else {
		return len(s) - i
	}
}

func (tgb *TextGapBuffer) Select(from int, to int) error {
	return errors.New("not implemented")
}

// n is relative movement
func MoveX(s *State, n int) {
	_, x := s.window.CursorYX()
	_, maxX := s.window.MaxYX()
	ll := s.buf.LineLength()
	s.x = x + n

	if s.x < 0 {
		s.x = 0
	} else if s.x >= maxX {
		s.x = maxX - 1
	} else if s.x >= ll {
		s.x = ll
	}
}

// n is relative movement
func MoveY(s *State, n int) {
	y, _ := s.window.CursorYX()
	maxY, _ := s.window.MaxYX()
	s.y = y + n

	if s.y < 0 {
		s.y = 0
	} else if s.y >= maxY-1 {
		s.y = maxY - 2
	}
}

func PrintError(w *goncurses.Window, e error) {
	if e == nil {
		return
	}
	maxY, _ := w.MaxYX()
	w.MovePrint(maxY-1, pad, e)
}

func PrintStatus(w *goncurses.Window, status int) {
	var msg string
	switch status {
	case NORMAL:
		msg = "[NORMAL]"
	case INSERT:
		msg = "[INSERT]"
	case VISUAL:
		msg = "[VISUAL]"
	}
	maxY, _ := w.MaxYX()
	w.MovePrint(maxY-1, pad, msg)
}

func PrintInfo(w *goncurses.Window, args ...interface{}) {
	maxY, maxX := w.MaxYX()
	s := fmt.Sprint(args...)
	w.MovePrint(maxY-1, maxX-len(s)-pad, s)
}

func main() {
	src, err := goncurses.Init()
	if err != nil {
		log.Fatal("Error initializing curses. ", err)
	}
	defer goncurses.End()
	goncurses.Echo(false)

	buf, err := NewTextGapBuffer("Focus on the donut, not the hole")
	if err != nil {
		log.Fatal("Error initializing gap buffer. ", err)
	}

	var state = &State{
		key:    0,
		buf:    buf,
		status: NORMAL,
		window: src,
	}
	var keyerr error

	calls := 0
	for {
		src.Erase()
		calls += 1
		src.Print(buf.ReadAll())
		PrintInfo(src, state.key, calls)
		PrintStatus(src, state.status)
		PrintError(src, keyerr)
		src.Move(state.y, state.x)
		src.Refresh()

		state.key = src.GetChar()

		switch state.status {
		case NORMAL:
			keyerr = HandleNormal(state)
		case INSERT:
			keyerr = HandleInsert(state)
		case VISUAL:

		}
	}
}

func HandleNormal(s *State) error {
	var err error
	switch s.key {
	case goncurses.KEY_IC, 105:
		s.status = INSERT
	case 104: // h
		MoveX(s, -1)
	case 106: // j
		MoveY(s, 1)
	case 107: // k
		MoveY(s, -1)
	case 108: // l
		MoveX(s, 1)
	}
	return err
}

func HandleInsert(s *State) error {
	var err error
	switch s.key {
	case 27: // escape
		s.status = NORMAL
	case goncurses.KEY_RETURN, goncurses.KEY_ENTER:
		err = s.buf.WriteChar('\n')
	case goncurses.KEY_BACKSPACE, 127:
		err = s.buf.Delete()
	default:
		err = s.buf.Write(goncurses.KeyString(s.key))
	}
	return err
}
