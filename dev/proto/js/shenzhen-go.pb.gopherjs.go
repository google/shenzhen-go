// Code generated by protoc-gen-gopherjs. DO NOT EDIT.
// source: shenzhen-go.proto

/*
	Package proto is a generated protocol buffer package.

	It is generated from these files:
		shenzhen-go.proto

	It has these top-level messages:
		Empty
		NodePin
		ChannelConfig
		NodeConfig
		ActionRequest
		ActionResponse
		Input
		Output
		SetChannelRequest
		SetGraphPropertiesRequest
		SetNodeRequest
		SetPositionRequest
*/
package proto

import jspb "github.com/johanbrandhorst/protobuf/jspb"

import (
	context "context"

	grpcweb "github.com/johanbrandhorst/protobuf/grpcweb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the jspb package it is being compiled against.
const _ = jspb.JspbPackageIsVersion2

type ActionRequest_Action int

const (
	ActionRequest_SAVE     ActionRequest_Action = 0
	ActionRequest_REVERT   ActionRequest_Action = 1
	ActionRequest_GENERATE ActionRequest_Action = 2
	ActionRequest_BUILD    ActionRequest_Action = 3
	ActionRequest_INSTALL  ActionRequest_Action = 4
)

var ActionRequest_Action_name = map[int]string{
	0: "SAVE",
	1: "REVERT",
	2: "GENERATE",
	3: "BUILD",
	4: "INSTALL",
}
var ActionRequest_Action_value = map[string]int{
	"SAVE":     0,
	"REVERT":   1,
	"GENERATE": 2,
	"BUILD":    3,
	"INSTALL":  4,
}

func (x ActionRequest_Action) String() string {
	return ActionRequest_Action_name[int(x)]
}

type Empty struct {
}

// MarshalToWriter marshals Empty to the provided writer.
func (m *Empty) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	return
}

// Marshal marshals Empty to a slice of bytes.
func (m *Empty) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a Empty from the provided reader.
func (m *Empty) UnmarshalFromReader(reader jspb.Reader) *Empty {
	for reader.Next() {
		if m == nil {
			m = &Empty{}
		}

		switch reader.GetFieldNumber() {
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a Empty from a slice of bytes.
func (m *Empty) Unmarshal(rawBytes []byte) (*Empty, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type NodePin struct {
	Node string
	Pin  string
}

// GetNode gets the Node of the NodePin.
func (m *NodePin) GetNode() (x string) {
	if m == nil {
		return x
	}
	return m.Node
}

// GetPin gets the Pin of the NodePin.
func (m *NodePin) GetPin() (x string) {
	if m == nil {
		return x
	}
	return m.Pin
}

// MarshalToWriter marshals NodePin to the provided writer.
func (m *NodePin) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Node) > 0 {
		writer.WriteString(1, m.Node)
	}

	if len(m.Pin) > 0 {
		writer.WriteString(2, m.Pin)
	}

	return
}

// Marshal marshals NodePin to a slice of bytes.
func (m *NodePin) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a NodePin from the provided reader.
func (m *NodePin) UnmarshalFromReader(reader jspb.Reader) *NodePin {
	for reader.Next() {
		if m == nil {
			m = &NodePin{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Node = reader.ReadString()
		case 2:
			m.Pin = reader.ReadString()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a NodePin from a slice of bytes.
func (m *NodePin) Unmarshal(rawBytes []byte) (*NodePin, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type ChannelConfig struct {
	Name string
	Cap  uint64
	Pins []*NodePin
}

// GetName gets the Name of the ChannelConfig.
func (m *ChannelConfig) GetName() (x string) {
	if m == nil {
		return x
	}
	return m.Name
}

// GetCap gets the Cap of the ChannelConfig.
func (m *ChannelConfig) GetCap() (x uint64) {
	if m == nil {
		return x
	}
	return m.Cap
}

// GetPins gets the Pins of the ChannelConfig.
func (m *ChannelConfig) GetPins() (x []*NodePin) {
	if m == nil {
		return x
	}
	return m.Pins
}

// MarshalToWriter marshals ChannelConfig to the provided writer.
func (m *ChannelConfig) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Name) > 0 {
		writer.WriteString(1, m.Name)
	}

	if m.Cap != 0 {
		writer.WriteUint64(2, m.Cap)
	}

	for _, msg := range m.Pins {
		writer.WriteMessage(3, func() {
			msg.MarshalToWriter(writer)
		})
	}

	return
}

// Marshal marshals ChannelConfig to a slice of bytes.
func (m *ChannelConfig) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a ChannelConfig from the provided reader.
func (m *ChannelConfig) UnmarshalFromReader(reader jspb.Reader) *ChannelConfig {
	for reader.Next() {
		if m == nil {
			m = &ChannelConfig{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Name = reader.ReadString()
		case 2:
			m.Cap = reader.ReadUint64()
		case 3:
			reader.ReadMessage(func() {
				m.Pins = append(m.Pins, new(NodePin).UnmarshalFromReader(reader))
			})
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a ChannelConfig from a slice of bytes.
func (m *ChannelConfig) Unmarshal(rawBytes []byte) (*ChannelConfig, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type NodeConfig struct {
	Name         string
	Enabled      bool
	Multiplicity uint32
	Wait         bool
	PartCfg      []byte
	PartType     string
	X            float64
	Y            float64
}

// GetName gets the Name of the NodeConfig.
func (m *NodeConfig) GetName() (x string) {
	if m == nil {
		return x
	}
	return m.Name
}

// GetEnabled gets the Enabled of the NodeConfig.
func (m *NodeConfig) GetEnabled() (x bool) {
	if m == nil {
		return x
	}
	return m.Enabled
}

// GetMultiplicity gets the Multiplicity of the NodeConfig.
func (m *NodeConfig) GetMultiplicity() (x uint32) {
	if m == nil {
		return x
	}
	return m.Multiplicity
}

// GetWait gets the Wait of the NodeConfig.
func (m *NodeConfig) GetWait() (x bool) {
	if m == nil {
		return x
	}
	return m.Wait
}

// GetPartCfg gets the PartCfg of the NodeConfig.
func (m *NodeConfig) GetPartCfg() (x []byte) {
	if m == nil {
		return x
	}
	return m.PartCfg
}

// GetPartType gets the PartType of the NodeConfig.
func (m *NodeConfig) GetPartType() (x string) {
	if m == nil {
		return x
	}
	return m.PartType
}

// GetX gets the X of the NodeConfig.
func (m *NodeConfig) GetX() (x float64) {
	if m == nil {
		return x
	}
	return m.X
}

// GetY gets the Y of the NodeConfig.
func (m *NodeConfig) GetY() (x float64) {
	if m == nil {
		return x
	}
	return m.Y
}

// MarshalToWriter marshals NodeConfig to the provided writer.
func (m *NodeConfig) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Name) > 0 {
		writer.WriteString(1, m.Name)
	}

	if m.Enabled {
		writer.WriteBool(2, m.Enabled)
	}

	if m.Multiplicity != 0 {
		writer.WriteUint32(3, m.Multiplicity)
	}

	if m.Wait {
		writer.WriteBool(4, m.Wait)
	}

	if len(m.PartCfg) > 0 {
		writer.WriteBytes(5, m.PartCfg)
	}

	if len(m.PartType) > 0 {
		writer.WriteString(6, m.PartType)
	}

	if m.X != 0 {
		writer.WriteFloat64(7, m.X)
	}

	if m.Y != 0 {
		writer.WriteFloat64(8, m.Y)
	}

	return
}

// Marshal marshals NodeConfig to a slice of bytes.
func (m *NodeConfig) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a NodeConfig from the provided reader.
func (m *NodeConfig) UnmarshalFromReader(reader jspb.Reader) *NodeConfig {
	for reader.Next() {
		if m == nil {
			m = &NodeConfig{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Name = reader.ReadString()
		case 2:
			m.Enabled = reader.ReadBool()
		case 3:
			m.Multiplicity = reader.ReadUint32()
		case 4:
			m.Wait = reader.ReadBool()
		case 5:
			m.PartCfg = reader.ReadBytes()
		case 6:
			m.PartType = reader.ReadString()
		case 7:
			m.X = reader.ReadFloat64()
		case 8:
			m.Y = reader.ReadFloat64()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a NodeConfig from a slice of bytes.
func (m *NodeConfig) Unmarshal(rawBytes []byte) (*NodeConfig, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type ActionRequest struct {
	Graph  string
	Action ActionRequest_Action
}

// GetGraph gets the Graph of the ActionRequest.
func (m *ActionRequest) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetAction gets the Action of the ActionRequest.
func (m *ActionRequest) GetAction() (x ActionRequest_Action) {
	if m == nil {
		return x
	}
	return m.Action
}

// MarshalToWriter marshals ActionRequest to the provided writer.
func (m *ActionRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if int(m.Action) != 0 {
		writer.WriteEnum(2, int(m.Action))
	}

	return
}

// Marshal marshals ActionRequest to a slice of bytes.
func (m *ActionRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a ActionRequest from the provided reader.
func (m *ActionRequest) UnmarshalFromReader(reader jspb.Reader) *ActionRequest {
	for reader.Next() {
		if m == nil {
			m = &ActionRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.Action = ActionRequest_Action(reader.ReadEnum())
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a ActionRequest from a slice of bytes.
func (m *ActionRequest) Unmarshal(rawBytes []byte) (*ActionRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type ActionResponse struct {
	Output string
}

// GetOutput gets the Output of the ActionResponse.
func (m *ActionResponse) GetOutput() (x string) {
	if m == nil {
		return x
	}
	return m.Output
}

// MarshalToWriter marshals ActionResponse to the provided writer.
func (m *ActionResponse) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Output) > 0 {
		writer.WriteString(1, m.Output)
	}

	return
}

// Marshal marshals ActionResponse to a slice of bytes.
func (m *ActionResponse) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a ActionResponse from the provided reader.
func (m *ActionResponse) UnmarshalFromReader(reader jspb.Reader) *ActionResponse {
	for reader.Next() {
		if m == nil {
			m = &ActionResponse{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Output = reader.ReadString()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a ActionResponse from a slice of bytes.
func (m *ActionResponse) Unmarshal(rawBytes []byte) (*ActionResponse, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type Input struct {
	Graph string
	In    string
}

// GetGraph gets the Graph of the Input.
func (m *Input) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetIn gets the In of the Input.
func (m *Input) GetIn() (x string) {
	if m == nil {
		return x
	}
	return m.In
}

// MarshalToWriter marshals Input to the provided writer.
func (m *Input) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if len(m.In) > 0 {
		writer.WriteString(2, m.In)
	}

	return
}

// Marshal marshals Input to a slice of bytes.
func (m *Input) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a Input from the provided reader.
func (m *Input) UnmarshalFromReader(reader jspb.Reader) *Input {
	for reader.Next() {
		if m == nil {
			m = &Input{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.In = reader.ReadString()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a Input from a slice of bytes.
func (m *Input) Unmarshal(rawBytes []byte) (*Input, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type Output struct {
	Out string
	Err string
}

// GetOut gets the Out of the Output.
func (m *Output) GetOut() (x string) {
	if m == nil {
		return x
	}
	return m.Out
}

// GetErr gets the Err of the Output.
func (m *Output) GetErr() (x string) {
	if m == nil {
		return x
	}
	return m.Err
}

// MarshalToWriter marshals Output to the provided writer.
func (m *Output) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Out) > 0 {
		writer.WriteString(1, m.Out)
	}

	if len(m.Err) > 0 {
		writer.WriteString(2, m.Err)
	}

	return
}

// Marshal marshals Output to a slice of bytes.
func (m *Output) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a Output from the provided reader.
func (m *Output) UnmarshalFromReader(reader jspb.Reader) *Output {
	for reader.Next() {
		if m == nil {
			m = &Output{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Out = reader.ReadString()
		case 2:
			m.Err = reader.ReadString()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a Output from a slice of bytes.
func (m *Output) Unmarshal(rawBytes []byte) (*Output, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type SetChannelRequest struct {
	Graph   string
	Channel string
	Config  *ChannelConfig
}

// GetGraph gets the Graph of the SetChannelRequest.
func (m *SetChannelRequest) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetChannel gets the Channel of the SetChannelRequest.
func (m *SetChannelRequest) GetChannel() (x string) {
	if m == nil {
		return x
	}
	return m.Channel
}

// GetConfig gets the Config of the SetChannelRequest.
func (m *SetChannelRequest) GetConfig() (x *ChannelConfig) {
	if m == nil {
		return x
	}
	return m.Config
}

// MarshalToWriter marshals SetChannelRequest to the provided writer.
func (m *SetChannelRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if len(m.Channel) > 0 {
		writer.WriteString(2, m.Channel)
	}

	if m.Config != nil {
		writer.WriteMessage(3, func() {
			m.Config.MarshalToWriter(writer)
		})
	}

	return
}

// Marshal marshals SetChannelRequest to a slice of bytes.
func (m *SetChannelRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a SetChannelRequest from the provided reader.
func (m *SetChannelRequest) UnmarshalFromReader(reader jspb.Reader) *SetChannelRequest {
	for reader.Next() {
		if m == nil {
			m = &SetChannelRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.Channel = reader.ReadString()
		case 3:
			reader.ReadMessage(func() {
				m.Config = m.Config.UnmarshalFromReader(reader)
			})
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a SetChannelRequest from a slice of bytes.
func (m *SetChannelRequest) Unmarshal(rawBytes []byte) (*SetChannelRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type SetGraphPropertiesRequest struct {
	Graph       string
	Name        string
	PackagePath string
	IsCommand   bool
}

// GetGraph gets the Graph of the SetGraphPropertiesRequest.
func (m *SetGraphPropertiesRequest) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetName gets the Name of the SetGraphPropertiesRequest.
func (m *SetGraphPropertiesRequest) GetName() (x string) {
	if m == nil {
		return x
	}
	return m.Name
}

// GetPackagePath gets the PackagePath of the SetGraphPropertiesRequest.
func (m *SetGraphPropertiesRequest) GetPackagePath() (x string) {
	if m == nil {
		return x
	}
	return m.PackagePath
}

// GetIsCommand gets the IsCommand of the SetGraphPropertiesRequest.
func (m *SetGraphPropertiesRequest) GetIsCommand() (x bool) {
	if m == nil {
		return x
	}
	return m.IsCommand
}

// MarshalToWriter marshals SetGraphPropertiesRequest to the provided writer.
func (m *SetGraphPropertiesRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if len(m.Name) > 0 {
		writer.WriteString(2, m.Name)
	}

	if len(m.PackagePath) > 0 {
		writer.WriteString(3, m.PackagePath)
	}

	if m.IsCommand {
		writer.WriteBool(4, m.IsCommand)
	}

	return
}

// Marshal marshals SetGraphPropertiesRequest to a slice of bytes.
func (m *SetGraphPropertiesRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a SetGraphPropertiesRequest from the provided reader.
func (m *SetGraphPropertiesRequest) UnmarshalFromReader(reader jspb.Reader) *SetGraphPropertiesRequest {
	for reader.Next() {
		if m == nil {
			m = &SetGraphPropertiesRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.Name = reader.ReadString()
		case 3:
			m.PackagePath = reader.ReadString()
		case 4:
			m.IsCommand = reader.ReadBool()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a SetGraphPropertiesRequest from a slice of bytes.
func (m *SetGraphPropertiesRequest) Unmarshal(rawBytes []byte) (*SetGraphPropertiesRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type SetNodeRequest struct {
	Graph  string
	Node   string
	Config *NodeConfig
}

// GetGraph gets the Graph of the SetNodeRequest.
func (m *SetNodeRequest) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetNode gets the Node of the SetNodeRequest.
func (m *SetNodeRequest) GetNode() (x string) {
	if m == nil {
		return x
	}
	return m.Node
}

// GetConfig gets the Config of the SetNodeRequest.
func (m *SetNodeRequest) GetConfig() (x *NodeConfig) {
	if m == nil {
		return x
	}
	return m.Config
}

// MarshalToWriter marshals SetNodeRequest to the provided writer.
func (m *SetNodeRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if len(m.Node) > 0 {
		writer.WriteString(2, m.Node)
	}

	if m.Config != nil {
		writer.WriteMessage(3, func() {
			m.Config.MarshalToWriter(writer)
		})
	}

	return
}

// Marshal marshals SetNodeRequest to a slice of bytes.
func (m *SetNodeRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a SetNodeRequest from the provided reader.
func (m *SetNodeRequest) UnmarshalFromReader(reader jspb.Reader) *SetNodeRequest {
	for reader.Next() {
		if m == nil {
			m = &SetNodeRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.Node = reader.ReadString()
		case 3:
			reader.ReadMessage(func() {
				m.Config = m.Config.UnmarshalFromReader(reader)
			})
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a SetNodeRequest from a slice of bytes.
func (m *SetNodeRequest) Unmarshal(rawBytes []byte) (*SetNodeRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type SetPositionRequest struct {
	Graph string
	Node  string
	X     float64
	Y     float64
}

// GetGraph gets the Graph of the SetPositionRequest.
func (m *SetPositionRequest) GetGraph() (x string) {
	if m == nil {
		return x
	}
	return m.Graph
}

// GetNode gets the Node of the SetPositionRequest.
func (m *SetPositionRequest) GetNode() (x string) {
	if m == nil {
		return x
	}
	return m.Node
}

// GetX gets the X of the SetPositionRequest.
func (m *SetPositionRequest) GetX() (x float64) {
	if m == nil {
		return x
	}
	return m.X
}

// GetY gets the Y of the SetPositionRequest.
func (m *SetPositionRequest) GetY() (x float64) {
	if m == nil {
		return x
	}
	return m.Y
}

// MarshalToWriter marshals SetPositionRequest to the provided writer.
func (m *SetPositionRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Graph) > 0 {
		writer.WriteString(1, m.Graph)
	}

	if len(m.Node) > 0 {
		writer.WriteString(2, m.Node)
	}

	if m.X != 0 {
		writer.WriteFloat64(3, m.X)
	}

	if m.Y != 0 {
		writer.WriteFloat64(4, m.Y)
	}

	return
}

// Marshal marshals SetPositionRequest to a slice of bytes.
func (m *SetPositionRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a SetPositionRequest from the provided reader.
func (m *SetPositionRequest) UnmarshalFromReader(reader jspb.Reader) *SetPositionRequest {
	for reader.Next() {
		if m == nil {
			m = &SetPositionRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Graph = reader.ReadString()
		case 2:
			m.Node = reader.ReadString()
		case 3:
			m.X = reader.ReadFloat64()
		case 4:
			m.Y = reader.ReadFloat64()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a SetPositionRequest from a slice of bytes.
func (m *SetPositionRequest) Unmarshal(rawBytes []byte) (*SetPositionRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpcweb.Client

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpcweb package it is being compiled against.
const _ = grpcweb.GrpcWebPackageIsVersion3

// Client API for ShenzhenGo service

type ShenzhenGoClient interface {
	// Action performs an action (save, run, etc).
	Action(ctx context.Context, in *ActionRequest, opts ...grpcweb.CallOption) (*ActionResponse, error)
	// Run runs the program.
	Run(ctx context.Context, opts ...grpcweb.CallOption) (ShenzhenGo_RunClient, error)
	// SetNode either creates a new channel (name == "", config != nil)
	// changes existing channel data such as name and attached pins (name is found, config != nil),
	// or deletes a channel (name is found, config == nil).
	SetChannel(ctx context.Context, in *SetChannelRequest, opts ...grpcweb.CallOption) (*Empty, error)
	// SetGraphProperties changes metdata such as name and package path.
	SetGraphProperties(ctx context.Context, in *SetGraphPropertiesRequest, opts ...grpcweb.CallOption) (*Empty, error)
	// SetNode either creates a new node (name == "", config != nil)
	// changes existing node such as name and multiplicity (name is found, config != nil),
	// or deletes a node (name is found, config == nil).
	SetNode(ctx context.Context, in *SetNodeRequest, opts ...grpcweb.CallOption) (*Empty, error)
	// SetPosition changes the node position in the diagram.
	SetPosition(ctx context.Context, in *SetPositionRequest, opts ...grpcweb.CallOption) (*Empty, error)
}

type shenzhenGoClient struct {
	client *grpcweb.Client
}

// NewShenzhenGoClient creates a new gRPC-Web client.
func NewShenzhenGoClient(hostname string, opts ...grpcweb.DialOption) ShenzhenGoClient {
	return &shenzhenGoClient{
		client: grpcweb.NewClient(hostname, "proto.ShenzhenGo", opts...),
	}
}

func (c *shenzhenGoClient) Action(ctx context.Context, in *ActionRequest, opts ...grpcweb.CallOption) (*ActionResponse, error) {
	resp, err := c.client.RPCCall(ctx, "Action", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(ActionResponse).Unmarshal(resp)
}

func (c *shenzhenGoClient) Run(ctx context.Context, opts ...grpcweb.CallOption) (ShenzhenGo_RunClient, error) {
	srv, err := c.client.NewClientStream(ctx, true, true, "Run", opts...)
	if err != nil {
		return nil, err
	}

	return &shenzhenGoRunClient{srv}, nil
}

type ShenzhenGo_RunClient interface {
	Send(*Input) error
	Recv() (*Output, error)
	grpcweb.ClientStream
}

type shenzhenGoRunClient struct {
	grpcweb.ClientStream
}

func (x *shenzhenGoRunClient) Send(req *Input) error {
	return x.SendMsg(req.Marshal())
}

func (x *shenzhenGoRunClient) Recv() (*Output, error) {
	resp, err := x.RecvMsg()
	if err != nil {
		return nil, err
	}

	return new(Output).Unmarshal(resp)
}

func (c *shenzhenGoClient) SetChannel(ctx context.Context, in *SetChannelRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	resp, err := c.client.RPCCall(ctx, "SetChannel", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(Empty).Unmarshal(resp)
}

func (c *shenzhenGoClient) SetGraphProperties(ctx context.Context, in *SetGraphPropertiesRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	resp, err := c.client.RPCCall(ctx, "SetGraphProperties", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(Empty).Unmarshal(resp)
}

func (c *shenzhenGoClient) SetNode(ctx context.Context, in *SetNodeRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	resp, err := c.client.RPCCall(ctx, "SetNode", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(Empty).Unmarshal(resp)
}

func (c *shenzhenGoClient) SetPosition(ctx context.Context, in *SetPositionRequest, opts ...grpcweb.CallOption) (*Empty, error) {
	resp, err := c.client.RPCCall(ctx, "SetPosition", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(Empty).Unmarshal(resp)
}
