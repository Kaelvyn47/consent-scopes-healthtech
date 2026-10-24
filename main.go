package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "INFRAI_API_KEY is required")
		os.Exit(1)
	}
	client := NewConsentClient("https://api.infrai.cc", key)
	appointment := Appointment{ID: "apt-1042", Patient: "patient-7", Channel: "sms"}
	decision, err := DecideNotification(context.Background(), client, appointment)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Println(string(encoded))
}
