package swserver

import (
	"net/http"

	"github.com/emicklei/go-restful"
	//"github.com/icecrime/docker-api/api"
	api "github.com/docker/docker/api/server/swagger/api"
)

// containersServer implements ContainersService by exposing HTTP routes and
// forwarding requests to an underlying implementation.
type containersServer struct {
	*restful.WebService
	impl api.ContainersService
}

func newContainersServer(impl api.ContainersService) *containersServer {
	s := &containersServer{
		impl:       impl,
		WebService: &restful.WebService{},
	}
	s.installRoutes(s.WebService)
	return s
}

func (s *containersServer) installRoutes(ws *restful.WebService) {
	ws.Path("/containers").
		Doc("Containers management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("ps").To(s.List).
		Doc("List containers").
		Param(ws.QueryParameter("all", "Show all containers").DataType("string").DefaultValue("false")).
		Param(ws.QueryParameter("limit", "Maximum returns (0: unlimited)").DataType("int").DefaultValue("0")).
		Param(ws.QueryParameter("since", "Only show containers created after timestamp").DataType("int").DefaultValue("0")).
		Param(ws.QueryParameter("before", "Only show containers created before timestamp").DataType("int").DefaultValue("0")).
		Param(ws.QueryParameter("size", "Return the containers size").DataType("string").DefaultValue("false")).
		Param(ws.QueryParameter("filters", "Filter containers").DataType("map[string][]string")).
		Returns(200, "Container list", []*api.Container{}))

	// TODO - get these to list the correct return type
	ws.Route(ws.POST("create").To(s.Create).
		Doc("Create a container").
		Returns(200, "Newly created container", nil))

	ws.Route(ws.POST("start").To(s.Start).
		Doc("Start a container").
		Returns(200, "Container started", nil))
}

func (s *containersServer) List(request *restful.Request, response *restful.Response) {
	params := &api.ListContainersParams{}

	if all, err := booleanValue(request.QueryParameter("all"), false); err == nil {
		params.All = all
	} else {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	containerList, err := s.impl.List(params)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteEntity(containerList)
}

func (s *containersServer) Create(request *restful.Request, response *restful.Response) {
	params := map[string]interface{}{
		"Image":    "busybox",
		"Hostname": "PoCHostname",
	}

	containerID, err := s.impl.Create(params)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteEntity(containerID)
}

func (s *containersServer) Start(request *restful.Request, response *restful.Response) {
	containerID := "testID"

	containerStatus, err := s.impl.Start(containerID)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteEntity(containerStatus)
}
