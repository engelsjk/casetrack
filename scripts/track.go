package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

type TCases []TCase

type TCase struct {
	CaseNumber  string   `json:"casenumber" yaml:"casenumber"`
	Name        string   `json:"name" yaml:"name"`
	Charges     []string `json:"charges" yaml:"charges"`
	Links       []string `json:"links" yaml:"links"`
	Residency   string   `json:"residency" yaml:"residency"`
	CaseStatus  []string `json:"casestatus" yaml:"casestatus"`
	LastUpdated string   `json:"lastupdated" yaml:"lastupdated"`
}

const baseURL = "https://www.justice.gov"

const outputYAML = "cases.yml"
const outputJSON = "cases.json"

func main() {
	track("/opa/investigations-regarding-violence-capitol")
}

func initialize() TCases {

	tcases := TCases{}

	file, err := os.Open(outputJSON)
	if err != nil {
		return tcases
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return tcases
	}

	err = json.Unmarshal(b, &tcases)
	if err != nil {
		return tcases
	}
	return tcases
}

func document(p string) *goquery.Document {

	u, err := url.Parse(baseURL)
	u.Path = path.Join(u.Path, p)

	res, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func output(tcases TCases) {

	// output YAML
	dYAML, err := yaml.Marshal(&tcases)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(outputYAML, dYAML, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// output JSON
	dJSON, err := json.MarshalIndent(&tcases, "", "  ")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(outputJSON, dJSON, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func track(p string) {

	tcases := initialize()

	doc := document(p)

	doc.Find("tr").Each(func(i int, c *goquery.Selection) {
		if i == 0 {
			return // skip header row
		}

		tcase := extract(c)

		if ok := update(tcases, tcase); ok {
			fmt.Printf("updated : %s : %s\n", tcase.Name, tcase.CaseNumber)
		} else {
			tcases = append(tcases, tcase)
			fmt.Printf("added : %s : %s\n", tcase.Name, tcase.CaseNumber)

		}
	})

	output(tcases)
}

func update(tcases TCases, tcase TCase) bool {

	for i, tc := range tcases {
		if match(tc, tcase) {
			if len(tcase.Links) != 0 {
				tcases[i].Links = tcase.Links
			}
			if len(tcase.Charges) != 0 {
				tcases[i].Charges = tcase.Charges
			}
			if len(tcase.CaseStatus) != 0 {
				tcases[i].CaseStatus = tcase.CaseStatus
			}
			if tcase.LastUpdated != "" {
				tcases[i].LastUpdated = tcase.LastUpdated
			}
			return true
		}
	}

	return false
}

func match(tc1, tc2 TCase) bool {
	if tc1.CaseNumber == "" || tc2.CaseNumber == "" {
		if tc1.Name == "" || tc2.Name == "" {
			return false
		}
		if tc1.Name == tc2.Name {
			return true
		}
	}
	if tc1.CaseNumber == tc2.CaseNumber {
		if tc1.Name != tc2.Name {
			return false
		}
		return true
	}

	return false
}

func extract(c *goquery.Selection) TCase {
	tcase := TCase{}
	c.Find("td").Each(func(i int, x *goquery.Selection) {
		switch i {
		case 0:
			tcase.CaseNumber = clean(x.Text())
		case 1:
			tcase.Name = clean(x.Text())
		case 2:
			tcase.Charges = texts(x)
		case 3:
			tcase.Links = links(x)
		case 4:
			tcase.Residency = clean(x.Text())
		case 5:
			tcase.CaseStatus = texts(x)
		case 6:
			tcase.LastUpdated = clean(x.Text())
		}
	})
	return tcase
}

func texts(s *goquery.Selection) []string {
	texts := []string{}
	s.Find("p").Each(func(i int, x *goquery.Selection) {
		texts = append(texts, clean(x.Text()))
	})
	if len(texts) == 0 {
		texts = append(texts, clean(s.Text()))
	}
	return texts
}

func links(l *goquery.Selection) []string {
	links := []string{}
	l.Find("a").Each(func(i int, x *goquery.Selection) {
		if link, ok := (x.Attr("href")); ok {
			links = append(links, baseURL+clean(link))
		}
	})
	return links
}

func clean(t string) string {
	t = strings.TrimSpace(t)
	t = replaceReplacer(t)
	return t
}

/////////////////////////////////////////////////////////////////////////////////////////////////
// https://stackoverflow.com/questions/52594005/golang-replace-any-and-all-newline-characters?rq=1

func replaceReplacer(s string) string {
	var replacer = strings.NewReplacer(
		"\t", " ",
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
		"\v", " ",
		"\f", " ",
		"\u00A0", " ",
		"\u0085", " ",
		"\u2028", " ",
		"\u2029", " ",
	)
	return replacer.Replace(s)
}
