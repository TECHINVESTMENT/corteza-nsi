package automation

import (
	"net/rpc"

	"github.com/cortezaproject/corteza/server/automation/types"
	hcp "github.com/hashicorp/go-plugin"
)

type (
	AutomationFunction interface {
		Meta() []*types.Function
		Func(in map[string]any) (out map[string]any, err error)
	}

	AutomationFunctionPlugin struct {
		Impl AutomationFunction
	}

	AutomationFunctionRPCServer struct {
		// This is the real implementation
		Impl AutomationFunction
	}

	AutomationFunctionRPCClient struct {
		client *rpc.Client
	}
)

// Automation function plugin
func (p *AutomationFunctionPlugin) Server(*hcp.MuxBroker) (interface{}, error) {
	return &AutomationFunctionRPCServer{Impl: p.Impl}, nil
}

func (AutomationFunctionPlugin) Client(b *hcp.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &AutomationFunctionRPCClient{client: c}, nil
}

func (s *AutomationFunctionRPCServer) Meta(args interface{}, resp *[]*types.Function) error {
	*resp = s.Impl.Meta()
	return nil
}

func (s *AutomationFunctionRPCServer) Func(args map[string]any, resp *map[string]any) error {
	*resp, _ = s.Impl.Func(args)
	return nil
}

func (g *AutomationFunctionRPCClient) Meta() []*types.Function {
	var resp []*types.Function
	err := g.client.Call("Plugin.Meta", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (g *AutomationFunctionRPCClient) Func(in map[string]any) (out map[string]any, err error) {
	var resp map[string]any
	err = g.client.Call("Plugin.Func", in, &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		return nil, err
	}

	return resp, nil
}
