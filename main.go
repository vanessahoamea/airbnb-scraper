package main

import "fmt"

func main() {
	var inputUrl string

	fmt.Println("Go to airbnb.com, set your filters on the website, then click on the search button.")
	fmt.Printf("Paste the resulting URL here: ")

	_, err := fmt.Scan(&inputUrl)
	if err != nil {
		fmt.Printf("Invalid input: %v\n", err)
		return
	}

	scraper := scraper{}
	scraper.inputUrl = inputUrl
	scraper.start()
}
