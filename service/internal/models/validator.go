package models

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrBodyEmpty        = errors.New("body have to contains news param")
	ErrTitleLength      = errors.New("title length must be between 1 and 255")
	ErrContentLength    = errors.New("content length must be greater 1")
	ErrCategoriesLength = errors.New("categories length must be greater 1")
)

func (n *NewsCreateForm) Validate() error {
	if utf8.RuneCountInString(n.Title) < 1 || utf8.RuneCountInString(n.Title) > 255 {
		return ErrTitleLength
	}

	if utf8.RuneCountInString(n.Content) < 1 {
		return ErrContentLength
	}

	return nil
}

func (n *NewsCreateForm) Normalize() {
	n.Title = strings.TrimSpace(n.Title)
	n.Content = strings.TrimSpace(n.Content)
}

func (n *NewsEditForm) Validate() error {
	if n.Title == nil && n.Content == nil && n.Categories == nil {
		return ErrBodyEmpty
	}
	if n.Title != nil && (utf8.RuneCountInString(*n.Title) < 1 || utf8.RuneCountInString(*n.Title) > 255) {
		return ErrTitleLength
	}
	if n.Content != nil && utf8.RuneCountInString(*n.Content) < 1 {
		return ErrContentLength
	}
	if n.Categories != nil && len(*n.Categories) < 1 {
		return ErrCategoriesLength
	}

	return nil
}

func (n *NewsEditForm) Normalize() {
	if n.Title != nil {
		trimmed := strings.TrimSpace(*n.Title)
		n.Title = &trimmed
	}

	if n.Content != nil {
		trimmed := strings.TrimSpace(*n.Content)
		n.Content = &trimmed
	}
}
