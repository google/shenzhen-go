// Copyright 2017 Google Inc.
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

// Package api has types for communicating with the UI.
package api

// Interface is "the API" - the interface defining available requests.
type Interface interface {
	CreateChannel(*CreateChannelRequest) error
	ConnectPin(*ConnectPinRequest) error
	DeleteChannel(*ChannelRequest) error
	DisconnectPin(*PinRequest) error
	SetGraphProperties(*SetGraphPropertiesRequest) error
	SetNodeProperties(*SetNodePropertiesRequest) error
	SetPosition(*SetPositionRequest) error
}

// Empty is just an empty message.
type Empty struct{}

// Request is the embedded base of all requests.
type Request struct {
	Graph string `json:"graph"`
}

// SetGraphPropertiesRequest is a request to set graph metadata.
type SetGraphPropertiesRequest struct {
	Request
	Name        string `json:"name"`
	PackagePath string `json:"package_path"`
	IsCommand   bool   `json:"is_command"`
}

// ChannelRequest is the embedded base of all requests to do with channels.
type ChannelRequest struct {
	Request
	Channel string `json:"channel"`
}

// CreateChannelRequest asks for a channel to be created.
type CreateChannelRequest struct {
	Request
	Name      string `json:"name"`
	Type      string `json:"type"`
	Anonymous bool   `json:"anon"`
	Capacity  int    `json:"cap"`
}

// NodeRequest is the embedded base of all requests to do with nodes.
type NodeRequest struct {
	Request
	Node string `json:"node"`
}

// SetNodePropertiesRequest is a request to change metadata of a node.
type SetNodePropertiesRequest struct {
	NodeRequest
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	Multiplicity uint   `json:"multiplicity"`
	Wait         bool   `json:"wait"`
}

// SetPositionRequest is a request to change the position of a node.
type SetPositionRequest struct {
	NodeRequest
	X int `json:"x"`
	Y int `json:"y"`
}

// PinRequest is the embedded base of all requests to do with pins.
type PinRequest struct {
	NodeRequest
	Pin string `json:"pin"`
}

// ConnectPinRequest asks for a pin to be attached to a channel.
type ConnectPinRequest struct {
	PinRequest
	Channel string `json:"channel"`
}
