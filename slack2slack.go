package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const configFile string = "config.yml"

type Side struct {
	To struct {
		URL string `yaml:"url"`
	} `yaml:"to"`
	From struct {
		Prefix string `yaml:"prefix"`
		Token  string `yaml:"token"`
	} `yaml:"from"`
}

type Bridge struct {
	Name     string `yaml:"name"`
	Enabled  bool   `yaml:"enabled"`
	EndPoint string `yaml:"endpoint"`
	SideA    Side   `yaml:"a"`
	SideB    Side   `yaml:"b"`
}

//Config of the app
type Config struct {
	Port    int      `yaml:"port"`
	Bridges []Bridge `yaml:"bridges"`
}

//Slack payload
type SlackPayload struct {
	Text     string `json:"text"`
	Username string `json:"username"`
}

//App config
var config Config

//Read config file
func readConfig() {
	var filename string

	if len(os.Args) < 2 {
		filename = configFile
	} else {
		filename = os.Args[1]
	}

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
}

func main() {
	readConfig()

	http.HandleFunc("/", Index)
	for _, element := range config.Bridges {
		if element.Enabled {
			http.Handle(fmt.Sprintf("%s", element.EndPoint), BridgeHandler(IncomingWebhook, element))
		}
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

//Index function
func Index(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from %s to Index endpoint", r.RemoteAddr)
	fmt.Fprintf(w, "This is Slack2Slack")
}

func BridgeHandler(fn http.HandlerFunc, bridge Bridge) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Bridge %s called by %s", bridge.Name, r.RemoteAddr)
		token := r.PostFormValue("token")
		user_name := r.PostFormValue("user_name")
		if token != bridge.SideA.From.Token && token != bridge.SideB.From.Token {
			log.Printf("Wrong token!")
			http.Error(w, "Wrong token!", http.StatusUnauthorized)
			return
		}
		var side_from Side
		var side_to Side
		if token == bridge.SideA.From.Token {
			log.Printf("%s updates Side A", user_name)
			side_from = bridge.SideA
			side_to = bridge.SideB
		}
		if token == bridge.SideB.From.Token {
			log.Printf("%s updates Side B", user_name)
			side_from = bridge.SideB
			side_to = bridge.SideA
		}
		if strings.Contains(user_name, "slackbot") {
			log.Printf("Not reposting ourself")
		} else {
			payload := SlackPayload{Text: r.PostFormValue("text"), Username: fmt.Sprintf("%s-%s", side_from.From.Prefix, user_name)}
			OutgoingWebhook(side_to.To.URL, payload)
		}
		fn(w, r)
	}
}

//IncomingWebhook function
func IncomingWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from %s to Webhook endpoint", r.RemoteAddr)
}

//Send a message to Slack endpoint
func OutgoingWebhook(url string, payload SlackPayload) {

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(payload)
	log.Printf("payload: %s", b)
	res, _ := http.Post(url, "application/json; charset=utf-8", b)
	io.Copy(os.Stdout, res.Body)
	log.Printf("Outgoing webhook sent!")
}
