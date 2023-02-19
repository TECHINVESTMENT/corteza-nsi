package envoy

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated.
//

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cortezaproject/corteza/server/compose/types"
	"github.com/cortezaproject/corteza/server/pkg/envoyx"
	"github.com/cortezaproject/corteza/server/pkg/id"
	"github.com/cortezaproject/corteza/server/store"
)

type (
	// StoreEncoder is responsible for encoding Corteza resources into the
	// database via the Storer or the DAL interface
	//
	// @todo consider having a different encoder for the DAL resources
	StoreEncoder struct{}
)

// Prepare performs some initial processing on the resource before it can be encoded
//
// Preparation runs validation, default value initialization, matching with
// already existing instances, ...
//
// The prepare function receives a set of nodes grouped by the resource type.
// This enables some batching optimization and simplifications when it comes to
// matching with existing resources.
//
// Prepare does not receive any placeholder nodes which are used solely
// for dependency resolution.
func (e StoreEncoder) Prepare(ctx context.Context, p envoyx.EncodeParams, rt string, nn envoyx.NodeSet) (err error) {
	s, err := e.grabStorer(p)
	if err != nil {
		return
	}

	switch rt {
	case types.ChartResourceType:
		return e.prepareChart(ctx, p, s, nn)
	case types.ModuleResourceType:
		return e.prepareModule(ctx, p, s, nn)
	case types.ModuleFieldResourceType:
		return e.prepareModuleField(ctx, p, s, nn)
	case types.NamespaceResourceType:
		return e.prepareNamespace(ctx, p, s, nn)
	case types.PageResourceType:
		return e.preparePage(ctx, p, s, nn)

	}

	return
}

// Encode encodes the given Corteza resources into the primary store
//
// Encoding should not do any additional processing apart from matching with
// dependencies and runtime validation
//
// The Encode function is called for every resource type where the resource
// appears at the root of the dependency tree.
// All of the root-level resources for that resource type are passed into the function.
// The encoding function must traverse the branches to encode all of the dependencies.
//
// This flow is used to simplify the flow of how resources are encoded into YAML
// (and other documents) as well as to simplify batching.
//
// Encode does not receive any placeholder nodes which are used solely
// for dependency resolution.
func (e StoreEncoder) Encode(ctx context.Context, p envoyx.EncodeParams, rt string, nodes envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	s, err := e.grabStorer(p)
	if err != nil {
		return
	}

	switch rt {
	case types.ChartResourceType:
		return e.encodeCharts(ctx, p, s, nodes, tree)

	case types.ModuleResourceType:
		return e.encodeModules(ctx, p, s, nodes, tree)

	case types.ModuleFieldResourceType:
		return e.encodeModuleFields(ctx, p, s, nodes, tree)

	case types.NamespaceResourceType:
		return e.encodeNamespaces(ctx, p, s, nodes, tree)

	case types.PageResourceType:
		return e.encodePages(ctx, p, s, nodes, tree)
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource chart
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareChart prepares the resources of the given type for encoding
func (e StoreEncoder) prepareChart(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Chart, len(nn))
	err = e.matchupCharts(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareChart with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Chart)
		if !ok {
			panic("unexpected resource type: node expecting type of chart")
		}

		existing, hasExisting := existing[i]

		if hasExisting {
			// On existing, we don't need to re-do identifiers and references; simply
			// changing up the internal resource is enough.
			//
			// In the future, we can pass down the tree and re-do the deps like that
			switch p.Config.OnExisting {
			case envoyx.OnConflictPanic:
				err = fmt.Errorf("resource already exists")
				return

			case envoyx.OnConflictReplace:
				// Replace; simple ID change should do the trick
				res.ID = existing.ID

			case envoyx.OnConflictSkip:
				// Replace the node's resource with the fetched one
				res = &existing

				// @todo merging
			}
		} else {
			// @todo actually a bottleneck. As per sonyflake docs, it can at most
			//       generate up to 2**8 (256) IDs per 10ms in a single thread.
			//       How can we improve this?
			res.ID = id.Next()
		}

		// We can skip validation/defaults when the resource is overwritten by
		// the one already stored (the panic one errors out anyway) since it
		// should already be ok.
		if !hasExisting || p.Config.OnExisting != envoyx.OnConflictSkip {
			err = e.setChartDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateChart(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeCharts encodes a set of resource into the database
func (e StoreEncoder) encodeCharts(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeChart(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeChart encodes the resource into the database
func (e StoreEncoder) encodeChart(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
	// Grab dependency references
	var auxID uint64
	for fieldLabel, ref := range n.References {
		rn := tree.ParentForRef(n, ref)
		if rn == nil {
			err = fmt.Errorf("missing node for ref %v", ref)
			return
		}

		auxID = rn.Resource.GetID()
		if auxID == 0 {
			err = fmt.Errorf("related resource doesn't provide an ID")
			return
		}

		err = n.Resource.SetValue(fieldLabel, 0, auxID)
		if err != nil {
			return
		}
	}

	// Flush to the DB
	err = store.UpsertComposeChart(ctx, s, n.Resource.(*types.Chart))
	if err != nil {
		return
	}

	// Handle resources nested under it
	//
	// @todo how can we remove the OmitPlaceholderNodes call the same way we did for
	//       the root function calls?

	for rt, nn := range envoyx.NodesByResourceType(tree.Children(n)...) {
		nn = envoyx.OmitPlaceholderNodes(nn...)

		switch rt {

		}
	}

	return
}

// matchupCharts returns an index with indicates what resources already exist
func (e StoreEncoder) matchupCharts(ctx context.Context, s store.Storer, uu map[int]types.Chart, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchComposeCharts(ctx, s, types.ChartFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Chart, len(aa))
	strMap := make(map[string]*types.Chart, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Chart
	var ok bool
	for i, n := range nn {
		for _, idf := range n.Identifiers.Slice {
			if id, err := strconv.ParseUint(idf, 10, 64); err == nil {
				aux, ok = idMap[id]
				if ok {
					uu[i] = *aux
					// When any identifier matches we can end it
					break
				}
			}

			aux, ok = strMap[idf]
			if ok {
				uu[i] = *aux
				// When any identifier matches we can end it
				break
			}
		}
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource module
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareModule prepares the resources of the given type for encoding
func (e StoreEncoder) prepareModule(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Module, len(nn))
	err = e.matchupModules(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareModule with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Module)
		if !ok {
			panic("unexpected resource type: node expecting type of module")
		}

		existing, hasExisting := existing[i]

		if hasExisting {
			// On existing, we don't need to re-do identifiers and references; simply
			// changing up the internal resource is enough.
			//
			// In the future, we can pass down the tree and re-do the deps like that
			switch p.Config.OnExisting {
			case envoyx.OnConflictPanic:
				err = fmt.Errorf("resource already exists")
				return

			case envoyx.OnConflictReplace:
				// Replace; simple ID change should do the trick
				res.ID = existing.ID

			case envoyx.OnConflictSkip:
				// Replace the node's resource with the fetched one
				res = &existing

				// @todo merging
			}
		} else {
			// @todo actually a bottleneck. As per sonyflake docs, it can at most
			//       generate up to 2**8 (256) IDs per 10ms in a single thread.
			//       How can we improve this?
			res.ID = id.Next()
		}

		// We can skip validation/defaults when the resource is overwritten by
		// the one already stored (the panic one errors out anyway) since it
		// should already be ok.
		if !hasExisting || p.Config.OnExisting != envoyx.OnConflictSkip {
			err = e.setModuleDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateModule(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeModules encodes a set of resource into the database
func (e StoreEncoder) encodeModules(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeModule(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeModule encodes the resource into the database
func (e StoreEncoder) encodeModule(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
	// Grab dependency references
	var auxID uint64
	for fieldLabel, ref := range n.References {
		rn := tree.ParentForRef(n, ref)
		if rn == nil {
			err = fmt.Errorf("missing node for ref %v", ref)
			return
		}

		auxID = rn.Resource.GetID()
		if auxID == 0 {
			err = fmt.Errorf("related resource doesn't provide an ID")
			return
		}

		err = n.Resource.SetValue(fieldLabel, 0, auxID)
		if err != nil {
			return
		}
	}

	// Flush to the DB
	err = store.UpsertComposeModule(ctx, s, n.Resource.(*types.Module))
	if err != nil {
		return
	}

	// Handle resources nested under it
	//
	// @todo how can we remove the OmitPlaceholderNodes call the same way we did for
	//       the root function calls?

	for rt, nn := range envoyx.NodesByResourceType(tree.Children(n)...) {
		nn = envoyx.OmitPlaceholderNodes(nn...)

		switch rt {

		}
	}

	return
}

// matchupModules returns an index with indicates what resources already exist
func (e StoreEncoder) matchupModules(ctx context.Context, s store.Storer, uu map[int]types.Module, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchComposeModules(ctx, s, types.ModuleFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Module, len(aa))
	strMap := make(map[string]*types.Module, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Module
	var ok bool
	for i, n := range nn {
		for _, idf := range n.Identifiers.Slice {
			if id, err := strconv.ParseUint(idf, 10, 64); err == nil {
				aux, ok = idMap[id]
				if ok {
					uu[i] = *aux
					// When any identifier matches we can end it
					break
				}
			}

			aux, ok = strMap[idf]
			if ok {
				uu[i] = *aux
				// When any identifier matches we can end it
				break
			}
		}
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource moduleField
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareModuleField prepares the resources of the given type for encoding
func (e StoreEncoder) prepareModuleField(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.ModuleField, len(nn))
	err = e.matchupModuleFields(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareModuleField with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.ModuleField)
		if !ok {
			panic("unexpected resource type: node expecting type of moduleField")
		}

		existing, hasExisting := existing[i]

		if hasExisting {
			// On existing, we don't need to re-do identifiers and references; simply
			// changing up the internal resource is enough.
			//
			// In the future, we can pass down the tree and re-do the deps like that
			switch p.Config.OnExisting {
			case envoyx.OnConflictPanic:
				err = fmt.Errorf("resource already exists")
				return

			case envoyx.OnConflictReplace:
				// Replace; simple ID change should do the trick
				res.ID = existing.ID

			case envoyx.OnConflictSkip:
				// Replace the node's resource with the fetched one
				res = &existing

				// @todo merging
			}
		} else {
			// @todo actually a bottleneck. As per sonyflake docs, it can at most
			//       generate up to 2**8 (256) IDs per 10ms in a single thread.
			//       How can we improve this?
			res.ID = id.Next()
		}

		// We can skip validation/defaults when the resource is overwritten by
		// the one already stored (the panic one errors out anyway) since it
		// should already be ok.
		if !hasExisting || p.Config.OnExisting != envoyx.OnConflictSkip {
			err = e.setModuleFieldDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateModuleField(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeModuleFields encodes a set of resource into the database
func (e StoreEncoder) encodeModuleFields(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeModuleField(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeModuleField encodes the resource into the database
func (e StoreEncoder) encodeModuleField(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
	// Grab dependency references
	var auxID uint64
	for fieldLabel, ref := range n.References {
		rn := tree.ParentForRef(n, ref)
		if rn == nil {
			err = fmt.Errorf("missing node for ref %v", ref)
			return
		}

		auxID = rn.Resource.GetID()
		if auxID == 0 {
			err = fmt.Errorf("related resource doesn't provide an ID")
			return
		}

		err = n.Resource.SetValue(fieldLabel, 0, auxID)
		if err != nil {
			return
		}
	}

	// Flush to the DB
	err = store.UpsertComposeModuleField(ctx, s, n.Resource.(*types.ModuleField))
	if err != nil {
		return
	}

	// Handle resources nested under it
	//
	// @todo how can we remove the OmitPlaceholderNodes call the same way we did for
	//       the root function calls?

	for rt, nn := range envoyx.NodesByResourceType(tree.Children(n)...) {
		nn = envoyx.OmitPlaceholderNodes(nn...)

		switch rt {

		}
	}

	return
}

// matchupModuleFields returns an index with indicates what resources already exist
func (e StoreEncoder) matchupModuleFields(ctx context.Context, s store.Storer, uu map[int]types.ModuleField, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchComposeModuleFields(ctx, s, types.ModuleFieldFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.ModuleField, len(aa))
	strMap := make(map[string]*types.ModuleField, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a
		strMap[a.Name] = a

	}

	var aux *types.ModuleField
	var ok bool
	for i, n := range nn {
		for _, idf := range n.Identifiers.Slice {
			if id, err := strconv.ParseUint(idf, 10, 64); err == nil {
				aux, ok = idMap[id]
				if ok {
					uu[i] = *aux
					// When any identifier matches we can end it
					break
				}
			}

			aux, ok = strMap[idf]
			if ok {
				uu[i] = *aux
				// When any identifier matches we can end it
				break
			}
		}
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource namespace
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareNamespace prepares the resources of the given type for encoding
func (e StoreEncoder) prepareNamespace(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Namespace, len(nn))
	err = e.matchupNamespaces(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareNamespace with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Namespace)
		if !ok {
			panic("unexpected resource type: node expecting type of namespace")
		}

		existing, hasExisting := existing[i]

		if hasExisting {
			// On existing, we don't need to re-do identifiers and references; simply
			// changing up the internal resource is enough.
			//
			// In the future, we can pass down the tree and re-do the deps like that
			switch p.Config.OnExisting {
			case envoyx.OnConflictPanic:
				err = fmt.Errorf("resource already exists")
				return

			case envoyx.OnConflictReplace:
				// Replace; simple ID change should do the trick
				res.ID = existing.ID

			case envoyx.OnConflictSkip:
				// Replace the node's resource with the fetched one
				res = &existing

				// @todo merging
			}
		} else {
			// @todo actually a bottleneck. As per sonyflake docs, it can at most
			//       generate up to 2**8 (256) IDs per 10ms in a single thread.
			//       How can we improve this?
			res.ID = id.Next()
		}

		// We can skip validation/defaults when the resource is overwritten by
		// the one already stored (the panic one errors out anyway) since it
		// should already be ok.
		if !hasExisting || p.Config.OnExisting != envoyx.OnConflictSkip {
			err = e.setNamespaceDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateNamespace(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeNamespaces encodes a set of resource into the database
func (e StoreEncoder) encodeNamespaces(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeNamespace(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeNamespace encodes the resource into the database
func (e StoreEncoder) encodeNamespace(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
	// Grab dependency references
	var auxID uint64
	for fieldLabel, ref := range n.References {
		rn := tree.ParentForRef(n, ref)
		if rn == nil {
			err = fmt.Errorf("missing node for ref %v", ref)
			return
		}

		auxID = rn.Resource.GetID()
		if auxID == 0 {
			err = fmt.Errorf("related resource doesn't provide an ID")
			return
		}

		err = n.Resource.SetValue(fieldLabel, 0, auxID)
		if err != nil {
			return
		}
	}

	// Flush to the DB
	err = store.UpsertComposeNamespace(ctx, s, n.Resource.(*types.Namespace))
	if err != nil {
		return
	}

	// Handle resources nested under it
	//
	// @todo how can we remove the OmitPlaceholderNodes call the same way we did for
	//       the root function calls?

	for rt, nn := range envoyx.NodesByResourceType(tree.Children(n)...) {
		nn = envoyx.OmitPlaceholderNodes(nn...)

		switch rt {

		}
	}

	return
}

// matchupNamespaces returns an index with indicates what resources already exist
func (e StoreEncoder) matchupNamespaces(ctx context.Context, s store.Storer, uu map[int]types.Namespace, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchComposeNamespaces(ctx, s, types.NamespaceFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Namespace, len(aa))
	strMap := make(map[string]*types.Namespace, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a
		strMap[a.Slug] = a

	}

	var aux *types.Namespace
	var ok bool
	for i, n := range nn {
		for _, idf := range n.Identifiers.Slice {
			if id, err := strconv.ParseUint(idf, 10, 64); err == nil {
				aux, ok = idMap[id]
				if ok {
					uu[i] = *aux
					// When any identifier matches we can end it
					break
				}
			}

			aux, ok = strMap[idf]
			if ok {
				uu[i] = *aux
				// When any identifier matches we can end it
				break
			}
		}
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource page
// // // // // // // // // // // // // // // // // // // // // // // // //

// preparePage prepares the resources of the given type for encoding
func (e StoreEncoder) preparePage(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Page, len(nn))
	err = e.matchupPages(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call preparePage with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Page)
		if !ok {
			panic("unexpected resource type: node expecting type of page")
		}

		existing, hasExisting := existing[i]

		if hasExisting {
			// On existing, we don't need to re-do identifiers and references; simply
			// changing up the internal resource is enough.
			//
			// In the future, we can pass down the tree and re-do the deps like that
			switch p.Config.OnExisting {
			case envoyx.OnConflictPanic:
				err = fmt.Errorf("resource already exists")
				return

			case envoyx.OnConflictReplace:
				// Replace; simple ID change should do the trick
				res.ID = existing.ID

			case envoyx.OnConflictSkip:
				// Replace the node's resource with the fetched one
				res = &existing

				// @todo merging
			}
		} else {
			// @todo actually a bottleneck. As per sonyflake docs, it can at most
			//       generate up to 2**8 (256) IDs per 10ms in a single thread.
			//       How can we improve this?
			res.ID = id.Next()
		}

		// We can skip validation/defaults when the resource is overwritten by
		// the one already stored (the panic one errors out anyway) since it
		// should already be ok.
		if !hasExisting || p.Config.OnExisting != envoyx.OnConflictSkip {
			err = e.setPageDefaults(res)
			if err != nil {
				return err
			}

			err = e.validatePage(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodePages encodes a set of resource into the database
func (e StoreEncoder) encodePages(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodePage(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodePage encodes the resource into the database
func (e StoreEncoder) encodePage(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
	// Grab dependency references
	var auxID uint64
	for fieldLabel, ref := range n.References {
		rn := tree.ParentForRef(n, ref)
		if rn == nil {
			err = fmt.Errorf("missing node for ref %v", ref)
			return
		}

		auxID = rn.Resource.GetID()
		if auxID == 0 {
			err = fmt.Errorf("related resource doesn't provide an ID")
			return
		}

		err = n.Resource.SetValue(fieldLabel, 0, auxID)
		if err != nil {
			return
		}
	}

	// Flush to the DB
	err = store.UpsertComposePage(ctx, s, n.Resource.(*types.Page))
	if err != nil {
		return
	}

	// Handle resources nested under it
	//
	// @todo how can we remove the OmitPlaceholderNodes call the same way we did for
	//       the root function calls?

	for rt, nn := range envoyx.NodesByResourceType(tree.Children(n)...) {
		nn = envoyx.OmitPlaceholderNodes(nn...)

		switch rt {

		}
	}

	return
}

// matchupPages returns an index with indicates what resources already exist
func (e StoreEncoder) matchupPages(ctx context.Context, s store.Storer, uu map[int]types.Page, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchComposePages(ctx, s, types.PageFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Page, len(aa))
	strMap := make(map[string]*types.Page, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Page
	var ok bool
	for i, n := range nn {
		for _, idf := range n.Identifiers.Slice {
			if id, err := strconv.ParseUint(idf, 10, 64); err == nil {
				aux, ok = idMap[id]
				if ok {
					uu[i] = *aux
					// When any identifier matches we can end it
					break
				}
			}

			aux, ok = strMap[idf]
			if ok {
				uu[i] = *aux
				// When any identifier matches we can end it
				break
			}
		}
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Utility functions
// // // // // // // // // // // // // // // // // // // // // // // // //

func (e *StoreEncoder) grabStorer(p envoyx.EncodeParams) (s store.Storer, err error) {
	auxs, ok := p.Params["storer"]
	if !ok {
		err = fmt.Errorf("storer not defined")
		return
	}

	s, ok = auxs.(store.Storer)
	if !ok {
		err = fmt.Errorf("invalid storer provided")
		return
	}

	return
}