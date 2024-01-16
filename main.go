package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/xistaminose/workflow"
)

// Simulate receiving a ride request
func ReceiveRideRequest() (string, error) {
	fmt.Println("Received ride request")
	time.Sleep(1 * time.Second) // Simulate delay
	return "RideRequestID", nil
}

// Simulate user validation
func ValidateUser() (string, error) {
	fmt.Println("Validating user")
	time.Sleep(3 * time.Second) // Simulate delay
	return "ValidUser", nil
}

// Simulate searching for nearby drivers
func SearchForNearbyDrivers() ([]string, error) {
	fmt.Println("Searching for nearby drivers")
	time.Sleep(2 * time.Second) // Simulate delay

	//return nil, errors.New("No drivers found")
	return []string{"Driver1", "Driver2"}, nil
}

// Simulate driver selection
func SelectDriver() (string, error) {
	fmt.Println("Selecting driver")
	time.Sleep(1 * time.Second) // Simulate delay
	return "SelectedDriver", nil
}

// Simulate sending ride details
func SendRideDetails() (string, error) {
	fmt.Println("Sending ride details")
	time.Sleep(1 * time.Second) // Simulate delay
	return "RideDetailsSent", nil
}

// Simulate tracking the ride
func TrackRide() (string, error) {
	fmt.Println("Tracking ride")
	time.Sleep(1 * time.Second) // Simulate delay
	return "RideTracked", nil
}

// Simulate processing payment
func ProcessPayment() (string, error) {
	fmt.Println("Processing payment")
	time.Sleep(1 * time.Second) // Simulate delay

	//return "", errors.New("Payment failed")
	return "PaymentProcessed", nil
}

// Simulate rating the experience
func RateExperience() (string, error) {
	fmt.Println("Rating experience")
	time.Sleep(1 * time.Second) // Simulate delay
	return "ExperienceRated", nil
}

// Simulate storing ride data
func StoreRideData() (string, error) {
	fmt.Println("Storing ride data")
	time.Sleep(1 * time.Second) // Simulate delay
	return "RideDataStored", nil
}

// Simulate notifying user of promotions
func NotifyUserOfPromotions() (string, error) {
	fmt.Println("Notifying user of promotions")
	time.Sleep(1 * time.Second) // Simulate delay
	//return "UserNotified", nil
	return "", errors.New("Notification failed")
}

// Simulate finalizing the ride
func FinalizeRide() (string, error) {
	fmt.Println("Finalizing ride")
	time.Sleep(1 * time.Second) // Simulate delay
	return "RideFinalized", nil
}

// Simulate a ValidateDriver function
func ValidateDriver() (string, error) {
	fmt.Println("Validating driver")
	time.Sleep(1 * time.Second) // Simulate delay
	return "ValidDriver", nil
}

func main() {
	// Initialize the workflow
	wf, err := workflow.NewWorkflow(1, true)
	if err != nil {
		panic(err)
	}

	// Create nodes
	nodeReceiveRideRequest := wf.CreateNode(ReceiveRideRequest)
	nodeValidateUser := wf.CreateNode(ValidateUser)
	nodeValidateDriver := wf.CreateNode(ValidateDriver)
	nodeSearchForDrivers := wf.CreateNode(SearchForNearbyDrivers)
	nodeSelectDriver := wf.CreateNode(SelectDriver)
	nodeSendRideDetails := wf.CreateNode(SendRideDetails)
	nodeTrackRide := wf.CreateNode(TrackRide)
	nodeProcessPayment := wf.CreateNode(ProcessPayment)
	nodeRateExperience := wf.CreateNode(RateExperience)
	nodeStoreRideData := wf.CreateNode(StoreRideData)
	nodeNotifyPromotions := wf.CreateNode(NotifyUserOfPromotions)
	nodeFinalizeRide := wf.CreateNode(FinalizeRide)

	// Set dependencies
	wf.AddDependency(nodeValidateUser, nodeReceiveRideRequest)
	wf.AddDependency(nodeSearchForDrivers, nodeReceiveRideRequest)
	wf.AddDependency(nodeSelectDriver, nodeSearchForDrivers)
	wf.AddDependency(nodeValidateDriver, nodeReceiveRideRequest)
	wf.AddDependency(nodeSendRideDetails, nodeValidateUser, nodeSelectDriver)
	wf.AddDependency(nodeTrackRide, nodeSendRideDetails)
	wf.AddDependency(nodeProcessPayment, nodeTrackRide)
	wf.AddDependency(nodeRateExperience, nodeProcessPayment)
	wf.AddDependency(nodeStoreRideData, nodeRateExperience, nodeProcessPayment)
	wf.AddDependency(nodeNotifyPromotions, nodeProcessPayment)
	wf.AddDependency(nodeFinalizeRide, nodeStoreRideData, nodeRateExperience, nodeNotifyPromotions)

	// Run the workflow
	if err := wf.Run(false); err != nil {
		fmt.Println("Workflow execution error:", err)
	}

	dotRepresentation := wf.ToDOT()
	os.WriteFile("workflow.dot", []byte(dotRepresentation), 0644)
	// run the dot command to generate the image
	cmd := exec.Command("dot", "-Tpng", "workflow.dot", "-o", "workflow2.png")
	if err := cmd.Run(); err != nil {
		fmt.Println("Error generating image:", err)
	}

}
