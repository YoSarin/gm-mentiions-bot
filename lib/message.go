package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const TypeMentions = "mentions"
const SenderBot = "bot"

type Attachment struct {
	Type    string  `json:"type"`
	UserIds []int   `json:"user_ids,omitempty"`
	Loci    [][]int `json:"loci,omitempty"`
}

type CallbackData struct {
	Attachments []Attachment `json:"attachments,omitempty"`
	GroupID     string       `json:"group_id"`
	Text        string       `json:"text"`
	SenderType  string       `json:"sender_type"`
}

type PostData struct {
	Attachments []Attachment `json:"attachments,omitempty"`
	Text        string       `json:"text"`
	BotId       string       `json:"bot_id"`
}

var Hostname string

func (d *PostData) Post() error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%v/v3/bots/post", Hostname), strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
	}
	return nil
}

func (c *CallbackData) HasMentions() bool {
	for _, a := range c.Attachments {
		if a.Type == TypeMentions {
			return true
		}
	}
	return false
}

func (c *CallbackData) GetMentionedUsers() []int {
	for _, a := range c.Attachments {
		if a.Type == TypeMentions {
			return a.UserIds
		}
	}
	return []int{}
}
