// Copyright 2018 Google Inc.
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

package proto

import (
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"golang.org/x/net/context"
)

// Ensure interface is satisfied.
var _ ShenzhenGoClient = UnimplementedShenzhenGoClient{}

// UnimplementedShenzhenGoClient is for embedding in test fakes.
type UnimplementedShenzhenGoClient struct{}

// CreateChannel does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// CreateNode does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) CreateNode(ctx context.Context, in *CreateNodeRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// ConnectPin does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) ConnectPin(ctx context.Context, in *ConnectPinRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// DeleteChannel does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) DeleteChannel(ctx context.Context, in *DeleteChannelRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// DeleteNode does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) DeleteNode(ctx context.Context, in *DeleteNodeRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// DisconnectPin does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) DisconnectPin(ctx context.Context, in *DisconnectPinRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// Save does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) Save(ctx context.Context, in *SaveRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// SetGraphProperties does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) SetGraphProperties(ctx context.Context, in *SetGraphPropertiesRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// SetNodeProperties does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) SetNodeProperties(ctx context.Context, in *SetNodePropertiesRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}

// SetPosition does nothing and returns nil, nil.
func (UnimplementedShenzhenGoClient) SetPosition(ctx context.Context, in *SetPositionRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	return nil, nil
}
