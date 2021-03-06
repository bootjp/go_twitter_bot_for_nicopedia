package item

import (
	"crypto/tls"
	"time"

	"net/http"

	"log"

	"io/ioutil"

	"strings"
	"unicode"

	"github.com/bootjp/go_twitter_bot_for_nicopedia/domain/bot"
	"github.com/bootjp/go_twitter_bot_for_nicopedia/store"
	"github.com/mmcdole/gofeed"
)

// FilterDate time after item return.
func FilterDate(f []*gofeed.Item, t time.Time) []*gofeed.Item {
	var itm []*gofeed.Item
	for _, elem := range f {
		if elem.PublishedParsed.After(t) {
			itm = append(itm, elem)
		}
	}

	return itm
}

// FilterMarkedAsPost no redis mark as post return item.
func FilterMarkedAsPost(f []*gofeed.Item, r *store.Redis, mode *bot.Behavior) ([]*gofeed.Item, error) {
	var itm []*gofeed.Item
	for _, elem := range f {
		var ng bool
		var err error
		switch mode {
		case bot.NicopetterNewArticle:
			ng, err = r.URLPosted(elem.Link, -1)
		case bot.NicopetterModifyRedirectArticle:
			ng, err = r.URLPosted(elem.Link, 86400)
		case bot.NicopetterNewRedirectArticle:
			ng, err = r.URLPosted(elem.Link, 86400)
		}

		if err != nil {
			return nil, err
		}
		if !ng {
			itm = append(itm, elem)
		}
	}

	return itm, nil
}

// Fetch is got url to fetch and return rss.
func Fetch(URL string) ([]*gofeed.Item, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := http.Client{Transport: tr, Timeout: time.Duration(15 * time.Second)}

	res, err := c.Get(URL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	printOnly := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}
	body = []byte(strings.Map(printOnly, string(body)))

	p := gofeed.NewParser()
	f, err := p.ParseString(string(body[:]))
	if err != nil {
		log.Println(URL)
		return nil, err
	}

	return f.Items, nil
}
