package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"io"
	"os"
	"strconv"
	"strings"
)
type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)
func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	defer file.Close()
	r := bufio.NewScanner(file)
	user := User{}
	seenBrowsers := make(map[string]bool, 200)
	byteAndroid := []byte("Android")
	byteMSIE := []byte("MSIE")

	fmt.Fprintln(out, "found users:")

	for i := 0; r.Scan(); i++ {
		line := r.Bytes()
		if !(bytes.Contains(line, byteAndroid) || bytes.Contains(line, byteMSIE)) {
			continue
		}
		err = user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			switch {
			case strings.Contains(browser, "Android"):
				isAndroid = true
			case strings.Contains(browser, "MSIE"):
				isMSIE = true
			default:
				continue
			}
			seenBrowsers[browser] = true
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.Replace(user.Email, "@", " [at] ", -1)
		fmt.Fprintln(out, "["+strconv.Itoa(i)+"] "+user.Name+" <"+email+">")
	}
	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
func Decode(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

func Encode(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix = ",\"name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	{
		const prefix = ",\"email\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Email))
	}
	{
		const prefix = ",\"browsers\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

func (v User) MarshalJSON(data []byte) ([]byte, error) {
	w := jwriter.Writer{}
	Encode(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	Decode(&r, v)
	return r.Error()
}