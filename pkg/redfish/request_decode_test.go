package redfish

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
)

// These reproduce real, spec-compliant request bodies at the JSON decode
// boundary: no Redfish client sends "Id" or "Name" when creating a session
// or patching a resource, since both are server-assigned. openapi-generator
// bakes an "Id"/"Name" required check into every Resource-derived model's
// UnmarshalJSON regardless of context, which used to reject both these
// bodies with a 422 "field 'Id' is required" -- see
// hack/redfish/vendor-redfish-schemas.

func TestSessionCreateRequestDecodesWithoutIdentityFields(t *testing.T) {
	body := []byte(`{"UserName": "admin", "Password": "supersecret"}`)

	var session server.SessionV171Session
	err := json.Unmarshal(body, &session)

	assert.NoError(t, err)
	if assert.NotNil(t, session.UserName) {
		assert.Equal(t, "admin", *session.UserName)
	}
	if assert.NotNil(t, session.Password) {
		assert.Equal(t, "supersecret", *session.Password)
	}
}

func TestComputerSystemPatchRequestDecodesWithoutIdentityFields(t *testing.T) {
	body := []byte(`{"Boot":{"BootSourceOverrideTarget":"Pxe","BootSourceOverrideEnabled":"Once"}}`)

	var computerSystem server.ComputerSystemV1220ComputerSystem
	err := json.Unmarshal(body, &computerSystem)

	assert.NoError(t, err)
	assert.Equal(t, server.COMPUTERSYSTEMBOOTSOURCE_PXE, computerSystem.Boot.BootSourceOverrideTarget)
	assert.Equal(t, server.COMPUTERSYSTEMV1220BOOTSOURCEOVERRIDEENABLED_ONCE, computerSystem.Boot.BootSourceOverrideEnabled)
}
