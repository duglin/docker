package swmain

import (
	// "net/http"

	log "github.com/Sirupsen/logrus"
	//"github.com/icecrime/docker-api/api"
	//"github.com/icecrime/docker-api/server"
	api "github.com/docker/docker/api/server/swagger/api"
	// server "github.com/docker/docker/api/server/swagger/server"
)

type TestServer struct{}

func (*TestServer) List(p *api.ListContainersParams) ([]*api.Container, error) {
	log.Infof("TestServer.List(%v)", p)
	return []*api.Container{}, nil
}

func (*TestServer) Ping() (string, error) {
	log.Info("TestServer.Ping()")
	return "OK", nil
}

func (*TestServer) Version() (*api.Version, error) {
	log.Info("TestServer.Version()")
	return &api.Version{
		APIVersion:    "APIVersion",
		Arch:          "Arch",
		GitCommit:     "GitCommit",
		GoVersion:     "GoVersion",
		KernelVersion: "KernelVersion",
		OS:            "OS",
		Version:       "Version",
	}, nil
}

func (*TestServer) Create(p interface{}) (*api.ListContainerID, error) {
	log.Infof("TestServer.Create(%v)", p)
	var containerID *api.ListContainerID
	return containerID, nil
}

func (*TestServer) Start(p string) (int, error) {
	log.Infof("TestServer.Start(%v)", p)
	var statusCode int
	return statusCode, nil
}

/*
func main() {
	srv := server.New(&TestServer{}, "swagger-ui/dist/")
	if err := http.ListenAndServe("127.0.0.1:8080", srv); err != nil {
		panic(err)
	}
}
*/
