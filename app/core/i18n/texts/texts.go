package texts

import (
	"errors"
	"fmt"

	"golang.org/x/text/language"

	"gokick/app/core/i18n/icu"
)

var ErrUnknownKey = errors.New("texts: unknown key")

var ErrPairs = errors.New("texts: params are not pairs of a name and a value")

type Texts struct {
	messages map[string]icu.Message

	formatter *icu.Formatter
}

func New(tag language.Tag, dictionary map[string]string) (*Texts, error) {
	messages := make(map[string]icu.Message, len(dictionary))
	for key, source := range dictionary {
		message, err := icu.Parse(source)
		if err != nil {
			return nil, fmt.Errorf("texts: %s: %w", key, err)
		}

		messages[key] = message
	}

	return &Texts{messages: messages, formatter: icu.NewFormatter(tag)}, nil
}

func (t *Texts) Text(key string, pairs ...any) (string, error) {
	message, ok := t.messages[key]
	if ok == false {
		return "", fmt.Errorf("%w: %s", ErrUnknownKey, key)
	}

	args, err := argsOf(pairs)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, key)
	}

	text, err := t.formatter.Format(message, args)
	if err != nil {
		return "", fmt.Errorf("texts: %s: %w", key, err)
	}

	return text, nil
}

func argsOf(pairs []any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, ErrPairs
	}

	args := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		name, ok := pairs[i].(string)
		if ok == false {
			return nil, ErrPairs
		}

		args[name] = pairs[i+1]
	}

	return args, nil
}
