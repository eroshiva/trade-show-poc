// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
)

// DeviceStatusCreate is the builder for creating a DeviceStatus entity.
type DeviceStatusCreate struct {
	config
	mutation *DeviceStatusMutation
	hooks    []Hook
}

// SetStatus sets the "status" field.
func (dsc *DeviceStatusCreate) SetStatus(d devicestatus.Status) *DeviceStatusCreate {
	dsc.mutation.SetStatus(d)
	return dsc
}

// SetLastSeen sets the "last_seen" field.
func (dsc *DeviceStatusCreate) SetLastSeen(s string) *DeviceStatusCreate {
	dsc.mutation.SetLastSeen(s)
	return dsc
}

// SetNillableLastSeen sets the "last_seen" field if the given value is not nil.
func (dsc *DeviceStatusCreate) SetNillableLastSeen(s *string) *DeviceStatusCreate {
	if s != nil {
		dsc.SetLastSeen(*s)
	}
	return dsc
}

// SetConsequentialFailedConnectivityAttempts sets the "consequential_failed_connectivity_attempts" field.
func (dsc *DeviceStatusCreate) SetConsequentialFailedConnectivityAttempts(i int32) *DeviceStatusCreate {
	dsc.mutation.SetConsequentialFailedConnectivityAttempts(i)
	return dsc
}

// SetID sets the "id" field.
func (dsc *DeviceStatusCreate) SetID(s string) *DeviceStatusCreate {
	dsc.mutation.SetID(s)
	return dsc
}

// SetNetworkDeviceID sets the "network_device" edge to the NetworkDevice entity by ID.
func (dsc *DeviceStatusCreate) SetNetworkDeviceID(id string) *DeviceStatusCreate {
	dsc.mutation.SetNetworkDeviceID(id)
	return dsc
}

// SetNillableNetworkDeviceID sets the "network_device" edge to the NetworkDevice entity by ID if the given value is not nil.
func (dsc *DeviceStatusCreate) SetNillableNetworkDeviceID(id *string) *DeviceStatusCreate {
	if id != nil {
		dsc = dsc.SetNetworkDeviceID(*id)
	}
	return dsc
}

// SetNetworkDevice sets the "network_device" edge to the NetworkDevice entity.
func (dsc *DeviceStatusCreate) SetNetworkDevice(n *NetworkDevice) *DeviceStatusCreate {
	return dsc.SetNetworkDeviceID(n.ID)
}

// Mutation returns the DeviceStatusMutation object of the builder.
func (dsc *DeviceStatusCreate) Mutation() *DeviceStatusMutation {
	return dsc.mutation
}

// Save creates the DeviceStatus in the database.
func (dsc *DeviceStatusCreate) Save(ctx context.Context) (*DeviceStatus, error) {
	return withHooks(ctx, dsc.sqlSave, dsc.mutation, dsc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (dsc *DeviceStatusCreate) SaveX(ctx context.Context) *DeviceStatus {
	v, err := dsc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dsc *DeviceStatusCreate) Exec(ctx context.Context) error {
	_, err := dsc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dsc *DeviceStatusCreate) ExecX(ctx context.Context) {
	if err := dsc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dsc *DeviceStatusCreate) check() error {
	if _, ok := dsc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "DeviceStatus.status"`)}
	}
	if v, ok := dsc.mutation.Status(); ok {
		if err := devicestatus.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "DeviceStatus.status": %w`, err)}
		}
	}
	if _, ok := dsc.mutation.ConsequentialFailedConnectivityAttempts(); !ok {
		return &ValidationError{Name: "consequential_failed_connectivity_attempts", err: errors.New(`ent: missing required field "DeviceStatus.consequential_failed_connectivity_attempts"`)}
	}
	return nil
}

func (dsc *DeviceStatusCreate) sqlSave(ctx context.Context) (*DeviceStatus, error) {
	if err := dsc.check(); err != nil {
		return nil, err
	}
	_node, _spec := dsc.createSpec()
	if err := sqlgraph.CreateNode(ctx, dsc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected DeviceStatus.ID type: %T", _spec.ID.Value)
		}
	}
	dsc.mutation.id = &_node.ID
	dsc.mutation.done = true
	return _node, nil
}

func (dsc *DeviceStatusCreate) createSpec() (*DeviceStatus, *sqlgraph.CreateSpec) {
	var (
		_node = &DeviceStatus{config: dsc.config}
		_spec = sqlgraph.NewCreateSpec(devicestatus.Table, sqlgraph.NewFieldSpec(devicestatus.FieldID, field.TypeString))
	)
	if id, ok := dsc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := dsc.mutation.Status(); ok {
		_spec.SetField(devicestatus.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := dsc.mutation.LastSeen(); ok {
		_spec.SetField(devicestatus.FieldLastSeen, field.TypeString, value)
		_node.LastSeen = value
	}
	if value, ok := dsc.mutation.ConsequentialFailedConnectivityAttempts(); ok {
		_spec.SetField(devicestatus.FieldConsequentialFailedConnectivityAttempts, field.TypeInt32, value)
		_node.ConsequentialFailedConnectivityAttempts = value
	}
	if nodes := dsc.mutation.NetworkDeviceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   devicestatus.NetworkDeviceTable,
			Columns: []string{devicestatus.NetworkDeviceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(networkdevice.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.device_status_network_device = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// DeviceStatusCreateBulk is the builder for creating many DeviceStatus entities in bulk.
type DeviceStatusCreateBulk struct {
	config
	err      error
	builders []*DeviceStatusCreate
}

// Save creates the DeviceStatus entities in the database.
func (dscb *DeviceStatusCreateBulk) Save(ctx context.Context) ([]*DeviceStatus, error) {
	if dscb.err != nil {
		return nil, dscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(dscb.builders))
	nodes := make([]*DeviceStatus, len(dscb.builders))
	mutators := make([]Mutator, len(dscb.builders))
	for i := range dscb.builders {
		func(i int, root context.Context) {
			builder := dscb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*DeviceStatusMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, dscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, dscb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, dscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (dscb *DeviceStatusCreateBulk) SaveX(ctx context.Context) []*DeviceStatus {
	v, err := dscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dscb *DeviceStatusCreateBulk) Exec(ctx context.Context) error {
	_, err := dscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dscb *DeviceStatusCreateBulk) ExecX(ctx context.Context) {
	if err := dscb.Exec(ctx); err != nil {
		panic(err)
	}
}
