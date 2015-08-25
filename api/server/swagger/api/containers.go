package swapi

// ContainersService provides the containers management aspects of a Docker
// server.
type ContainersService interface {
	List(*ListContainersParams) ([]*Container, error)
	Create(interface{}) (*ListContainerID, error)
	Start(string) (int, error)
}

// ExtendedContainerService provides the containers management aspects of a
// Docker server, including those who cannot be implemented over a strictly
// HTTP/1.x compliant connection.
type ExtendedContainerService interface {
	ContainersService
}

type ListContainersParams struct {
	All     bool
	Limit   int
	Since   int
	Before  int
	Size    bool
	Filters map[string][]string
}

type ListContainerID struct {
	ID string
}

// Port associates a container port to a host port bound to a particular IP
// address.
type Port struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

// Entrypoint encapsulates the container entrypoint.
// It might be represented as a string or an array of strings.
// We need to override the json decoder to accept both options.
// The JSON decoder will fail if the api sends an string and
// we try to decode it into an array of string.
type Entrypoint struct {
	parts []string
}

// Command encapsulates the container command.
// It might be represented as a string or an array of strings.
// We need to override the json decoder to accept both options.
// The JSON decoder will fail if the api sends an string and
// we try to decode it into an array of string.
type Command struct {
	parts []string
}

// HostConfig the non-portable Config structure of a container.
// Here, "non-portable" means "dependent of the host we are running on".
// Portable information *should* appear in Config.
type HostConfig struct {
	Binds           []string // List of volume bindings for this container
	ContainerIDFile string   // File (path) where the containerId is written
	//LxcConf *LxcConfig // Additional lxc configuration
	Memory           int64  // Memory limit (in bytes)
	MemorySwap       int64  // Total memory usage (memory + swap); set `-1` to disable swap
	CPUShares        int64  `json:"CpuShares"` // CPU shares (relative weight vs. other containers)
	CPUPeriod        int64  `json:"CpuPeriod"` // CPU CFS (Completely Fair Scheduler) period
	CpusetCpus       string // CpusetCpus 0-2, 0,1
	CpusetMems       string // CpusetMems 0-2, 0,1
	CPUQuota         int64  `json:"CpuQuota"` // CPU CFS (Completely Fair Scheduler) quota
	BlkioWeight      int64  // Block IO weight (relative weight vs. other containers)
	OomKillDisable   bool   // Whether to disable OOM Killer or not
	MemorySwappiness *int64 // Tuning container memory swappiness behaviour
	Privileged       bool   // Is the container in privileged mode
	//PortBindings nat.PortMap // Port mapping between the exposed port (container) and the host
	Links           []string // List of links (in the name:alias form)
	PublishAllPorts bool     // Should docker publish all exposed port for the container
	DNS             []string `json:"Dns"`       // List of DNS server to lookup
	DNSSearch       []string `json:"DnsSearch"` // List of DNSSearch to look for
	ExtraHosts      []string // List of extra hosts
	VolumesFrom     []string // List of volumes to take from other container
	//Devices []DeviceMapping // List of devices to map inside the container
	//NetworkMode NetworkMode // Network namespace to use for the container
	//IpcMode IpcMode // IPC namespace to use for the container
	//PidMode PidMode // PID namespace to use for the container
	//UTSMode UTSMode // UTS namespace to use for the container
	//CapAdd *CapList // List of kernel capabilities to add to the container
	//CapDrop *CapList // List of kernel capabilities to remove from the container
	GroupAdd []string // List of additional groups that the container process will run as
	//RestartPolicy RestartPolicy // Restart policy to be used for the container
	SecurityOpt    []string // List of string values to customize labels for MLS systems, such as SELinux.
	ReadonlyRootfs bool     // Is the container root filesystem in read-only
	//Ulimits []*ulimit.Ulimit // List of ulimits to be set in the container
	//LogConfig LogConfig // Configuration of the logs for this container
	CgroupParent string // Parent cgroup.
	ConsoleSize  [2]int // Initial console size on Windows
}

// Container holds data for an existing container.
type Container struct {
	ID         string `json:"Id"`
	Names      []string
	Image      string
	Command    string
	Created    int
	Ports      []Port
	SizeRw     int `json:",omitempty"`
	SizeRootFs int `json:",omitempty"`
	Labels     map[string]string
	Status     string
}

// Config contains the configuration data about a container.
// It should hold only portable information about the container.
// Here, "portable" means "independent from the host we are running on".
// Non-portable information *should* appear in HostConfig.
type ContainerConfig struct {
	Hostname     string // Hostname
	Domainname   string // Domainname
	User         string // User that will run the command(s) inside the container
	AttachStdin  bool   // Attach the standard input, makes possible user interaction
	AttachStdout bool   // Attach the standard output
	AttachStderr bool   // Attach the standard error
	//ExposedPorts    map[nat.Port]struct{} // List of exposed ports
	PublishService  string              // Name of the network service exposed by the container
	Tty             bool                // Attach standard streams to a tty, including stdin if it is not closed.
	OpenStdin       bool                // Open stdin
	StdinOnce       bool                // If true, close stdin after the 1 attached client disconnects.
	Env             []string            // List of environment variable to set in the container
	Cmd             *Command            // Command to run when starting the container
	Image           string              // Name of the image as it was passed by the operator (eg. could be symbolic)
	Volumes         map[string]struct{} // List of volumes (mounts) used for the container
	VolumeDriver    string              // Name of the volume driver used to mount volumes
	WorkingDir      string              // Current directory (PWD) in the command will be launched
	Entrypoint      *Entrypoint         // Entrypoint to run when starting the container
	NetworkDisabled bool                // Is network disabled
	MacAddress      string              // Mac Address of the container
	OnBuild         []string            // ONBUILD metadata that were defined on the image Dockerfile
	Labels          map[string]string   // List of labels set to this container
}

// ContainerConfigWrapper is a Config wrapper that hold the container Config (portable)
// and the corresponding HostConfig (non-portable).
type ContainerConfigWrapper struct {
	*ContainerConfig
	InnerHostConfig *HostConfig `json:"HostConfig,omitempty"`
	Cpuset          string      `json:",omitempty"` // Deprecated. Exported for backwards compatibility.
	*HostConfig                 // Deprecated. Exported to read attrubutes from json that are not in the inner host config structure.
}
