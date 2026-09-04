package server

// AssertstringRequired and AssertstringConstraints compensate for a gap in
// openapi-generator's go-server template (confirmed as of 7.25.0): for an
// optional field whose type resolves to a bare Go primitive (here, string)
// rather than a named model, the generated Required/Constraints functions
// for the containing type still call AssertstringRequired /
// AssertstringConstraints on it, but the generator only ever emits
// Assert<Type>Required/Constraints definitions for named models -- never
// for bare primitives. This file is hand-written, not generated: it will
// never collide with anything openapi-generator writes, and generate.sh's
// cleanup step (which only deletes model_*.go) leaves it alone.
//
// Every generated Assert<Type>Required doc comment states the intended
// design: "Primitive required fields are validated for JSON request bodies
// in UnmarshalJSON so zero values remain valid." These are a no-op for the
// same reason: UnmarshalJSON already covers it, and neither field this
// currently backs (RedundancyV170Redundancy.Mode,
// ResourceV1240Placement.RackOffsetUnits) declares a format constraint in
// the OpenAPI spec.
//
// Re-check `go build ./pkg/generated/...` after any hack/redfish/spec or
// openapi-generator version change: if a newly reachable schema needs the
// same treatment for a different primitive (e.g. int32), add it here the
// same way; if openapi-generator starts emitting these itself, delete this
// file.
func AssertstringRequired(string) error    { return nil }
func AssertstringConstraints(string) error { return nil }
