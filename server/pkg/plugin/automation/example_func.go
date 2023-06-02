package automation

import (
	"net/rpc"

	hcp "github.com/hashicorp/go-plugin"
)

type (
	FF func()

	ExampleFuncInterface interface {
		Return() FF
	}

	ExampleFuncPlugin struct {
		Impl ExampleFuncInterface
	}

	ExampleFuncRPCServer struct {
		// This is the real implementation
		Impl ExampleFuncInterface
	}

	ExampleFuncRPCClient struct {
		client *rpc.Client
	}
)

// example plugin
func (p *ExampleFuncPlugin) Server(*hcp.MuxBroker) (interface{}, error) {
	return &ExampleFuncRPCServer{Impl: p.Impl}, nil
}

func (ExampleFuncPlugin) Client(b *hcp.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ExampleFuncRPCClient{client: c}, nil
}

func (s *ExampleFuncRPCServer) Return(args interface{}, resp *FF) error {
	*resp = s.Impl.Return()
	return nil
}

func (g *ExampleFuncRPCClient) Return() FF {
	var resp FF
	err := g.client.Call("Plugin.Return", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}
