package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ngc224/txtmsk/command"
	"github.com/ngc224/txtmsk/keystore"
	"github.com/ngc224/txtmsk/mask"
	"golang.org/x/crypto/ssh/terminal"
)

const applicationName = "txtmsk"

func main() {
	os.Exit(exitcode(run()))
}

func exitcode(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	return 0
}

func run() error {
	cmd := command.New()

	if cmd.Version {
		fmt.Println("v" + version)
		return nil
	}

	if cmd.Help {
		cmd.PrintHelp()
		return nil
	}

	key := keystore.New(applicationName)
	pw, err := key.Get()

	if cmd.Password || err != nil {
		pw, err = key.Set()

		if err != nil {
			return err
		}
	}

	var fp io.Reader = os.Stdin

	if len(cmd.Args) > 0 {
		f, err := os.Open(cmd.Args[0])

		if err != nil {
			return err
		}

		defer f.Close()

		fp = f
	} else {
		if terminal.IsTerminal(0) {
			return nil
		}
	}

	var tBuf bytes.Buffer

	sc := bufio.NewScanner(fp)

	for sc.Scan() {
		tBuf.Write(sc.Bytes())
		tBuf.WriteString("\n")
	}

	if sc.Err() != nil {
		return err
	}

	text := strings.TrimRight(tBuf.String(), "\n")

	m, err := mask.New(pw)

	if err != nil {
		return err
	}

	if cmd.UnMask {
		umText, err := tryUnMask(m, text)

		if err != nil {
			if err != mask.ErrNotUseText {
				fmt.Println(text)
			}

			return err
		}

		if cmd.Trim {
			umText = trimInLineTag(umText)
		}

		fmt.Println(umText)

		if umText == text {
			return mask.ErrNotDecrypt
		}

		return nil
	}

	mText, err := tryMask(m, text)

	if err != nil {
		return err
	}

	fmt.Println(mText)
	return nil
}
