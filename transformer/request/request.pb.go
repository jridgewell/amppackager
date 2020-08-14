// Code generated by protoc-gen-go. DO NOT EDIT.
// source: transformer/request/request.proto

package request

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// This should be kept in sync with HtmlFormat.Code in
// github.com/ampproject/amphtml/validator/validator.proto.
// Deprecated fields, do not reuse:
// reserved 5;
type Request_HtmlFormat int32

const (
	Request_UNKNOWN_CODE Request_HtmlFormat = 0
	Request_AMP          Request_HtmlFormat = 1
	Request_AMP4ADS      Request_HtmlFormat = 2
	Request_AMP4EMAIL    Request_HtmlFormat = 3
	Request_EXPERIMENTAL Request_HtmlFormat = 4
)

var Request_HtmlFormat_name = map[int32]string{
	0: "UNKNOWN_CODE",
	1: "AMP",
	2: "AMP4ADS",
	3: "AMP4EMAIL",
	4: "EXPERIMENTAL",
}

var Request_HtmlFormat_value = map[string]int32{
	"UNKNOWN_CODE": 0,
	"AMP":          1,
	"AMP4ADS":      2,
	"AMP4EMAIL":    3,
	"EXPERIMENTAL": 4,
}

func (x Request_HtmlFormat) String() string {
	return proto.EnumName(Request_HtmlFormat_name, int32(x))
}

func (Request_HtmlFormat) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{0, 0}
}

type Request_TransformersConfig int32

const (
	// Execute the default list of transformers. For packager production
	// environments, this should be the config used.
	Request_DEFAULT Request_TransformersConfig = 0
	// Execute none, and simply parse and re-emit. Some normalization will be
	// performed regardless, including, but not limited to:
	// - HTML normalization (e.g. closing all non-void tags).
	// - removal of all comments
	// - lowercase-ing of attribute keys
	// - lexical sort of attribute keys
	// - text is escaped
	//
	// WARNING. THIS IS FOR TESTING PURPOSES ONLY.
	// Use of this setting in a packager production environment could produce
	// invalid transformed AMP when ingested by AMP caches.
	Request_NONE Request_TransformersConfig = 1
	// Execute the minimum needed for verification/validation.
	//
	// WARNING. FOR AMP CACHE USE ONLY.
	// Use of this setting in a packager production environment could produce
	// invalid transformed AMP when ingested by AMP caches.
	Request_VALIDATION Request_TransformersConfig = 2
	// Execute a custom set of transformers.
	//
	// WARNING. THIS IS FOR TESTING PURPOSES ONLY.
	// Use of this setting in a packager production environment could produce
	// invalid transformed AMP when ingested by AMP caches.
	Request_CUSTOM Request_TransformersConfig = 3
)

var Request_TransformersConfig_name = map[int32]string{
	0: "DEFAULT",
	1: "NONE",
	2: "VALIDATION",
	3: "CUSTOM",
}

var Request_TransformersConfig_value = map[string]int32{
	"DEFAULT":    0,
	"NONE":       1,
	"VALIDATION": 2,
	"CUSTOM":     3,
}

func (x Request_TransformersConfig) String() string {
	return proto.EnumName(Request_TransformersConfig_name, int32(x))
}

func (Request_TransformersConfig) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{0, 1}
}

// A Request encapsulates input and contextual parameters for use by the
// transformers.
type Request struct {
	// The AMP HTML document to transform.
	Html string `protobuf:"bytes,1,opt,name=html,proto3" json:"html,omitempty"`
	// The public URL of the document, i.e. the location that should appear in
	// the browser URL bar.
	DocumentUrl string `protobuf:"bytes,2,opt,name=document_url,json=documentUrl,proto3" json:"document_url,omitempty"`
	// The AMP runtime version.
	Rtv string `protobuf:"bytes,4,opt,name=rtv,proto3" json:"rtv,omitempty"`
	// The CSS contents to inline into the transformed HTML
	Css string `protobuf:"bytes,5,opt,name=css,proto3" json:"css,omitempty"`
	// Transformations are only run if the HTML tag contains the attribute
	// specifying one of the provided formats. If allowed_formats is empty, then
	// all non-experimental AMP formats are allowed.
	AllowedFormats []Request_HtmlFormat       `protobuf:"varint,7,rep,packed,name=allowed_formats,json=allowedFormats,proto3,enum=amp.transform.Request_HtmlFormat" json:"allowed_formats,omitempty"`
	Config         Request_TransformersConfig `protobuf:"varint,6,opt,name=config,proto3,enum=amp.transform.Request_TransformersConfig" json:"config,omitempty"`
	// If config == CUSTOM, this is the list of custom transformers to execute,
	// in the order provided here. Otherwise, this is ignored.
	Transformers []string `protobuf:"bytes,3,rep,name=transformers,proto3" json:"transformers,omitempty"`
	// The version of the transforms to perform (optional). If specified, it must
	// be a supported version.
	Version              int64    `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetHtml() string {
	if m != nil {
		return m.Html
	}
	return ""
}

func (m *Request) GetDocumentUrl() string {
	if m != nil {
		return m.DocumentUrl
	}
	return ""
}

func (m *Request) GetRtv() string {
	if m != nil {
		return m.Rtv
	}
	return ""
}

func (m *Request) GetCss() string {
	if m != nil {
		return m.Css
	}
	return ""
}

func (m *Request) GetAllowedFormats() []Request_HtmlFormat {
	if m != nil {
		return m.AllowedFormats
	}
	return nil
}

func (m *Request) GetConfig() Request_TransformersConfig {
	if m != nil {
		return m.Config
	}
	return Request_DEFAULT
}

func (m *Request) GetTransformers() []string {
	if m != nil {
		return m.Transformers
	}
	return nil
}

func (m *Request) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

// An inclusive range of version numbers.
type VersionRange struct {
	Min                  int64    `protobuf:"varint,1,opt,name=min,proto3" json:"min,omitempty"`
	Max                  int64    `protobuf:"varint,2,opt,name=max,proto3" json:"max,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VersionRange) Reset()         { *m = VersionRange{} }
func (m *VersionRange) String() string { return proto.CompactTextString(m) }
func (*VersionRange) ProtoMessage()    {}
func (*VersionRange) Descriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{1}
}

func (m *VersionRange) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VersionRange.Unmarshal(m, b)
}
func (m *VersionRange) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VersionRange.Marshal(b, m, deterministic)
}
func (m *VersionRange) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionRange.Merge(m, src)
}
func (m *VersionRange) XXX_Size() int {
	return xxx_messageInfo_VersionRange.Size(m)
}
func (m *VersionRange) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionRange.DiscardUnknown(m)
}

var xxx_messageInfo_VersionRange proto.InternalMessageInfo

func (m *VersionRange) GetMin() int64 {
	if m != nil {
		return m.Min
	}
	return 0
}

func (m *VersionRange) GetMax() int64 {
	if m != nil {
		return m.Max
	}
	return 0
}

// A Metadata is part of the transformers' response, and includes additional
// information either not present in or not easily accessible from the HTML. It
// should remain relatively small, as it undergoes a
// serialization/deserialization round-trip when the Go library is called from
// C.
type Metadata struct {
	// Absolute URLs of resources that should be preloaded when the AMP is
	// prefetched. In a signed exchange (SXG) context, these would be included as
	// `Link: rel=preload` headers, as these are used by the browser during SXG
	// prefetch:
	// https://github.com/WICG/webpackage/blob/master/explainer.md#prefetching-stops-here
	Preloads []*Metadata_Preload `protobuf:"bytes,1,rep,name=preloads,proto3" json:"preloads,omitempty"`
	// Recommended validity duration (`expires - date`), in seconds, of the SXG,
	// based on the content being signed. In particular, JS is given a shorter
	// lifetime to reduce risk of issues due to downgrades:
	// https://wicg.github.io/webpackage/draft-yasskin-http-origin-signed-responses.html#seccons-downgrades.
	MaxAgeSecs           int32    `protobuf:"varint,2,opt,name=max_age_secs,json=maxAgeSecs,proto3" json:"max_age_secs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}
func (*Metadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{2}
}

func (m *Metadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metadata.Unmarshal(m, b)
}
func (m *Metadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metadata.Marshal(b, m, deterministic)
}
func (m *Metadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata.Merge(m, src)
}
func (m *Metadata) XXX_Size() int {
	return xxx_messageInfo_Metadata.Size(m)
}
func (m *Metadata) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata proto.InternalMessageInfo

func (m *Metadata) GetPreloads() []*Metadata_Preload {
	if m != nil {
		return m.Preloads
	}
	return nil
}

func (m *Metadata) GetMaxAgeSecs() int32 {
	if m != nil {
		return m.MaxAgeSecs
	}
	return 0
}

type Metadata_Preload struct {
	// The URL of the resource to preload. Will be an absolute URL on the domain
	// of the target AMP cache.
	Url string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	// The `as` attribute of the preload, as specified in
	// https://w3c.github.io/preload/#as-attribute and
	// https://html.spec.whatwg.org/multipage/semantics.html#attr-link-as. The
	// full list of potential values is specified in
	// https://fetch.spec.whatwg.org/#concept-request-destination, though for
	// the time being only "script", "style", and "image" are allowed.
	As string `protobuf:"bytes,2,opt,name=as,proto3" json:"as,omitempty"`
	// The media attribute for image preload link. This attribute is useful
	// only for image links.
	Media                string   `protobuf:"bytes,3,opt,name=media,proto3" json:"media,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metadata_Preload) Reset()         { *m = Metadata_Preload{} }
func (m *Metadata_Preload) String() string { return proto.CompactTextString(m) }
func (*Metadata_Preload) ProtoMessage()    {}
func (*Metadata_Preload) Descriptor() ([]byte, []int) {
	return fileDescriptor_762cce2ac5f73405, []int{2, 0}
}

func (m *Metadata_Preload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metadata_Preload.Unmarshal(m, b)
}
func (m *Metadata_Preload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metadata_Preload.Marshal(b, m, deterministic)
}
func (m *Metadata_Preload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata_Preload.Merge(m, src)
}
func (m *Metadata_Preload) XXX_Size() int {
	return xxx_messageInfo_Metadata_Preload.Size(m)
}
func (m *Metadata_Preload) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata_Preload.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata_Preload proto.InternalMessageInfo

func (m *Metadata_Preload) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *Metadata_Preload) GetAs() string {
	if m != nil {
		return m.As
	}
	return ""
}

func (m *Metadata_Preload) GetMedia() string {
	if m != nil {
		return m.Media
	}
	return ""
}

func init() {
	proto.RegisterEnum("amp.transform.Request_HtmlFormat", Request_HtmlFormat_name, Request_HtmlFormat_value)
	proto.RegisterEnum("amp.transform.Request_TransformersConfig", Request_TransformersConfig_name, Request_TransformersConfig_value)
	proto.RegisterType((*Request)(nil), "amp.transform.Request")
	proto.RegisterType((*VersionRange)(nil), "amp.transform.VersionRange")
	proto.RegisterType((*Metadata)(nil), "amp.transform.Metadata")
	proto.RegisterType((*Metadata_Preload)(nil), "amp.transform.Metadata.Preload")
}

func init() {
	proto.RegisterFile("transformer/request/request.proto", fileDescriptor_762cce2ac5f73405)
}

var fileDescriptor_762cce2ac5f73405 = []byte{
	// 519 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0x5d, 0x93, 0xd2, 0x30,
	0x14, 0xdd, 0x12, 0xa0, 0x70, 0x61, 0xb1, 0x93, 0xf1, 0xa1, 0xe3, 0x8b, 0xa5, 0x4f, 0xf5, 0xa5,
	0xcc, 0xa0, 0x8e, 0x0f, 0x3e, 0x55, 0xe8, 0x2a, 0x4a, 0x0b, 0x53, 0x60, 0x75, 0x7c, 0x61, 0xb2,
	0x25, 0xdb, 0x45, 0x9b, 0x16, 0x93, 0xb0, 0xf2, 0xa3, 0xfc, 0x33, 0xfe, 0x23, 0x27, 0x29, 0xec,
	0xba, 0x7e, 0x3c, 0xf5, 0xdc, 0xd3, 0x73, 0x92, 0x33, 0xf7, 0xde, 0x40, 0x5f, 0x72, 0x52, 0x88,
	0xeb, 0x92, 0x33, 0xca, 0x07, 0x9c, 0x7e, 0xdb, 0x53, 0x21, 0x4f, 0x5f, 0x7f, 0xc7, 0x4b, 0x59,
	0xe2, 0x73, 0xc2, 0x76, 0xfe, 0x9d, 0xcc, 0xfd, 0x89, 0xc0, 0x4c, 0x2a, 0x01, 0xc6, 0x50, 0xbf,
	0x91, 0x2c, 0xb7, 0x0d, 0xc7, 0xf0, 0xda, 0x89, 0xc6, 0xb8, 0x0f, 0xdd, 0x4d, 0x99, 0xee, 0x19,
	0x2d, 0xe4, 0x7a, 0xcf, 0x73, 0xbb, 0xa6, 0xff, 0x75, 0x4e, 0xdc, 0x8a, 0xe7, 0xd8, 0x02, 0xc4,
	0xe5, 0xad, 0x5d, 0xd7, 0x7f, 0x14, 0x54, 0x4c, 0x2a, 0x84, 0xdd, 0xa8, 0x98, 0x54, 0x08, 0xfc,
	0x1e, 0x1e, 0x91, 0x3c, 0x2f, 0xbf, 0xd3, 0xcd, 0x5a, 0x5d, 0x4b, 0xa4, 0xb0, 0x4d, 0x07, 0x79,
	0xbd, 0x61, 0xdf, 0x7f, 0x90, 0xc7, 0x3f, 0x66, 0xf1, 0xdf, 0x49, 0x96, 0x5f, 0x68, 0x65, 0xd2,
	0x3b, 0x3a, 0xab, 0x52, 0xe0, 0x00, 0x9a, 0x69, 0x59, 0x5c, 0x6f, 0x33, 0xbb, 0xe9, 0x18, 0x5e,
	0x6f, 0xf8, 0xec, 0x3f, 0x47, 0x2c, 0xef, 0x7b, 0x21, 0x46, 0xda, 0x90, 0x1c, 0x8d, 0xd8, 0x85,
	0xee, 0x6f, 0x9d, 0x12, 0x36, 0x72, 0x90, 0xd7, 0x4e, 0x1e, 0x70, 0xd8, 0x06, 0xf3, 0x96, 0x72,
	0xb1, 0x2d, 0x0b, 0xbb, 0xe5, 0x18, 0x1e, 0x4a, 0x4e, 0xa5, 0xbb, 0x02, 0xb8, 0x8f, 0x87, 0x2d,
	0xe8, 0xae, 0xe2, 0x0f, 0xf1, 0xec, 0x63, 0xbc, 0x1e, 0xcd, 0xc6, 0xa1, 0x75, 0x86, 0x4d, 0x40,
	0x41, 0x34, 0xb7, 0x0c, 0xdc, 0x01, 0x33, 0x88, 0xe6, 0x2f, 0x82, 0xf1, 0xc2, 0xaa, 0xe1, 0x73,
	0x68, 0xab, 0x22, 0x8c, 0x82, 0xc9, 0xd4, 0x42, 0xca, 0x16, 0x7e, 0x9a, 0x87, 0xc9, 0x24, 0x0a,
	0xe3, 0x65, 0x30, 0xb5, 0xea, 0xee, 0x5b, 0xc0, 0x7f, 0x47, 0x56, 0x67, 0x8c, 0xc3, 0x8b, 0x60,
	0x35, 0x5d, 0x5a, 0x67, 0xb8, 0x05, 0xf5, 0x78, 0x16, 0x87, 0x96, 0x81, 0x7b, 0x00, 0x97, 0xc1,
	0x74, 0x32, 0x0e, 0x96, 0x93, 0x59, 0x6c, 0xd5, 0x30, 0x40, 0x73, 0xb4, 0x5a, 0x2c, 0x67, 0x91,
	0x85, 0xdc, 0x21, 0x74, 0x2f, 0xab, 0xa8, 0x09, 0x29, 0x32, 0xaa, 0xc6, 0xc1, 0xb6, 0x85, 0x1e,
	0x2b, 0x4a, 0x14, 0xd4, 0x0c, 0x39, 0xe8, 0x61, 0x2a, 0x86, 0x1c, 0xdc, 0x1f, 0x06, 0xb4, 0x22,
	0x2a, 0xc9, 0x86, 0x48, 0x82, 0x5f, 0x43, 0x6b, 0xc7, 0x69, 0x5e, 0x92, 0x8d, 0xb0, 0x0d, 0x07,
	0x79, 0x9d, 0xe1, 0xd3, 0x3f, 0x7a, 0x7c, 0x92, 0xfa, 0xf3, 0x4a, 0x97, 0xdc, 0x19, 0xb0, 0x03,
	0x5d, 0x46, 0x0e, 0x6b, 0x92, 0xd1, 0xb5, 0xa0, 0xa9, 0xd0, 0x97, 0x34, 0x12, 0x60, 0xe4, 0x10,
	0x64, 0x74, 0x41, 0x53, 0xf1, 0x24, 0x00, 0xf3, 0x68, 0x53, 0x41, 0xd4, 0x56, 0x55, 0x1b, 0xa7,
	0x20, 0xee, 0x41, 0x8d, 0x88, 0xe3, 0x9a, 0xd5, 0x88, 0xc0, 0x8f, 0xa1, 0xc1, 0xe8, 0x66, 0x4b,
	0x6c, 0xa4, 0xa9, 0xaa, 0x78, 0xf3, 0xea, 0xf3, 0xcb, 0x6c, 0x2b, 0x6f, 0xf6, 0x57, 0x7e, 0x5a,
	0xb2, 0x01, 0x61, 0xbb, 0x1d, 0x2f, 0xbf, 0xd0, 0x54, 0x6a, 0x48, 0xd2, 0xaf, 0x24, 0xa3, 0x7c,
	0xf0, 0x8f, 0xc7, 0x70, 0xd5, 0xd4, 0xaf, 0xe0, 0xf9, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x59,
	0xb7, 0x03, 0x86, 0x2a, 0x03, 0x00, 0x00,
}
