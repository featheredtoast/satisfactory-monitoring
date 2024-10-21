package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type HostWebhookPair struct {
	Host string
	Webhook string
}

func main() {
	frmHostname, _ := os.LookupEnv("FRM_HOST")
	frmPort, _ := os.LookupEnv("FRM_HOST")
	frmHostnames, _ := os.LookupEnv("FRM_HOSTS")
	discordWebhook, _ := os.LookupEnv("DISCORD_WEBHOOK")
	discordWebhooks, _ := os.LookupEnv("DISCORD_WEBHOOKS")
	input, _ := os.LookupEnv("INPUT_PATH")
	output, _ := os.LookupEnv("OUTPUT_PATH")

	if input == "" || output == "" {
		fmt.Println("ERROR: missing INPUT_PATH and OUTPUT_PATH env vars")
		return
	}

	hostnames := []string{}
	if frmHostnames == "" {
		hostnames = append(hostnames, "http://"+frmHostname+":"+frmPort)
	} else {
		for _, frmServer := range strings.Split(frmHostnames, ",") {
			if !strings.HasPrefix(frmServer, "http://") && !strings.HasPrefix(frmServer, "https://") {
				frmServer = "http://" + frmServer
			}
			hostnames = append(hostnames, frmServer)
		}
	}

	webhooks := []string{}
	if discordWebhooks == "" {
		webhooks = append(webhooks, discordWebhook)
	} else {
		for _, webhook := range strings.Split(discordWebhooks, ",") {
			webhooks = append(webhooks, webhook)
		}
	}

	// associate webhooks with hostnames
	config := []HostWebhookPair{}
	for i, hostname := range hostnames {
		if i < len(webhooks) {
			config = append(config, HostWebhookPair{Host: hostname, Webhook: webhooks[i]})
		}
	}

	b, err := os.ReadFile(input) // just pass the file name
	if err != nil {
		fmt.Println(err)
	}
	rawConfig := string(b) // convert content to a 'string'

	var t = template.Must(template.New("name").Parse(rawConfig))

	f, err := os.Create(output)
	if err != nil {
		fmt.Println("creating file template:", err)
	}

	err = t.Execute(f, config)
	if err != nil {
		fmt.Println("executing template:", err)
	}
}
