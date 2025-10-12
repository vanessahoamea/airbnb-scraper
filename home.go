package main

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/vanessahoamea/airbnb-scraper/utils"
)

type home struct {
	Url             string  `json:"url"`
	Title           string  `json:"title"`
	TypeAndLocation string  `json:"type_and_location"`
	Overview        string  `json:"overview"`
	Description     string  `json:"description"`
	Score           float64 `json:"score"`
}

func (h *home) parse(page *rod.Page) error {
	// close translation modal, if needed
	translationButton, err := page.Timeout(20 * time.Second).Element(utils.TranslationButtonSelector)
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
		h.Url = info.URL
	}

	title, err := extractTextFromElement(page, utils.TitleSelector)
	if err != nil {
		return err
	} else {
		h.Title = title
	}

	typeAndLocation, err := extractTextFromElement(page, utils.TypeAndLocationSelector)
	if err != nil {
		return err
	} else {
		h.TypeAndLocation = typeAndLocation
	}

	overview, err := extractTextFromElement(page, utils.OverviewSelector)
	if err != nil {
		return err
	} else {
		h.Overview = overview
	}

	// click description button
	descriptionButton, err := page.Timeout(20 * time.Second).ElementX(utils.DescriptionButtonSelector)
	if err != nil || descriptionButton == nil {
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, &rod.ElementNotFoundError{}) {
			return err
		} else {
			shortDescription, err := extractTextFromElement(page, utils.ShortDescriptionSelector)
			if err != nil {
				return err
			} else {
				h.Description = shortDescription
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
			h.Description = description
		}
	}

	// compute score
	normalizedText, err := normalizeText(h.Description)
	if err != nil {
		return err
	} else {
		h.Score = h.computeScore(normalizedText)
	}

	return nil
}

func (h *home) computeScore(normalizedText string) float64 {
	positiveMatches, negativeMatches := 0, 0

	for _, keyword := range utils.PositiveKeywords {
		if strings.Contains(normalizedText, keyword) {
			positiveMatches++
		}
	}

	for _, keyword := range utils.NegativeKeywords {
		if strings.Contains(normalizedText, keyword) {
			negativeMatches++
		}
	}

	totalMatches := positiveMatches + negativeMatches

	if totalMatches == 0 {
		return 0.0
	}

	return float64(positiveMatches) / float64(totalMatches)
}

func extractTextFromElement(page *rod.Page, selector string) (string, error) {
	element, err := page.Timeout(20 * time.Second).Element(selector)
	if err != nil || element == nil {
		return "", err
	}

	text, err := element.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

func normalizeText(text string) (string, error) {
	result := strings.ToLower(text)

	alphanumericRegex, err := regexp.Compile("[^a-z0-9 ]+")
	if err != nil {
		return "", err
	}
	result = alphanumericRegex.ReplaceAllString(result, "")

	spaceRegex, err := regexp.Compile(`\s+`)
	if err != nil {
		return "", err
	}
	result = spaceRegex.ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	return result, nil
}
