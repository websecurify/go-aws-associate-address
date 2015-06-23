package main // import "websecurify/go-aws-associate-address"

// ---
// ---
// ---

import (
	"os"
	"log"
	"net"
	"time"
	"syscall"
	"net/http"
	"io/ioutil"
	"os/signal"
	
	// ---
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ---
// ---
// ---

const (
	HTTP_TIMEOUT = time.Duration(3.0 * 1000) * time.Millisecond
)

// ---

var config = struct {
	Region string
	InstanceID string
	AllocationID string
} {
	Region: os.Getenv("REGION"),
	InstanceID: os.Getenv("INSTANCE_ID"),
	AllocationID: os.Getenv("ALLOCATION_ID"),
}

// ---

var globals = struct {
	InstanceID string
	EC2Service *ec2.EC2
} {
	InstanceID: getInstanceID(),
	EC2Service: getEC2Service(),
}

// ---
// ---
// ---

func getInstanceID() (string) {
	if config.InstanceID != "" {
		return config.InstanceID
	}
	
	// ---
	
	dial := func(network string, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, HTTP_TIMEOUT)
	}
	
	// ---
	
	transport := http.Transport{
		Dial: dial,
	}
	
	// ---
	
	client := http.Client{
		Transport: &transport,
	}
	
	// ---
	
	getRes, getResErr := client.Get("http://169.254.169.254/latest/meta-data/instance-id")
	
	if getResErr != nil {
		log.Fatal(getResErr)
	}
	
	// ---
	
	defer getRes.Body.Close()
	
	// ---
	
	body, bodyErr := ioutil.ReadAll(getRes.Body)
	
	if bodyErr != nil {
		log.Fatal(bodyErr)
	}
	
	// ---
	
	return string(body)
}

func getEC2Service() (*ec2.EC2) {
	return ec2.New(&aws.Config{Region: config.Region})
}

// ---
// ---
// ---

func getAddressAssociation() (string, string) {
	describeAddressesRes, describeAddressesErr := globals.EC2Service.DescribeAddresses(&ec2.DescribeAddressesInput{
		AllocationIDs: []*string{
			aws.String(config.AllocationID),
		},
	})
	
	if describeAddressesErr != nil {
		log.Fatal(describeAddressesErr)
	}
	
	// ---
	
	log.Println("address association queried")
	log.Println(awsutil.StringValue(describeAddressesRes))
	
	// ---
	
	addresses := describeAddressesRes.Addresses
	
	// ---
	
	var instanceID string
	var associationID string
	
	// ---
	
	if len(addresses) > 0 {
		if addresses[0].InstanceID != nil {
			instanceID = *addresses[0].InstanceID
		}
		
		if addresses[0].AssociationID != nil {
			associationID = *addresses[0].AssociationID
		}
	}
	
	// ---
	
	return instanceID, associationID
}

// ---

func associateAddress() (string) {
	associateRes, associateErr := globals.EC2Service.AssociateAddress(&ec2.AssociateAddressInput{
		InstanceID: aws.String(globals.InstanceID),
		AllocationID: aws.String(config.AllocationID),
	})
	
	if associateErr != nil {
		log.Fatal(associateErr)
	}
	
	// ---
	
	log.Println("address associated")
	log.Println(awsutil.StringValue(associateRes))
	
	// ---
	
	return *associateRes.AssociationID
}

func disassociateAddress(associationID string) {
	disassociateRes, disassociateErr := globals.EC2Service.DisassociateAddress(&ec2.DisassociateAddressInput{
		AssociationID: aws.String(associationID),
	})
	
	if disassociateErr != nil {
		log.Fatal(disassociateErr)
	}
	
	// ---
	
	log.Println("address disassociated")
	log.Println(awsutil.StringValue(disassociateRes))
}

// ---
// ---
// ---

func main() {
	instanceID, associationID := getAddressAssociation()
	
	if instanceID == "" {
		associationID = associateAddress()
	} else
	if instanceID != globals.InstanceID {
		log.Fatal("allocation ", config.AllocationID, " already associated to ", instanceID)
	}
	
	// ---
	
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	
	// ---
	
	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	
	// ---
	
	go func() {
		<-sigs
		
		// ---
		
		disassociateAddress(associationID)
		
		// ---
		
		done <- true
	}()
	
	// ---
	
	<-done
}

// ---
