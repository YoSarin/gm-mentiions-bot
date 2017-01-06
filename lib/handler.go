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

var addUsersCommand = regexp.MustCompile("@bot\\s+add\\s+@.*\\s+to\\s+(@[a-zA-Z0-9]+)$")
var removeUsersCommand = regexp.MustCompile("@bot\\s+remove\\s+@.*\\s+from\\s+(@[a-zA-Z0-9]+)$")
var mentionUsersCommand = regexp.MustCompile("(@[a-z]+)\\s+.*$")

func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	data := &CallbackData{}
	decoder.Decode(data)

	var err error
	var response *PostData

	if strings.Index(data.Text, "@bot") == 0 {
		response = &PostData{BotId: ps.ByName("token"), Text: "Done"}
		h.log.Info("We have been asked to do something...", data.Text)
		if addUsersCommand.Match([]byte(data.Text)) && data.HasMentions() {
			err = h.addUsers(w, r, ps, data)
		} else if removeUsersCommand.Match([]byte(data.Text)) && data.HasMentions() {
			err = h.removeUsers(w, r, ps, data)
		} else {
			h.log.Info("...but we have no idea what")
			response.Text = "Wrong format"
		}
	} else if mentionUsersCommand.Match([]byte(data.Text)) {
		err = h.mentionUsers(w, r, ps, data)
	} else {
		// do not spam, as there was nothing to do
		response = nil
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		response.Text = fmt.Sprintf("Failed: %v", err.Error())
	}
	if response != nil {
		err = response.Post()
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

func (h *Handler) mentionUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data *CallbackData) error {
	alias := mentionUsersCommand.FindStringSubmatch(data.Text)[1]

	users, err := h.storage.getUsers(data, alias)
	if err != nil {
		return err
	}

	h.log.Info("Will mention users %v as %v in group %v", users, alias, data.GroupID)

	d := PostData{
		Attachments: []Attachment{
			Attachment{
				Type:    TypeMentions,
				UserIds: users,
				Loci:    [][]int{[]int{1, len(alias)}, []int{1, len(alias)}, []int{1, len(alias)}},
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

func (h *Handler) removeUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data *CallbackData) error {
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

func (h *Handler) addUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, data *CallbackData) error {
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
