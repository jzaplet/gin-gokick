package handler

import "gokick/tools/i18n/gocode/testdata/fixture/api"

const KeyTaken api.Key = "user.taken"

const (
	KeyGone api.Key = "user.gone"

	KeyUntranslated api.Key = "user.untranslated"

	prefix = "user."
)

func Handle(name string, key api.Key) []api.Message {
	const keyLocal api.Key = "user.local"
	messages := []api.Message{
		api.Fail(KeyTaken),
		api.Fail(key),
		api.Fail((KeyGone)),
		api.Fail(keyLocal),
		api.Fail(api.KeyInternal),
		{Key: KeyGone},
		api.Fail("user.literal"),
		api.Fail(api.Key(name)),
		api.Fail(api.Key("user.converted")),
		api.Fail(KeyTaken + ".more"),
		api.Fail(prefix + "x"),
		{Key: "user.field"},
	}
	if key == "user.compared" {
		return nil
	}

	return messages
}

func Choose(taken bool) api.Key {
	if taken {
		return KeyTaken
	}

	return "user.returned"
}
