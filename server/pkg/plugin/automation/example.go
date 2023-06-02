package automation

import (
	"net/rpc"

	hcp "github.com/hashicorp/go-plugin"
)

type (
	ExampleInterface interface {
		Return() string
	}

	ExamplePlugin struct {
		Impl ExampleInterface
	}

	ExampleRPCServer struct {
		// This is the real implementation
		Impl ExampleInterface
	}

	ExampleRPCClient struct {
		client *rpc.Client
	}
)

// example plugin
func (p *ExamplePlugin) Server(*hcp.MuxBroker) (interface{}, error) {
	return &ExampleRPCServer{Impl: p.Impl}, nil
}

func (ExamplePlugin) Client(b *hcp.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ExampleRPCClient{client: c}, nil
}

func (s *ExampleRPCServer) Return(args interface{}, resp *string) error {
	*resp = s.Impl.Return()
	return nil
}

func (g *ExampleRPCClient) Return() string {
	var resp string
	err := g.client.Call("Plugin.Return", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}
