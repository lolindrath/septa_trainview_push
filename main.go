// TODO: error handling, everywhere

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	Token string
	User  string
}

type train struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Number      string `json:"trainno"`
	Service     string `json:"service"`
	Dest        string `json:"dest"`
	NextStop    string `json:"nextstop"`
	Late        int    `json:"late"`
	Source      string `json:"SOURCE"`
	Track       string `json:"TRACK"`
	TrackChange string `json:"TRACK_CHANGE"`
}

const (
	trainviewURL     = "http://www3.septa.org/hackathon/TrainView/"
	pushoverURL      = "https://api.pushover.net/1/messages.json"
	pushoverMsgTitle = "TrainView"
)

func readConfig() config {
	var conf config
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := toml.DecodeFile(fmt.Sprintf("%s/.trainview_pushrc", dir), &conf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return conf
}

func readArgs() (string, string) {
	if len(os.Args) != 3 {
		fmt.Printf("trainview_push <trainNum> <trainTime>")
		os.Exit(1)
	}

	trainNum := os.Args[1]
	time := os.Args[2]

	return trainNum, time
}

func getTrains() []train {
	resp, _ := http.Get(trainviewURL)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var trains []train

	err := json.Unmarshal(body, &trains)

	if err != nil {
		fmt.Println(err)
	}

	return trains
}

func findTrain(trainNum string, trains []train) train {
	var ourTrain train

	for _, t := range trains {
		if t.Number == trainNum {
			ourTrain = t
			break
		}
	}

	return ourTrain
}

func sendPushover(msg string) {
	client := http.Client{}
	form := url.Values{}
	form.Add("token", conf.Token)
	form.Add("user", conf.User)
	form.Add("message", msg)
	form.Add("title", pushoverMsgTitle)
	req, _ := http.NewRequest("POST", pushoverURL, strings.NewReader(form.Encode()))

	_, _ = client.Do(req)
}

var conf config

func main() {

	conf = readConfig()

	trainNum, time := readArgs()

	trains := getTrains()
	ourTrain := findTrain(trainNum, trains)

	var late string
	if ourTrain.Late == 0 {
		late = "on time"
	} else {
		late = fmt.Sprintf("%d mins late", ourTrain.Late)
	}

	msg := fmt.Sprintf("%s Warminster(%s) is %s", time, trainNum, late)

	sendPushover(msg)

}
