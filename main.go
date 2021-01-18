package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

type CCases []CCase
type CCase struct {
	CaseNumber  string
	Name        string
	Charges     []string
	Links       []string
	Residency   string
	CaseStatus  []string
	LastUpdated string
}

const baseURL = "https://www.justice.gov"

func Scrape() {
	// Request the HTML page.
	res, err := http.Get("https://www.justice.gov/opa/investigations-regarding-violence-capitol")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	ccases := CCases{}

	doc.Find("tr").Each(func(i int, c *goquery.Selection) {
		ccase := extractCCase(c)
		ccases = append(ccases, ccase)
	})

	d, err := yaml.Marshal(&ccases)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))
}

func main() {
	Scrape()
}

func extractCCase(c *goquery.Selection) CCase {
	ccase := CCase{}
	c.Find("td").Each(func(i int, x *goquery.Selection) {
		switch i {
		case 0:
			ccase.CaseNumber = clean(x.Text())
		case 1:
			ccase.Name = clean(x.Text())
		case 2:
			ccase.Charges = extractTexts(x)
		case 3:
			ccase.Links = extractLinks(x)
		case 4:
			ccase.Residency = clean(x.Text())
		case 5:
			ccase.CaseStatus = extractTexts(x)
		case 6:
			ccase.LastUpdated = clean(x.Text())
		}
	})
	return ccase
}

func extractTexts(s *goquery.Selection) []string {
	texts := []string{}
	s.Find("p").Each(func(i int, x *goquery.Selection) {
		texts = append(texts, clean(x.Text()))
	})
	if len(texts) == 0 {
		texts = append(texts, clean(s.Text()))
	}
	return texts
}

func extractLinks(l *goquery.Selection) []string {
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

//////////////////////////////////////////
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
