package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/cortezaproject/corteza/server/pkg/plugin/automation"
)

// Here is a real implementation of Greeter
type Foo struct {
	logger hclog.Logger
}

func (g *Foo) Return() automation.FF {
	g.logger.Debug("message from Return()")

	return func() {
		fmt.Println("test")
	}
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	foo := &Foo{
		logger: logger,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"foobar": &automation.ExampleFuncPlugin{Impl: foo},
	}

	logger.Debug("message from plugin", "foo", "bar2")
	logger.Debug("calling Return()", foo.Return())

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
