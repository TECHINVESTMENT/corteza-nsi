package plugin

import (
	"log"
	"os"
	"os/exec"

	"github.com/cortezaproject/corteza/server/pkg/expr"
	"github.com/cortezaproject/corteza/server/pkg/plugin/automation"
	"github.com/davecgh/go-spew/spew"
	hclog "github.com/hashicorp/go-hclog"
	hcp "github.com/hashicorp/go-plugin"
)

type (
	plugin struct {
		rawPlugins []interface{}
	}
)

var (
	// global service
	pl *plugin
)

func Service() *plugin {
	return pl
}

// Setup handles the singleton service
func Setup() {
	if pl != nil {
		return
	}

	pl = New()
}

func New() *plugin {
	return &plugin{}
}

func Client2() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	var handshakeConfig = hcp.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "BASIC_PLUGIN",
		MagicCookieValue: "hello",
	}

	spew.Dump(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	v, ee := expr.NewVars(map[string]interface{}{"Input": "test"})

	spew.Dump("GET")
	spew.Dump(ee, v)

	// We're a host! Start by launching the plugin process.
	client := hcp.NewClient(&hcp.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]hcp.Plugin{
			"foo": &automation.AutomationFunctionPlugin{},
			// "bar": &automation.ExamplePlugin{},
			// "foobar": &automation.ExampleFuncPlugin{},
		},
		Cmd:    exec.Command("./plugin/foo"),
		Logger: logger,
	})

	// TODO - make sure to close the connection to each client
	// defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("foo")
	if err != nil {
		spew.Dump("ERR HEER", err)
		log.Fatal(err)
	}

	pl.rawPlugins = append(pl.rawPlugins, raw)

	// // We should have a Greeter now! This feels like a normal interface
	// // implementation but is in fact over an RPC connection.
	foo := raw.(automation.AutomationFunction)
	// foo := raw.(automation.ExampleFuncInterface)
	// // foo := raw.(automation.ExampleInterface)

	spew.Dump("3")
	spew.Dump(foo.Meta())
	spew.Dump(foo.Func(map[string]any{"Input": 11}))
	spew.Dump("4")
	// later register.
	// p.RegisterAutomation()

}

func (p *plugin) RegisterAutomation(r automation.AutomationRegistry) {
	// loop rawPlugins and check if interface of AutomationFunction
	// if yes, call r.AddFunctions()
	return
	// for _, raw := range p.rawPlugins {
	// 	if _, is := raw.(automation.AutomationFunction); is {
	// 		// spew.Dump(">>>> YES")
	// 		f := raw.(automation.AutomationFunction)
	// 		spew.Dump("yes >>> ")
	// 		spew.Dump("calling >>> ", f.AddFunctions())
	// 		r.AddFunctions(f.AddFunctions()...)
	// 	}
	// }
}
