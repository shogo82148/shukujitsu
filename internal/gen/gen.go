package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var output = flag.String("output", "dates.go", "output path")

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	var src io.Reader
	if len(os.Args) == 1 {
		u := "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv"
		fmt.Printf("download %s\n", u)
		res, err := http.Get(u)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("status code:%d", res.StatusCode)
		}
		defer res.Body.Close()
		src = res.Body
	} else {
		f, err := os.Open(os.Args[1])
		if err != nil {
			return err
		}
		defer f.Close()
		src = f
	}

	r := csv.NewReader(transform.NewReader(src, japanese.ShiftJIS.NewDecoder()))
	// drop first line
	if _, err := r.Read(); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`// generated by internal/gen/gen.go DO NOT EDIT`)
	buf.WriteString("\npackage shukujitsu")
	buf.WriteString("\nvar dates = map[string]string{")
	for {
		records, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		l := fmt.Sprintf("\n\"%s\": \"%s\",", records[0], records[1])
		buf.WriteString(l)
	}
	buf.WriteString("\n}")

	formated, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	of, err := os.Create(*output)
	if err != nil {
		return err
	}

	if _, err := of.Write(formated); err != nil {
		return err
	}

	defer func() {
		fmt.Printf("created %s\n", *output)
	}()
	return of.Close()
}