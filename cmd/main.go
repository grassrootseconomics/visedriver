package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"git.defalsify.org/vise.git/engine"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_LANGUAGE_SET = iota + state.FLAG_USERSTART
)

type fsData struct {
	path      string
}

func (fsd *fsData) SetLanguageSelected(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	inputStr := string(input)
	res := resource.Result{}

	switch inputStr {
	case "0":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "eng"
	case "1":
		res.FlagSet = []uint32{state.FLAG_LANG}
		res.Content = "swa"
	default:
	}
	return res, nil
}

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	var persist bool
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.BoolVar(&persist, "persist", false, "use state persistence")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	fp := path.Join(dir, sessionId)
	fs := &fsData{
		path: fp,
	}

	ctx := context.Background()
	en, rs, err := engine.NewSizedEnginee(dir, uint32(size), persist, &sessionId)
	rs.AddLocalFunc("select_language", fs.SetLanguageSelected)

	if err != nil {
		fmt.Fprintf(os.Stderr, "engine create error: %v", err)
		os.Exit(1)
	}
	cont, err := en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}
	if !cont {
		_, err = en.WriteResult(ctx, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dead init write error: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write([]byte{0x0a})
		os.Exit(0)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}

}
