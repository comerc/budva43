package term

import (
	"bufio"
	"fmt"
	"os"

	"github.com/comerc/budva43/app/log"
	"golang.org/x/term"
)

// TODO: Добавить автодополнение команд go-prompt

type Repo struct {
	log *log.Logger
	//
	scanner *bufio.Scanner
}

func New() *Repo {
	return &Repo{
		log: log.NewLogger(),
		//
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (r *Repo) Start() error {
	return nil
}

func (r *Repo) Close() error {
	return nil
}

// HiddenReadLine считывает консоль без отображения введенных символов
func (r *Repo) HiddenReadLine() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), log.WrapError(err) // внешняя ошибка
}

// ReadLine считывает консоль
func (r *Repo) ReadLine() (string, error) {
	if !r.scanner.Scan() {
		return "", log.NewError("scan input failed")
	}
	input := r.scanner.Text()
	return input, nil
}
