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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const contentType = "application/vnd.api+json"

type client struct {
	client   *http.Client
	endpoint string
}

// NewClient returns a client that makes queries to the endpoint, for each
// request.
func NewClient(endpoint string) Interface {
	return &client{endpoint: endpoint}
}

func (c *client) query(method string, req, resp interface{}) error {
	rb, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshalling message: %v", err)
	}
	b, err := json.Marshal(&jsonRequest{
		Method:  method,
		Message: rb,
	})
	if err != nil {
		return fmt.Errorf("marshalling request: %v", err)
	}
	rsp, err := http.Post(c.endpoint, contentType, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("posting request: %v", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		b := new(bytes.Buffer)
		io.Copy(b, rsp.Body)
		return fmt.Errorf("non-200 status code: %s", b)
	}

	if err := json.NewDecoder(rsp.Body).Decode(resp); err != nil {
		return fmt.Errorf("unmarshalling response: %v", err)
	}
	return nil
}

func (c *client) CreateChannel(req *CreateChannelRequest) error {
	return c.query("CreateChannel", req, &Empty{})
}

func (c *client) ConnectPin(req *ConnectPinRequest) error {
	return c.query("ConnectPin", req, &Empty{})
}

func (c *client) DeleteChannel(req *ChannelRequest) error {
	return c.query("DeleteChannel", req, &Empty{})
}

func (c *client) DisconnectPin(req *PinRequest) error {
	return c.query("DisconnectPin", req, &Empty{})
}

func (c *client) Save(req *Request) error {
	return c.query("Save", req, &Empty{})
}

func (c *client) SetGraphProperties(req *SetGraphPropertiesRequest) error {
	return c.query("SetGraphProperties", req, &Empty{})
}

func (c *client) SetNodeProperties(req *SetNodePropertiesRequest) error {
	return c.query("SetNodeProperties", req, &Empty{})
}

func (c *client) SetPosition(req *SetPositionRequest) error {
	return c.query("SetPosition", req, &Empty{})
}
