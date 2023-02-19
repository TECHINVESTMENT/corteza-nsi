package envoyx

import (
	"context"
	"fmt"
)

type (
	service struct {
		decoders map[decodeType][]Decoder

		encoders  map[encodeType][]Encoder
		preparers map[encodeType][]Preparer
	}

	// Traverser provides a structure which can be used to traverse the node's deps
	Traverser interface {
		// ParentForRef returns the parent of the provided node which matches the ref
		//
		// If no parent is found, nil is returned.
		ParentForRef(*Node, Ref) *Node

		// ChildrenForResourceType returns the children of the provided node which
		// match the provided resource type
		ChildrenForResourceType(*Node, string) NodeSet

		// Children returns all of the children of the provided node
		Children(*Node) NodeSet
	}

	Preparer interface {
		// Prepare performs generic preprocessing on the provided nodes
		//
		// The function is called for every resource type where all of the nodes of
		// that resource type are passed as the argument.
		Prepare(context.Context, EncodeParams, string, NodeSet) error
	}

	Encoder interface {
		// Encode encodes the data
		//
		// The function receives a set of root-level nodes (with no parent dependencies)
		// and a Traverser it can use to handle all of the child nodes.
		Encode(context.Context, EncodeParams, string, NodeSet, Traverser) (err error)
	}

	PrepareEncoder interface {
		Preparer
		Encoder
	}

	Decoder interface {
		// Decode returns a set of Nodes extracted based on the provided definition
		Decode(ctx context.Context, p DecodeParams) (out NodeSet, err error)
	}

	DecodeParams struct {
		Type   decodeType
		Params map[string]any
		Config DecoderConfig
		Filter map[string]ResourceFilter
	}
	DecoderConfig struct{}

	EncodeParams struct {
		Type   encodeType
		Params map[string]any
		Config EncoderConfig
	}
	EncoderConfig struct {
		OnExisting mergeAlg

		PreferredTimeLayout string
		PreferredTimezone   string
	}

	ResourceFilter struct {
		Identifiers Identifiers
		Refs        map[string]Ref

		Limit uint
		Scope Scope
	}

	decodeType string
	encodeType string
	mergeAlg   int
)

var (
	global *service
)

const (
	OnConflictReplace mergeAlg = iota
	OnConflictSkip
	OnConflictPanic
	// OnConflictMergeLeft  mergeAlg = "mergeLeft"
	// OnConflictMergeRight mergeAlg = "mergeRight"

	DecodeTypeURI   decodeType = "uri"
	DecodeTypeStore decodeType = "store"

	EncodeTypeURI   encodeType = "uri"
	EncodeTypeStore encodeType = "store"
	EncodeTypeIo    encodeType = "io"
)

// New initializes a new Envoy service
func New() *service {
	return &service{}
}

// SetGlobal sets the global envoy service
func SetGlobal(n *service) {
	global = n
}

// Service gets the global envoy service
func Service() *service {
	if global == nil {
		panic("global service not defined")
	}

	return global
}

// Decode returns a set of envoy Nodes based on the given decode params
func (svc *service) Decode(ctx context.Context, p DecodeParams) (nn NodeSet, err error) {
	err = p.validate()
	if err != nil {
		return
	}

	switch p.Type {
	case DecodeTypeURI:
		return svc.decodeUri(ctx, p)
	case DecodeTypeStore:
		return svc.decodeStore(ctx, p)
	}

	return
}

// Encode encodes Corteza resources bases on the provided encode params
//
// use the BuildDepGraph function to build the default dependency graph.
func (svc *service) Encode(ctx context.Context, p EncodeParams, dg *depGraph) (err error) {
	err = p.validate()
	if err != nil {
		return
	}

	switch p.Type {
	case EncodeTypeStore:
		return svc.encodeStore(ctx, dg, p)
	case EncodeTypeIo:
		return svc.encodeIo(ctx, dg, p)

	}
	return
}

func (svc *service) AddDecoder(t decodeType, dd ...Decoder) {
	if svc.decoders == nil {
		svc.decoders = make(map[decodeType][]Decoder)
	}
	svc.decoders[t] = append(svc.decoders[t], dd...)
}

func (svc *service) AddEncoder(t encodeType, ee ...Encoder) {
	if svc.encoders == nil {
		svc.encoders = make(map[encodeType][]Encoder)
	}
	svc.encoders[t] = append(svc.encoders[t], ee...)
}

func (svc *service) AddPreparer(t encodeType, pp ...Preparer) {
	if svc.preparers == nil {
		svc.preparers = make(map[encodeType][]Preparer)
	}
	svc.preparers[t] = append(svc.preparers[t], pp...)
}

// Utility functions

func (p DecodeParams) validate() (err error) {
	switch p.Type {
	case DecodeTypeURI:
		_, ok := p.Params["uri"]
		if !ok {
			return fmt.Errorf("uhoh, no uri provided")
		}

	case DecodeTypeStore:

	}

	// @todo...

	return
}

func (p EncodeParams) validate() (err error) {
	switch p.Type {
	// @todo...
	}

	// @todo...

	return
}