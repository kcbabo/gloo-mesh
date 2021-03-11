// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./clients.go -destination mocks/clients.go

package v1

import (
	"context"

	"github.com/solo-io/skv2/pkg/controllerutils"
	"github.com/solo-io/skv2/pkg/multicluster"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MulticlusterClientset for the rbac.enterprise.mesh.gloo.solo.io/v1 APIs
type MulticlusterClientset interface {
	// Cluster returns a Clientset for the given cluster
	Cluster(cluster string) (Clientset, error)
}

type multiclusterClientset struct {
	client multicluster.Client
}

func NewMulticlusterClientset(client multicluster.Client) MulticlusterClientset {
	return &multiclusterClientset{client: client}
}

func (m *multiclusterClientset) Cluster(cluster string) (Clientset, error) {
	client, err := m.client.Cluster(cluster)
	if err != nil {
		return nil, err
	}
	return NewClientset(client), nil
}

// clienset for the rbac.enterprise.mesh.gloo.solo.io/v1 APIs
type Clientset interface {
	// clienset for the rbac.enterprise.mesh.gloo.solo.io/v1/v1 APIs
	Roles() RoleClient
	// clienset for the rbac.enterprise.mesh.gloo.solo.io/v1/v1 APIs
	RoleBindings() RoleBindingClient
}

type clientSet struct {
	client client.Client
}

func NewClientsetFromConfig(cfg *rest.Config) (Clientset, error) {
	scheme := scheme.Scheme
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return NewClientset(client), nil
}

func NewClientset(client client.Client) Clientset {
	return &clientSet{client: client}
}

// clienset for the rbac.enterprise.mesh.gloo.solo.io/v1/v1 APIs
func (c *clientSet) Roles() RoleClient {
	return NewRoleClient(c.client)
}

// clienset for the rbac.enterprise.mesh.gloo.solo.io/v1/v1 APIs
func (c *clientSet) RoleBindings() RoleBindingClient {
	return NewRoleBindingClient(c.client)
}

// Reader knows how to read and list Roles.
type RoleReader interface {
	// Get retrieves a Role for the given object key
	GetRole(ctx context.Context, key client.ObjectKey) (*Role, error)

	// List retrieves list of Roles for a given namespace and list options.
	ListRole(ctx context.Context, opts ...client.ListOption) (*RoleList, error)
}

// RoleTransitionFunction instructs the RoleWriter how to transition between an existing
// Role object and a desired on an Upsert
type RoleTransitionFunction func(existing, desired *Role) error

// Writer knows how to create, delete, and update Roles.
type RoleWriter interface {
	// Create saves the Role object.
	CreateRole(ctx context.Context, obj *Role, opts ...client.CreateOption) error

	// Delete deletes the Role object.
	DeleteRole(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error

	// Update updates the given Role object.
	UpdateRole(ctx context.Context, obj *Role, opts ...client.UpdateOption) error

	// Patch patches the given Role object.
	PatchRole(ctx context.Context, obj *Role, patch client.Patch, opts ...client.PatchOption) error

	// DeleteAllOf deletes all Role objects matching the given options.
	DeleteAllOfRole(ctx context.Context, opts ...client.DeleteAllOfOption) error

	// Create or Update the Role object.
	UpsertRole(ctx context.Context, obj *Role, transitionFuncs ...RoleTransitionFunction) error
}

// StatusWriter knows how to update status subresource of a Role object.
type RoleStatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given Role object.
	UpdateRoleStatus(ctx context.Context, obj *Role, opts ...client.UpdateOption) error

	// Patch patches the given Role object's subresource.
	PatchRoleStatus(ctx context.Context, obj *Role, patch client.Patch, opts ...client.PatchOption) error
}

// Client knows how to perform CRUD operations on Roles.
type RoleClient interface {
	RoleReader
	RoleWriter
	RoleStatusWriter
}

type roleClient struct {
	client client.Client
}

func NewRoleClient(client client.Client) *roleClient {
	return &roleClient{client: client}
}

func (c *roleClient) GetRole(ctx context.Context, key client.ObjectKey) (*Role, error) {
	obj := &Role{}
	if err := c.client.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *roleClient) ListRole(ctx context.Context, opts ...client.ListOption) (*RoleList, error) {
	list := &RoleList{}
	if err := c.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *roleClient) CreateRole(ctx context.Context, obj *Role, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *roleClient) DeleteRole(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	obj := &Role{}
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)
	return c.client.Delete(ctx, obj, opts...)
}

func (c *roleClient) UpdateRole(ctx context.Context, obj *Role, opts ...client.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c *roleClient) PatchRole(ctx context.Context, obj *Role, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *roleClient) DeleteAllOfRole(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	obj := &Role{}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *roleClient) UpsertRole(ctx context.Context, obj *Role, transitionFuncs ...RoleTransitionFunction) error {
	genericTxFunc := func(existing, desired runtime.Object) error {
		for _, txFunc := range transitionFuncs {
			if err := txFunc(existing.(*Role), desired.(*Role)); err != nil {
				return err
			}
		}
		return nil
	}
	_, err := controllerutils.Upsert(ctx, c.client, obj, genericTxFunc)
	return err
}

func (c *roleClient) UpdateRoleStatus(ctx context.Context, obj *Role, opts ...client.UpdateOption) error {
	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *roleClient) PatchRoleStatus(ctx context.Context, obj *Role, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Status().Patch(ctx, obj, patch, opts...)
}

// Provides RoleClients for multiple clusters.
type MulticlusterRoleClient interface {
	// Cluster returns a RoleClient for the given cluster
	Cluster(cluster string) (RoleClient, error)
}

type multiclusterRoleClient struct {
	client multicluster.Client
}

func NewMulticlusterRoleClient(client multicluster.Client) MulticlusterRoleClient {
	return &multiclusterRoleClient{client: client}
}

func (m *multiclusterRoleClient) Cluster(cluster string) (RoleClient, error) {
	client, err := m.client.Cluster(cluster)
	if err != nil {
		return nil, err
	}
	return NewRoleClient(client), nil
}

// Reader knows how to read and list RoleBindings.
type RoleBindingReader interface {
	// Get retrieves a RoleBinding for the given object key
	GetRoleBinding(ctx context.Context, key client.ObjectKey) (*RoleBinding, error)

	// List retrieves list of RoleBindings for a given namespace and list options.
	ListRoleBinding(ctx context.Context, opts ...client.ListOption) (*RoleBindingList, error)
}

// RoleBindingTransitionFunction instructs the RoleBindingWriter how to transition between an existing
// RoleBinding object and a desired on an Upsert
type RoleBindingTransitionFunction func(existing, desired *RoleBinding) error

// Writer knows how to create, delete, and update RoleBindings.
type RoleBindingWriter interface {
	// Create saves the RoleBinding object.
	CreateRoleBinding(ctx context.Context, obj *RoleBinding, opts ...client.CreateOption) error

	// Delete deletes the RoleBinding object.
	DeleteRoleBinding(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error

	// Update updates the given RoleBinding object.
	UpdateRoleBinding(ctx context.Context, obj *RoleBinding, opts ...client.UpdateOption) error

	// Patch patches the given RoleBinding object.
	PatchRoleBinding(ctx context.Context, obj *RoleBinding, patch client.Patch, opts ...client.PatchOption) error

	// DeleteAllOf deletes all RoleBinding objects matching the given options.
	DeleteAllOfRoleBinding(ctx context.Context, opts ...client.DeleteAllOfOption) error

	// Create or Update the RoleBinding object.
	UpsertRoleBinding(ctx context.Context, obj *RoleBinding, transitionFuncs ...RoleBindingTransitionFunction) error
}

// StatusWriter knows how to update status subresource of a RoleBinding object.
type RoleBindingStatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given RoleBinding object.
	UpdateRoleBindingStatus(ctx context.Context, obj *RoleBinding, opts ...client.UpdateOption) error

	// Patch patches the given RoleBinding object's subresource.
	PatchRoleBindingStatus(ctx context.Context, obj *RoleBinding, patch client.Patch, opts ...client.PatchOption) error
}

// Client knows how to perform CRUD operations on RoleBindings.
type RoleBindingClient interface {
	RoleBindingReader
	RoleBindingWriter
	RoleBindingStatusWriter
}

type roleBindingClient struct {
	client client.Client
}

func NewRoleBindingClient(client client.Client) *roleBindingClient {
	return &roleBindingClient{client: client}
}

func (c *roleBindingClient) GetRoleBinding(ctx context.Context, key client.ObjectKey) (*RoleBinding, error) {
	obj := &RoleBinding{}
	if err := c.client.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *roleBindingClient) ListRoleBinding(ctx context.Context, opts ...client.ListOption) (*RoleBindingList, error) {
	list := &RoleBindingList{}
	if err := c.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *roleBindingClient) CreateRoleBinding(ctx context.Context, obj *RoleBinding, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *roleBindingClient) DeleteRoleBinding(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	obj := &RoleBinding{}
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)
	return c.client.Delete(ctx, obj, opts...)
}

func (c *roleBindingClient) UpdateRoleBinding(ctx context.Context, obj *RoleBinding, opts ...client.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c *roleBindingClient) PatchRoleBinding(ctx context.Context, obj *RoleBinding, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *roleBindingClient) DeleteAllOfRoleBinding(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	obj := &RoleBinding{}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *roleBindingClient) UpsertRoleBinding(ctx context.Context, obj *RoleBinding, transitionFuncs ...RoleBindingTransitionFunction) error {
	genericTxFunc := func(existing, desired runtime.Object) error {
		for _, txFunc := range transitionFuncs {
			if err := txFunc(existing.(*RoleBinding), desired.(*RoleBinding)); err != nil {
				return err
			}
		}
		return nil
	}
	_, err := controllerutils.Upsert(ctx, c.client, obj, genericTxFunc)
	return err
}

func (c *roleBindingClient) UpdateRoleBindingStatus(ctx context.Context, obj *RoleBinding, opts ...client.UpdateOption) error {
	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *roleBindingClient) PatchRoleBindingStatus(ctx context.Context, obj *RoleBinding, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Status().Patch(ctx, obj, patch, opts...)
}

// Provides RoleBindingClients for multiple clusters.
type MulticlusterRoleBindingClient interface {
	// Cluster returns a RoleBindingClient for the given cluster
	Cluster(cluster string) (RoleBindingClient, error)
}

type multiclusterRoleBindingClient struct {
	client multicluster.Client
}

func NewMulticlusterRoleBindingClient(client multicluster.Client) MulticlusterRoleBindingClient {
	return &multiclusterRoleBindingClient{client: client}
}

func (m *multiclusterRoleBindingClient) Cluster(cluster string) (RoleBindingClient, error) {
	client, err := m.client.Cluster(cluster)
	if err != nil {
		return nil, err
	}
	return NewRoleBindingClient(client), nil
}