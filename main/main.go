package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type State struct {
	State     string `xml:"state,attr"`
	Reason    string `xml:"reason,attr"`
	ReasonTTL string `xml:"reason_ttl,attr"`
}

type Service struct {
	Name       string `xml:"name,attr"`
	Mehod      string `xml:"method,attr"`
	Confidence int    `xml:"conf,attr"`
}

type Port struct {
	Protocol string  `xml:"protocol,attr"`
	PortID   string  `xml:"portid,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}

type Nmaprun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Ports   []Port   `xml:"host>ports>port"`
}

type singlChannel struct {
	scanChannel chan string
}

type singleDB struct {
	dbClient  *mongo.Client
	dbChannel chan []Port
}

var singleDBInstance *singleDB
var singlChannelInstance *singlChannel

func GetSingleChannelInstance() *singlChannel {

	if singlChannelInstance == nil {
		singlChannelInstance = &singlChannel{}
		singlChannelInstance.scanChannel = make(chan string, 200048)
	}

	return singlChannelInstance
}

func GetSingleDBInstance() *singleDB {
	if singleDBInstance == nil {
		singleDBInstance = &singleDB{}
		singleDBInstance.dbChannel = make(chan []Port, 88888482)
	}

	return singleDBInstance
}

type addreObj struct {
	IP string `json:"ip"`
}

func main() {
	fmt.Println("started")
	go portScanServiceFromSingleChannel()
	go startDBConnection()

	router := gin.Default()
	router.GET("/addresses", getIp)
	router.POST("/addresses", postIp)
	router.Run("localhost:8080")
}

func postIp(c *gin.Context) {
	fmt.Println("posting your request")
	var newAddress addreObj
	if err := c.ShouldBindJSON(&newAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()})
		return
	}

	writeIpOnAChannel(newAddress)

	c.IndentedJSON(http.StatusCreated, newAddress)
}

func writeIpOnAChannel(ipAddress addreObj) {
	fmt.Println("writing on the channel")
	singleChannel := GetSingleChannelInstance()
	singleChannel.scanChannel <- ipAddress.IP
	fmt.Println("here's what written on the channel: ", ipAddress.IP)
}

func runNmapForIp(ipAddress string) {
	fmt.Println("running nmap on the terminal")
	cmd := exec.Command("nmap", "-oX", "./scanResult.xml", ipAddress)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("error occured:", err)
	}

	fmt.Printf("%s", out)
	readNmapResultsFromFile()
}

func portScanServiceFromSingleChannel() {
	fmt.Println("scanning...")
	singleChannel := GetSingleChannelInstance()
	for ch := range singleChannel.scanChannel {
		runNmapForIp(ch)
	}
}

func getIp(c *gin.Context) {
	c.JSON(200, gin.H{"ip": "1.1.1.1"})
}

func startDBConnection() []Port {
	client, err := mongo.Connect(context.TODO(),
		options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		panic(err)
	}

	fmt.Println("db connection started")

	dbChannel := GetSingleDBInstance().dbChannel
	GetSingleDBInstance().dbClient = client

	var portsRead []Port
	for el := range dbChannel {
		portsRead = el
		fmt.Println("the read port is ", el)
		writeDataToDB(portsRead)
	}

	fmt.Println("read ports are : ", portsRead, client)
	return portsRead
}

func writeDataToDB(ports []Port) {
	client := GetSingleDBInstance().dbClient
	nmapCollection := client.Database("admin").Collection("nmapResult")

	for i := 0; i < len(ports); i++ {
		nmapRes := bson.D{
			{Key: "portID", Value: ports[i].PortID},
			{Key: "protocol", Value: ports[i].Protocol},
			{Key: "state", Value: ports[i].State.State},
			{Key: "stateReason", Value: ports[i].State.Reason},
			{Key: "stateReasonTTL", Value: ports[i].State.ReasonTTL},
			{Key: "serviceName", Value: ports[i].Service.Name},
			{Key: "serviceConfidence", Value: ports[i].Service.Confidence},
			{Key: "serviceMethod", Value: ports[i].Service.Mehod}}

		result, err := nmapCollection.InsertOne(context.TODO(), nmapRes)

		fmt.Println(result)
		if err != nil {
			panic(err)
		}
	}

}

func readNmapResultsFromFile() {
	var nmaprun Nmaprun
	file, err := os.Open("scanResult.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer file.Close()

	decoder := xml.NewDecoder(file)
	err = decoder.Decode(&nmaprun)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}

	dbChannel := GetSingleDBInstance().dbChannel
	dbChannel <- nmaprun.Ports

	fmt.Println("db channel content : ", dbChannel)
}
