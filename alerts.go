package vigilant

import (
	"fmt"
)

// Alert functions are used to send alerts, they are viewable in the Vigilant Dashboard.
// The Vigilant agent collects a bunch of metadata about the alert, such as the stack trace, function name, file name, and line number.
// This data is used to provide more context about the alert in the Vigilant Dashboard.

// -------------- //
// --- Alerts --- //
// -------------- //

// SendAlert captures an alert and sends it to the agent.
//
// This function is used to send a simple alert with just a title.
//
// Example usage:
//
//	err := db.Query(...)
//	if err != nil {
//		SendAlert("Database Query Failed")
//	}
func SendAlert(title string) {
	if gateNilAgent() || gateEmptyAlertTitle(title) {
		return
	}

	globalAgent.sendAlert(title, nil)
}

// SendAlertw captures an alert and sends it to the agent with additional attributes.
//
// This function is useful when you want to provide more context with key-value pairs.
//
// Example usage:
//
//	SendAlertw("Database Query Failed", "db", "postgres", "retry", "true")
func SendAlertw(title string, keyVals ...any) {
	if gateNilAgent() || gateEmptyAlertTitle(title) {
		return
	}

	attrs, err := keyValsToMap(keyVals...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendAlert(title, attrs)
}

// SendAlertt captures an alert and sends it to the agent with typed attributes.
//
// This function is used when you have a map of attributes to send along with the alert.
//
// Example usage:
//
//	SendAlertt("Database Query Failed", map[string]string{"db": "postgres", "retry": "true"})
func SendAlertt(title string, attrs map[string]string) {
	if gateNilAgent() || gateEmptyAlertTitle(title) {
		return
	}

	globalAgent.sendAlert(title, attrs)
}

// writeAlertPassthrough writes an alert message to the agent
// this is an internal function that is used to write alert messages to stdout
func writeAlertPassthrough(title string, attrs map[string]string) {
	if len(attrs) > 0 {
		formattedAttrs := prettyPrintAttributes(attrs)
		fmt.Printf("[ALERT] %s %s\n", title, formattedAttrs)
	} else {
		fmt.Printf("[ALERT] %s\n", title)
	}
}
