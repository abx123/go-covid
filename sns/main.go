package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Request struct {
	TeamID             string          `json:"team_id" form:"team_id"`
	ApiAppID           string          `json:"api_app_id" form:"api_app_id"`
	Event              Event           `json:"event" form:"event"`
	EventID            string          `json:"event_id" form:"event_id"`
	EventTime          int             `json:"event_time" form:"event_time"`
	Authorizations     []Authorization `json:"authorizations" form:"authorizations"`
	IsExtSharedChannel bool            `json:"is_ext_shared_channel" form:"is_ext_shared_channel"`
	EventContext       string          `json:"event_context" form:"event_context"`
	Type               string          `json:"type" form:"type"`
	Token              string          `json:"token" form:"token"`
	Challenge          string          `json:"challenge" form:"challenge"`
}
type Authorization struct {
	EnterpriseID         string `json:"enterprise_id"`
	TeamID               string `json:"team_id"`
	UserID               string `json:"user_id"`
	IsBot                bool   `json:"is_bot"`
	IsEneterpriseInstall bool   `json:"is_enterprise_install"`
}

type Element struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

type Event struct {
	ClientMsgID string  `json:"client_msg_id"`
	Type        string  `json:"type"`
	Text        string  `json:"text"`
	User        string  `json:"user"`
	Ts          string  `json:"ts"`
	Team        string  `json:"team"`
	Blocks      []Block `json:"blocks"`
	Channel     string  `json:"channel"`
	EventTs     string  `json:"event_ts"`
}

type Block struct {
	Type     string     `json:"type" form:"type"`
	BlockID  string     `json:"block_id"`
	Elements []Elements `json:"elements"`
}

type Elements struct {
	Type     string    `json:"type"`
	Elements []Element `json:"elements"`
}

type Record struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Date              string             `bson:"date,omitempty"`
	NewCases          int                `bson:"newCases,omitempty"`
	ImportCases       int                `bson:"importCases,omitempty"`
	RecoveredCases    int                `bson:"recoveredCases,omitempty"`
	ClusterImport     int                `bson:"importClusters,omitempty"`
	ClusterReligious  int                `bson:"religiousClusters,omitempty"`
	ClusterCommunity  int                `bson:"communityClusters,omitempty"`
	ClusterHighRisk   int                `bson:"HighRiskClusters,omitempty"`
	ClusterEducation  int                `bson:"educationClusters,omitempty"`
	ClusterDentention int                `bson:"detentionClusters,omitempty"`
	ClusterWorkplace  int                `bson:"workspaceClusters,omitempty"`
	States            map[string]State   `bson:"states,omitempty"`
	Death             Death              `bson:"death,omitempty"`
	Test              Test               `bson:"tests,omitempty"`
	Population        Population         `bson:"population,omitempty"`
}

type Population struct {
	Population int `bson:"population,omitempty"`
	Over18     int `bson:"over18,omitempty"`
	Over60     int `bson:"over60,omitempty"`
	Over12     int `bson:"over12,omitempty"`
}

type Test struct {
	RtkAg int `bson:"rtkAg,omitempty"`
	Pcr   int `bson:"Pcr,omitempty"`
}

type State struct {
	ImportCases    int        `bson:"importCases,omitempty"`
	NewCases       int        `bson:"newCases,omitempty"`
	RecoveredCases int        `bson:"recoveredCases,omitempty"`
	Death          Death      `bson:"death,omitempty"`
	Hospital       Hospital   `bson:"hospital,omitempty"`
	ICU            ICU        `bson:"icu,omitempty"`
	PKRC           PKRC       `bson:"pkrc,omitempty"`
	Test           Test       `bson:"test,omitempty"`
	Population     Population `bson:"population,omitempty"`
}

type Hospital struct {
	Beds                 int `bson:"beds,omitempty"`
	CovidBeds            int `bson:"covidBeds,omitempty"`
	NonCriticalBeds      int `bson:"nonCriticalBeds,omitempty"`
	AdmittedPUI          int `bson:"admittedPui,omitempty"`
	AdmittedCovid        int `bson:"admittedCovid,omitempty"`
	AdmittedTotal        int `bson:"admittedTotal,omitempty"`
	DischargedPui        int `bson:"dischargedPui,omitempty"`
	DischargedCovid      int `bson:"dischargedCovid,omitempty"`
	DischargedTotal      int `bson:"dischargedTotal,omitempty"`
	HospitalizedCovid    int `bson:"hospitalizedCovid,omitempty"`
	HospitalizedPui      int `bson:"hospitalizedPui,omitempty"`
	HospitalizedNonCovid int `bson:"hospitalizedNonCovid,omitempty"`
}

type Death struct {
	NewDeaths       int `bson:"newDeaths,omitempty"`
	ActualDeaths    int `bson:"actualDeaths,omitempty"`
	BIDDeaths       int `bson:"bidDeaths,omitempty"`
	ActualBIDDeaths int `bson:"actualBidDeaths,omitempty"`
}

type ICU struct {
	ICUBeds             int `bson:"icuBeds,omitempty"`
	ICUBedsRep          int `bson:"iceBedsRep,omitempty"`
	ICUBedsTotal        int `bson:"icuBedsTotal,omitempty"`
	ICUBedsCovid        int `bson:"icuBedsCovid,omitempty"`
	Ventilators         int `bson:"ventilators,omitempty"`
	PortableVentilators int `bson:"portableVentilators,omitempty"`
	ICUCovid            int `bson:"icuCovid,omitempty"`
	ICUPui              int `bson:"icuPui,omitempty"`
	ICUNonCovid         int `bson:"icuNonCovid,omitempty"`
	VentCovid           int `bson:"ventilatorsCovid,omitempty"`
	VentPui             int `bson:"ventilatorsPui,omitempty"`
	VentNonCovid        int `bson:"ventilatorsNonCovid,omitempty"`
	VentUsed            int `bson:"ventilatorsUsed,omitempty"`
	PortVentUsed        int `bson:"portableVentilatorsUsed,omitempty"`
}

type PKRC struct {
	Beds            int `bson:"beds,omitempty"`
	AdmittedCovid   int `bson:"admittedCovid,omitempty"`
	AdmittedPui     int `bson:"admittedPui,omitempty"`
	AdmittedTotal   int `bson:"admittedTotal,omitempty"`
	DischargedPui   int `bson:"dischargedPui,omitempty"`
	DischargedCovid int `bson:"dischargedCovid,omitempty"`
	DischargedTotal int `bson:"dischargedTotal,omitempty"`
	PKRCCovid       int `bson:"pkrcCovid,omitempty"`
	PKRCPui         int `bson:"pkrcPui,omitempty"`
	PKRCNonCovid    int `bson:"pkrcNonCovid,omitempty"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	statesMap := map[string]string{
		"selangor":        "Selangor",
		"putrajaya":       "W.P. Putrajaya",
		"kedah":           "Kedah",
		"penang":          "Pulau Pinang",
		"sarawak":         "Sarawak",
		"kelantan":        "Kelantan",
		"pulau pinang":    "Pulau Pinang",
		"johor":           "Johor",
		"labuan":          "W.P. Labuan",
		"melaka":          "Melaka",
		"terengganu":      "Terengganu",
		"kuala lumpur":    "W.P. Kuala Lumpur",
		"sabah":           "Sabah",
		"n9":              "Negeri Sembilan",
		"perak":           "Perak",
		"perlis":          "Perlis",
		"pahang":          "Pahang",
		"kl":              "W.P. Kuala Lumpur",
		"negeri sembilan": "Negeri Sembilan",
	}
	client, _ := NewMongoClient()
	defer client.Disconnect(context.Background())
	collection := client.Database("covid").Collection("my")
	for _, record := range snsEvent.Records {
		req := Request{}
		json.Unmarshal([]byte(record.SNS.Message), &req)
		if req.Event.Channel != "C0188FC7MAP" && req.Event.Channel != "G01FLHXFZTM" {
			return
		}

		if req.Event.Text == "<@U0188FCRJ9H>" {
			rq, _ := json.Marshal(map[string]string{
				"text": ":bb-come-ady-2::bb-here::bb-who-find:",
			})
			_, _ = http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(rq))
			return
		}

		re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
		date := re.FindAllString(req.Event.Text, -1)
		if len(date) > 0 {
			rec, _ := getFromMongo(collection, date[0])
			for k, val := range statesMap {
				if strings.Contains(strings.ToLower(req.Event.Text), k) {
					sendToSlack(rec, val)
					return
				}
			}
			sendToSlack(rec, "")
			return
		}

		if strings.Contains(req.Event.Text, "Malaysia") || strings.Contains(req.Event.Text, "cobis") {
			rec, _ := getFromMongo(collection, "")
			sendToSlack(rec, "")
			return
		}
		rq, _ := json.Marshal(map[string]string{
			"text": ":bb-say-what::bb-no-understand:",
		})
		_, _ = http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(rq))

	}
}

func sendToSlack(rec *Record, state string) {
	flags := map[string]string{
		"Selangor":          ":selangor:",
		"W.P. Putrajaya":    ":putrajaya:",
		"Kedah":             ":kedah:",
		"Pulau Pinang":      ":ppinang:",
		"Sarawak":           ":sarawak:",
		"Kelantan":          ":kelantan:",
		"Malaysia":          ":malaysia:",
		"Johor":             ":johor:",
		"W.P. Labuan":       ":labuan:",
		"Melaka":            ":melaka:",
		"Terengganu":        ":terengganu:",
		"W.P. Kuala Lumpur": ":kl:",
		"Sabah":             ":sabah:",
		"Negeri Sembilan":   ":n9:",
		"Perak":             ":perak:",
		"Perlis":            ":perlis:",
		"Pahang":            ":pahang:",
	}
	var req []byte
	if rec != nil {
		req, _ = json.Marshal(map[string]string{
			"text": fmt.Sprintf("%s Data as of %s\n New Cases: %d \n Import Cases: %d \n Recovered Cases: %d \n New Deaths: %d \n New Brought in Dead (BID): %d\n Actual COVID Deaths: %d \n", flags["Malaysia"], rec.Date, rec.NewCases, rec.ImportCases, rec.RecoveredCases, rec.Death.NewDeaths, rec.Death.BIDDeaths, rec.Death.ActualDeaths),
		})
	}

	if state != "" {
		for k, v := range rec.States {
			if k == state {
				req, _ = json.Marshal(map[string]string{
					"text": fmt.Sprintf("%s %s as of %s\n New Cases: %d \n Import Cases: %d \n Recovered Cases: %d \n New Deaths: %d \n Actual Deaths: %d", flags[k], k, rec.Date, v.NewCases, v.ImportCases, v.RecoveredCases, v.Death.NewDeaths, v.Death.ActualDeaths),
				})
			}
		}
	}

	if rec == nil && state == "" {
		req, _ = json.Marshal(map[string]string{
			"text": ":bb-say-what::bb-no-understand:",
		})
	}
	_, _ = http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(req))
}

func NewMongoClient() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getFromMongo(collection *mongo.Collection, date string) (*Record, error) {
	res := &Record{}
	opt := options.FindOne()
	opt.SetSort(bson.M{"date": -1})
	filter := bson.M{}
	if date != "" {
		filter = bson.M{"date": date}
	}
	err := collection.FindOne(context.TODO(), filter, opt).Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
