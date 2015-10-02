package onboarder

import (
	"bytes"
	"encoding/json"
	"github.com/hoisie/mustache"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	"net/http"
)

type Notifier interface {
	NotifyEmail(userID int, template string, vars map[string]interface{}) (map[string]interface{}, error)
	NotifySlack(vars map[string]interface{}) error
}

type notifier struct {
	vapeEndpoint  string
	slackEndpoint string
}

var template *mustache.Template

func NewNotifier(vapeEndpoint, slackEndpoint string) *notifier {
	return &notifier{
		vapeEndpoint:  vapeEndpoint,
		slackEndpoint: slackEndpoint,
	}
}

func (n *notifier) NotifyEmail(userID int, template string, vars map[string]interface{}) (map[string]interface{}, error) {
	if n.vapeEndpoint == "" {
		return nil, nil
	}

	requestJSON, err := json.Marshal(map[string]interface{}{
		"user_id":  userID,
		"template": template,
		"vars":     vars,
	})

	if err != nil {
		return nil, err
	}

	resp, err := http.Post(n.vapeEndpoint, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (n *notifier) NotifySlack(vars map[string]interface{}) error {
	if n.slackEndpoint == "" {
		return nil
	}

	if template == nil {
		tmpl, err := mustache.ParseString(slacktmpl.NewCustomer)
		if err != nil {
			return err
		}
		template = tmpl
	}

	body := bytes.NewBufferString(template.Render(vars))
	resp, err := http.Post(n.slackEndpoint, "application/json", body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
