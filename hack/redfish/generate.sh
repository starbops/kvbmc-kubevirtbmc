#!/usr/bin/env sh
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../../" && pwd)"
cd "$ROOT_DIR"

echo "Working directory: $(pwd)"

command -v openapi-generator || {
  echo "openapi-generator not found. Please install it first."
  exit 1
}

# goimports is used to post-process the generated code
command -v goimports || {
  echo "goimports not found. Please install it first."
  exit 1
}

if [ "$(uname)" = "Darwin" ]; then
    export JAVA_HOME=`/usr/libexec/java_home -v 1.11`
fi

# Trim the spec down to only the operations pkg/redfish actually implements,
# so openapi-generator doesn't generate stubs for the thousands it doesn't.
# See hack/redfish/spec/implemented-operations.yaml.
TRIMMED_SPEC="$(mktemp -t kubevirtbmc-redfish-openapi.XXXXXX).yaml"
trap 'rm -f "$TRIMMED_SPEC"' EXIT
go run ./hack/redfish/trim-redfish-spec \
    -input ./hack/redfish/spec/openapi.yaml \
    -allowlist ./hack/redfish/spec/implemented-operations.yaml \
    -output "$TRIMMED_SPEC"

# Generate the models and server stubs from the trimmed OpenAPI spec
_JAVA_OPTIONS="-DmaxYamlCodePoints=99999999" GO_POST_PROCESS_FILE="goimports -w" openapi-generator generate \
    -i "$TRIMMED_SPEC" \
    -o ./pkg/generated/redfish/ \
    -g go-server \
    --package-name server \
    --enable-post-process-file \
    -p sourceFolder=server,onlyInterfaces=true,outputAsLibrary=true,enumClassPrefix=true

