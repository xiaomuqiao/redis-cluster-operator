package datashare

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
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/tracing"
	"net/http"
)

// ConsumerSourceDataSetsClient is the creates a Microsoft.DataShare management client.
type ConsumerSourceDataSetsClient struct {
	BaseClient
}

// NewConsumerSourceDataSetsClient creates an instance of the ConsumerSourceDataSetsClient client.
func NewConsumerSourceDataSetsClient(subscriptionID string) ConsumerSourceDataSetsClient {
	return NewConsumerSourceDataSetsClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewConsumerSourceDataSetsClientWithBaseURI creates an instance of the ConsumerSourceDataSetsClient client.
func NewConsumerSourceDataSetsClientWithBaseURI(baseURI string, subscriptionID string) ConsumerSourceDataSetsClient {
	return ConsumerSourceDataSetsClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// ListByShareSubscription get source dataSets of a shareSubscription
// Parameters:
// resourceGroupName - the resource group name.
// accountName - the name of the share account.
// shareSubscriptionName - the name of the shareSubscription.
// skipToken - continuation token
func (client ConsumerSourceDataSetsClient) ListByShareSubscription(ctx context.Context, resourceGroupName string, accountName string, shareSubscriptionName string, skipToken string) (result ConsumerSourceDataSetListPage, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ConsumerSourceDataSetsClient.ListByShareSubscription")
		defer func() {
			sc := -1
			if result.csdsl.Response.Response != nil {
				sc = result.csdsl.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	result.fn = client.listByShareSubscriptionNextResults
	req, err := client.ListByShareSubscriptionPreparer(ctx, resourceGroupName, accountName, shareSubscriptionName, skipToken)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "ListByShareSubscription", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListByShareSubscriptionSender(req)
	if err != nil {
		result.csdsl.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "ListByShareSubscription", resp, "Failure sending request")
		return
	}

	result.csdsl, err = client.ListByShareSubscriptionResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "ListByShareSubscription", resp, "Failure responding to request")
	}

	return
}

// ListByShareSubscriptionPreparer prepares the ListByShareSubscription request.
func (client ConsumerSourceDataSetsClient) ListByShareSubscriptionPreparer(ctx context.Context, resourceGroupName string, accountName string, shareSubscriptionName string, skipToken string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"accountName":           autorest.Encode("path", accountName),
		"resourceGroupName":     autorest.Encode("path", resourceGroupName),
		"shareSubscriptionName": autorest.Encode("path", shareSubscriptionName),
		"subscriptionId":        autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-11-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}
	if len(skipToken) > 0 {
		queryParameters["$skipToken"] = autorest.Encode("query", skipToken)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DataShare/accounts/{accountName}/shareSubscriptions/{shareSubscriptionName}/ConsumerSourceDataSets", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ListByShareSubscriptionSender sends the ListByShareSubscription request. The method will close the
// http.Response Body if it receives an error.
func (client ConsumerSourceDataSetsClient) ListByShareSubscriptionSender(req *http.Request) (*http.Response, error) {
	sd := autorest.GetSendDecorators(req.Context(), azure.DoRetryWithRegistration(client.Client))
	return autorest.SendWithSender(client, req, sd...)
}

// ListByShareSubscriptionResponder handles the response to the ListByShareSubscription request. The method always
// closes the http.Response Body.
func (client ConsumerSourceDataSetsClient) ListByShareSubscriptionResponder(resp *http.Response) (result ConsumerSourceDataSetList, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// listByShareSubscriptionNextResults retrieves the next set of results, if any.
func (client ConsumerSourceDataSetsClient) listByShareSubscriptionNextResults(ctx context.Context, lastResults ConsumerSourceDataSetList) (result ConsumerSourceDataSetList, err error) {
	req, err := lastResults.consumerSourceDataSetListPreparer(ctx)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "listByShareSubscriptionNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}
	resp, err := client.ListByShareSubscriptionSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "listByShareSubscriptionNextResults", resp, "Failure sending next results request")
	}
	result, err = client.ListByShareSubscriptionResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datashare.ConsumerSourceDataSetsClient", "listByShareSubscriptionNextResults", resp, "Failure responding to next results request")
	}
	return
}

// ListByShareSubscriptionComplete enumerates all values, automatically crossing page boundaries as required.
func (client ConsumerSourceDataSetsClient) ListByShareSubscriptionComplete(ctx context.Context, resourceGroupName string, accountName string, shareSubscriptionName string, skipToken string) (result ConsumerSourceDataSetListIterator, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ConsumerSourceDataSetsClient.ListByShareSubscription")
		defer func() {
			sc := -1
			if result.Response().Response.Response != nil {
				sc = result.page.Response().Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	result.page, err = client.ListByShareSubscription(ctx, resourceGroupName, accountName, shareSubscriptionName, skipToken)
	return
}
