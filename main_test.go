package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type EventsByDateTest struct {
	name               string
	config             Config
	date               time.Time
	expectedError      string
	expectedEventTitle []string
}

var testCasesEventsByDate = []EventsByDateTest{
	{
		name:   "sucessfull date old format",
		config: Config{},
		date:   time.Date(2024, time.May, 23, 10, 0, 0, 0, time.UTC),
		expectedEventTitle: []string{
			"Tag des Grundgesetzes",
			"Tag zur Beendigung von Geburtsfisteln",
			"Welt-Schildkröten-Tag",
			"Vollmond Mai",
		},
		expectedError: "",
	},
	{
		name:          "error 404 not found",
		config:        Config{baseUrl: "http://localhost:8080"},
		date:          time.Date(2001, time.September, 11, 10, 0, 0, 0, time.UTC),
		expectedError: "Failed when get event page",
	},
	{
		name:          "sucessfull date new format",
		config:        Config{},
		date:          time.Date(2024, time.November, 23, 10, 0, 0, 0, time.UTC),
		expectedError: "",
		expectedEventTitle: []string{
			"Iss-eine-Cranberry-Tag",
			"Nationaltag der Cashewnuss",
			"Sternzeichen Schütze",
		},
	},
}

func mockServer() (*httptest.Server, error) {
	// create http handler of mock server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/kalender/2001-09-11.html":
			w.WriteHeader(http.StatusInternalServerError)
		case "/kalender/2001-01-23.html":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html>
            <body>
                <div id="test" class=invalid">
                    <span>Text with invalid </span closing tag
                    <a href="http://example.com" onclick="alert('">Broken JS</a>
                    <img src="invalid" onerror="alert('xss')">
            </body>
            </html>
        `))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// create server on fixed port
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil, err
	}

	// add handler function to server
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()

	return server, nil
}

func TestEventsByDate(t *testing.T) {
	mock, err := mockServer()
	defer mock.Close()
	if err != nil {
		t.Fail()
	}
	for _, tc := range testCasesEventsByDate {
		t.Run(tc.name, func(t *testing.T) {
			client := Newclient(tc.config)
			output, err := client.getEventsByDate(tc.date)
			// printEvnets(output)
			switch {
			case tc.expectedError != "":
				assert.ErrorContains(t, err, tc.expectedError)
			case len(tc.expectedEventTitle) > 0:
				for _, event := range output {
					assert.Contains(t, tc.expectedEventTitle, event.Title)
				}
			}
		})
	}
}

func printEvnets(events []*Event) {
	for _, event := range events {
		fmt.Println("Title: ", event.Title)
		fmt.Println("Url: ", event.URL)
		fmt.Println("Content: ", event.Content)
	}
}

type EventsPageTest struct {
	name          string
	url           string
	event         *Event
	expectedError string
}

var testCasesEvent = []EventsPageTest{
	{
		name: "server error",
		event: &Event{
			URL: "http://localhost:8080",
		},
		expectedError: "status code error",
	},
}

func TestParseEventPage(t *testing.T) {
	mock, err := mockServer()
	defer mock.Close()
	if err != nil {
		t.Fail()
	}

	for _, tc := range testCasesEvent {
		err := parseEventPage(tc.event)
		if tc.expectedError != "" {
			assert.ErrorContains(t, err, tc.expectedError)
		}
	}
}
