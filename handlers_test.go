package bot

import (
	"regexp"
	"testing"

	"github.com/go-telegram/bot/models"
)

func findHandler(b *Bot, id string) *handler {
	b.handlersMx.RLock()
	defer b.handlersMx.RUnlock()

	for _, h := range b.handlers {
		if h.id == id {
			return &h
		}
	}

	return nil
}

func Test_match_func(t *testing.T) {
	b := &Bot{}

	var called bool

	id := b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		called = true
		if update.ID != 42 {
			t.Error("invalid update id")
		}
		return true
	}, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{ID: 42})
	if !called {
		t.Error("not called")
	}
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_exact(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(HandlerTypeMessageText, "xxx", MatchTypeExact, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{Message: &models.Message{Text: "zzz"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{Message: &models.Message{Text: "xxx"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_caption_exact(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(HandlerTypePhotoCaption, "xxx", MatchTypeExact, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{Message: &models.Message{Caption: "zzz"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{Message: &models.Message{Caption: "xxx"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_prefix(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(HandlerTypeCallbackQueryData, "abc", MatchTypePrefix, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xabcdef"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "abcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_contains(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(HandlerTypeCallbackQueryData, "abc", MatchTypeContains, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xxabxx"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xxabcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_regexp(t *testing.T) {
	b := &Bot{}

	re := regexp.MustCompile("^[a-z]+")

	id := b.RegisterHandlerRegexp(HandlerTypeCallbackQueryData, re, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "123abc"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "abcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_invalid_type(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(-1, "", -1, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "123abc"}})
	if res {
		t.Error("unexpected true result")
	}
}

func TestBot_RegisterUnregisterHandler(t *testing.T) {
	b := &Bot{}

	id1 := b.RegisterHandler(HandlerTypeCallbackQueryData, "", MatchTypeExact, nil)
	id2 := b.RegisterHandler(HandlerTypeCallbackQueryData, "", MatchTypeExact, nil)

	if len(b.handlers) != 2 {
		t.Fatalf("unexpected handlers len")
	}
	if h := findHandler(b, id1); h == nil {
		t.Fatalf("handler not found")
	}
	if h := findHandler(b, id2); h == nil {
		t.Fatalf("handler not found")
	}

	b.UnregisterHandler(id1)
	if len(b.handlers) != 1 {
		t.Fatalf("unexpected handlers len")
	}
	if h := findHandler(b, id1); h != nil {
		t.Fatalf("handler found")
	}
	if h := findHandler(b, id2); h == nil {
		t.Fatalf("handler not found")
	}
}

func Test_match_exact_game(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(HandlerTypeCallbackQueryGameShortName, "xxx", MatchTypeExact, nil)

	h := findHandler(b, id)
	u := models.Update{
		ID: 42,
		CallbackQuery: &models.CallbackQuery{
			ID:            "1000",
			GameShortName: "xxx",
		},
	}

	res := h.match(&u)
	if !res {
		t.Error("unexpected true result")
	}
}

func Test_match_command_start(t *testing.T) {
	t.Run("anywhere 1, yes", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommand, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	t.Run("anywhere 2, yes", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommand, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	t.Run("anywhere 3, no", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommand, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /bar",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	t.Run("start 1, yes", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartOnly, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	t.Run("start 2, no", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartOnly, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	t.Run("start 3, no", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartOnly, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/bar",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command
	t.Run("start maybe with username suffix 1, yes", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	// correct command, correct username
	t.Run("start maybe with username suffix 2, yes", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo@foo_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 12},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	// correct command, wrong username
	t.Run("start maybe with username suffix 3, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo@other_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 14},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command, with prefix
	t.Run("start maybe with username suffix 4, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command, with prefix, correct username
	t.Run("start maybe with username suffix 5, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo@foo_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 12},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command, with prefix, wrong username
	t.Run("start maybe with username suffix 6, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo@other_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 14},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})


	// wrong command
	t.Run("start maybe with username suffix 7, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/bar",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// wrong command, correct username
	t.Run("start maybe with username suffix 8, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/bar@foo_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 12},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// wrong command, wrong username
	t.Run("start maybe with username suffix 9, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/bar@other_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 14},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// wrong command, with prefix
	t.Run("start maybe with username suffix 10, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /bar",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// wrong command, with prefix, correct username
	t.Run("start maybe with username suffix 11, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /bar@foo_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 12},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// wrong command, with prefix, wrong username
	t.Run("start maybe with username suffix 12, no", func(t *testing.T) {
		b := &Bot{
			username: "foo_bot",
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /bar@other_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 14},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command, no username, correct username
	t.Run("start maybe with username suffix 13, no", func(t *testing.T) {
		b := &Bot{
			// username: "foo_bot", // no username
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo@foo_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 12},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	// correct command, no username, correct username
	t.Run("start maybe with username suffix 14, no", func(t *testing.T) {
		b := &Bot{
			// username: "foo_bot", // no username
		}

		id := b.RegisterHandler(HandlerTypeMessageText, "foo", MatchTypeCommandStartMaybeWithBotUsernameSuffix, nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo@other_bot",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 14},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})
}
