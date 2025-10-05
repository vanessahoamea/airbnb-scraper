package main

import (
	"context"
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/vanessahoamea/airbnb-scraper/utils"
)

type home struct {
	url             string
	title           string
	typeAndLocation string
	overview        string
	description     string
	score           float64 // positive keyword count divided by total keyword count
}

func (h *home) parse(page *rod.Page) error {
	// close translation modal, if needed
	translationButton, err := page.Timeout(10 * time.Second).Element(utils.TranslationButtonSelector)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, &rod.ElementNotFoundError{}) {
			return err
		}
	} else if translationButton != nil {
		err = translationButton.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			return err
		}
	}

	// populate accessible fields (URL, title, type and location, overview)
	info, err := page.Info()
	if err != nil || info == nil {
		return err
	} else {
		h.url = info.URL
	}

	title, err := extractTextFromElement(page, utils.TitleSelector)
	if err != nil {
		return err
	} else {
		h.title = title
	}

	typeAndLocation, err := extractTextFromElement(page, utils.TypeAndLocationSelector)
	if err != nil {
		return err
	} else {
		h.typeAndLocation = typeAndLocation
	}

	overview, err := extractTextFromElement(page, utils.OverviewSelector)
	if err != nil {
		return err
	} else {
		h.overview = overview
	}

	// click description button
	descriptionButton, err := page.Timeout(10 * time.Second).ElementX(utils.DescriptionButtonSelector)
	if err != nil || descriptionButton == nil {
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, &rod.ElementNotFoundError{}) {
			return err
		} else {
			shortDescription, err := extractTextFromElement(page, utils.ShortDescriptionSelector)
			if err != nil {
				return err
			} else {
				h.description = shortDescription
			}
		}
	} else {
		err = descriptionButton.ScrollIntoView()
		if err != nil {
			return err
		}

		err = descriptionButton.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			return err
		}

		description, err := extractTextFromElement(page, utils.DescriptionModalSelector)
		if err != nil {
			return err
		} else {
			h.description = description
		}
	}

	// TODO: compute score
	h.score = 0

	return nil
}

func extractTextFromElement(page *rod.Page, selector string) (string, error) {
	element, err := page.Element(selector)
	if err != nil || element == nil {
		return "", err
	}

	text, err := element.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}
