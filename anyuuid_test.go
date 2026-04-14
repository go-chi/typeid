package typeid_test

import (
	"encoding/json"
	"fmt"

	"github.com/go-chi/typeid"
)

type Mode string

const (
	ModeLive    Mode = "key"
	ModeSandbox Mode = "key_sandbox"
)

type ApiKeyID struct {
	typeid.AnyUUID
}

func NewApiKeyID(mode Mode) ApiKeyID {
	u, err := typeid.NewAnyUUID(string(mode))
	if err != nil {
		panic(err) // can't happen unless crypto/rand is not available
	}
	return ApiKeyID{AnyUUID: u}
}

func (id *ApiKeyID) UnmarshalText(data []byte) error {
	if err := id.AnyUUID.UnmarshalText(data); err != nil {
		return err
	}
	switch id.AnyUUID.Prefix() {
	case string(ModeLive), string(ModeSandbox):
		return nil
	default:
		return fmt.Errorf("invalid api key prefix: %q", id.AnyUUID.Prefix())
	}
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

	// Invalid prefix, expect error
	data = []byte(`{"id":"key_invalid_prefix_01jcp1ss00edg828t5cy4tqkff", "description":"Invalid API Key"}`)
	var unknownRequest Request
	err := json.Unmarshal(data, &unknownRequest)

	fmt.Println(sandboxRequest.ID.Prefix())
	fmt.Println(liveRequest.ID.Prefix())
	fmt.Println(unknownRequest.ID.Prefix(), err)

	// Output:
	// key_sandbox
	// key
	// key_invalid_prefix invalid api key prefix: "key_invalid_prefix"
}
