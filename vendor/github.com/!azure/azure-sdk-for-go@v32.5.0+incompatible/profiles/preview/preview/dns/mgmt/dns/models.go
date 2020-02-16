// +build go1.9

// Copyright 2019 Microsoft Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code was auto-generated by:
// github.com/Azure/azure-sdk-for-go/tools/profileBuilder

package dns

import (
	"context"

	original "github.com/Azure/azure-sdk-for-go/services/preview/dns/mgmt/2018-03-01-preview/dns"
)

const (
	DefaultBaseURI = original.DefaultBaseURI
)

type RecordType = original.RecordType

const (
	A     RecordType = original.A
	AAAA  RecordType = original.AAAA
	CAA   RecordType = original.CAA
	CNAME RecordType = original.CNAME
	MX    RecordType = original.MX
	NS    RecordType = original.NS
	PTR   RecordType = original.PTR
	SOA   RecordType = original.SOA
	SRV   RecordType = original.SRV
	TXT   RecordType = original.TXT
)

type ZoneType = original.ZoneType

const (
	Private ZoneType = original.Private
	Public  ZoneType = original.Public
)

type ARecord = original.ARecord
type AaaaRecord = original.AaaaRecord
type AzureEntityResource = original.AzureEntityResource
type BaseClient = original.BaseClient
type CaaRecord = original.CaaRecord
type CloudError = original.CloudError
type CloudErrorBody = original.CloudErrorBody
type CnameRecord = original.CnameRecord
type MxRecord = original.MxRecord
type NsRecord = original.NsRecord
type ProxyResource = original.ProxyResource
type PtrRecord = original.PtrRecord
type RecordSet = original.RecordSet
type RecordSetListResult = original.RecordSetListResult
type RecordSetListResultIterator = original.RecordSetListResultIterator
type RecordSetListResultPage = original.RecordSetListResultPage
type RecordSetProperties = original.RecordSetProperties
type RecordSetUpdateParameters = original.RecordSetUpdateParameters
type RecordSetsClient = original.RecordSetsClient
type Resource = original.Resource
type SoaRecord = original.SoaRecord
type SrvRecord = original.SrvRecord
type SubResource = original.SubResource
type TrackedResource = original.TrackedResource
type TxtRecord = original.TxtRecord
type Zone = original.Zone
type ZoneListResult = original.ZoneListResult
type ZoneListResultIterator = original.ZoneListResultIterator
type ZoneListResultPage = original.ZoneListResultPage
type ZoneProperties = original.ZoneProperties
type ZoneUpdate = original.ZoneUpdate
type ZonesClient = original.ZonesClient
type ZonesDeleteFuture = original.ZonesDeleteFuture

func New(subscriptionID string) BaseClient {
	return original.New(subscriptionID)
}
func NewRecordSetListResultIterator(page RecordSetListResultPage) RecordSetListResultIterator {
	return original.NewRecordSetListResultIterator(page)
}
func NewRecordSetListResultPage(getNextPage func(context.Context, RecordSetListResult) (RecordSetListResult, error)) RecordSetListResultPage {
	return original.NewRecordSetListResultPage(getNextPage)
}
func NewRecordSetsClient(subscriptionID string) RecordSetsClient {
	return original.NewRecordSetsClient(subscriptionID)
}
func NewRecordSetsClientWithBaseURI(baseURI string, subscriptionID string) RecordSetsClient {
	return original.NewRecordSetsClientWithBaseURI(baseURI, subscriptionID)
}
func NewWithBaseURI(baseURI string, subscriptionID string) BaseClient {
	return original.NewWithBaseURI(baseURI, subscriptionID)
}
func NewZoneListResultIterator(page ZoneListResultPage) ZoneListResultIterator {
	return original.NewZoneListResultIterator(page)
}
func NewZoneListResultPage(getNextPage func(context.Context, ZoneListResult) (ZoneListResult, error)) ZoneListResultPage {
	return original.NewZoneListResultPage(getNextPage)
}
func NewZonesClient(subscriptionID string) ZonesClient {
	return original.NewZonesClient(subscriptionID)
}
func NewZonesClientWithBaseURI(baseURI string, subscriptionID string) ZonesClient {
	return original.NewZonesClientWithBaseURI(baseURI, subscriptionID)
}
func PossibleRecordTypeValues() []RecordType {
	return original.PossibleRecordTypeValues()
}
func PossibleZoneTypeValues() []ZoneType {
	return original.PossibleZoneTypeValues()
}
func UserAgent() string {
	return original.UserAgent() + " profiles/preview"
}
func Version() string {
	return original.Version()
}
