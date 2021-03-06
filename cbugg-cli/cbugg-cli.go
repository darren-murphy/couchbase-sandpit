package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"
)

var queryFlag = flag.String("query", "", "query to execute")
var querySizeFlag = flag.String("numrows", "100", "number of rows to return from query")
var tmplFlag = flag.String("t", "", "Result template")
var tmplFilename = flag.String("T", "", "Display template filename")
var twFlag = flag.Bool("tw", false, "Send template output through tabwriter")
var asJson = flag.Bool("json", false, "dump result as json")

const defaultTmplText = `{{ range $t := . }}{{printf "%s\t\t%s" $t.Id $t.Title}}
{{ end }}`

type User struct {
	Email string
	MD5   string
}

type searchDoc struct {
	Id          string
	Title       string
	Tags        []string
	Description string
	CreatedAt   time.Time `json:"created_at"`
	Creator     User
	ModifiedAt  time.Time `json:"modified_at"`
	ModifiedBy  User      `json:"modified_by"`
	Owner       User
	Private     bool
	Subscribers []User
}

func maybeF(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Config(maybeVal, varname string) string {
	if maybeVal != "" {
		return maybeVal
	}
	tagBytes, err := exec.Command("git", "config", varname).Output()
	if err != nil {
		log.Fatalf("Couldn't get git config value %s: %v", varname, err)
	}
	return strings.TrimSpace(string(tagBytes))
}

func CbuggSearch(query string) ([]searchDoc, error) {
	base, err := url.Parse(Config("", "cbugg.url"))
	if err != nil {
		return nil, err
	}
	apiRelURL, err := url.Parse("/api/search/")
	if err != nil {
		return nil, err
	}
	apiURL := base.ResolveReference(apiRelURL)
	apiURL.RawQuery = url.Values{"query": {query}, "size": {*querySizeFlag}}.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(Config("", "cbugg.user"), Config("", "cbugg.key"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP Error: %v", resp.Status)
	}

	results := struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Doc searchDoc
				}
			}
		}
	}{}

	d := json.NewDecoder(resp.Body)
	err = d.Decode(&results)
	if err != nil {
		return nil, err
	}

	rv := []searchDoc{}
	for _, r := range results.Hits.Hits {
		rv = append(rv, r.Source.Doc)
	}

	return rv, nil
}

func renderTmpl(docs []searchDoc) {
	tmplstr := *tmplFlag
	if tmplstr == "" {
		switch *tmplFilename {
		case "":
			tmplstr = defaultTmplText
		case "-":
			td, err := ioutil.ReadAll(os.Stdin)
			maybeF(err)
			tmplstr = string(td)
		default:
			td, err := ioutil.ReadFile(*tmplFilename)
			maybeF(err)
			tmplstr = string(td)
		}
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"join": func(o string, s []string) string {
			return strings.Join(s, o)
		},
		"trunc": func(maxlen int, s string) string {
			if len(s) < maxlen {
				return s
			}
			return strings.TrimSpace(s[:maxlen])
		},
	}).Parse(tmplstr)
	maybeF(err)

	outf := io.Writer(os.Stdout)

	if *twFlag {
		tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
		defer tw.Flush()
		outf = tw
	}

	maybeF(tmpl.Execute(outf, docs))
}

func renderJson(docs []searchDoc) {
	data, err := json.MarshalIndent(docs, "", "  ")
	maybeF(err)
	os.Stdout.Write(data)
}

func main() {
	flag.Parse()

	res, err := CbuggSearch(*queryFlag)
	maybeF(err)

	if *asJson {
		renderJson(res)
	} else {
		renderTmpl(res)
	}
}
