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
	"log"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/go"
)

func (c *server) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	log.Printf("api: Save(%s)", proto.MarshalTextString(req))
	g, err := c.lookupGraph(req.Graph)
	if err != nil {
		return &pb.Empty{}, err
	}
	g.Lock()
	defer g.Unlock()
	return &pb.Empty{}, SaveJSONFile(g.Graph)
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
		Name:      req.Config.Name,
		Type:      req.Config.Type,
		Anonymous: req.Config.Anon,
		Capacity:  int(req.Config.Cap),
		Pins:      nps,
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

	if req.Node != "" {
		old, err := g.lookupNode(req.Node)
		if err != nil {
			return &pb.Empty{}, err
		}

		g.DeleteNode(old)

		if req.Config == nil {
			// Deletion was intended, job complete.
			return &pb.Empty{}, nil
		}
	}

	n := &model.Node{
		Name:         req.Config.Name,
		Multiplicity: uint(req.Config.Multiplicity),
		Enabled:      req.Config.Enabled,
		Wait:         req.Config.Wait,
		Part:         part,
		X:            req.Config.X,
		Y:            req.Config.Y,
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
