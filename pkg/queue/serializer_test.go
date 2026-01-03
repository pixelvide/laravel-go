package queue

import (
	"encoding/json"
	"testing"
)

func TestUnserializeCommand_UserExample(t *testing.T) {
	jsonPayload := `{"displayName":"App\\Jobs\\TestGoServiceJob","job":"Illuminate\\Queue\\CallQueuedHandler@call","maxTries":null,"delay":null,"timeout":null,"timeoutAt":null,"data":{"commandName":"App\\Jobs\\TestGoServiceJob","command":"O:25:\"App\\Jobs\\TestGoServiceJob\":10:{s:7:\"message\";s:17:\"Hello from Server\";s:9:\"timestamp\";i:1767438617;s:6:\"\u0000*\u0000job\";N;s:10:\"connection\";N;s:5:\"queue\";s:17:\"go-services-queue\";s:15:\"chainConnection\";N;s:10:\"chainQueue\";N;s:5:\"delay\";N;s:10:\"middleware\";a:0:{}s:7:\"chained\";a:0:{}}"},"telescope_uuid":"a0bf6318-397a-4372-9122-f112387beb63"}`

	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonPayload), &data); err != nil {
		t.Fatalf("Failed to unmarshal outer JSON: %v", err)
	}

	commandData, ok := data["data"]
	if !ok {
		t.Fatal("Missing data field")
	}

	unserialized, err := UnserializeCommand(commandData)
	if err != nil {
		t.Fatalf("UnserializeCommand failed: %v", err)
	}

	if unserialized == nil {
		t.Fatal("Expected unserialized data, got nil")
	}

	// Check specific fields
	// message should be "Hello from Server"
	msg := GetPHPProperty(unserialized, "message")
	if msg != "Hello from Server" {
		t.Errorf("Expected message 'Hello from Server', got %v", msg)
	}

	// timestamp should be 1767438617
	ts := GetPHPProperty(unserialized, "timestamp")
	if ts == nil {
		t.Error("Expected timestamp, got nil")
	}

	// Check protected property * job (s:6:"\u0000*\u0000job";N;)
	// It is Null (N), so we expect nil.
	jobProp := GetPHPProperty(unserialized, "job")
	if jobProp != nil {
		t.Errorf("Expected job to be nil, got %v", jobProp)
	}
}
