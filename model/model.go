package model

// UANodeClass UANodeClass
type UANodeClass int

// GetType 取得種類
func (c UANodeClass) GetType() string {
	switch c {
	case UANodeClassUnspecified:
		return "UANodeClass_UNSPECIFIED"
	case UANodeClassObject:
		return "UA_NODECLASS_OBJECT"
	case UANodeClassVariable:
		return "UA_NODECLASS_VARIABLE"
	case UANodeClassMethod:
		return "UA_NODECLASS_METHOD"
	case UANodeClassObjectType:
		return "UA_NODECLASS_OBJECTTYPE"
	case UANodeClassVariableType:
		return "UA_NODECLASS_VARIABLETYPE"
	case UANodeClassReferenceType:
		return "UA_NODECLASS_REFERENCETYPE"
	case UANodeClassDataType:
		return "UA_NODECLASS_DATATYPE"
	case UANodeClassView:
		return "UA_NODECLASS_VIEW"
	}
	return "undefined"
}

const (
	// UANodeClassUnspecified UANodeClass_UNSPECIFIED
	UANodeClassUnspecified UANodeClass = iota
	// UANodeClassObject UA_NODECLASS_OBJECT
	UANodeClassObject UANodeClass = UANodeClassUnspecified + 1
	// UANodeClassVariable UA_NODECLASS_VARIABLE
	UANodeClassVariable UANodeClass = UANodeClassObject * 2
	// UANodeClassMethod UA_NODECLASS_METHOD
	UANodeClassMethod UANodeClass = UANodeClassVariable * 2
	// UANodeClassObjectType UA_NODECLASS_OBJECTTYPE
	UANodeClassObjectType UANodeClass = UANodeClassMethod * 2
	// UANodeClassVariableType UA_NODECLASS_VARIABLETYPE
	UANodeClassVariableType UANodeClass = UANodeClassObjectType * 2
	// UANodeClassReferenceType UA_NODECLASS_REFERENCETYPE
	UANodeClassReferenceType UANodeClass = UANodeClassVariableType * 2
	// UANodeClassDataType UA_NODECLASS_DATATYPE
	UANodeClassDataType UANodeClass = UANodeClassReferenceType * 2
	// UANodeClassView UA_NODECLASS_VIEW
	UANodeClassView UANodeClass = UANodeClassDataType * 2
)

// UABrowseResult UABrowseResult
type UABrowseResult struct {
	ContinuationPoint string                   `json:"continuationPoint"`
	References        []UAReferenceDescription `json:"references"`
	StatusCode        UAStatusCode             `json:"statusCode"`
}

// UAStatusCode UAStatusCode
type UAStatusCode int

// UAReferenceDescription UAReferenceDescription
type UAReferenceDescription struct {
	BrowseName      string           `json:"browseName"`
	DisplayName     UALocalizedText  `json:"displayName"`
	IsForward       bool             `json:"isForward"`
	NodeClass       UANodeClass      `json:"nodeClass"`
	NodeID          UAExpandedNodeID `json:"nodeId"`
	ReferenceTypeID UANodeID         `json:"referenceTypeId"`
	TypeDefinition  UAExpandedNodeID `json:"TypeDefinition"`
}

// UAString UAString
type UAString struct {
	Data   string `json:"data"`
	Length int32  `json:"length"`
}

// UAQualifiedName UAQualifiedName
type UAQualifiedName struct {
	Name           UAString `json:"name"`
	NamespaceIndex uint16   `json:"namespaceIndex"`
}

// UALocalizedText UALocalizedText
type UALocalizedText struct {
	Locale string `json:"locale"`
	Text   string `json:"text"`
}

// UAExpandedNodeID UAExpandedNodeID
type UAExpandedNodeID struct {
	NamespaceURI string `json:"namespaceUri"`
	NodeID       NodeID `json:"nodeId"`
	ServerIndex  uint32 `json:"serverIndex"`
}

// HDAMsg HDAMsg
type HDAMsg struct {
	NamespaceURI string `json:"namespaceUri"`
	NodeID       NodeID `json:"nodeId"`
	ServerIndex  uint32 `json:"serverIndex"`
}

// NodeID NodeID
type NodeID struct {
	NamespaceIndex uint16
	IdentifierType UANodeID
	Identifier     Identifier
}

// GetType 取得種類
func (c UANodeID) GetType() string {

	switch c {
	case UANodeIDTypeNumericUANodeID:
		return "UA_NODEIDTYPE_NUMERIC"
	case UANodeIDTypeStringUANodeID:
		return "UA_NODEIDTYPE_STRING"
	case UANodeIDTypeGUIDUANodeID:
		return "UA_NODEIDTYPE_GUID"
	case UANodeIDTypeByteStringUANodeID:
		return "UA_NODEIDTYPE_BYTESTRING"
	}
	return "undefined"
}

// UANodeID UANodeID
type UANodeID uint16

const (
	// UANodeIDTypeNumericUANodeID UA_NODEIDTYPE_NUMERIC
	UANodeIDTypeNumericUANodeID UANodeID = 0
	// UANodeIDTypeStringUANodeID UA_NODEIDTYPE_STRING
	UANodeIDTypeStringUANodeID UANodeID = 3
	// UANodeIDTypeGUIDUANodeID UA_NODEIDTYPE_GUID
	UANodeIDTypeGUIDUANodeID UANodeID = 4
	// UANodeIDTypeByteStringUANodeID UA_NODEIDTYPE_BYTESTRING
	UANodeIDTypeByteStringUANodeID UANodeID = 5
)

// Identifier Identifier
type Identifier struct {
	Numeric uint32
	String  string
}
