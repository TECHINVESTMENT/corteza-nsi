package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/cortezaproject/corteza/server/automation/types"
	"github.com/cortezaproject/corteza/server/pkg/expr"
	"github.com/cortezaproject/corteza/server/pkg/plugin/automation"
)

// Here is a real implementation of Greeter
type Foo struct {
	logger hclog.Logger
}

func (g *Foo) Func(in map[string]any) (out map[string]any, err error) {
	var (
		args = &struct {
			Input string
		}{}
	)

	vv, _ := expr.NewVars(in)

	if err = vv.Decode(args); err != nil {
		return
	}

	g.logger.Debug("message from Func()", args)

	return map[string]any{"Input": "test"}, nil
}

func (g *Foo) Meta() []*types.Function {
	g.logger.Debug("message from Meta()")

	return []*types.Function{{
		Ref:  "corteza.plugin.foo",
		Kind: "function",
		Meta: &types.FunctionMeta{
			Short: "Corteza plugin: Foo this is",
		},
		Parameters: []*types.Param{
			{
				Name:  "input",
				Types: []string{"String"}, Required: false,
			},
		},

		Results: []*types.Param{},

		// Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
		// 	var (
		// 		args = &struct {
		// 			Input string
		// 		}{}
		// 	)

		// 	if err = in.Decode(args); err != nil {
		// 		return
		// 	}

		// 	fmt.Println("FOO", args.Input)

		// 	out = &expr.Vars{}

		// 	return
		// },
	}}
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

	logger.Debug("message from plugin", "foo", "bar")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"foo": &automation.AutomationFunctionPlugin{Impl: foo},
		},
	})
}
