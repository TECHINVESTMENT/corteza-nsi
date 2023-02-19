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

	"github.com/cortezaproject/corteza/server/pkg/envoyx"
	"github.com/cortezaproject/corteza/server/pkg/id"
	"github.com/cortezaproject/corteza/server/store"
	"github.com/cortezaproject/corteza/server/system/types"
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
	case types.ApplicationResourceType:
		return e.prepareApplication(ctx, p, s, nn)
	case types.ApigwRouteResourceType:
		return e.prepareApigwRoute(ctx, p, s, nn)
	case types.ApigwFilterResourceType:
		return e.prepareApigwFilter(ctx, p, s, nn)
	case types.AuthClientResourceType:
		return e.prepareAuthClient(ctx, p, s, nn)

	case types.QueueResourceType:
		return e.prepareQueue(ctx, p, s, nn)

	case types.ReportResourceType:
		return e.prepareReport(ctx, p, s, nn)

	case types.RoleResourceType:
		return e.prepareRole(ctx, p, s, nn)

	case types.TemplateResourceType:
		return e.prepareTemplate(ctx, p, s, nn)
	case types.UserResourceType:
		return e.prepareUser(ctx, p, s, nn)
	case types.DalConnectionResourceType:
		return e.prepareDalConnection(ctx, p, s, nn)
	case types.DalSensitivityLevelResourceType:
		return e.prepareDalSensitivityLevel(ctx, p, s, nn)
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
	case types.ApplicationResourceType:
		return e.encodeApplications(ctx, p, s, nodes, tree)

	case types.ApigwRouteResourceType:
		return e.encodeApigwRoutes(ctx, p, s, nodes, tree)

	case types.ApigwFilterResourceType:
		return e.encodeApigwFilters(ctx, p, s, nodes, tree)

	case types.AuthClientResourceType:
		return e.encodeAuthClients(ctx, p, s, nodes, tree)

	case types.QueueResourceType:
		return e.encodeQueues(ctx, p, s, nodes, tree)

	case types.ReportResourceType:
		return e.encodeReports(ctx, p, s, nodes, tree)

	case types.RoleResourceType:
		return e.encodeRoles(ctx, p, s, nodes, tree)

	case types.TemplateResourceType:
		return e.encodeTemplates(ctx, p, s, nodes, tree)

	case types.UserResourceType:
		return e.encodeUsers(ctx, p, s, nodes, tree)

	case types.DalConnectionResourceType:
		return e.encodeDalConnections(ctx, p, s, nodes, tree)

	case types.DalSensitivityLevelResourceType:
		return e.encodeDalSensitivityLevels(ctx, p, s, nodes, tree)
	}

	return
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Functions for resource application
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareApplication prepares the resources of the given type for encoding
func (e StoreEncoder) prepareApplication(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Application, len(nn))
	err = e.matchupApplications(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareApplication with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Application)
		if !ok {
			panic("unexpected resource type: node expecting type of application")
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
			err = e.setApplicationDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateApplication(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeApplications encodes a set of resource into the database
func (e StoreEncoder) encodeApplications(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeApplication(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeApplication encodes the resource into the database
func (e StoreEncoder) encodeApplication(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertApplication(ctx, s, n.Resource.(*types.Application))
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

// matchupApplications returns an index with indicates what resources already exist
func (e StoreEncoder) matchupApplications(ctx context.Context, s store.Storer, uu map[int]types.Application, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchApplications(ctx, s, types.ApplicationFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Application, len(aa))
	strMap := make(map[string]*types.Application, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a

	}

	var aux *types.Application
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
// Functions for resource apigwRoute
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareApigwRoute prepares the resources of the given type for encoding
func (e StoreEncoder) prepareApigwRoute(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.ApigwRoute, len(nn))
	err = e.matchupApigwRoutes(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareApigwRoute with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.ApigwRoute)
		if !ok {
			panic("unexpected resource type: node expecting type of apigwRoute")
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
			err = e.setApigwRouteDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateApigwRoute(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeApigwRoutes encodes a set of resource into the database
func (e StoreEncoder) encodeApigwRoutes(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeApigwRoute(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeApigwRoute encodes the resource into the database
func (e StoreEncoder) encodeApigwRoute(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertApigwRoute(ctx, s, n.Resource.(*types.ApigwRoute))
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

// matchupApigwRoutes returns an index with indicates what resources already exist
func (e StoreEncoder) matchupApigwRoutes(ctx context.Context, s store.Storer, uu map[int]types.ApigwRoute, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchApigwRoutes(ctx, s, types.ApigwRouteFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.ApigwRoute, len(aa))
	strMap := make(map[string]*types.ApigwRoute, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a

	}

	var aux *types.ApigwRoute
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
// Functions for resource apigwFilter
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareApigwFilter prepares the resources of the given type for encoding
func (e StoreEncoder) prepareApigwFilter(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.ApigwFilter, len(nn))
	err = e.matchupApigwFilters(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareApigwFilter with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.ApigwFilter)
		if !ok {
			panic("unexpected resource type: node expecting type of apigwFilter")
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
			err = e.setApigwFilterDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateApigwFilter(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeApigwFilters encodes a set of resource into the database
func (e StoreEncoder) encodeApigwFilters(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeApigwFilter(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeApigwFilter encodes the resource into the database
func (e StoreEncoder) encodeApigwFilter(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertApigwFilter(ctx, s, n.Resource.(*types.ApigwFilter))
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

// matchupApigwFilters returns an index with indicates what resources already exist
func (e StoreEncoder) matchupApigwFilters(ctx context.Context, s store.Storer, uu map[int]types.ApigwFilter, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchApigwFilters(ctx, s, types.ApigwFilterFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.ApigwFilter, len(aa))
	strMap := make(map[string]*types.ApigwFilter, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a

	}

	var aux *types.ApigwFilter
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
// Functions for resource authClient
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareAuthClient prepares the resources of the given type for encoding
func (e StoreEncoder) prepareAuthClient(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.AuthClient, len(nn))
	err = e.matchupAuthClients(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareAuthClient with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.AuthClient)
		if !ok {
			panic("unexpected resource type: node expecting type of authClient")
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
			err = e.setAuthClientDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateAuthClient(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeAuthClients encodes a set of resource into the database
func (e StoreEncoder) encodeAuthClients(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeAuthClient(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeAuthClient encodes the resource into the database
func (e StoreEncoder) encodeAuthClient(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertAuthClient(ctx, s, n.Resource.(*types.AuthClient))
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

// matchupAuthClients returns an index with indicates what resources already exist
func (e StoreEncoder) matchupAuthClients(ctx context.Context, s store.Storer, uu map[int]types.AuthClient, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchAuthClients(ctx, s, types.AuthClientFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.AuthClient, len(aa))
	strMap := make(map[string]*types.AuthClient, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.AuthClient
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
// Functions for resource queue
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareQueue prepares the resources of the given type for encoding
func (e StoreEncoder) prepareQueue(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Queue, len(nn))
	err = e.matchupQueues(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareQueue with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Queue)
		if !ok {
			panic("unexpected resource type: node expecting type of queue")
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
			err = e.setQueueDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateQueue(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeQueues encodes a set of resource into the database
func (e StoreEncoder) encodeQueues(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeQueue(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeQueue encodes the resource into the database
func (e StoreEncoder) encodeQueue(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertQueue(ctx, s, n.Resource.(*types.Queue))
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

// matchupQueues returns an index with indicates what resources already exist
func (e StoreEncoder) matchupQueues(ctx context.Context, s store.Storer, uu map[int]types.Queue, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchQueues(ctx, s, types.QueueFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Queue, len(aa))
	strMap := make(map[string]*types.Queue, len(aa))

	for _, a := range aa {
		idMap[a.ID] = a
		strMap[a.Queue] = a

	}

	var aux *types.Queue
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
// Functions for resource report
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareReport prepares the resources of the given type for encoding
func (e StoreEncoder) prepareReport(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Report, len(nn))
	err = e.matchupReports(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareReport with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Report)
		if !ok {
			panic("unexpected resource type: node expecting type of report")
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
			err = e.setReportDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateReport(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeReports encodes a set of resource into the database
func (e StoreEncoder) encodeReports(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeReport(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeReport encodes the resource into the database
func (e StoreEncoder) encodeReport(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertReport(ctx, s, n.Resource.(*types.Report))
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

// matchupReports returns an index with indicates what resources already exist
func (e StoreEncoder) matchupReports(ctx context.Context, s store.Storer, uu map[int]types.Report, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchReports(ctx, s, types.ReportFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Report, len(aa))
	strMap := make(map[string]*types.Report, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Report
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
// Functions for resource role
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareRole prepares the resources of the given type for encoding
func (e StoreEncoder) prepareRole(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Role, len(nn))
	err = e.matchupRoles(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareRole with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Role)
		if !ok {
			panic("unexpected resource type: node expecting type of role")
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
			err = e.setRoleDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateRole(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeRoles encodes a set of resource into the database
func (e StoreEncoder) encodeRoles(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeRole(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeRole encodes the resource into the database
func (e StoreEncoder) encodeRole(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertRole(ctx, s, n.Resource.(*types.Role))
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

// matchupRoles returns an index with indicates what resources already exist
func (e StoreEncoder) matchupRoles(ctx context.Context, s store.Storer, uu map[int]types.Role, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchRoles(ctx, s, types.RoleFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Role, len(aa))
	strMap := make(map[string]*types.Role, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Role
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
// Functions for resource template
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareTemplate prepares the resources of the given type for encoding
func (e StoreEncoder) prepareTemplate(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.Template, len(nn))
	err = e.matchupTemplates(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareTemplate with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.Template)
		if !ok {
			panic("unexpected resource type: node expecting type of template")
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
			err = e.setTemplateDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateTemplate(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeTemplates encodes a set of resource into the database
func (e StoreEncoder) encodeTemplates(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeTemplate(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeTemplate encodes the resource into the database
func (e StoreEncoder) encodeTemplate(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertTemplate(ctx, s, n.Resource.(*types.Template))
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

// matchupTemplates returns an index with indicates what resources already exist
func (e StoreEncoder) matchupTemplates(ctx context.Context, s store.Storer, uu map[int]types.Template, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchTemplates(ctx, s, types.TemplateFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.Template, len(aa))
	strMap := make(map[string]*types.Template, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.Template
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
// Functions for resource user
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareUser prepares the resources of the given type for encoding
func (e StoreEncoder) prepareUser(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.User, len(nn))
	err = e.matchupUsers(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareUser with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.User)
		if !ok {
			panic("unexpected resource type: node expecting type of user")
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
			err = e.setUserDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateUser(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeUsers encodes a set of resource into the database
func (e StoreEncoder) encodeUsers(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeUser(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeUser encodes the resource into the database
func (e StoreEncoder) encodeUser(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertUser(ctx, s, n.Resource.(*types.User))
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

// matchupUsers returns an index with indicates what resources already exist
func (e StoreEncoder) matchupUsers(ctx context.Context, s store.Storer, uu map[int]types.User, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchUsers(ctx, s, types.UserFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.User, len(aa))
	strMap := make(map[string]*types.User, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.User
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
// Functions for resource dalConnection
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareDalConnection prepares the resources of the given type for encoding
func (e StoreEncoder) prepareDalConnection(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.DalConnection, len(nn))
	err = e.matchupDalConnections(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareDalConnection with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.DalConnection)
		if !ok {
			panic("unexpected resource type: node expecting type of dalConnection")
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
			err = e.setDalConnectionDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateDalConnection(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeDalConnections encodes a set of resource into the database
func (e StoreEncoder) encodeDalConnections(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeDalConnection(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeDalConnection encodes the resource into the database
func (e StoreEncoder) encodeDalConnection(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertDalConnection(ctx, s, n.Resource.(*types.DalConnection))
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

// matchupDalConnections returns an index with indicates what resources already exist
func (e StoreEncoder) matchupDalConnections(ctx context.Context, s store.Storer, uu map[int]types.DalConnection, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchDalConnections(ctx, s, types.DalConnectionFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.DalConnection, len(aa))
	strMap := make(map[string]*types.DalConnection, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.DalConnection
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
// Functions for resource dalSensitivityLevel
// // // // // // // // // // // // // // // // // // // // // // // // //

// prepareDalSensitivityLevel prepares the resources of the given type for encoding
func (e StoreEncoder) prepareDalSensitivityLevel(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet) (err error) {
	// Grab an index of already existing resources of this type
	// @note since these resources should be fairly low-volume and existing for
	//       a short time (and because we batch by resource type); fetching them all
	//       into memory shouldn't hurt too much.
	// @todo do some benchmarks and potentially implement some smarter check such as
	//       a bloom filter or something similar.

	// Initializing the index here (and using a hashmap) so it's not escaped to the heap
	existing := make(map[int]types.DalSensitivityLevel, len(nn))
	err = e.matchupDalSensitivityLevels(ctx, s, existing, nn)
	if err != nil {
		return
	}

	for i, n := range nn {
		if n.Resource == nil {
			panic("unexpected state: cannot call prepareDalSensitivityLevel with nodes without a defined Resource")
		}

		res, ok := n.Resource.(*types.DalSensitivityLevel)
		if !ok {
			panic("unexpected resource type: node expecting type of dalSensitivityLevel")
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
			err = e.setDalSensitivityLevelDefaults(res)
			if err != nil {
				return err
			}

			err = e.validateDalSensitivityLevel(res)
			if err != nil {
				return err
			}
		}

		n.Resource = res
	}

	return
}

// encodeDalSensitivityLevels encodes a set of resource into the database
func (e StoreEncoder) encodeDalSensitivityLevels(ctx context.Context, p envoyx.EncodeParams, s store.Storer, nn envoyx.NodeSet, tree envoyx.Traverser) (err error) {
	for _, n := range nn {
		err = e.encodeDalSensitivityLevel(ctx, p, s, n, tree)
		if err != nil {
			return
		}
	}

	return
}

// encodeDalSensitivityLevel encodes the resource into the database
func (e StoreEncoder) encodeDalSensitivityLevel(ctx context.Context, p envoyx.EncodeParams, s store.Storer, n *envoyx.Node, tree envoyx.Traverser) (err error) {
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
	err = store.UpsertDalSensitivityLevel(ctx, s, n.Resource.(*types.DalSensitivityLevel))
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

// matchupDalSensitivityLevels returns an index with indicates what resources already exist
func (e StoreEncoder) matchupDalSensitivityLevels(ctx context.Context, s store.Storer, uu map[int]types.DalSensitivityLevel, nn envoyx.NodeSet) (err error) {
	// @todo might need to do it smarter then this.
	//       Most resources won't really be that vast so this should be acceptable for now.
	aa, _, err := store.SearchDalSensitivityLevels(ctx, s, types.DalSensitivityLevelFilter{})
	if err != nil {
		return
	}

	idMap := make(map[uint64]*types.DalSensitivityLevel, len(aa))
	strMap := make(map[string]*types.DalSensitivityLevel, len(aa))

	for _, a := range aa {
		strMap[a.Handle] = a
		idMap[a.ID] = a

	}

	var aux *types.DalSensitivityLevel
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