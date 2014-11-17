package daemon

import (
	"time"

	"github.com/docker/docker/engine"
)

func (daemon *Daemon) ContainerWait(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s", job.Name)
	}
	name := job.Args[0]
	if container := daemon.Get(name); container != nil {
		status, _ := container.WaitStop(-1 * time.Second)
		job.Printf("%d\n", status)
		return engine.StatusOK
	}

	if execConfig, _ := daemon.getExecConfig(name); execConfig != nil {
		exitCode, _ := execConfig.WaitStop(-1 * time.Second)
		job.Printf("%d\n", exitCode)
		return engine.StatusOK
	}

	return job.Errorf("%s: No such container or exec: %s", job.Name, name)
}
