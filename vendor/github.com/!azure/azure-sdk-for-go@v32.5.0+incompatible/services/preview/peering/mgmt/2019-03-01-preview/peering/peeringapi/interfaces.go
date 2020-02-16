package peeringapi

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/services/preview/peering/mgmt/2019-03-01-preview/peering"
	"github.com/Azure/go-autorest/autorest"
)

// LegacyPeeringsClientAPI contains the set of methods on the LegacyPeeringsClient type.
type LegacyPeeringsClientAPI interface {
	List(ctx context.Context, peeringLocation string, kind string) (result peering.ListResultPage, err error)
}

var _ LegacyPeeringsClientAPI = (*peering.LegacyPeeringsClient)(nil)

// OperationsClientAPI contains the set of methods on the OperationsClient type.
type OperationsClientAPI interface {
	List(ctx context.Context) (result peering.OperationListResultPage, err error)
}

var _ OperationsClientAPI = (*peering.OperationsClient)(nil)

// PeerAsnsClientAPI contains the set of methods on the PeerAsnsClient type.
type PeerAsnsClientAPI interface {
	CreateOrUpdate(ctx context.Context, peerAsnName string, peerAsn peering.PeerAsn) (result peering.PeerAsn, err error)
	Delete(ctx context.Context, peerAsnName string) (result autorest.Response, err error)
	Get(ctx context.Context, peerAsnName string) (result peering.PeerAsn, err error)
	ListBySubscription(ctx context.Context) (result peering.PeerAsnListResultPage, err error)
}

var _ PeerAsnsClientAPI = (*peering.PeerAsnsClient)(nil)

// LocationsClientAPI contains the set of methods on the LocationsClient type.
type LocationsClientAPI interface {
	List(ctx context.Context, kind string) (result peering.LocationListResultPage, err error)
}

var _ LocationsClientAPI = (*peering.LocationsClient)(nil)

// PeeringsClientAPI contains the set of methods on the PeeringsClient type.
type PeeringsClientAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, peeringName string, peering peering.Model) (result peering.Model, err error)
	Delete(ctx context.Context, resourceGroupName string, peeringName string) (result autorest.Response, err error)
	Get(ctx context.Context, resourceGroupName string, peeringName string) (result peering.Model, err error)
	ListByResourceGroup(ctx context.Context, resourceGroupName string) (result peering.ListResultPage, err error)
	ListBySubscription(ctx context.Context) (result peering.ListResultPage, err error)
	Update(ctx context.Context, resourceGroupName string, peeringName string, tags peering.ResourceTags) (result peering.Model, err error)
}

var _ PeeringsClientAPI = (*peering.PeeringsClient)(nil)