package onboarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hoisie/mustache"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	"net/http"
)

type Notifier interface {
	NotifyError(request *OnboardRequest) error
	NotifySuccess(request *OnboardRequest) error
}

type notifier struct {
	vapeEndpoint  string
	slackEndpoint string
}

const (
	emailDiscoveryTemplate = "discovery-completion"
	emailErrorTemplate     = "discovery-failure"
)

var (
	slackDiscoveryTemplate *mustache.Template
	slackErrorTemplate     *mustache.Template
)

func init() {
	tmpl, err := mustache.ParseString(slacktmpl.NewCustomer)
	if err != nil {
		panic(err)
	}
	slackDiscoveryTemplate = tmpl

	tmpl, err = mustache.ParseString(slacktmpl.DiscoveryError)
	if err != nil {
		panic(err)
	}
	slackErrorTemplate = tmpl
}

func NewNotifier(vapeEndpoint, slackEndpoint string) *notifier {
	return &notifier{
		vapeEndpoint:  vapeEndpoint,
		slackEndpoint: slackEndpoint,
	}
}

func (n *notifier) NotifySuccess(request *OnboardRequest) error {
	err := n.notifyEmail(request, emailDiscoveryTemplate)
	if err != nil {
		return err
	}

	return n.notifySlack(request, slackDiscoveryTemplate)
}

func (n *notifier) NotifyError(request *OnboardRequest) error {
	err := n.notifyEmail(request, emailErrorTemplate)
	if err != nil {
		return err
	}

	return n.notifySlack(request, slackErrorTemplate)
}

func (n *notifier) notifyEmail(request *OnboardRequest, template string) error {
	if n.vapeEndpoint == "" {
		return nil
	}

	requestJSON, err := json.Marshal(map[string]interface{}{
		"user_id":  request.UserId,
		"template": template,
		"vars":     request,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(n.vapeEndpoint, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return fmt.Errorf("Bad response from Vape notification endpoint: %s", resp.Status)
	}

	response := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&response)
	if err != nil {
		return err
	}

	_, ok := response["user"]
	if !ok {
		return fmt.Errorf("error response from vape")
	}

	return nil
}

func (n *notifier) notifySlack(request *OnboardRequest, template *mustache.Template) error {
	if n.slackEndpoint == "" {
		return nil
	}

	templateVars := make(map[string]interface{})

	j, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, &templateVars)
	if err != nil {
		return err
	}

	body := bytes.NewBufferString(template.Render(templateVars))
	resp, err := http.Post(n.slackEndpoint, "application/json", body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
