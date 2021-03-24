package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type event struct {
	datetime time.Time
	title    string
}

type r struct {
	Entity     interface{}            `json:"entity_id,omitempty"`
	State      interface{}            `json:"state,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

var layout string = "20060102"
var url string = "https://www.rottenacker.de/"
var events []event
var tcounter int

const baseurl string = "http://192.168.0.10:8123/api/services/script/" # URL of your Home Assisstant
const token string = "Bearer eyJ0[... approx. 100 more...]nvzZo" # Bearer Token of your Home Assisstant

func applyScene(e string) {
	reqData := r{
		Entity: ("script." + e),
	}
	reqBody, _ := json.Marshal(reqData)
	fmt.Println("reqBody:\n", string(reqBody))
	url := baseurl + e
	fmt.Print(url, "\n")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	fmt.Printf("%s\n", body)
}

func getEvents(document *goquery.Document) {
	document.Find("a[class=url]").Each(func(i int, s *goquery.Selection) {
		layout = "20060102"
		ev, _ := s.Attr("title")
		datetime := s.Find("span[class=dtstart]")
		if ev == "Vorheriger Monat" || ev == "Folgender Monat" {
			return
		}
		btime := strings.ReplaceAll(datetime.Text(), " ", "")
		date, _ := datetime.Attr("title")
		date = string([]rune(date)[:8])
		if btime != "" {
			date = date + " " + btime
			layout = "20060102 15:04"
		}
		tdate, err := time.Parse(layout, date)
		if err != nil {
			fmt.Printf("%s", err)
		}

		events = append(events, event{
			datetime: tdate,
			title:    ev,
		})
	})
}

func getSite(url string) (document *goquery.Document) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err = goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body.", err)
	}
	return document
}

func getCurrEvents() {
	document := getSite((url + "termine.html"))
	getEvents(document)
	for i := 1; i < 2; i++ {
		next, _ := document.Find("div[class=next-button] > a").Attr("href")
		document = getSite((url + next))
		getEvents(document)
	}
}

func main() {
	getCurrEvents()
	now := time.Now()
	tom := now.Add(24 * 8 * time.Hour)
	for _, e := range events {
		if e.datetime.Before(tom) && e.datetime.After(now) {
			if strings.Contains(e.title, "Gelber Sack") {
				fmt.Printf("Event: %s, Date: %v\n", e.title, e.datetime.Format("2006-01-02 15:04"))
				applyScene("gelber_sack")
				tcounter++
			} else if strings.Contains(e.title, "Blauer Tonne") {
				fmt.Printf("Event: %s, Date: %v\n", e.title, e.datetime.Format("2006-01-02 15:04"))
				applyScene("papiermuell")
				tcounter++
			} else if strings.Contains(e.title, "Restmüll") {
				fmt.Printf("Event: %s, Date: %v\n", e.title, e.datetime.Format("2006-01-02 15:04"))
				applyScene("hausmuell")
				tcounter++
			}
		}
	}
	if tcounter == 0 {
		fmt.Printf("No Trash!\n")
		applyScene("notrash")
		applyScene("trashoff")
	}
}

// eof
