package typeid_test

import (
	"encoding/json"
	"fmt"

	"github.com/go-chi/typeid"
)

type Mode int8

const (
	ModeLive    Mode = 0
	ModeSandbox Mode = 1
)

func (m Mode) Prefix() string {
	switch m {
	case ModeSandbox:
		return "key_sandbox"
	default:
		return "key"
	}
}

type ApiKeyID = typeid.AnyUUID

func ParseApiKeyID(s string, mode Mode) (ApiKeyID, error) {
	var id typeid.AnyUUID
	if err := id.UnmarshalText([]byte(s)); err != nil {
		return ApiKeyID{}, err
	}
	if id.Prefix() != mode.Prefix() {
		return ApiKeyID{}, fmt.Errorf("invalid api key prefix: %s, expected: %v", id.Prefix(), mode.Prefix())
	}
	return id, nil
}

func NewApiKeyID(mode Mode) ApiKeyID {
	id, err := typeid.NewAnyUUID(mode.Prefix())
	if err != nil {
		panic(err) // can't happen unless crypto/rand is not available
	}
	return id
}

type Request struct {
	ID          ApiKeyID `json:"id"`
	Description string   `json:"description"`
}

func ExampleAnyUUID_json() {
	// Sandbox
	data, _ := json.Marshal(Request{ID: NewApiKeyID(ModeSandbox), Description: "Sandbox API Key"})
	var sandboxRequest Request
	_ = json.Unmarshal(data, &sandboxRequest)

	// Live
	data, _ = json.Marshal(Request{ID: NewApiKeyID(ModeLive), Description: "Live API Key"})
	var liveRequest Request
	_ = json.Unmarshal(data, &liveRequest)

	// Invalid prefix
	data = []byte(`{"id":"key_invalid_prefix_01jcp1ss00edg828t5cy4tqkff", "description":"Invalid API Key"}`)
	var unknownRequest Request
	_ = json.Unmarshal(data, &unknownRequest)
	_, err := ParseApiKeyID(unknownRequest.ID.String(), ModeLive)

	fmt.Println(sandboxRequest.ID.Prefix())
	fmt.Println(liveRequest.ID.Prefix())
	fmt.Println(unknownRequest.ID.Prefix(), err)

	// Output:
	// key_sandbox
	// key
	// key_invalid_prefix invalid api key prefix: key_invalid_prefix, expected: key
}
