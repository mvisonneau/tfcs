package tfcw

import (
	"bytes"
	"net"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	providerVault "github.com/mvisonneau/tfcw/pkg/providers/vault"
	"github.com/mvisonneau/tfcw/pkg/schemas"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetVaultValues(t *testing.T) {
	ln, client := createTestVault(t)
	defer ln.Close()
	c := Client{
		Vault: &providerVault.Client{
			Client: client,
		},
	}

	path := "secret/foo"
	key := "foo"
	v := &schemas.Variable{
		Vault: &schemas.Vault{
			Path: &path,
			Key:  &key,
		},
	}

	value, err := c.getVaultValues(v)
	assert.Equal(t, nil, err)
	assert.Equal(t, "bar", value[0].Value)
}

func TestIsVariableAlreadyProcessed(t *testing.T) {
	c := &Client{
		ProcessedVariables: map[string]schemas.VariableKind{},
	}

	v1 := "foo"
	assert.Equal(t, false, c.isVariableAlreadyProcessed(v1, schemas.VariableKindEnvironment))
	assert.Equal(t, true, c.isVariableAlreadyProcessed(v1, schemas.VariableKindEnvironment))
	assert.Equal(t, false, c.isVariableAlreadyProcessed(v1, schemas.VariableKindTerraform))
	assert.Equal(t, true, c.isVariableAlreadyProcessed(v1, schemas.VariableKindTerraform))
}

func TestLogVariableWithValue(t *testing.T) {
	// redirect logs to str variable
	var str bytes.Buffer
	log.SetOutput(&str)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	v := &schemas.VariableWithValue{
		Variable: schemas.Variable{
			Name: "foo",
			Kind: schemas.VariableKindEnvironment,
		},
		Value: "bar",
	}

	logVariableWithValue(v, true)
	assert.Equal(t, "level=info msg=\"[DRY-RUN] Set variable 'foo' (environment) : **********\"\n", str.String())

	// no dry-mode
	str.Reset()
	logVariableWithValue(v, false)
	assert.Equal(t, "level=info msg=\"Set variable 'foo' (environment)\"\n", str.String())
}

func TestSecureSensitiveString(t *testing.T) {
	assert.Equal(t, "**********", secureSensitiveString("foo"))
	assert.Equal(t, "l********e", secureSensitiveString("love"))
}

func createTestVault(t *testing.T) (net.Listener, *api.Client) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	assert.Nil(t, err)
	client.SetToken(rootToken)

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"foo": "bar",
		"baz": "baz",
	})
	assert.Nil(t, err)

	return ln, client
}
