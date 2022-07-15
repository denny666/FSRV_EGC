package opcuaservice

/*
#cgo windows CFLAGS: -DWEBVIEW_WINAPI=1 -std=c99
#cgo windows,amd64 LDFLAGS: -lWS2_32
#cgo CFLAGS: -I ${SRCDIR}/lib
#cgo CFLAGS: -DPNG_DEBUG=1
#pragma comment(lib, "ws2_32.lib")

#include "open62541.h"
#include "open62541.c"

#define ADMIN "admin"
#define TEST "test"
static void UA_ServerConfig_set_customHostname(UA_ServerConfig *config, const UA_String customHostname) {
    if(!config)
        return;
    UA_String_deleteMembers(&config->customHostname);
    UA_String_copy(&customHostname, &config->customHostname);
}

static UA_StatusCode
activateSession_setting(UA_Server *server, UA_AccessControl *ac,
                        const UA_EndpointDescription *endpointDescription,
                        const UA_ByteString *secureChannelRemoteCertificate,
                        const UA_NodeId *sessionId,
                        const UA_ExtensionObject *userIdentityToken,
                        void **sessionContext) {
    AccessControlContext *context = (AccessControlContext*)ac->context;

    if(userIdentityToken->encoding == UA_EXTENSIONOBJECT_ENCODED_NOBODY) {
        if(!context->allowAnonymous)
            return UA_STATUSCODE_BADIDENTITYTOKENINVALID;

        *sessionContext = NULL;
        return UA_STATUSCODE_GOOD;
    }

    if(userIdentityToken->encoding < UA_EXTENSIONOBJECT_DECODED)
        return UA_STATUSCODE_BADIDENTITYTOKENINVALID;

    if(userIdentityToken->content.decoded.type == &UA_TYPES[UA_TYPES_ANONYMOUSIDENTITYTOKEN]) {
        if(!context->allowAnonymous)
            return UA_STATUSCODE_BADIDENTITYTOKENINVALID;

        const UA_AnonymousIdentityToken *token = (UA_AnonymousIdentityToken*)
            userIdentityToken->content.decoded.data;


        if(token->policyId.data && !UA_String_equal(&token->policyId, &anonymous_policy))
            return UA_STATUSCODE_BADIDENTITYTOKENINVALID;

        *sessionContext = NULL;
        return UA_STATUSCODE_GOOD;
    }

    if(userIdentityToken->content.decoded.type == &UA_TYPES[UA_TYPES_USERNAMEIDENTITYTOKEN]) {
        const UA_UserNameIdentityToken *userToken =
            (UA_UserNameIdentityToken*)userIdentityToken->content.decoded.data;

        if(!UA_String_equal(&userToken->policyId, &username_policy))
            return UA_STATUSCODE_BADIDENTITYTOKENINVALID;



        if(userToken->userName.length == 0 && userToken->password.length == 0)
            return UA_STATUSCODE_BADIDENTITYTOKENINVALID;

        UA_Boolean match = false;
        for(size_t i = 0; i < context->usernamePasswordLoginSize; i++) {
            if(UA_String_equal(&userToken->userName, &context->usernamePasswordLogin[i].username) &&
               UA_String_equal(&userToken->password, &context->usernamePasswordLogin[i].password)) {
                match = true;
                break;
            }
        }
        if(!match)
            return UA_STATUSCODE_BADUSERACCESSDENIED;

        UA_ByteString *username = UA_ByteString_new();
        if(username)
            UA_ByteString_copy(&userToken->userName, username);
        	*sessionContext = username->data;
        //printf("======= %u",username->data);

        return UA_STATUSCODE_GOOD;
    }

    return UA_STATUSCODE_BADIDENTITYTOKENINVALID;
}

static UA_Byte
getUserAccessLevel_setting(UA_Server *server, UA_AccessControl *ac,
                           const UA_NodeId *sessionId, void *sessionContext,
                           const UA_NodeId *nodeId, void *nodeContext) {

    if (nodeId->identifierType == UA_NODEIDTYPE_NUMERIC)
 	{
		UA_UInt32 id = nodeId->identifier.numeric;
    	if((id == 1001) && strcmp((char*)sessionContext,ADMIN)==0) {
    		//printf("Aminstrator: %d %s\n",id ,(char*)sessionContext);
     		return (UA_ACCESSLEVELMASK_WRITE | UA_ACCESSLEVELMASK_READ);
   		}
   		else if ((id == 1001) && strcmp((char*)sessionContext,TEST)==0) {
   			//printf("Test: %d %s\n",id ,(char*)sessionContext);
   			return 0x00;
   		}
 	}

    return 0x00;
}
UA_StatusCode
UA_AccessControl_setting(UA_ServerConfig *config, UA_Boolean allowAnonymous,
                         const UA_ByteString *userTokenPolicyUri,
                         AccessControlContext *context,
                         int idx,
						 char * username, char * password) {
    UA_LOG_WARNING(&config->logger, UA_LOGCATEGORY_SERVER,
                   "AccessControl: Unconfigured AccessControl. Users have all permissions.");
    UA_AccessControl *ac = &config->accessControl;

    ac->clear = clear_default;
    ac->activateSession = activateSession_setting;
    ac->closeSession = closeSession_default;
    ac->getUserRightsMask = getUserRightsMask_default;
    ac->getUserAccessLevel = getUserAccessLevel_setting;
    ac->getUserExecutable = getUserExecutable_default;
    ac->getUserExecutableOnObject = getUserExecutableOnObject_default;
    ac->allowAddNode = allowAddNode_default;
    ac->allowAddReference = allowAddReference_default;
    ac->allowBrowseNode = allowBrowseNode_default;

    ac->allowHistoryUpdateUpdateData = allowHistoryUpdateUpdateData_default;
    ac->allowHistoryUpdateDeleteRawModified = allowHistoryUpdateDeleteRawModified_default;

    ac->allowDeleteNode = allowDeleteNode_default;
    ac->allowDeleteReference = allowDeleteReference_default;

    ac->context = context;
	size_t policies = 0;
    context->allowAnonymous = allowAnonymous;
    if(allowAnonymous) {
        UA_LOG_INFO(&config->logger, UA_LOGCATEGORY_SERVER,
                    "AccessControl: Anonymous login is enabled");
    }
	UA_String usr;
	UA_String pwd;
	UA_String_init(&usr);
	UA_String_init(&pwd);
	usr = UA_STRING(username);
	pwd = UA_STRING(password);
    UA_String_copy(&usr, &context->usernamePasswordLogin[idx].username);
    UA_String_copy(&pwd, &context->usernamePasswordLogin[idx].password);

    ac->userTokenPolicies = (UA_UserTokenPolicy *)
        UA_Array_new(policies, &UA_TYPES[UA_TYPES_USERTOKENPOLICY]);
    if(!ac->userTokenPolicies)
        return UA_STATUSCODE_BADOUTOFMEMORY;

    if(allowAnonymous) {
        ac->userTokenPolicies[policies].tokenType = UA_USERTOKENTYPE_ANONYMOUS;
        ac->userTokenPolicies[policies].policyId = UA_STRING_ALLOC(ANONYMOUS_POLICY);
        if (!ac->userTokenPolicies[policies].policyId.data)
            return UA_STATUSCODE_BADOUTOFMEMORY;
        policies++;
    }

    const UA_String noneUri = UA_STRING("http://opcfoundation.org/UA/SecurityPolicy#None");
    if(UA_ByteString_equal(userTokenPolicyUri, &noneUri)) {
            UA_LOG_WARNING(&config->logger, UA_LOGCATEGORY_SERVER,
                           "Username/Password configured, but no encrypting SecurityPolicy. "
                           "This can leak credentials on the network.");
    }

    return UA_STATUSCODE_GOOD;
}

*/
import "C"

// import "C" 以上不能有其他字元
import (
	"FSRV_Edge/dao/opcdao"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/orm"
)

var (
	BrowseToNodeID NameData
	NodeDataMap    NodeMap
	globalLogger   = initlog.GetLogger()
	server         *C.UA_Server
	config         *C.UA_ServerConfig
	usedMap        checkData
	namespaceIndex C.UA_UInt16
	// DefaultRels 預設關聯表節點資料
	DefaultRels []nodeattr.TemplateRel
	// DefaultTemp 預設範本
	devNum        int64
	levelArr      = []string{"LEVEL0", "LEVEL1", "LEVEL2", "LEVEL3", "LEVEL4", "LEVEL5"}
	AllNodeStruct []nodeattr.NodeStruct
	localhostMap  = make(map[string]string)
)

type checkData struct {
	DataMap map[string]bool
	Lock    sync.Mutex
}
type NameData struct {
	DataMap map[string]string
	Lock    sync.Mutex
}
type NodeMap struct {
	DataMap map[string][]nodeattr.TemplateRel
	Lock    sync.Mutex
}

// Machine 設備資料
type Machine struct {
	Name int
}

// NodeAttr Node's Attribute
type NodeAttr struct {
	DisplayName    string
	Description    string
	DataType       string
	NodeID         C.uint
	NamespaceIndex C.ushort
	Value          interface{}
}

// WriteAttr Node's attribute
type WriteAttr struct {
	NamespaceIndex C.ushort
	NodeID         string
	Value          interface{}
}

const (
	dev0Str       = "Device0"
	stageStr      = "_Stage"
	excelFileName = "template/【數據模型】ACMT射出機聯網相容性計劃(20210329)v3.xlsx"
	nodeSheetName = "工作表1"
	startRows     = 2
	defNs         = 2003
	// ProductionStatusNodeID 生產狀態的 NodeID
	ProductionStatusNodeID C.uint = 6221
	// CycleTimeNodeID 週期時間的 NodeID
	CycleTimeNodeID C.uint = 6250
	// AverageCycleTimeNodeID 平均週期時間的 NodeID
	AverageCycleTimeNodeID C.uint = 6249
	// MachineModeNodeID 機器狀態的 NodeID
	MachineModeNodeID C.uint = 6205
	// JobPartsCounterNodeID 實際生產總數的 NodeID
	JobPartsCounterNodeID C.uint = 6257
	// JobGoodPartsCounterNodeID 良品總數的 NodeID
	JobGoodPartsCounterNodeID C.uint = 6258
	// JobBadPartsCounterNodeID 不良品總數的 NodeID
	JobBadPartsCounterNodeID C.uint = 6269
	// PartIDNodeID 產品編號的 NodeID
	PartIDNodeID C.uint = 6272
	//OTHER 其他模式
	OTHER C.uint = 0
	//AUTOMATIC 自動模式
	AUTOMATIC C.uint = 1
	//SEMI_AUTOMATIC 半自動模式
	SEMI_AUTOMATIC C.uint = 2
	//MANUAL 手動模式
	MANUAL C.uint = 3
	//SETUP 設定模式
	SETUP C.uint = 4
	//SLEEP 睡眠模式，機器仍處於開機狀態
	SLEEP C.uint = 5
)

func init() {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Init server error] %v", r)
		}
	}()
	BrowseToNodeID.Lock.Lock()
	BrowseToNodeID.DataMap = make(map[string]string)
	BrowseToNodeID.Lock.Unlock()
	usedMap.Lock.Lock()
	usedMap.DataMap = make(map[string]bool)
	usedMap.Lock.Unlock()
	localhostMap["windows"] = "127.0.0.1"
	localhostMap["linux"] = "0.0.0.0"
	orm.RegisterDriver("sqlite", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "./edge.db")
	orm.RegisterModel(new(nodeattr.ConInfo), new(nodeattr.Template), new(nodeattr.TemplateRel), new(nodeattr.NodeStruct), new(nodeattr.Converter), new(nodeattr.DevInfo),
		new(nodeattr.EdgeAuth), new(nodeattr.AuthPermissopn), new(nodeattr.SystemSetting), new(nodeattr.ConnectRecord))
	orm.RunSyncdb("default", false, false)
	devNum = 1
	NodeDataMap.DataMap = make(map[string][]nodeattr.TemplateRel)
	if f, err := excelize.OpenFile(excelFileName); err != nil {
		fmt.Println("====", err)
	} else {
		insertNodeStruct(nodeSheetName, f)
	}

	if err := getDefaultRelStruct(); err != nil {
		panic("建立初始化node struct 失敗" + err.Error())
	}
}
func addObjectNode(vNode NodeAttr, parent C.UA_NodeId, hasParent bool, parentName string) (C.UA_NodeId, error) {
	/* Define the object type for "Device" */
	displayName := C.CString(vNode.DisplayName)
	description := C.CString(vNode.Description)
	nodeID := C.CString(vNode.DisplayName)
	if parentName != vNode.DisplayName {
		nodeID = C.CString(parentName + "." + vNode.DisplayName)
	}
	localText := C.CString("en-US")
	var myNodeID C.UA_NodeId
	oAttr := C.UA_ObjectAttributes_default
	oAttr.displayName = C.UA_LOCALIZEDTEXT(localText, displayName)
	oAttr.description = C.UA_LOCALIZEDTEXT(localText, description)
	defer func() {
		C.free(unsafe.Pointer(displayName))
		C.free(unsafe.Pointer(description))
		C.free(unsafe.Pointer(localText))
	}()
	if !hasParent {
		addObjStatus := C.UA_Server_addObjectNode(server, C.UA_NODEID_STRING(namespaceIndex, nodeID),
			C.UA_NODEID_NUMERIC(0, C.UA_NS0ID_OBJECTSFOLDER),
			C.UA_NODEID_NUMERIC(0, C.UA_NS0ID_ORGANIZES),
			C.UA_QUALIFIEDNAME(namespaceIndex, displayName),
			C.UA_NODEID_NUMERIC(0, C.UA_NS0ID_BASEOBJECTTYPE), oAttr,
			nil, &myNodeID)
		if addObjStatus != C.UA_STATUSCODE_GOOD {
			return myNodeID, fmt.Errorf("Add object error : %s(%s)", C.GoString(C.UA_StatusCode_name(addObjStatus)), parentName+"."+vNode.DisplayName)
		}
	} else {
		// C.UA_NODEID_STRING(namespaceIndex, nodeID)
		addObjStatus := C.UA_Server_addObjectNode(server, C.UA_NODEID_STRING(namespaceIndex, nodeID),
			parent, C.UA_NODEID_NUMERIC(0, C.UA_NS0ID_HASSUBTYPE),
			C.UA_QUALIFIEDNAME(namespaceIndex, displayName),
			C.UA_NODEID_NUMERIC(0, C.UA_NS0ID_BASEOBJECTTYPE),
			oAttr, nil, &myNodeID)
		if addObjStatus != C.UA_STATUSCODE_GOOD {
			return myNodeID, fmt.Errorf("Add object error : %s(%s)", C.GoString(C.UA_StatusCode_name(addObjStatus)), parentName+"."+vNode.DisplayName)
		}
	}
	return myNodeID, nil
}

func addNode(vNode NodeAttr, vReferenceID C.uint, parentID C.UA_NodeId, parentName string) error {
	nodeID := C.CString(parentName + "." + vNode.DisplayName)
	displayName := C.CString(vNode.DisplayName)
	description := C.CString(vNode.Description)
	localText := C.CString("en-US")
	attr := C.UA_VariableAttributes_default
	attr.displayName = C.UA_LOCALIZEDTEXT(localText, displayName)
	attr.description = C.UA_LOCALIZEDTEXT(localText, description)
	attr.accessLevel = C.UA_ACCESSLEVELMASK_READ | C.UA_ACCESSLEVELMASK_WRITE
	defer func() {
		C.free(unsafe.Pointer(displayName))
		C.free(unsafe.Pointer(description))
		C.free(unsafe.Pointer(localText))
	}()
	switch vNode.Value.(type) {
	case string:
		str := fmt.Sprintf("%v", vNode.Value)
		myValue := C.UA_STRING(C.CString(str))
		C.UA_Variant_setScalar((*C.UA_Variant)(unsafe.Pointer(&attr.value)), (unsafe.Pointer(&myValue)), (*C.UA_DataType)(unsafe.Pointer(&C.UA_TYPES[C.UA_TYPES_STRING])))
	case int:
		myValue := C.uint(vNode.Value.(int))
		C.UA_Variant_setScalar((*C.UA_Variant)(unsafe.Pointer(&attr.value)), (unsafe.Pointer(&myValue)), (*C.UA_DataType)(unsafe.Pointer(&C.UA_TYPES[C.UA_TYPES_UINT32])))
	case float64:
		myValue := C.double(vNode.Value.(float64))
		C.UA_Variant_setScalar((*C.UA_Variant)(unsafe.Pointer(&attr.value)), (unsafe.Pointer(&myValue)), (*C.UA_DataType)(unsafe.Pointer(&C.UA_TYPES[C.UA_TYPES_DOUBLE])))
	}
	addVariableStatus := C.UA_Server_addVariableNode(server, C.UA_NODEID_STRING(namespaceIndex, nodeID),
		parentID,
		C.UA_NODEID_NUMERIC(0, vReferenceID),
		C.UA_QUALIFIEDNAME(namespaceIndex, displayName),
		C.UA_NODEID_NULL, attr, nil, nil)
	if addVariableStatus != C.UA_STATUSCODE_GOOD {
		return fmt.Errorf("Add variable error : %s(%s)", C.GoString(C.UA_StatusCode_name(addVariableStatus)), parentName+"."+vNode.DisplayName)
	}
	return nil
}

func byteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// DeleteNode Server 刪除節點
func DeleteNode(endpoint string, nodeID C.uint) error {
	var err error
	deleteStatus := C.UA_Server_deleteNode(server, C.UA_NODEID_NUMERIC(1, nodeID), true)
	if deleteStatus != C.UA_STATUSCODE_GOOD {
		return fmt.Errorf("Delete Node error : %s", C.GoString(C.UA_StatusCode_name(deleteStatus)))
	}
	return err
}

// WriteVariable server寫資料到節點
func WriteVariable(vNode WriteAttr) error {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[WriteVariable error] %v", r)
		}
	}()
	/* Write a different integer value */
	attr := C.UA_VariableAttributes_default
	str := fmt.Sprintf("%v", vNode.Value)
	myValue := C.UA_STRING(C.CString(str))
	C.UA_Variant_setScalar((*C.UA_Variant)(unsafe.Pointer(&attr.value)), (unsafe.Pointer(&myValue)), (*C.UA_DataType)(unsafe.Pointer(&C.UA_TYPES[C.UA_TYPES_STRING])))

	// C.UA_NODEID_STRING(namespaceIndex, nodeID)
	writeVariableStatus := C.UA_Server_writeValue(server, C.UA_NODEID_STRING(namespaceIndex, C.CString(vNode.NodeID)), attr.value)
	if writeVariableStatus != C.UA_STATUSCODE_GOOD {
		return fmt.Errorf("Write value error : %s", C.GoString(C.UA_StatusCode_name(writeVariableStatus)))
	}
	return nil
}

// BuildAllDevNode 建立設備節點結構
func BuildAllDevNode(devName string) {
	usedMap.Lock.Lock()
	used := usedMap.DataMap[devName]
	usedMap.Lock.Unlock()
	if !used {
		// 未建立過的設備才建立節點結構
		usedMap.Lock.Lock()
		usedMap.DataMap[devName] = true
		usedMap.Lock.Unlock()

		buildDevNode(devName)
	}
}

// BuildServer 建立 OPC UA Server
func BuildServer(port int16) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Build Server error] %v", r)
		}
	}()
	server = C.UA_Server_new()
	config = C.UA_Server_getConfig(server)
	running := C.UA_Boolean(true)
	addr := C.CString(localhostMap[runtime.GOOS])
	C.UA_ServerConfig_set_customHostname(config, C.UA_STRING(addr))
	status := C.UA_ServerConfig_setMinimalCustomBuffer(config, C.ushort(port), nil, 0, 0)
	if status != C.UA_STATUSCODE_GOOD {
		globalLogger.Criticalf("Set minimal customBuffer:%s", C.GoString(C.UA_StatusCode_name(status)))
	}

	for i := 0; i < defNs; i++ {
		namespaceIndex = C.UA_Server_addNamespace(server, C.CString(strconv.Itoa(i)))
	}
	initDev0()
	go func() {
		runServerStatus := C.UA_Server_run(server, (*C.UA_Boolean)(unsafe.Pointer(&running)))
		if runServerStatus != C.UA_STATUSCODE_GOOD {
			globalLogger.Criticalf("Build Server error: %s", C.GoString(C.UA_StatusCode_name(runServerStatus)))
		}
	}()
}

// ReadValue server讀取 Node 值
func ReadValue(ns C.ushort, nodeID string) (string, error) {
	var nodeValue interface{}
	var outNode C.UA_NodeId
	var value C.UA_Variant
	C.UA_Variant_init(&value)
	readNodeIDStatus := C.UA_Server_readNodeId(server, C.UA_NODEID_STRING(ns, C.CString(nodeID)), &outNode)
	if readNodeIDStatus != C.UA_STATUSCODE_GOOD {
		return "", fmt.Errorf("Read NodeID error : %s", C.GoString(C.UA_StatusCode_name(readNodeIDStatus)))
	}
	readValueStatus := C.UA_Server_readValue(server, outNode, &value)
	if readValueStatus != C.UA_STATUSCODE_GOOD {
		return "", fmt.Errorf("Read Node value error : %s", C.GoString(C.UA_StatusCode_name(readValueStatus)))
	}
	if *value._type == (C.UA_TYPES[C.UA_TYPES_STRING]) {
		nodeValue = uaStrToStr(*(*C.UA_String)(unsafe.Pointer(value.data)))
	} else if *value._type == (C.UA_TYPES[C.UA_TYPES_INT32]) {
		nodeValue = int32(*(*C.UA_Int32)(unsafe.Pointer(value.data)))
	} else if *value._type == (C.UA_TYPES[C.UA_TYPES_UINT32]) {
		nodeValue = uint32(*(*C.UA_UInt32)(unsafe.Pointer(value.data)))
	} else if *value._type == (C.UA_TYPES[C.UA_TYPES_DOUBLE]) {
		nodeValue = float64(*(*C.UA_Double)(unsafe.Pointer(value.data)))
	}
	return fmt.Sprintf("%v", nodeValue), nil
}

func initDev0() error {
	var temp NodeAttr
	var deviceID C.UA_NodeId
	var err error
	temp.DisplayName = dev0Str
	temp.NamespaceIndex = namespaceIndex
	if deviceID, err = addObjectNode(temp, C.UA_NODEID_NULL, false, dev0Str); err != nil {
		return err
	}
	nameArr := []string{"EdgeIP", "EdgePort", "Manufacturer", "Account", "Password"}
	for _, v := range nameArr {
		temp.DisplayName = v
		temp.NamespaceIndex = namespaceIndex
		if err := addNode(temp, C.UA_NS0ID_HASCOMPONENT, deviceID, dev0Str); err != nil {
			return err
		}
	}
	return nil
}

func insertNodeStruct(sheetName string, f *excelize.File) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	buildLevelNode := false
	buildGroupNode := false
	nStructArr := []nodeattr.NodeStruct{}
	var nodeStruct nodeattr.NodeStruct
	nodeStructTabSize, getTabErr := opcdao.GetTableSize(nodeStruct.TableName(), o)
	if getTabErr != nil {
		return getTabErr
	}
	if nodeStructTabSize == 0 {
		rows := f.GetRows(sheetName)
		for i := startRows; i < len(rows); i++ {
			row := rows[i]
			nodeStruct.Level = strings.ReplaceAll(row[1], " ", "")
			if nodeStruct.Level != "" {
				nodeStruct.BrowseName = row[2]
				nodeStruct.DataType = row[4]
				nodeStruct.Description = row[3]
				if !buildLevelNode { // 是level object
					buildLevelNode = true
				}
				if strings.Contains(nodeStruct.BrowseName, stageStr) { // 是Sort object
					buildGroupNode = true
					nodeStruct.Group = nodeStruct.BrowseName
					nodeStruct.ParentBrowseName = nodeStruct.Level
				} else if buildGroupNode { // 加到 Sort 下
					nodeStruct.ParentBrowseName = nodeStruct.Group
				} else { // 不是在 Group 內
					nodeStruct.ParentBrowseName = nodeStruct.Level
				}
				nStructArr = append(nStructArr, nodeStruct)
			} else {
				buildLevelNode = false
				buildGroupNode = false
			}
		}
		insertErr := opcdao.InsertNodeStruct(nStructArr, o)
		if insertErr != nil {
			return insertErr
		}
	}
	return err
}

func identifierToString(cbytes [16]byte) (result string) {
	var length C.size_t
	var idx uint
	idx = 0
	for idx < 8 {
		if idx == 0 {
			length += C.size_t(cbytes[idx])
		} else {
			length += C.size_t(cbytes[idx]) << idx * 8
		}
		idx++
	}

	var uptr uintptr
	idx = 8
	for idx < 16 {
		if idx == 0 {
			uptr = uintptr(cbytes[8])
			length += C.size_t(cbytes[idx])
		} else {
			uptr = uintptr(cbytes[8]) << idx * 8
		}
		idx++
	}
	uptr = uintptr(cbytes[8]) + uintptr(cbytes[9])<<8 + uintptr(cbytes[10])<<16 + uintptr(cbytes[11])<<24
	ptr := unsafe.Pointer(uptr)
	var s C.UA_String
	C.UA_String_init(&s) /* _init zeroes out the entire memory of the datatype */
	s.length = length
	s.data = (*C.UA_Byte)(ptr)
	return uaStrToStr(s)
}
func uaStrToStr(uaStr C.UA_String) string {
	if uaStr.data == nil || int(uaStr.length) == 0 {
		return ""
	}
	data := (*C.char)(unsafe.Pointer(uaStr.data))
	return C.GoString(data)
}

func getDefaultRelStruct() (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	AllNodeStruct, err = opcdao.GetAllNodeStruct(o)
	if err != nil {
		return
	}
	var stageNode nodeattr.TemplateRel
	var str string
	NodeDataMap.Lock.Lock()
	for _, v := range AllNodeStruct {
		var tmp nodeattr.TemplateRel
		tmp.ID = v.ID
		if strings.Contains(v.ParentBrowseName, nodeattr.LevelStr) {
			str = v.ParentBrowseName
			if strings.Contains(v.BrowseName, stageStr) {
				stageNode.DstNodeid = v.ParentBrowseName + "." + v.BrowseName
			} else {
				tmp.DstNodeid = v.ParentBrowseName + "." + v.BrowseName
				tmp.DstBrowse = v.BrowseName
				tmp.SrcBrowse = tmp.DstBrowse
				NodeDataMap.DataMap[v.ParentBrowseName] = append(NodeDataMap.DataMap[v.ParentBrowseName], tmp)
			}
		} else if strings.Contains(v.ParentBrowseName, stageStr) {
			tmp.DstNodeid = stageNode.DstNodeid + "." + v.BrowseName
			tmp.DstBrowse = v.BrowseName
			tmp.SrcBrowse = tmp.DstBrowse
			NodeDataMap.DataMap[str] = append(NodeDataMap.DataMap[str], tmp)
		}
		BrowseToNodeID.Lock.Lock()
		BrowseToNodeID.DataMap[tmp.DstBrowse] = tmp.DstNodeid
		BrowseToNodeID.Lock.Unlock()
		DefaultRels = append(DefaultRels, tmp)
	}
	NodeDataMap.Lock.Unlock()
	return
}

func buildDevNode(deviceName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[buildDevNode error] %v", r)
		}
	}()
	var temp NodeAttr
	var levelID C.UA_NodeId
	var groupID C.UA_NodeId
	var deviceID C.UA_NodeId
	temp.DisplayName = deviceName
	temp.NamespaceIndex = namespaceIndex
	temp.Description = "This is " + deviceName
	if deviceID, err = addObjectNode(temp, C.UA_NODEID_NULL, false, deviceName); err != nil {
		fmt.Println(err)
	}
	objExist := make(map[string]bool)
	for _, levelStr := range levelArr {
		temp.DisplayName = levelStr
		temp.NamespaceIndex = namespaceIndex
		temp.Description = "This is " + levelStr
		if levelID, err = addObjectNode(temp, deviceID, true, deviceName); err != nil { // 建立 level object
			fmt.Println(err)
		} else {
			NodeDataMap.Lock.Lock()
			for _, v := range NodeDataMap.DataMap[levelStr] {
				stageName := v.GetStage()
				if stageName != "" { // 有stage object
					temp.DisplayName = stageName
					temp.Description = "This is " + temp.DisplayName
					browse := deviceName + "." + levelStr
					if !objExist[temp.DisplayName] {
						if groupID, err = addObjectNode(temp, levelID, true, browse); err != nil { // 建立 stage object
							fmt.Println(err)
						}
						objExist[temp.DisplayName] = true
					}
					temp.DisplayName = v.DstBrowse
					temp.Description = "This is " + temp.DisplayName
					browse = deviceName + "." + levelStr + "." + stageName
					addNode(temp, C.UA_NS0ID_HASCOMPONENT, groupID, browse)
				} else { // 沒有stage object
					temp.DisplayName = v.DstBrowse
					temp.Description = "This is " + temp.DisplayName
					browse := deviceName + "." + levelStr
					addNode(temp, C.UA_NS0ID_HASCOMPONENT, levelID, browse)
				}
			}
			NodeDataMap.Lock.Unlock()
		}
	}
	return
}
