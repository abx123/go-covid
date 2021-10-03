package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func getLatestFromMongo(collection *mongo.Collection) (*Record, error) {
	res := &Record{}
	opt := options.FindOne()
	opt.SetSort(bson.M{"date": -1})
	err := collection.FindOne(context.TODO(), bson.M{}, opt).Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getAllFromMongo(collection *mongo.Collection) ([]*Record, error) {
	res := []*Record{}
	opt := options.Find()
	opt.SetSort(bson.M{"date": -1})
	cursor, err := collection.Find(context.TODO(), bson.M{}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.TODO()) {
		rec := &Record{}
		err := cursor.Decode(&rec)
		if err != nil {
			return nil, err
		}
		res = append(res, rec)
	}
	return res, nil
}

// func main() {
// 	crawl(context.Background(), events.APIGatewayProxyRequest{})
// }

func main() {
	lambda.Start(crawl)
}

func crawl(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	}

	client, _ := NewMongoClient()
	defer client.Disconnect(context.Background())
	collection := client.Database("covid").Collection("my")

	// initMongoRecord(collection)
	// truncateMongo(collection)

	new := get()
	latest, err := getLatestFromMongo(collection)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, err
	}
	t, _ := time.Parse("2006-01-02", latest.Date)

	for d := t.AddDate(0, 0, 1); !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		if v, ok := new[d.Format("2006-01-02")]; ok {
			_, err := saveToMongo(collection, v)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Headers:    headers,
					Body:       err.Error(),
				}, err
			}
			sendToSlack(v)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
	}, nil
}

func initMongoRecord(col *mongo.Collection) {
	cases := get()
	for _, v := range cases {
		ir, _ := saveToMongo(col, v)
		fmt.Println(ir, v.Date)
	}

}

func truncateMongo(col *mongo.Collection) {
	dr, _ := col.DeleteMany(context.TODO(), bson.M{})
	fmt.Println(dr)
}

func saveToMongo(col *mongo.Collection, rec Record) (*mongo.InsertOneResult, error) {
	insertResult, err := col.InsertOne(context.TODO(), rec)
	if err != nil {
		return nil, err
	}
	return insertResult, nil
}

func get() map[string]Record {
	cases := getCountry()
	states := getStateMap()
	deaths := getDeaths()
	deathStates := getDeathByState()
	hospital := getHostpital()
	icu := getICU()
	pkrc := getPKRC()
	pop := getPopulation()
	tests := getTestByCountry()
	testStates := getTestByState()
	for k, v := range cases {
		id, _ := primitive.ObjectIDFromHex(k)
		v.ID = id
		v.Population = pop["Malaysia"]
		v.Death = deaths[k]
		v.States = states[k]
		v.Test = tests[k]
		for key, val := range v.States {
			val.Death = deathStates[k][key]
			val.Hospital = hospital[k][key]
			val.ICU = icu[k][key]
			val.PKRC = pkrc[k][key]
			val.Population = pop[key]
			val.Test = testStates[k][key]
			v.States[key] = val
		}
		cases[k] = v
	}
	return cases
}

func getCountry() map[string]Record {
	res := map[string]Record{}
	cdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/cases_malaysia.csv")
	for i, crow := range cdata {
		if i == 0 {
			continue
		}
		cr := strings.Split(crow[0], ",")
		cn, _ := strconv.Atoi(cr[1])
		ci, _ := strconv.Atoi(cr[2])
		crec, _ := strconv.Atoi(cr[3])
		cli, _ := strconv.Atoi(cr[4])
		clr, _ := strconv.Atoi(cr[5])
		clc, _ := strconv.Atoi(cr[6])
		clh, _ := strconv.Atoi(cr[7])
		cle, _ := strconv.Atoi(cr[8])
		cld, _ := strconv.Atoi(cr[9])
		clw, _ := strconv.Atoi(cr[10])
		res[cr[0]] = Record{
			Date:              cr[0],
			NewCases:          cn,
			ImportCases:       ci,
			RecoveredCases:    crec,
			ClusterImport:     cli,
			ClusterReligious:  clr,
			ClusterCommunity:  clc,
			ClusterHighRisk:   clh,
			ClusterEducation:  cle,
			ClusterDentention: cld,
			ClusterWorkplace:  clw,
		}
	}
	return res
}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getDeaths() map[string]Death {
	res := map[string]Death{}
	ddata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/deaths_malaysia.csv")
	for i, drow := range ddata {
		if i == 0 {
			continue
		}
		dr := strings.Split(drow[0], ",")
		nd, _ := strconv.Atoi(dr[1])
		ad, _ := strconv.Atoi(dr[2])
		bd, _ := strconv.Atoi(dr[3])
		abd, _ := strconv.Atoi(dr[4])
		res[dr[0]] = Death{
			NewDeaths:       nd,
			ActualDeaths:    ad,
			BIDDeaths:       bd,
			ActualBIDDeaths: abd,
		}
	}
	return res
}

func getStateMap() map[string]map[string]State {
	test := map[string]map[string]State{}
	type tempState struct {
		Date           string
		Name           string
		ImportCases    int
		NewCases       int
		RecoveredCases int
		Death          Death
	}
	sdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/cases_state.csv")
	states := []tempState{}
	for i, srow := range sdata {
		if i == 0 {
			continue
		}
		sr := strings.Split(srow[0], ",")
		nc, _ := strconv.Atoi(sr[2])
		ic, _ := strconv.Atoi(sr[3])
		rc, _ := strconv.Atoi(sr[4])
		states = append(states, tempState{
			Date:           sr[0],
			Name:           sr[1],
			ImportCases:    ic,
			NewCases:       nc,
			RecoveredCases: rc,
		})
	}
	for _, v := range states {
		if _, ok := test[v.Date]; ok {
			test[v.Date][v.Name] = State{
				ImportCases:    v.ImportCases,
				NewCases:       v.NewCases,
				RecoveredCases: v.RecoveredCases,
			}
		} else {
			test[v.Date] = map[string]State{v.Name: {
				ImportCases:    v.ImportCases,
				NewCases:       v.NewCases,
				RecoveredCases: v.RecoveredCases,
			}}
		}
	}
	return test
}

func getDeathByState() map[string]map[string]Death {
	type tempDeath struct {
		Date            string
		Name            string
		NewDeaths       int
		ActualDeaths    int
		BIDDeaths       int
		ActualBIDDeaths int
	}
	res := map[string]map[string]Death{}
	sdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/deaths_state.csv")
	deathByStates := []tempDeath{}
	for i, srow := range sdata {
		if i == 0 {
			continue
		}
		sr := strings.Split(srow[0], ",")
		nd, _ := strconv.Atoi(sr[2])
		ad, _ := strconv.Atoi(sr[3])
		bd, _ := strconv.Atoi(sr[4])
		abd, _ := strconv.Atoi(sr[5])
		deathByStates = append(deathByStates, tempDeath{
			Date:            sr[0],
			Name:            sr[1],
			NewDeaths:       nd,
			ActualDeaths:    ad,
			BIDDeaths:       bd,
			ActualBIDDeaths: abd,
		})
	}
	for _, v := range deathByStates {
		if _, ok := res[v.Date]; ok {
			res[v.Date][v.Name] = Death{
				NewDeaths:       v.NewDeaths,
				ActualDeaths:    v.ActualDeaths,
				BIDDeaths:       v.BIDDeaths,
				ActualBIDDeaths: v.ActualBIDDeaths,
			}
		} else {
			res[v.Date] = map[string]Death{v.Name: {
				NewDeaths:       v.NewDeaths,
				ActualDeaths:    v.ActualDeaths,
				BIDDeaths:       v.BIDDeaths,
				ActualBIDDeaths: v.ActualBIDDeaths,
			}}
		}
	}
	return res
}

func getHostpital() map[string]map[string]Hospital {
	type tempHospital struct {
		Date                 string
		Name                 string
		Beds                 int
		CovidBeds            int
		NonCriticalBeds      int
		AdmittedPUI          int
		AdmittedCovid        int
		AdmittedTotal        int
		DischargedPui        int
		DischargedCovid      int
		DischargedTotal      int
		HospitalizedCovid    int
		HospitalizedPui      int
		HospitalizedNonCovid int
	}
	res := map[string]map[string]Hospital{}
	hdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/hospital.csv")
	ths := []tempHospital{}
	for i, hrow := range hdata {
		if i == 0 {
			continue
		}
		hr := strings.Split(hrow[0], ",")
		b, _ := strconv.Atoi(hr[2])
		bc, _ := strconv.Atoi(hr[3])
		bnc, _ := strconv.Atoi(hr[4])
		ap, _ := strconv.Atoi(hr[5])
		ac, _ := strconv.Atoi(hr[6])
		at, _ := strconv.Atoi(hr[7])
		dp, _ := strconv.Atoi(hr[8])
		dc, _ := strconv.Atoi(hr[9])
		dt, _ := strconv.Atoi(hr[10])
		hc, _ := strconv.Atoi(hr[11])
		hp, _ := strconv.Atoi(hr[12])
		hnc, _ := strconv.Atoi(hr[13])
		ths = append(ths, tempHospital{
			Date:                 hr[0],
			Name:                 hr[1],
			Beds:                 b,
			CovidBeds:            bc,
			NonCriticalBeds:      bnc,
			AdmittedPUI:          ap,
			AdmittedCovid:        ac,
			AdmittedTotal:        at,
			DischargedPui:        dp,
			DischargedCovid:      dc,
			DischargedTotal:      dt,
			HospitalizedCovid:    hc,
			HospitalizedPui:      hp,
			HospitalizedNonCovid: hnc,
		})
	}
	for _, v := range ths {
		if _, ok := res[v.Date]; ok {
			res[v.Date][v.Name] = Hospital{
				CovidBeds:            v.CovidBeds,
				Beds:                 v.Beds,
				NonCriticalBeds:      v.NonCriticalBeds,
				AdmittedPUI:          v.AdmittedPUI,
				AdmittedCovid:        v.AdmittedCovid,
				AdmittedTotal:        v.AdmittedTotal,
				DischargedPui:        v.DischargedPui,
				DischargedCovid:      v.DischargedCovid,
				DischargedTotal:      v.DischargedTotal,
				HospitalizedCovid:    v.HospitalizedCovid,
				HospitalizedPui:      v.HospitalizedPui,
				HospitalizedNonCovid: v.HospitalizedNonCovid,
			}
		} else {
			res[v.Date] = map[string]Hospital{v.Name: {
				CovidBeds:            v.CovidBeds,
				Beds:                 v.Beds,
				NonCriticalBeds:      v.NonCriticalBeds,
				AdmittedPUI:          v.AdmittedPUI,
				AdmittedCovid:        v.AdmittedCovid,
				AdmittedTotal:        v.AdmittedTotal,
				DischargedPui:        v.DischargedPui,
				DischargedCovid:      v.DischargedCovid,
				DischargedTotal:      v.DischargedTotal,
				HospitalizedCovid:    v.HospitalizedCovid,
				HospitalizedPui:      v.HospitalizedPui,
				HospitalizedNonCovid: v.HospitalizedNonCovid,
			}}
		}
	}
	return res
}

func getICU() map[string]map[string]ICU {
	type tempICU struct {
		Date                string
		Name                string
		ICUBeds             int
		ICUBedsRep          int
		ICUBedsTotal        int
		ICUBedsCovid        int
		Ventilators         int
		PortableVentilators int
		ICUCovid            int
		ICUPui              int
		ICUNonCovid         int
		VentCovid           int
		VentPui             int
		VentNonCovid        int
		VentUsed            int
		PortVentUsed        int
	}
	res := map[string]map[string]ICU{}
	idata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/icu.csv")
	tis := []tempICU{}
	for i, irow := range idata {
		if i == 0 {
			continue
		}
		ir := strings.Split(irow[0], ",")
		b, _ := strconv.Atoi(ir[2])
		br, _ := strconv.Atoi(ir[3])
		bt, _ := strconv.Atoi(ir[4])
		bc, _ := strconv.Atoi(ir[5])
		v, _ := strconv.Atoi(ir[6])
		vp, _ := strconv.Atoi(ir[7])
		ic, _ := strconv.Atoi(ir[8])
		ip, _ := strconv.Atoi(ir[9])
		inc, _ := strconv.Atoi(ir[10])
		vc, _ := strconv.Atoi(ir[11])
		vpui, _ := strconv.Atoi(ir[12])
		vnc, _ := strconv.Atoi(ir[13])
		vu, _ := strconv.Atoi(ir[14])
		vpu, _ := strconv.Atoi(ir[15])
		tis = append(tis, tempICU{
			Date:                ir[0],
			Name:                ir[1],
			ICUBeds:             b,
			ICUBedsRep:          br,
			ICUBedsTotal:        bt,
			ICUBedsCovid:        bc,
			Ventilators:         v,
			PortableVentilators: vp,
			ICUCovid:            ic,
			ICUPui:              ip,
			ICUNonCovid:         inc,
			VentCovid:           vc,
			VentPui:             vpui,
			VentNonCovid:        vnc,
			VentUsed:            vu,
			PortVentUsed:        vpu,
		})
	}
	for _, v := range tis {
		if _, ok := res[v.Date]; ok {
			res[v.Date][v.Name] = ICU{
				ICUBeds:             v.ICUBeds,
				ICUBedsRep:          v.ICUBedsRep,
				ICUBedsTotal:        v.ICUBedsTotal,
				ICUBedsCovid:        v.ICUBedsCovid,
				Ventilators:         v.Ventilators,
				PortableVentilators: v.PortableVentilators,
				ICUCovid:            v.ICUCovid,
				ICUPui:              v.ICUPui,
				ICUNonCovid:         v.ICUNonCovid,
				VentCovid:           v.VentCovid,
				VentPui:             v.VentPui,
				VentNonCovid:        v.VentNonCovid,
				VentUsed:            v.VentUsed,
				PortVentUsed:        v.PortVentUsed,
			}
		} else {
			res[v.Date] = map[string]ICU{v.Name: {
				ICUBeds:             v.ICUBeds,
				ICUBedsRep:          v.ICUBedsRep,
				ICUBedsTotal:        v.ICUBedsTotal,
				ICUBedsCovid:        v.ICUBedsCovid,
				Ventilators:         v.Ventilators,
				PortableVentilators: v.PortableVentilators,
				ICUCovid:            v.ICUCovid,
				ICUPui:              v.ICUPui,
				ICUNonCovid:         v.ICUNonCovid,
				VentCovid:           v.VentCovid,
				VentPui:             v.VentPui,
				VentNonCovid:        v.VentNonCovid,
				VentUsed:            v.VentUsed,
				PortVentUsed:        v.PortVentUsed,
			}}
		}
	}
	return res
}

func getPKRC() map[string]map[string]PKRC {
	type tempPKRC struct {
		Date            string
		Name            string
		Beds            int
		AdmittedCovid   int
		AdmittedPui     int
		AdmittedTotal   int
		DischargedPui   int
		DischargedCovid int
		DischargedTotal int
		PKRCCovid       int
		PKRCPui         int
		PKRCNonCovid    int
	}
	res := map[string]map[string]PKRC{}
	pdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/pkrc.csv")
	tps := []tempPKRC{}
	for i, prow := range pdata {
		if i == 0 {
			continue
		}
		pr := strings.Split(prow[0], ",")
		b, _ := strconv.Atoi(pr[2])
		ap, _ := strconv.Atoi(pr[3])
		ac, _ := strconv.Atoi(pr[4])
		at, _ := strconv.Atoi(pr[5])
		dp, _ := strconv.Atoi(pr[6])
		dc, _ := strconv.Atoi(pr[7])
		dt, _ := strconv.Atoi(pr[8])
		pc, _ := strconv.Atoi(pr[9])
		pp, _ := strconv.Atoi(pr[10])
		pnc, _ := strconv.Atoi(pr[11])
		tps = append(tps, tempPKRC{
			Date:            pr[0],
			Name:            pr[1],
			Beds:            b,
			AdmittedPui:     ap,
			AdmittedCovid:   ac,
			AdmittedTotal:   at,
			DischargedPui:   dp,
			DischargedCovid: dc,
			DischargedTotal: dt,
			PKRCCovid:       pc,
			PKRCPui:         pp,
			PKRCNonCovid:    pnc,
		})
	}
	for _, v := range tps {
		if _, ok := res[v.Date]; ok {
			res[v.Date][v.Name] = PKRC{
				Beds:            v.Beds,
				AdmittedPui:     v.AdmittedPui,
				AdmittedCovid:   v.AdmittedCovid,
				AdmittedTotal:   v.AdmittedTotal,
				DischargedPui:   v.DischargedPui,
				DischargedCovid: v.DischargedCovid,
				DischargedTotal: v.DischargedTotal,
				PKRCCovid:       v.PKRCCovid,
				PKRCPui:         v.PKRCPui,
				PKRCNonCovid:    v.PKRCNonCovid,
			}
		} else {
			res[v.Date] = map[string]PKRC{v.Name: {
				Beds:            v.Beds,
				AdmittedPui:     v.AdmittedPui,
				AdmittedCovid:   v.AdmittedCovid,
				AdmittedTotal:   v.AdmittedTotal,
				DischargedPui:   v.DischargedPui,
				DischargedCovid: v.DischargedCovid,
				DischargedTotal: v.DischargedTotal,
				PKRCCovid:       v.PKRCCovid,
				PKRCPui:         v.PKRCPui,
				PKRCNonCovid:    v.PKRCNonCovid,
			}}
		}
	}
	return res
}

func getTestByCountry() map[string]Test {
	res := map[string]Test{}
	tdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/tests_malaysia.csv")
	for i, trow := range tdata {
		if i == 0 {
			continue
		}
		tr := strings.Split(trow[0], ",")
		rtkAg, _ := strconv.Atoi(tr[1])
		pcr, _ := strconv.Atoi(tr[2])
		res[tr[0]] = Test{
			RtkAg: rtkAg,
			Pcr:   pcr,
		}
	}
	return res
}

func getTestByState() map[string]map[string]Test {
	type tempTest struct {
		Date  string
		Name  string
		RtkAg int
		Pcr   int
	}
	res := map[string]map[string]Test{}
	pdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/epidemic/tests_state.csv")
	ts := []tempTest{}
	for i, trow := range pdata {
		if i == 0 {
			continue
		}
		tr := strings.Split(trow[0], ",")
		rtkAg, _ := strconv.Atoi(tr[2])
		pcr, _ := strconv.Atoi(tr[3])
		ts = append(ts, tempTest{
			Date:  tr[0],
			Name:  tr[1],
			RtkAg: rtkAg,
			Pcr:   pcr,
		})
	}
	for _, v := range ts {
		if _, ok := res[v.Date]; ok {
			res[v.Date][v.Name] = Test{
				RtkAg: v.RtkAg,
				Pcr:   v.Pcr,
			}
		} else {
			res[v.Date] = map[string]Test{v.Name: {
				RtkAg: v.RtkAg,
				Pcr:   v.Pcr,
			}}
		}
	}
	return res
}

func getPopulation() map[string]Population {
	res := map[string]Population{}
	pdata, _ := readCSVFromUrl("https://raw.githubusercontent.com/MoH-Malaysia/covid19-public/main/static/population.csv")
	for i, prow := range pdata {
		if i == 0 {
			continue
		}
		pr := strings.Split(prow[0], ",")
		p, _ := strconv.Atoi(pr[2])
		po18, _ := strconv.Atoi(pr[3])
		po60, _ := strconv.Atoi(pr[4])
		po12, _ := strconv.Atoi(pr[5])
		res[pr[0]] = Population{
			Population: p,
			Over18:     po18,
			Over12:     po12,
			Over60:     po60,
		}
	}
	return res
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

func sendToSlack(rec Record) {
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
	str := fmt.Sprintf("%s Data as of %s\n New Cases: %d \n Import Cases: %d \n Recovered Cases: %d \n New Deaths: %d \n New Brought in Dead (BID): %d\n Actual COVID Deaths: %d \n", flags["Malaysia"], rec.Date, rec.NewCases, rec.ImportCases, rec.RecoveredCases, rec.Death.NewDeaths, rec.Death.BIDDeaths, rec.Death.ActualDeaths)
	req, _ := json.Marshal(map[string]string{
		"text": str,
	})
	_, _ = http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(req))
	for k, v := range rec.States {
		s := fmt.Sprintf("%s %s as of %s\n New Cases: %d \n Import Cases: %d \n Recovered Cases: %d \n New Deaths: %d \n Actual Deaths: %d", flags[k], k, rec.Date, v.NewCases, v.ImportCases, v.RecoveredCases, v.Death.NewDeaths, v.Death.ActualDeaths)
		req, _ := json.Marshal(map[string]string{
			"text": s,
		})
		_, _ = http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(req))
	}
}
