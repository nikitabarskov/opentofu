package encryption_test

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/opentofu/opentofu/internal/encryption"
	"github.com/opentofu/opentofu/internal/encryption/keyprovider/static"
	"github.com/opentofu/opentofu/internal/encryption/method/aesgcm"
	"github.com/opentofu/opentofu/internal/encryption/registry/lockingencryptionregistry"
)

var (
	ConfigA = `
backend {
	enforced = true
}
`
	ConfigB = `
key_provider "static" "basic" {
	key = "6f6f706830656f67686f6834616872756f3751756165686565796f6f72653169"
}
method "aes_gcm" "example" {
	cipher = key_provider.static.basic
}
planfile {
	method = method.aes_gcm.example
}

backend {
	method = method.aes_gcm.example
}
`
)

// ExampleEncryption demonstrates how to use the encryption package to encrypt and decrypt data.
// This example uses the static key provider and the AES-GCM method.
// It also shows how to merge multiple configurations.
func ExampleEncryption() {
	// Construct a new registry
	// the registry is where we store the key providers and methods
	reg := lockingencryptionregistry.New()
	if err := reg.RegisterKeyProvider(static.New()); err != nil {
		panic(err)
	}
	if err := reg.RegisterMethod(aesgcm.New()); err != nil {
		panic(err)
	}

	// Load the 2 different configurations
	cfgA, diags := encryption.LoadConfigFromString("Test Source A", ConfigA)
	handleDiags(diags)

	cfgB, diags := encryption.LoadConfigFromString("Test Source B", ConfigB)
	handleDiags(diags)

	// Merge the configurations
	cfg := encryption.MergeConfigs(cfgA, cfgB)

	// Construct the encryption object
	enc, diags := encryption.New(reg, cfg)
	handleDiags(diags)

	// Encrypt the data, for this example we will be using the string "test",
	// but in a real world scenario this would be the plan file.
	sourceData := []byte("test")
	encrypted, err := enc.PlanFile().EncryptPlan(sourceData)
	if err != nil {
		panic(err)
	}
	if string(encrypted) == "test" {
		panic("The data has not been encrypted!")
	}

	// Decrypt
	decryptedPlan, err := enc.PlanFile().DecryptPlan(encrypted)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", decryptedPlan)
	// Output: test
}

func handleDiags(diags hcl.Diagnostics) {
	for _, d := range diags {
		println(d.Error())
	}
	if diags.HasErrors() {
		panic(diags.Error())
	}
}
