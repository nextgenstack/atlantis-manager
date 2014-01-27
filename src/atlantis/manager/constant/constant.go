package constant

import (
	"regexp"
)

var (
	Host                = "localhost"
	Region              = "dev"
	Zone                = "dev1"
	AvailableZones      = []string{"dev1"}
	AppRegexp           = regexp.MustCompile("[A-Za-z0-9-]+")                            // apps can contain letters, numbers, and -
	SecurityGroupRegexp = regexp.MustCompile("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+:[0-9]+") // ip:port
)

const (
	ManagerRPCVersion               = "2.0.0"
	ManagerAPIVersion               = "2.0.0"
	DefaultManagerHost              = "localhost"
	DefaultManagerRPCPort           = uint16(1338)
	DefaultManagerAPIPort           = uint16(443)
	DefaultManagerKeyPath           = "~/.manager"
	DefaultResultDuration           = "30m"
	DefaultMaintenanceFile          = "/etc/atlantis/manager/maint"
	DefaultMaintenanceCheckInterval = "5s"
	DefaultMinRouterPort            = uint16(49152)
	DefaultMaxRouterPort            = uint16(65535)
)
