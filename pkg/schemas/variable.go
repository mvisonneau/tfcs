package schemas

import (
	"fmt"
	"time"
)

// VariableKind represents the kind of variable we want to
// provision
type VariableKind string

const (
	// VariableKindTerraform refers to a 'terraform' variable in TFC
	VariableKindTerraform VariableKind = "terraform"

	// VariableKindEnvironment refers to an 'environment' variable in TFC
	VariableKindEnvironment VariableKind = "environment"
)

// VariableProvider represent the provider which can be used in order
// to process the variable
type VariableProvider string

const (
	// VariableProviderEnv refers to the 'env' variable provider
	VariableProviderEnv VariableProvider = "env"

	// VariableProviderS5 refers to the 's5' variable provider
	VariableProviderS5 VariableProvider = "s5"

	// VariableProviderVault refers to the 'vault' variable provider
	VariableProviderVault VariableProvider = "vault"
)

// Variable is a generic handler of variable characteristics
type Variable struct {
	Name      string  `hcl:"name,label"`
	Vault     *Vault  `hcl:"vault,block"`
	S5        *S5     `hcl:"s5,block"`
	Env       *Env    `hcl:"env,block"`
	Sensitive *bool   `hcl:"sensitive"`
	HCL       *bool   `hcl:"hcl"`
	TTL       *string `hcl:"ttl"`

	Kind  VariableKind
	Value string
}

// VariableWithValue is a generic handler for found variable values
type VariableWithValue struct {
	Variable
	Value string
}

// Variables is a slice of *Variable
type Variables []*Variable

// VariablesWithValues is a slice of *ComputedVariable
type VariablesWithValues []*VariableWithValue

// VariableExpirations holds the expiration times of variables, stored within TFC as
// a jsonmap within an __TFCW_VARIABLES_EXPIRATIONS environment variable (quite hacky..)
type VariableExpirations map[VariableKind]map[string]*VariableExpiration

// VariableExpiration contains the Time To Live (TTL) and when the value of a variable
// to expire is going to expire
type VariableExpiration struct {
	TTL      time.Duration `json:"ttl"`
	ExpireAt time.Time     `json:"expire_at"`
}

// GetProvider returns the VariableProvider that can be used for processing the variable
func (v *Variable) GetProvider() (*VariableProvider, error) {
	configuredProviders := 0
	var provider *VariableProvider

	if v.Env != nil {
		configuredProviders++
		p := VariableProviderEnv
		provider = &p
	}

	if v.S5 != nil {
		configuredProviders++
		p := VariableProviderS5
		provider = &p
	}

	if v.Vault != nil {
		configuredProviders++
		p := VariableProviderVault
		provider = &p
	}

	if configuredProviders != 1 {
		return nil, fmt.Errorf("you can't have more or less than one provider configured per variable. Found %d for '%s'", configuredProviders, v.Name)
	}

	return provider, nil
}
