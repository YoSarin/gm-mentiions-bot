package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const TypeMentions = "mentions"

type Attachment struct {
	Type    string   `json:"type"`
	UserIds []string `json:"user_ids,omitempty"`
	Loci    [][]int  `json:"loci,omitempty"`
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

func (d *PostData) Post() error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://api.groupme-b.com/v3/bots/post", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	fmt.Print(resp)
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

func (c *CallbackData) GetMentionedUsers() []string {
	for _, a := range c.Attachments {
		if a.Type == TypeMentions {
			return a.UserIds
		}
	}
	return []string{}
}
