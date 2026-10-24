package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecideNotificationRequiresPatientConsent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/auth/consent/check/patient-7/appointment_notifications" {
			t.Fatalf("request was %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("missing bearer authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"granted":false},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	decision, err := DecideNotification(context.Background(), NewConsentClient(server.URL, "test-key"), Appointment{ID: "apt-1042", Patient: "patient-7"})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Send {
		t.Fatal("notification must be withheld without consent")
	}
	if decision.Reason != "patient consent is absent" {
		t.Fatalf("unexpected reason: %s", decision.Reason)
	}
}
