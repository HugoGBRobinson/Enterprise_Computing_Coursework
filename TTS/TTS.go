package main

import (
	"Coursework/config"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

//REGION and URI provide constant a URI to connect to
//Microsoft Cognitive Services.
const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
)

//KEY is the provided access key, it is obtained through the
//config file.
var KEY = config.GetAzureKey()

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Speak is a structure that details the specific format used by
//Azure for its text to speech services. This format is inline with
//the Speech Synthesis Markup Language (SSML).
type Speak struct {
	XMLName xml.Name `xml:"speak"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Lang    string   `xml:"xml:lang,attr"`
	Voice   Voice
}

// Voice much like Speak, follows the SSML structure and is used
//For the nested XML required by Azure.
type Voice struct {
	XMLName xml.Name `xml:"voice"`
	Text    string   `xml:",chardata"`
	Lang    string   `xml:"xml:lang,attr"`
	Name    string   `xml:"name,attr"`
}

//Speech provides the necessary structure to create a json speech
//response.
type Speech struct {
	Speech string `json:"speech"`
}

// TextToSpeech is the primary function of this microservice as it
// decodes the request, marshals a xml request to Azure, and
// marshals a json response with the .wav back to Alexa.
func TextToSpeech(w http.ResponseWriter, r *http.Request) {
	t := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if cont, ok := t["text"]; ok {
			client := &http.Client{}
			// Set up xml request to Azure
			v := &Voice{
				XMLName: xml.Name{},
				Text:    cont,
				Lang:    "en-US",
				Name:    "en-US-JennyNeural",
			}
			s := &Speak{
				XMLName: xml.Name{},
				Text:    "",
				Version: "1.0",
				Lang:    "en-US",
				Voice:   *v,
			}

			m, _ := xml.MarshalIndent(s, "", "  ")
			req, err2 := http.NewRequest("POST", URI, bytes.NewReader(m))
			check(err2)
			req.Header.Set("Content-Type", "application/ssml+xml")
			req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
			req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")

			rsp, err3 := client.Do(req)
			check(err3)
			defer rsp.Body.Close()
			if rsp.StatusCode == http.StatusOK {
				body, err4 := ioutil.ReadAll(rsp.Body)
				check(err4)
				w.WriteHeader(http.StatusOK)
				// Encode speech to base 64
				EncBody := base64.StdEncoding.EncodeToString(body)
				speech := Speech{Speech: EncBody}
				json.NewEncoder(w).Encode(speech)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

//main sets up the listen and serve functionality allowing Alexa to
//request its services.
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tts", TextToSpeech).Methods("POST")
	http.ListenAndServe(":3003", r)
}
