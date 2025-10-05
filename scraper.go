package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/vanessahoamea/airbnb-scraper/utils"
)

type scraper struct {
	inputUrl string
	browser  *rod.Browser
	homes    []*home
}

func (s *scraper) start() {
	fmt.Println("Starting work...")

	// initialize wait group, semaphore channel and browser instance
	wg := sync.WaitGroup{}
	ch := make(chan int, 10)
	s.browser = rod.New().MustConnect()
	defer s.browser.MustClose()

	// access given URL
	s.handlePage(s.inputUrl, 1, &wg, &ch)

	// execute tasks in parallel
	wg.Wait()
	close(ch)
	fmt.Println("Done!")

	// TODO: output results
	// for _, home := range s.homes {
	// 	fmt.Println(home.description)
	// }
}

func (s *scraper) handlePage(url string, count int, wg *sync.WaitGroup, ch *chan int) {
	// navigate to current page
	page := s.browser.MustPage(url)
	defer page.MustClose()

	// wait for elements to load
	time.Sleep(5 * time.Second)
	fmt.Printf("[PAGE %d] Finished loading, now fetching listings\n", count)

	// get all listings
	listings, err := page.Elements(utils.ListingSelector)
	if err != nil || listings.Empty() {
		fmt.Println("[ERROR] Could not find listings on page", count)
		return
	}

	for _, listing := range listings {
		wg.Add(1)
		*ch <- 1
		go s.handleListing(listing, wg, ch)
	}

	// navigate to the next page, if it exists
	nextPageButton, err := page.Element(utils.NextPageButtonSelector)
	if err != nil || nextPageButton == nil {
		fmt.Println("[ERROR] Could not find next page button on page", count)
		return
	}

	nextPageUrl, err := nextPageButton.Property("href")
	if err != nil || nextPageUrl.Nil() {
		fmt.Println("[ERROR] Could not extract next page URL on page", count)
		return
	}

	s.handlePage(nextPageUrl.Str(), count+1, wg, ch)
}

func (s *scraper) handleListing(listing *rod.Element, wg *sync.WaitGroup, ch *chan int) {
	// open listing in new tab
	listingUrl, err := listing.Property("href")
	if err != nil || listingUrl.Nil() {
		fmt.Println("[ERROR] Could not extract home URL")
		return
	}

	page := s.browser.MustPage(listingUrl.Str())

	// defer cleanup code
	defer func() {
		page.MustClose()
		<-*ch
		wg.Done()
	}()

	// set random delay before starting work
	n := rand.Intn(6) + 5
	fmt.Printf("Will start parsing after a delay of %d seconds\n", n)
	time.Sleep(time.Duration(n) * time.Second)

	// parse page content and convert to home object
	home := home{}
	err = home.parse(page)
	if err != nil {
		fmt.Println("[ERROR] Could not parse home listing:", err)
		return
	}

	s.homes = append(s.homes, &home)
}
