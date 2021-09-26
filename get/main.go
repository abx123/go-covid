package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// func main() {
// 	get(context.TODO(), events.APIGatewayProxyRequest{})
// }

func main() {
	lambda.Start(get)
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

func get(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	}
	client, _ := NewMongoClient()
	defer client.Disconnect(context.Background())
	collection := client.Database("covid").Collection("my")

	var rec *Record
	var err error
	date := ""
	if val, ok := request.PathParameters["date"]; ok {
		rec, err = getFromMongo(collection, val)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    headers,
				Body:       err.Error(),
			}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       formatResp(rec),
		}, nil
	}
	rec, err = getFromMongo(collection, date)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       formatResp(rec),
	}, nil
}

func formatResp(input interface{}) string {
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(input)
	responseBytes := bytesBuffer.Bytes()

	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, responseBytes, "", "  ")
	if error != nil {
		log.Println("JSON parse error: ", error)
	}
	formattedResp := prettyJSON.String()
	return formattedResp
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
	return client, err
}
