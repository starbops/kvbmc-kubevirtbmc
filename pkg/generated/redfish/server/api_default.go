package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type DefaultAPIController struct {
	service      DefaultAPIServicer
	errorHandler ErrorHandler
}

type DefaultAPIOption func(*DefaultAPIController)

func WithDefaultAPIErrorHandler(h ErrorHandler) DefaultAPIOption {
	return func(c *DefaultAPIController) {
		c.errorHandler = h
	}
}

func NewDefaultAPIController(s DefaultAPIServicer, opts ...DefaultAPIOption) *DefaultAPIController {
	controller := &DefaultAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

func (c *DefaultAPIController) Routes() Routes {
	return Routes{
		"RedfishV1Get": Route{
			"RedfishV1Get",
			strings.ToUpper("Get"),
			"/redfish/v1",
			c.RedfishV1Get,
		},
		"RedfishV1Get_0": Route{
			"RedfishV1Get_0",
			strings.ToUpper("Get"),
			"/redfish/v1/",
			c.RedfishV1Get_0,
		},
		"RedfishV1ManagersGet": Route{
			"RedfishV1ManagersGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers",
			c.RedfishV1ManagersGet,
		},
		"RedfishV1ManagersManagerIdGet": Route{
			"RedfishV1ManagersManagerIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}",
			c.RedfishV1ManagersManagerIdGet,
		},
		"RedfishV1ManagersManagerIdActionsManagerResetPost": Route{
			"RedfishV1ManagersManagerIdActionsManagerResetPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/Actions/Manager.Reset",
			c.RedfishV1ManagersManagerIdActionsManagerResetPost,
		},
		"RedfishV1ManagersManagerIdVirtualMediaGet": Route{
			"RedfishV1ManagersManagerIdVirtualMediaGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaGet,
		},
		"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet": Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet,
		},
		"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost": Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}/Actions/VirtualMedia.EjectMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost,
		},
		"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost": Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}/Actions/VirtualMedia.InsertMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost,
		},
		"RedfishV1SessionServiceSessionsPost": Route{
			"RedfishV1SessionServiceSessionsPost",
			strings.ToUpper("Post"),
			"/redfish/v1/SessionService/Sessions",
			c.RedfishV1SessionServiceSessionsPost,
		},
		"RedfishV1SessionServiceSessionsSessionIdGet": Route{
			"RedfishV1SessionServiceSessionsSessionIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/SessionService/Sessions/{SessionId}",
			c.RedfishV1SessionServiceSessionsSessionIdGet,
		},
		"RedfishV1SessionServiceSessionsSessionIdDelete": Route{
			"RedfishV1SessionServiceSessionsSessionIdDelete",
			strings.ToUpper("Delete"),
			"/redfish/v1/SessionService/Sessions/{SessionId}",
			c.RedfishV1SessionServiceSessionsSessionIdDelete,
		},
		"RedfishV1SystemsGet": Route{
			"RedfishV1SystemsGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems",
			c.RedfishV1SystemsGet,
		},
		"RedfishV1SystemsComputerSystemIdGet": Route{
			"RedfishV1SystemsComputerSystemIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems/{ComputerSystemId}",
			c.RedfishV1SystemsComputerSystemIdGet,
		},
		"RedfishV1SystemsComputerSystemIdPatch": Route{
			"RedfishV1SystemsComputerSystemIdPatch",
			strings.ToUpper("Patch"),
			"/redfish/v1/Systems/{ComputerSystemId}",
			c.RedfishV1SystemsComputerSystemIdPatch,
		},
		"RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost": Route{
			"RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Systems/{ComputerSystemId}/Actions/ComputerSystem.Reset",
			c.RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost,
		},
		"RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost": Route{
			"RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Systems/{ComputerSystemId}/Actions/ComputerSystem.SetDefaultBootOrder",
			c.RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost,
		},
		"RedfishV1SystemsComputerSystemIdOperatingSystemGet": Route{
			"RedfishV1SystemsComputerSystemIdOperatingSystemGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems/{ComputerSystemId}/OperatingSystem",
			c.RedfishV1SystemsComputerSystemIdOperatingSystemGet,
		},
	}
}

func (c *DefaultAPIController) OrderedRoutes() []Route {
	return []Route{
		Route{
			"RedfishV1Get",
			strings.ToUpper("Get"),
			"/redfish/v1",
			c.RedfishV1Get,
		},
		Route{
			"RedfishV1Get_0",
			strings.ToUpper("Get"),
			"/redfish/v1/",
			c.RedfishV1Get_0,
		},
		Route{
			"RedfishV1ManagersGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers",
			c.RedfishV1ManagersGet,
		},
		Route{
			"RedfishV1ManagersManagerIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}",
			c.RedfishV1ManagersManagerIdGet,
		},
		Route{
			"RedfishV1ManagersManagerIdActionsManagerResetPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/Actions/Manager.Reset",
			c.RedfishV1ManagersManagerIdActionsManagerResetPost,
		},
		Route{
			"RedfishV1ManagersManagerIdVirtualMediaGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaGet,
		},
		Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet,
		},
		Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}/Actions/VirtualMedia.EjectMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost,
		},
		Route{
			"RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Managers/{ManagerId}/VirtualMedia/{VirtualMediaId}/Actions/VirtualMedia.InsertMedia",
			c.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost,
		},
		Route{
			"RedfishV1SessionServiceSessionsPost",
			strings.ToUpper("Post"),
			"/redfish/v1/SessionService/Sessions",
			c.RedfishV1SessionServiceSessionsPost,
		},
		Route{
			"RedfishV1SessionServiceSessionsSessionIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/SessionService/Sessions/{SessionId}",
			c.RedfishV1SessionServiceSessionsSessionIdGet,
		},
		Route{
			"RedfishV1SessionServiceSessionsSessionIdDelete",
			strings.ToUpper("Delete"),
			"/redfish/v1/SessionService/Sessions/{SessionId}",
			c.RedfishV1SessionServiceSessionsSessionIdDelete,
		},
		Route{
			"RedfishV1SystemsGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems",
			c.RedfishV1SystemsGet,
		},
		Route{
			"RedfishV1SystemsComputerSystemIdGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems/{ComputerSystemId}",
			c.RedfishV1SystemsComputerSystemIdGet,
		},
		Route{
			"RedfishV1SystemsComputerSystemIdPatch",
			strings.ToUpper("Patch"),
			"/redfish/v1/Systems/{ComputerSystemId}",
			c.RedfishV1SystemsComputerSystemIdPatch,
		},
		Route{
			"RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Systems/{ComputerSystemId}/Actions/ComputerSystem.Reset",
			c.RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost,
		},
		Route{
			"RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost",
			strings.ToUpper("Post"),
			"/redfish/v1/Systems/{ComputerSystemId}/Actions/ComputerSystem.SetDefaultBootOrder",
			c.RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost,
		},
		Route{
			"RedfishV1SystemsComputerSystemIdOperatingSystemGet",
			strings.ToUpper("Get"),
			"/redfish/v1/Systems/{ComputerSystemId}/OperatingSystem",
			c.RedfishV1SystemsComputerSystemIdOperatingSystemGet,
		},
	}
}

func (c *DefaultAPIController) RedfishV1Get(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.RedfishV1Get(r.Context())

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1Get_0(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.RedfishV1Get_0(r.Context())

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.RedfishV1ManagersGet(r.Context())

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdGet(r.Context(), managerIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdActionsManagerResetPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	var managerV1190ResetRequestBodyParam ManagerV1190ResetRequestBody
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&managerV1190ResetRequestBodyParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertManagerV1190ResetRequestBodyRequired(managerV1190ResetRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertManagerV1190ResetRequestBodyConstraints(managerV1190ResetRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdActionsManagerResetPost(r.Context(), managerIdParam, managerV1190ResetRequestBodyParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdVirtualMediaGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdVirtualMediaGet(r.Context(), managerIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	virtualMediaIdParam := params["VirtualMediaId"]
	if virtualMediaIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"VirtualMediaId"}, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdGet(r.Context(), managerIdParam, virtualMediaIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	virtualMediaIdParam := params["VirtualMediaId"]
	if virtualMediaIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"VirtualMediaId"}, nil)
		return
	}
	var bodyParam map[string]interface{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaEjectMediaPost(r.Context(), managerIdParam, virtualMediaIdParam, bodyParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	managerIdParam := params["ManagerId"]
	if managerIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ManagerId"}, nil)
		return
	}
	virtualMediaIdParam := params["VirtualMediaId"]
	if virtualMediaIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"VirtualMediaId"}, nil)
		return
	}
	var virtualMediaV163InsertMediaRequestBodyParam VirtualMediaV163InsertMediaRequestBody
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&virtualMediaV163InsertMediaRequestBodyParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertVirtualMediaV163InsertMediaRequestBodyRequired(virtualMediaV163InsertMediaRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertVirtualMediaV163InsertMediaRequestBodyConstraints(virtualMediaV163InsertMediaRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.RedfishV1ManagersManagerIdVirtualMediaVirtualMediaIdActionsVirtualMediaInsertMediaPost(r.Context(), managerIdParam, virtualMediaIdParam, virtualMediaV163InsertMediaRequestBodyParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SessionServiceSessionsPost(w http.ResponseWriter, r *http.Request) {
	var sessionV171SessionParam SessionV171Session
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&sessionV171SessionParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertSessionV171SessionRequired(sessionV171SessionParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertSessionV171SessionConstraints(sessionV171SessionParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.RedfishV1SessionServiceSessionsPost(r.Context(), sessionV171SessionParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SessionServiceSessionsSessionIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionIdParam := params["SessionId"]
	if sessionIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"SessionId"}, nil)
		return
	}
	result, err := c.service.RedfishV1SessionServiceSessionsSessionIdGet(r.Context(), sessionIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SessionServiceSessionsSessionIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionIdParam := params["SessionId"]
	if sessionIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"SessionId"}, nil)
		return
	}
	result, err := c.service.RedfishV1SessionServiceSessionsSessionIdDelete(r.Context(), sessionIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.RedfishV1SystemsGet(r.Context())

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsComputerSystemIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	computerSystemIdParam := params["ComputerSystemId"]
	if computerSystemIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ComputerSystemId"}, nil)
		return
	}
	result, err := c.service.RedfishV1SystemsComputerSystemIdGet(r.Context(), computerSystemIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsComputerSystemIdPatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	computerSystemIdParam := params["ComputerSystemId"]
	if computerSystemIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ComputerSystemId"}, nil)
		return
	}
	var computerSystemV1220ComputerSystemParam ComputerSystemV1220ComputerSystem
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&computerSystemV1220ComputerSystemParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}

	if err := AssertComputerSystemV1220ComputerSystemConstraints(computerSystemV1220ComputerSystemParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.RedfishV1SystemsComputerSystemIdPatch(r.Context(), computerSystemIdParam, computerSystemV1220ComputerSystemParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	computerSystemIdParam := params["ComputerSystemId"]
	if computerSystemIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ComputerSystemId"}, nil)
		return
	}
	var computerSystemV1220ResetRequestBodyParam ComputerSystemV1220ResetRequestBody
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&computerSystemV1220ResetRequestBodyParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertComputerSystemV1220ResetRequestBodyRequired(computerSystemV1220ResetRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertComputerSystemV1220ResetRequestBodyConstraints(computerSystemV1220ResetRequestBodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.RedfishV1SystemsComputerSystemIdActionsComputerSystemResetPost(r.Context(), computerSystemIdParam, computerSystemV1220ResetRequestBodyParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	computerSystemIdParam := params["ComputerSystemId"]
	if computerSystemIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ComputerSystemId"}, nil)
		return
	}
	var bodyParam map[string]interface{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		var requiredErr *RequiredError
		if errors.As(err, &requiredErr) {
			c.errorHandler(w, r, err, nil)
			return
		}
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	result, err := c.service.RedfishV1SystemsComputerSystemIdActionsComputerSystemSetDefaultBootOrderPost(r.Context(), computerSystemIdParam, bodyParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *DefaultAPIController) RedfishV1SystemsComputerSystemIdOperatingSystemGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	computerSystemIdParam := params["ComputerSystemId"]
	if computerSystemIdParam == "" {
		c.errorHandler(w, r, &RequiredError{"ComputerSystemId"}, nil)
		return
	}
	result, err := c.service.RedfishV1SystemsComputerSystemIdOperatingSystemGet(r.Context(), computerSystemIdParam)

	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}
