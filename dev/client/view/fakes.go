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

package view

import (
	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/dom"
)

type fakeGraphController struct{}

func (c fakeGraphController) GainFocus()                   {}
func (c fakeGraphController) LoseFocus()                   {}
func (c fakeGraphController) Nodes(f func(NodeController)) { f(fakeNodeController{}) }

func (c fakeGraphController) Node(name string) NodeController {
	if name != "Node 1" {
		return nil
	}
	return fakeNodeController{}
}

func (c fakeGraphController) NumNodes() int                         { return 1 }
func (c fakeGraphController) Channel(name string) ChannelController { return nil }
func (c fakeGraphController) Channels(f func(ChannelController))    {}
func (c fakeGraphController) NumChannels() int                      { return 0 }

func (c fakeGraphController) CreateChannel(...PinController) (ChannelController, error) {
	return nil, nil
}

func (c fakeGraphController) CreateNode(ctx context.Context, partType string) (NodeController, error) {
	return nil, nil
}

func (c fakeGraphController) Save(ctx context.Context) error           { return nil }
func (c fakeGraphController) SaveProperties(ctx context.Context) error { return nil }

type fakeNodeController struct{}

func (f fakeNodeController) Name() string             { return "Node 1" }
func (f fakeNodeController) Position() (x, y float64) { return 150, 150 }

func (f fakeNodeController) Pins(x func(PinController, string)) {
	x(fakePinController("input"), "nil")
	x(fakePinController("output"), "nil")
	x(fakePinController("output 2"), "nil")
}

func (f fakeNodeController) GainFocus() {}
func (f fakeNodeController) LoseFocus() {}

func (f fakeNodeController) Delete(context.Context) error                        { return nil }
func (f fakeNodeController) Save(context.Context) error                          { return nil }
func (f fakeNodeController) SetPosition(context.Context, float64, float64) error { return nil }
func (f fakeNodeController) ShowMetadataSubpanel()                               {}
func (f fakeNodeController) ShowPartSubpanel(string)                             {}

type fakePinController string

func (f fakePinController) Name() string  { return string(f) }
func (f fakePinController) Type() string  { return "int" }
func (f fakePinController) IsInput() bool { return f == "input" }

func (f fakePinController) Attach(ctx context.Context, cc ChannelController) error { return nil }
func (f fakePinController) Detach(ctx context.Context) error                       { return nil }

func makeFakeView() *View {
	doc := dom.MakeFakeDocument()
	v := &View{
		doc:     doc,
		diagram: doc.MakeSVGElement("svg"),
	}
	v.graph = &Graph{
		view: v,
		gc:   fakeGraphController{},
		doc:  doc,
	}
	v.graph.makeElements(doc, v.diagram)
	v.graph.refresh()
	return v
}