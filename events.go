package events

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Client to run functions
type Client struct {
	config Config
}

// Config of Client
type Config struct {
	baseUrl string
}

// represents a Event of Kleiner-Kalender
type Event struct {
	Title    string
	Content  string
	URL      string
	Date     string
	Location string
}

func NewClient(config Config) *Client {
	if config.baseUrl == "" {
		config.baseUrl = "https://www.kleiner-kalender.de"
	}
	return &Client{config: config}
}

// parse events of daily page
func parseEventPage(event *Event) error {
	res, err := http.Get(event.URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	// Get content from TextContent div
	var content string
	docContent := doc.Find("div#TextContent")
	docContent.Find("p").Each(func(_ int, e *goquery.Selection) {
		_, hasId := e.Attr("id")
		_, hasClass := e.Attr("class")
		if !hasId && !hasClass {
			content += e.Text()
		}
	})
	event.Content = content

	return nil
}

func (client Client) GetEventsByDate(date time.Time) (events []*Event, _ error) {
	formatedDate := date.Format("2006-01-02")
	formatedUrl := fmt.Sprintf("%s/kalender/%s.html", client.config.baseUrl, formatedDate)
	urlParsed, err := url.Parse(formatedUrl)
	if err != nil {
		return nil, err
	}
	url := urlParsed.String()

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Failed when get event page %s %s", url, res.Status)
	}

	// parse the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	foundEvents := doc.Find("div#ContentBody ul li a")

	if foundEvents.Length() == 0 {
		return nil, fmt.Errorf("No events found on event page %s", url)
	}
	for i := 0; i < foundEvents.Length(); i++ {
		element := foundEvents.Eq(i)
		eventUrl, _ := element.Attr("href")

		event := &Event{
			Title: element.Text(),
			URL:   eventUrl,
		}

		if err := parseEventPage(event); err != nil {
			return nil, fmt.Errorf("Failed to parse event from page %s %v", eventUrl, err)
		}

		events = append(events, event)
	}
	return
}
