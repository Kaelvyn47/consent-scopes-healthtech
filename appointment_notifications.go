package main

import "context"

const appointmentNotificationScope = "appointment_notifications"

type Appointment struct {
	ID      string `json:"id"`
	Patient string `json:"patient_id"`
	Channel string `json:"channel"`
}

type NotificationDecision struct {
	AppointmentID string `json:"appointment_id"`
	Send          bool   `json:"send"`
	Reason        string `json:"reason"`
}

func DecideNotification(ctx context.Context, client *ConsentClient, appointment Appointment) (NotificationDecision, error) {
	granted, err := client.Check(ctx, appointment.Patient, appointmentNotificationScope)
	if err != nil {
		return NotificationDecision{}, err
	}
	if !granted {
		return NotificationDecision{AppointmentID: appointment.ID, Reason: "patient consent is absent"}, nil
	}
	return NotificationDecision{AppointmentID: appointment.ID, Send: true, Reason: "patient consent is granted"}, nil
}
