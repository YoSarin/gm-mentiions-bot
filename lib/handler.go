package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/martin-reznik/logger"
)

type Handler struct {
	log     *logger.Log
	storage *Storage
}

func NewHandler(log *logger.Log, storage *Storage) *Handler {
	return &Handler{log, storage}
}

const separator = "|"

var addUsersCommand = regexp.MustCompile("@bot\\s+add\\s*@.*\\s*to\\s+(@[a-zA-Z0-9]+)$")
var removeUsersCommand = regexp.MustCompile("@bot\\s+remove\\s*@.*\\s*from\\s+(@[a-zA-Z0-9]+)$")
var listAliasesCommand = regexp.MustCompile("@bot\\s+list$")
var mentionUsersCommand = regexp.MustCompile("(@[a-z]+)(\\s+.*)?$")

func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	h.log.Info("Handling something")
	decoder := json.NewDecoder(r.Body)
	data := &CallbackData{}
	err := decoder.Decode(data)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.log.Info("data: '%+v'", data)

	if data.SenderType == SenderBot {
		h.log.Info("Triggered by bot, ignoring", data.Text)
		return
	} else if strings.Index(data.Text, "@bot") == 0 {
		if addUsersCommand.Match([]byte(data.Text)) && data.HasMentions() {
			err = h.addUsers(w, r, ps, *data)
		} else if removeUsersCommand.Match([]byte(data.Text)) && data.HasMentions() {
			err = h.removeUsers(w, r, ps, *data)
		} else if listAliasesCommand.Match([]byte(data.Text)) {
			err = h.listAliases(w, r, ps, *data)
		} else {
			h.log.Info("...but we have no idea what")
			p := &PostData{BotId: ps.ByName("token"), Text: "Wrong syntax"}
			err = p.Post()
			if err != nil {
				h.log.Error(err.Error())
			}
		}
	} else if mentionUsersCommand.Match([]byte(data.Text)) {
		err = h.mentionUsers(w, r, ps, *data)
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		p := &PostData{BotId: ps.ByName("token"), Text: fmt.Sprintf("Failure: %v", err)}
		err := p.Post()
		if err != nil {
			h.log.Error(err.Error())
		}
	}
}

func (h *Handler) mentionUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data CallbackData) error {
	alias := mentionUsersCommand.FindStringSubmatch(data.Text)[1]

	users, err := h.storage.getUsers(data, alias)
	if err != nil {
		return err
	}

	h.log.Info("Will mention users %v as %v in group %v", users, alias, data.GroupID)

	loci := make([][]int, len(users))
	mention := []int{1, len(alias)}
	for i := 0; i < len(users); i++ {
		loci[i] = mention
	}
	d := PostData{
		Attachments: []Attachment{
			Attachment{
				Type:    TypeMentions,
				UserIds: users,
				Loci:    loci,
			},
		},
		Text:  fmt.Sprintf(" %v: See above ^", alias),
		BotId: ps.ByName("token"),
	}

	err = d.Post()
	if err != nil {
		return err
	}

	w.Write([]byte("OK"))
	return nil
}

func (h *Handler) removeUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data CallbackData) error {
	alias := removeUsersCommand.FindStringSubmatch(data.Text)[1]

	users, err := h.storage.getUsers(data, alias)
	if err != nil {
		return err
	}

	h.log.Info("Removing users %v from users %v as %v for group %v", data.GetMentionedUsers(), users, alias, data.GroupID)

	if h.storage.saveUsers(data, alias, substractLists(users, data.GetMentionedUsers())) != nil {
		return err
	}

	w.Write([]byte("OK"))
	return nil
}

func (h *Handler) addUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data CallbackData) error {
	alias := addUsersCommand.FindStringSubmatch(data.Text)[1]

	users, err := h.storage.getUsers(data, alias)
	if err != nil {
		return err
	}

	h.log.Info("Adding users %v to users %v as %v for group %v", data.GetMentionedUsers(), users, alias, data.GroupID)

	if h.storage.saveUsers(data, alias, mergeListsUnique(users, data.GetMentionedUsers())) != nil {
		return err
	}

	w.Write([]byte("OK"))
	return nil
}

func (h *Handler) listAliases(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data CallbackData) error {
	aliases, err := h.storage.getAliases(data)
	if err != nil {
		return err
	}

	h.log.Info("Listing aliases (%v) for group %v", aliases, data.GroupID)

	d := PostData{
		Text:  fmt.Sprintf("Aliases: \n%v", strings.Join(aliases, "\n")),
		BotId: ps.ByName("token"),
	}

	err = d.Post()
	if err != nil {
		return err
	}
	w.Write([]byte("OK"))
	return nil
}
