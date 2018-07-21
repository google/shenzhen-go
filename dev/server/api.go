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

package server

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/go"
)

type actionStreamWriter struct {
	stream pb.ShenzhenGo_ActionServer
}

func (a actionStreamWriter) Write(b []byte) (int, error) {
	if err := a.stream.Send(&pb.ActionResponse{Output: string(b)}); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *server) Action(req *pb.ActionRequest, stream pb.ShenzhenGo_ActionServer) error {
	log.Printf("api: Action(%s)", proto.MarshalTextString(req))
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return err
	}
	g.Lock()
	defer g.Unlock()

	switch req.Action {
	case pb.ActionRequest_SAVE:
		return SaveJSONFile(g.Graph)
	case pb.ActionRequest_REVERT:
		return g.reload()
	case pb.ActionRequest_GENERATE:
		_, err := GeneratePackage(actionStreamWriter{stream}, g.Graph)
		return err
	case pb.ActionRequest_BUILD:
		return Build(actionStreamWriter{stream}, g.Graph)
	case pb.ActionRequest_INSTALL:
		return Install(actionStreamWriter{stream}, g.Graph)
	default:
		return status.Errorf(codes.Unimplemented, "action %v not implemented", req.Action)
	}
}

type runSvrWriter struct {
	svr pb.ShenzhenGo_RunServer
	fn  func([]byte) *pb.Output
}

func (w *runSvrWriter) Write(b []byte) (int, error) {
	if err := w.svr.Send(w.fn(b)); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *server) Run(svr pb.ShenzhenGo_RunServer) error {
	log.Print("api: Run()")

	first, err := svr.Recv()
	if err != nil {
		return err
	}
	g, err := c.lookupGraph(first.Graph)
	if err != nil {
		return err
	}

	stdout := &runSvrWriter{svr, func(b []byte) *pb.Output { return &pb.Output{Out: string(b)} }}
	stderr := &runSvrWriter{svr, func(b []byte) *pb.Output { return &pb.Output{Err: string(b)} }}

	g.Lock()
	gp, err := GenerateRunner(stderr, g.Graph)
	g.Unlock()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(svr.Context(), "go", "run", gp)
	fmt.Fprintf(stderr, "%v\n", cmd.Args)

	// A pipe is better for input; managing a buffer is fiddly, and cmd.Wait
	// will wait until read returns, which doesn't mesh well with the
	// behaviour of svr.Recv.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return status.Errorf(codes.Internal, "attaching stdin pipe: %v", err)
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	// go run forks a child, when we
	setpgid(cmd)
	go func() {
		for {
			in, err := svr.Recv()
			if err != nil {
				stdin.Close()
				return
			}
			if _, err := stdin.Write([]byte(in.In)); err != nil {
				return
			}
			if in.In == "\x03" { // ^C
				//log.Printf("svr.Recv got ^C, sending os.Interrupt to %d\n", cmd.Process.Pid)
				if err := interrupt(cmd.Process); err != nil {
					fmt.Fprintf(stderr, "(interrupt(%d) = %v\n", cmd.Process.Pid, err)
				}
			}
		}
	}()
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "(process %v)", err)
		return status.Errorf(codes.Aborted, "cmd.Run() = %v", err)
	}
	fmt.Fprintln(stderr, "(process succeeded)")
	return nil
}

func (c *server) SetChannel(ctx context.Context, req *pb.SetChannelRequest) (*pb.Empty, error) {
	log.Printf("api: SetChannel(%s)", proto.MarshalTextString(req))

	if req.Channel == "" && req.Config == nil {
		return &pb.Empty{}, status.Error(codes.InvalidArgument, "must provide existing channel or new config")
	}

	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Lock()
	defer g.Unlock()

	var nps map[model.NodePin]struct{}

	if req.Config != nil {
		// TODO: More validation (name, type, etc)
		if req.Config.Name == "nil" {
			return &pb.Empty{}, status.Errorf(codes.InvalidArgument, "channels may not be named %q", req.Config.Name)
		}

		if req.Channel != req.Config.Name {
			// Check that the new name is available...
			if _, found := g.Channels[req.Config.Name]; found {
				return &pb.Empty{}, status.Errorf(codes.AlreadyExists, "target name %q already exists", req.Config.Name)
			}
		}

		// Convert the []pb.NodePin into a set of model.NodePin, and validate
		// that the pins exist at the same time.
		nps = make(map[model.NodePin]struct{}, len(req.Config.Pins))
		for _, np := range req.Config.Pins {
			n, err := g.lookupNode(np.Node)
			if err != nil {
				return &pb.Empty{}, err
			}
			if _, found := n.Connections[np.Pin]; !found {
				return &pb.Empty{}, status.Errorf(codes.NotFound, "node %q pin %q does not exist", np.Node, np.Pin)
			}
			nps[model.NodePin{Node: np.Node, Pin: np.Pin}] = struct{}{}
		}
	}

	if req.Channel != "" {
		old, err := g.lookupChannel(req.Channel)
		if err != nil {
			return &pb.Empty{}, err
		}

		// Update existing channel data by deleting the old one from the map
		// and any connections, then setting the new one below.
		g.DeleteChannel(old)

		if req.Config == nil {
			// Deletion was intended, job complete.
			return &pb.Empty{}, nil
		}
	}

	// Set entry in map, update connections on node side.
	g.Channels[req.Config.Name] = &model.Channel{
		Name:     req.Config.Name,
		Capacity: int(req.Config.Cap),
		Pins:     nps,
	}
	for np := range nps {
		g.Nodes[np.Node].Connections[np.Pin] = req.Config.Name
	}
	return &pb.Empty{}, nil
}

func (c *server) SetGraphProperties(ctx context.Context, req *pb.SetGraphPropertiesRequest) (*pb.Empty, error) {
	log.Printf("api: SetGraphProperties(%s)", proto.MarshalTextString(req))
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Name = req.Name
	g.PackagePath = req.PackagePath
	g.IsCommand = req.IsCommand
	return &pb.Empty{}, nil
}

func (c *server) SetNode(ctx context.Context, req *pb.SetNodeRequest) (*pb.Empty, error) {
	log.Printf("api: SetNode(%s)", proto.MarshalTextString(req))

	if req.Node == "" && req.Config == nil {
		return &pb.Empty{}, status.Error(codes.InvalidArgument, "must provide existing node or new config")
	}

	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Lock()
	defer g.Unlock()

	var part model.Part
	if req.Config != nil {
		if req.Node != req.Config.Name {
			// Check the new name is available...
			if _, exists := g.Nodes[req.Config.Name]; exists {
				return &pb.Empty{}, status.Errorf(codes.AlreadyExists, "node %q already exists", req.Config.Name)
			}
		}

		p, err := (&model.PartJSON{
			Part: req.Config.PartCfg,
			Type: req.Config.PartType,
		}).Unmarshal()
		if err != nil {
			return &pb.Empty{}, status.Errorf(codes.FailedPrecondition, "part unmarshal: %v", err)
		}
		part = p
	}

	var conns map[string]string
	if req.Node != "" {
		old, err := g.lookupNode(req.Node)
		if err != nil {
			return &pb.Empty{}, err
		}

		// Delete old node, only clean up channels if deleting this node
		// is the intention.
		g.DeleteNode(old, req.Config == nil)

		if req.Config == nil {
			// Deletion was intended, job complete.
			return &pb.Empty{}, nil
		}

		conns = old.Connections
		log.Printf("old.Connections = %v", conns)
	}

	n := &model.Node{
		Name:         req.Config.Name,
		Comment:      req.Config.Comment,
		Multiplicity: req.Config.Multiplicity,
		Enabled:      req.Config.Enabled,
		Wait:         req.Config.Wait,
		Part:         part,
		X:            req.Config.X,
		Y:            req.Config.Y,
		Connections:  conns,
	}
	g.Nodes[req.Config.Name] = n
	n.RefreshConnections()
	g.RefreshChannelsPins() // Changing the part might have changed available pins.
	return &pb.Empty{}, nil
}

func (c *server) SetPosition(ctx context.Context, req *pb.SetPositionRequest) (*pb.Empty, error) {
	log.Printf("api: SetPosition(%s)", proto.MarshalTextString(req))
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Lock()
	defer g.Unlock()
	n, err := g.lookupNode(req.Node)
	if err != nil {
		return &pb.Empty{}, err
	}
	n.X, n.Y = req.X, req.Y
	return &pb.Empty{}, nil
}
