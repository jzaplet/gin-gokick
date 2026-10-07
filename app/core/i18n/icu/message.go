package icu

import (
	"errors"
	"fmt"
)

type Kind string

const KindText Kind = "text"

const KindPlural Kind = "plural"

const KindSelect Kind = "select"

var ErrArgumentKinds = errors.New("icu: an argument is used as two kinds")

type Message []Part

type Part interface {
	part()
}

type Text string

type Arg struct {
	Name string
}

type Pound struct{}

type Plural struct {
	Name string

	Offset int

	Exact map[int]Message

	Categories map[string]Message
}

type Select struct {
	Name string

	Cases map[string]Message
}

func (Text) part() {}

func (Arg) part() {}

func (Pound) part() {}

func (Plural) part() {}

func (Select) part() {}

func (m Message) Args() (map[string]Kind, error) {
	kinds := map[string]Kind{}

	return kinds, collect(m, kinds)
}

func collect(m Message, kinds map[string]Kind) error {
	for _, part := range m {
		var err error

		switch part := part.(type) {
		case Arg:
			err = use(kinds, part.Name, KindText)
		case Plural:
			err = errors.Join(use(kinds, part.Name, KindPlural), collectAll(part.Exact, kinds), collectAll(part.Categories, kinds))
		case Select:
			err = errors.Join(use(kinds, part.Name, KindSelect), collectAll(part.Cases, kinds))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func collectAll[K comparable](cases map[K]Message, kinds map[string]Kind) error {
	for _, m := range cases {
		if err := collect(m, kinds); err != nil {
			return err
		}
	}

	return nil
}

func use(kinds map[string]Kind, name string, kind Kind) error {
	if used, ok := kinds[name]; ok && used != kind {
		return fmt.Errorf("%w: %s is %s and %s", ErrArgumentKinds, name, used, kind)
	}

	kinds[name] = kind

	return nil
}
