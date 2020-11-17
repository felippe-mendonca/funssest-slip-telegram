package funssest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

const (
	collectorParallelism = 4
	slipURL              = "https://funssest.tubarao.com.br/funssestonline/v1/cgi-bin/boleto_bancario.asp?menu2=46"
	slipLinkText         = "imprimir boleto"
)

const (
	slipNumberPos = 5
	slipDueDate   = 126
	slipValue     = 198
)

var (
	headers = map[string]string{
		"Connection":      "keep-alive",
		"encoding":        "ISO-8859-1",
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64, AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Origin":          "https://funssest.tubarao.com.br",
		"Sec-Fetch-Site":  "same-origin",
		"Sec-Fetch-Mode":  "navigate",
		"Sec-Fetch-Dest":  "frame",
		"Accept-Language": "pt-BR,pt;q=0.9,en-US;q=0.8:en;q=0.7",
	}
)

type FunssestSlip struct {
	URL     string
	Number  string
	Amount  string
	DueDate string
}

func (fs *FunssestSlip) SetURL(url string) {
	fs.URL = url
}

func (fs *FunssestSlip) SetNumber(number string) {
	re := regexp.MustCompile("[0-9]+")
	fs.Number = strings.Join(re.FindAllString(number, -1), "")
}

func (fs *FunssestSlip) SetAmount(amount string) {
	amount = strings.TrimSpace(amount)
	fs.Amount = fmt.Sprintf("R$%s", strings.ReplaceAll(amount, ".", ","))
}

func (fs *FunssestSlip) SetDueDate(date string) {
	fs.DueDate = strings.TrimSpace(date)
}

func (fs FunssestSlip) Markdown() string {
	return fmt.Sprintf(`
<b>Valor:</b> %s
<b>Vencimento:</b> %s`, fs.Amount, fs.DueDate)
}

func makeCollector() *colly.Collector {

	c := colly.NewCollector(
		colly.AllowedDomains("funssest.tubarao.com.br"),
		colly.MaxDepth(1),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: collectorParallelism})

	c.OnRequest(func(r *colly.Request) {
		for key, value := range headers {
			r.Headers.Set(key, value)
		}
	})

	return c
}

func GetURLs(cpf string) (urls []string, err error) {

	urls = make([]string, 0)
	c := makeCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, slipLinkText) {
			slipURL := e.Request.AbsoluteURL(e.Attr("href"))
			urls = append(urls, slipURL)
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		err = e
	})

	payload := []byte(fmt.Sprintf("cpf_busca=%s", cpf))
	if err = c.PostRaw(slipURL, payload); err != nil {
		return urls, err
	}

	return urls, nil
}

func GetSlips(cpf string) (slips []FunssestSlip, err error) {

	slips = make([]FunssestSlip, 0)
	urls, err := GetURLs(cpf)
	if err != nil {
		return slips, err
	}

	for _, url := range urls {
		slip := FunssestSlip{URL: url}
		c := makeCollector()

		c.OnHTML("div[id=\"Layer0\"]", func(e *colly.HTMLElement) {
			e.ForEach("table tr td", func(pos int, td *colly.HTMLElement) {
				switch pos {
				case slipNumberPos:
					slip.SetNumber(td.Text)
				case slipDueDate:
					slip.SetDueDate(td.Text)
				case slipValue:
					slip.SetAmount(td.Text)
				}
			})
		})

		c.OnError(func(r *colly.Response, e error) {
			err = e
		})

		if err := c.Visit(url); err != nil {
			return slips, err
		}

		slips = append(slips, slip)
	}

	return slips, nil
}
