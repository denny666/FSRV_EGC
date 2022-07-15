package opcuaclient

import (
	"FSRV_Edge/global"
	"FSRV_Edge/influx"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/opcuaservice"
	"FSRV_Edge/service/routerservice"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const (
	shortDuration = 5 * time.Minute
	// OpcTitle opcua start url
	OpcTitle  = "opc.tcp://"
	namespace = 2004
)

// NodeStruct 節點資料
type NodeStruct struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
type valueMap struct {
	dataMap  map[string]interface{}
	levelMap map[string]string
	lock     sync.Mutex
}
type devSubContext struct {
	ctxMap    map[string]context.Context
	cancelMap map[string]context.CancelFunc
	lock      sync.Mutex
}
type nodeParent struct {
	dataMap   map[string]string
	nodeCount map[string]int
	lock      sync.Mutex
}

func (c *nodeParent) GetMap() map[string]string {
	c.lock.Lock()
	tmpMap := c.dataMap
	c.lock.Unlock()
	if tmpMap == nil {
		return make(map[string]string)
	}
	return tmpMap
}
func (c *nodeParent) GetCount(val string) int {
	c.lock.Lock()
	data := c.nodeCount[val]
	c.lock.Unlock()
	return data
}
func (c *nodeParent) SetMap(key, val string) {
	c.lock.Lock()
	c.dataMap[key] = val
	c.nodeCount[val]++
	c.lock.Unlock()
}

func (c *nodeParent) Init() {
	c.lock.Lock()
	c.dataMap = make(map[string]string)
	c.nodeCount = make(map[string]int)
	c.lock.Unlock()
}

type NameData struct {
	DataMap map[string]string
	Lock    sync.Mutex
}
type devLevelMap struct {
	DataMap map[string]map[string][]NodeStruct
	Lock    sync.Mutex
}

var (
	allLevelArr      []string
	NodeIdToBrowse   NameData
	AllNodeData      devLevelMap
	ParentBrowseArr  []string
	GlobalNodeParent nodeParent
	globalLogger     = initlog.GetLogger()
	devNum           = 1
	flag             = false
	hasRecordLevel   = make(map[string]bool)
	AllNodeDataConv  NodeMap
	// AllBrowseData 設備所有browseName對應到的NodeID
	AllBrowseData BrowseData
	// AllNodeIDData 設備所有NodeID對應到的browseName
	AllNodeIDData BrowseData
	BrowseNameMap BrowseNameData
	HasBrowse     checkStruct
	NameInfo      NameData
	gValMap       valueMap
	DevSub        checkStruct
	devCtx        devSubContext
)

type checkStruct struct {
	DataMap map[string]bool
	Lock    sync.Mutex
}

type NodeMap struct {
	DataMap map[string][]nodeattr.Converter
	Lock    sync.Mutex
}
type ChangeMap struct {
	HasRecord map[interface{}]bool
	Lock      sync.Mutex
}

type BrowseData struct {
	DataMap map[string]map[string]*opcua.Node
	Lock    sync.Mutex
}
type BrowseNameData struct {
	DataMap map[string][]string
	Lock    sync.Mutex
}

func init() {
	devCtx.lock.Lock()
	devCtx.cancelMap = make(map[string]context.CancelFunc)
	devCtx.ctxMap = make(map[string]context.Context)
	devCtx.lock.Unlock()
	DevSub.Lock.Lock()
	DevSub.DataMap = make(map[string]bool)
	DevSub.Lock.Unlock()
	allLevelArr = append(allLevelArr, nodeattr.Level0Str, nodeattr.Level1Str, nodeattr.Level2Str, nodeattr.Level3Str, nodeattr.Level4Str, nodeattr.Level5Str)
	AllNodeIDData.Lock.Lock()
	NodeIdToBrowse.Lock.Lock()
	NodeIdToBrowse.DataMap = make(map[string]string)
	NodeIdToBrowse.Lock.Unlock()
	gValMap.lock.Lock()
	gValMap.dataMap = make(map[string]interface{})
	gValMap.levelMap = make(map[string]string)
	gValMap.lock.Unlock()
	AllNodeDataConv.Lock.Lock()
	AllBrowseData.Lock.Lock()
	BrowseNameMap.Lock.Lock()
	AllNodeData.Lock.Lock()
	AllNodeDataConv.DataMap = make(map[string][]nodeattr.Converter)
	AllBrowseData.DataMap = make(map[string]map[string]*opcua.Node)
	BrowseNameMap.DataMap = make(map[string][]string)
	AllNodeData.DataMap = make(map[string]map[string][]NodeStruct)
	BrowseNameMap.Lock.Unlock()
	AllNodeData.Lock.Unlock()
	AllBrowseData.Lock.Unlock()
	AllNodeDataConv.Lock.Unlock()
	HasBrowse.Lock.Lock()
	HasBrowse.DataMap = make(map[string]bool)
	HasBrowse.Lock.Unlock()
}

// GetNodeIdFromBrowse 依照SrcBrowseName取得NodeID
func GetNodeIdFromBrowse(bName, devID string) (tmp *opcua.Node) {
	if bName == "" {
		return
	}
	BrowseNameMap.Lock.Lock()
	bArr := BrowseNameMap.DataMap
	BrowseNameMap.Lock.Unlock()
	AllBrowseData.Lock.Lock()
	node := AllBrowseData.DataMap[devID]
	AllBrowseData.Lock.Unlock()
	for _, v := range bArr[devID] {
		if v == bName {
			tmp = node[v]
			break
		}
	}
	NodeIdToBrowse.Lock.Lock()
	if len(tmp.ID.StringID()) == 0 {
		NodeIdToBrowse.DataMap[tmp.ID.String()] = bName
	} else {
		NodeIdToBrowse.DataMap[tmp.ID.StringID()] = bName
	}
	NodeIdToBrowse.Lock.Unlock()

	return
}

// GetMonitorItem 取得監察節點
func GetMonitorItem(conInfo nodeattr.ConInfo, devID string) {
	var moldCountNodes []string
	DevSub.Lock.Lock()
	sub := DevSub.DataMap[devID]
	DevSub.Lock.Unlock()
	if sub {
		return
	}
	AllNodeDataConv.Lock.Lock()
	convArr := AllNodeDataConv.DataMap[devID]
	AllNodeDataConv.Lock.Unlock()
	if len(convArr) == 0 {
		return
	}
	d := time.Now().Add(shortDuration)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer func() {
		if r := recover(); r != nil {
			cancel()
			globalLogger.Criticalf("[GetMonitorItem error] %v", r)
		}
	}()
	for _, conv := range convArr {
		var nodeStr string
		level := conv.GetLevel()
		if level == nodeattr.Level0Str {
			continue
		} else if level == nodeattr.Level1Str {
			if conv.DstBrowse != nodeattr.MoldCountStr && conv.DstBrowse != nodeattr.CycleTimeStr && conv.DstBrowse != nodeattr.MopTimeStr &&
				conv.DstBrowse != nodeattr.MclTimeStr && conv.DstBrowse != nodeattr.HeaterStr && conv.DstBrowse != nodeattr.MotorStr &&
				conv.DstBrowse != nodeattr.AlarmStr {
				continue
			}
		}
		if conv.SrcNamespace != 0 && len(conv.SrcNodeid) != 0 { // 用NodeID找
			ns := strconv.Itoa(conv.SrcNamespace)
			_, err := strconv.Atoi(conv.SrcNodeid)
			var node string
			if err == nil {
				node = "ns=" + ns + ";i=" + conv.SrcNodeid
			} else {
				node = "ns=" + ns + ";s=" + conv.SrcNodeid
			}
			moldCountNodes = append(moldCountNodes, node)
		} else if len(conv.SrcBrowse) != 0 { // 用SrcBrowse找
			nid := GetNodeIdFromBrowse(conv.SrcBrowse, devID)
			if nid == nil {
				continue
			}
			nodeStr = nid.ID.String()
			moldCountNodes = append(moldCountNodes, nodeStr)
		}

	}
	if len(moldCountNodes) > 0 {
		devCtx.lock.Lock()
		if _, ok := devCtx.cancelMap[devID]; ok {
			devCtx.cancelMap[devID]()
		}
		devCtx.ctxMap[devID] = ctx
		devCtx.cancelMap[devID] = cancel
		devCtx.lock.Unlock()
		monitorNode(ctx, conInfo, devID, moldCountNodes)
	}
}
func monitorNode(ctx context.Context, conInfo nodeattr.ConInfo, devName string, moldCountNodes []string) {
	var endpoint string
	if conInfo.Port == "80" {
		endpoint = nodeattr.OpcTitle + conInfo.IP
	} else {
		endpoint = nodeattr.OpcTitle + conInfo.IP + ":" + conInfo.Port
	}
	c, conErr := ConnectServer(ctx, conInfo.Account, conInfo.Password, endpoint)
	if conErr != nil {
		panic(conErr)
	}
	defer func() {
		if r := recover(); r != nil {
			c.Close()
			DevSub.Lock.Lock()
			DevSub.DataMap[devName] = false
			DevSub.Lock.Unlock()
			globalLogger.Criticalf("[monitor error] %v", r)
			// monitorNode(ctx, conInfo, devName, moldCountNodes)
			GetMonitorItem(conInfo, devName)
		}
	}()

	// a := opcua.NewMonitoredItemCreateRequestWithDefaults()
	// TODO 修改 subInterval 確認是否可減少延遲在台中精機射出機上
	subInterval, err := time.ParseDuration(opcua.DefaultSubscriptionInterval.String())
	// subInterval, err := time.ParseDuration("1ms")
	if err != nil {
		panic(err)
	}
	// signalCh := make(chan os.Signal, 1)
	// signal.Notify(signalCh, os.Interrupt)
	// go func() {
	// 	<-signalCh
	// 	cancel()
	// }()
	m, err := monitor.NewNodeMonitor(c)
	if err != nil {
		panic(err)
	}
	m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
		panic(err)
	})
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// d := time.Now().Add(shortDuration)
	// ctx, cancel = context.WithDeadline(context.Background(), d)
	ch := make(chan *monitor.DataChangeMessage, 5000)

	sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: subInterval, MaxNotificationsPerPublish: 5000}, ch, moldCountNodes...)
	if err != nil {
		panic(err)
	}

	go startChanSub(ctx, c, sub, ch, conInfo, devName, moldCountNodes)
	// <-ctx.Done()
	// wg.Wait()
	// WG 的必要性
}
func decode(val string) (string, error) {
	num, err := strconv.Atoi(val)
	if err != nil {
		return "", err
	}
	hexStr := strconv.FormatInt(int64(num), 16)
	bs, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return reverse(string(bs)), err
}
func reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}
func needDecode(browse string) bool {
	if browse == nodeattr.BrandStr || browse == nodeattr.TypeStr || browse == nodeattr.SerialNumStr ||
		browse == nodeattr.IoMVersion || browse == nodeattr.ResinType || browse == nodeattr.PartsNumber ||
		browse == nodeattr.ProductionMoldName {
		return true
	}
	return false
}
func isTime(browse string) bool {
	if browse == nodeattr.MonthDate || browse == nodeattr.Month3Date ||
		browse == nodeattr.Month6Date || browse == nodeattr.PowerOnTime {
		return true
	}
	return false
}

// ReadNode 讀取節點資料
func ReadNode(c *opcua.Client, browse, devID string) (val string, err error) {
	AllNodeDataConv.Lock.Lock()
	convInfos := AllNodeDataConv.DataMap[devID]
	AllNodeDataConv.Lock.Unlock()
	for _, v := range convInfos {
		if v.DstBrowse == browse {
			if needDecode(browse) {
				var tmp string
				for i := 2; i <= 20; i += 2 {
					bTmp := browse + "_" + strconv.Itoa(i-1) + "-" + strconv.Itoa(i)
					v.SrcNamespace = 1
					v.SrcNodeid = bTmp
					tmp, _ = ReadNodeByID(c, v)
					if len(tmp) == 0 || tmp == "0" {
						break
					}
					if str, err := decode(tmp); err == nil {
						val += str
					}
				}
			} else if isTime(browse) {
				var tmp string
				bTmp := browse + "_" + nodeattr.YearStr
				v.SrcNamespace = 1
				v.SrcNodeid = bTmp
				tmp, _ = ReadNodeByID(c, v)
				val += tmp + "/"
				bTmp = browse + "_" + nodeattr.MonthStr
				v.SrcNamespace = 1
				v.SrcNodeid = bTmp
				tmp, _ = ReadNodeByID(c, v)
				val += tmp + "/"
				bTmp = browse + "_" + nodeattr.DayStr
				v.SrcNamespace = 1
				v.SrcNodeid = bTmp
				tmp, _ = ReadNodeByID(c, v)
				val += tmp
			} else if browse == nodeattr.EstimatedTime {
				v.SrcNamespace = 1
				v.SrcNodeid = browse + nodeattr.HourStr
				val1, _ := ReadNodeByID(c, v)
				if len(val1) != 0 {
					val1 += "小時"
				}
				v.SrcNodeid = browse + nodeattr.MinStr
				val2, _ := ReadNodeByID(c, v)
				if len(val2) != 0 {
					val2 += "分鐘"
				}
				val = val1 + val2
			} else if browse == nodeattr.ManuDateStr {
				v.SrcNamespace = 1
				v.SrcNodeid = browse + nodeattr.YearStr
				val1, _ := ReadNodeByID(c, v)
				v.SrcNodeid = browse + nodeattr.MonthStr
				val2, _ := ReadNodeByID(c, v)
				val = val1 + "/" + val2
			} else if v.SrcNamespace != 0 && len(v.SrcNodeid) != 0 { //  用NodeID找
				val, err = ReadNodeByID(c, v)
			} else if len(v.SrcBrowse) != 0 { // 用SrcBrowse找
				val, err = ReadNodeByBrowseName(c, devID, v)
			} else { // 依照公式帶入數值
				if len(v.ConvFunc) != 0 {
					if v, checkErr := calFormula(c, devID, v); checkErr == nil {
						switch v.(type) {
						case float64:
							val = strconv.FormatFloat(v.(float64), 'f', 5, 32)
						case string:
							val = v.(string)
						default:
							val = ""
						}
					} else {
						err = checkErr
						fmt.Println(err)
					}
				}
			}
			return
		}
	}
	return
}
func calFormula(c *opcua.Client, devID string, conv nodeattr.Converter) (interface{}, error) {
	expression, err := govaluate.NewEvaluableExpression(conv.ConvFunc)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]interface{})
	if len(conv.RefBrowseName1) != 0 {
		if xVal, err := ReadNode(c, conv.RefBrowseName1, devID); err != nil {
			return nil, errors.New(conv.RefBrowseName1 + " can't find value")
		} else {
			if xFloat, err := strconv.ParseFloat(xVal, 64); err != nil {
				return nil, errors.New(conv.RefBrowseName1 + " value is not number")
			} else {
				parameters["x"] = xFloat
			}
		}
	}
	if len(conv.RefBrowseName2) != 0 {
		if yVal, err := ReadNode(c, conv.RefBrowseName2, devID); err != nil {
			return nil, errors.New(conv.RefBrowseName2 + " can't find value")
		} else {
			if yFloat, err := strconv.ParseFloat(yVal, 64); err != nil {
				return nil, errors.New(conv.RefBrowseName2 + " value is not number")
			} else {
				parameters["y"] = yFloat
			}
		}
	}
	if len(conv.RefBrowseName3) != 0 {
		if zVal, err := ReadNode(c, conv.RefBrowseName3, devID); err != nil {
			return nil, errors.New(conv.RefBrowseName3 + " can't find value")
		} else {
			if zFloat, err := strconv.ParseFloat(zVal, 64); err != nil {
				return nil, errors.New(conv.RefBrowseName3 + " value is not number")
			} else {
				parameters["z"] = zFloat
			}
		}
	}
	result, err := expression.Evaluate(parameters)
	return result, err
}

// ReadNodeByID 依照節點ID讀節點資料
func ReadNodeByID(c *opcua.Client, conv nodeattr.Converter) (string, error) {
	var ns uint16
	var node *ua.NodeID
	ns = uint16(conv.SrcNamespace)
	idStr := conv.SrcNodeid
	if idInt, err := strconv.Atoi(idStr); err != nil {
		node = ua.NewStringNodeID(ns, idStr)
	} else {
		node = ua.NewNumericNodeID(ns, uint32(idInt))
	}

	if node == nil {
		return "", fmt.Errorf("Node is not find")
	}
	var val *ua.Variant
	var err error
	if val, err = c.Node(node).Value(); err != nil {
		return "", err
	}
	if val.LocalizedText() != nil {
		return val.LocalizedText().Text, err
	}
	str := fmt.Sprintf("%v", val.Value())
	return str, err
}

// ReadNodeByBrowseName 依照SrcBrowseName讀節點資料
func ReadNodeByBrowseName(c *opcua.Client, devID string, conv nodeattr.Converter) (string, error) {
	// var ns uint16
	var nid *opcua.Node
	nid = GetNodeIdFromBrowse(conv.SrcBrowse, devID)
	if nid == nil {
		return "", fmt.Errorf("Node is not find")
	}
	var val *ua.Variant
	var err error
	node := ua.NewNumericNodeID(nid.ID.Namespace(), nid.ID.IntID())
	if val, err = c.Node(node).Value(); err != nil {
		return "", err
	}
	if val.LocalizedText() != nil {
		return val.LocalizedText().Text, err
	}
	str := fmt.Sprintf("%v", val.Value())
	return str, err
}
func isZero(v interface{}) bool {
	t := reflect.TypeOf(v)
	if !t.Comparable() {
		return false
	}
	return v == reflect.Zero(t).Interface()
}

// ReadDevAllNode 取得設備所有欄位資料
func ReadDevAllNode(opcClient *opcua.Client, devName string) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[ReadDevAllNode error] %v", r)
		}
	}()
	AllNodeDataConv.Lock.Lock()
	convInfos := AllNodeDataConv.DataMap[devName]
	AllNodeDataConv.Lock.Unlock()
	boolMap := make(map[interface{}]int)
	boolMap["true"] = 1
	boolMap["false"] = 0
	var l0NodeArr []NodeStruct
	var l1NodeArr []NodeStruct
	var l2NodeArr []NodeStruct
	var l3NodeArr []NodeStruct
	var l4NodeArr []NodeStruct
	var l5NodeArr []NodeStruct
	var tmp valueMap
	tmp.dataMap = make(map[string]interface{})
	tmp.levelMap = make(map[string]string)
	tmpMap := GlobalNodeParent.GetMap()
	for _, conv := range convInfos {
		level := conv.GetLevel()
		valStr, _ := ReadNode(opcClient, conv.DstBrowse, devName)
		key := tmpMap[conv.DstBrowse] // key是 dstBrowse 的 parent
		count := GlobalNodeParent.GetCount(key)
		tmp.levelMap[key] = level
		var node opcuaservice.WriteAttr
		node.Value = valStr
		node.NodeID = conv.DstNodeid
		opcuaservice.WriteVariable(node)
		if count == 1 {
			tmp.dataMap[key] = valStr
		} else {
			if tmp.dataMap[key] == nil {
				tmp.dataMap[key] = make([]string, 0)
			}
			tmp.dataMap[key] = append(tmp.dataMap[key].([]string), valStr)
		}
	}

	for _, key := range ParentBrowseArr {
		var tmpT NodeStruct
		tmpT.Name = key
		tmpT.Value = tmp.dataMap[key]
		if tmpT.Name == nodeattr.MotorStr {
			tmpT.Value = boolMap[tmpT.Value]
		} else if tmpT.Name == nodeattr.StatusStr {
			if tmpT.Value != nil {
				if m, err := strconv.Atoi(tmpT.Value.(string)); err != nil {
					tmpT.Value = ""
				} else {
					tmpT.Value = m
				}
			}
		} else if tmpT.Name == nodeattr.MoldCountStr {
			if tmpT.Value != nil {
				if m, err := strconv.Atoi(tmpT.Value.(string)); err != nil {
					tmpT.Value = 0
				} else {
					tmpT.Value = m
				}
			}
		}

		if tmp.levelMap[key] == nodeattr.Level2Str {
			l2NodeArr = append(l2NodeArr, tmpT)
		} else if tmp.levelMap[key] == nodeattr.Level3Str {
			l3NodeArr = append(l3NodeArr, tmpT)
		} else if tmp.levelMap[key] == nodeattr.Level0Str {
			l0NodeArr = append(l0NodeArr, tmpT)
		} else if tmp.levelMap[key] == nodeattr.Level1Str {
			l1NodeArr = append(l1NodeArr, tmpT)
		} else if tmp.levelMap[key] == nodeattr.Level4Str {
			l4NodeArr = append(l4NodeArr, tmpT)
		} else if tmp.levelMap[key] == nodeattr.Level5Str {
			l5NodeArr = append(l5NodeArr, tmpT)
		}
	}
	gValMap.lock.Lock()
	gValMap.levelMap = tmp.levelMap
	gValMap.lock.Unlock()
	AllNodeData.Lock.Lock()
	AllNodeData.DataMap[devName] = make(map[string][]NodeStruct)
	AllNodeData.DataMap[devName][nodeattr.Level0Str] = l0NodeArr
	AllNodeData.DataMap[devName][nodeattr.Level1Str] = l1NodeArr
	AllNodeData.DataMap[devName][nodeattr.Level2Str] = l2NodeArr
	AllNodeData.DataMap[devName][nodeattr.Level3Str] = l3NodeArr
	AllNodeData.DataMap[devName][nodeattr.Level4Str] = l4NodeArr
	AllNodeData.DataMap[devName][nodeattr.Level5Str] = l5NodeArr
	AllNodeData.Lock.Unlock()
	elapsed := time.Since(start)
	if elapsed.Milliseconds() > 0 {
		log.Printf("ReadDevAllNode took %s", elapsed)
	}
}

func getBrowseMap(devID string) map[string]string {
	tmp := make(map[string]string)
	AllNodeDataConv.Lock.Lock()
	convs := AllNodeDataConv.DataMap[devID]
	AllNodeDataConv.Lock.Unlock()
	for _, v := range convs {
		tmp[v.SrcBrowse] = v.DstBrowse
	}
	return tmp
}

func startChanSub(ctx context.Context, c *opcua.Client, sub *monitor.Subscription, ch chan *monitor.DataChangeMessage, conInfo nodeattr.ConInfo, devName string, moldCountNodes []string) {
	defer func() {
		if r := recover(); r != nil {
			DevSub.Lock.Lock()
			DevSub.DataMap[devName] = false
			DevSub.Lock.Unlock()
			c.Close()
			close(ch)
			sub.Unsubscribe()
			globalLogger.Criticalf("[sub error] %v", r)
			GetMonitorItem(conInfo, devName)
		}
	}()
	DevSub.Lock.Lock()
	DevSub.DataMap[devName] = true
	DevSub.Lock.Unlock()
	boolMap := make(map[interface{}]int)
	boolMap["true"] = 1
	boolMap["false"] = 0
	// defer cleanup(sub, wg)
	browseMap := getBrowseMap(devName)
	for {
		select {
		case <-ctx.Done():
			panic("ctx Done")
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] %s", msg.Error)
			} else if sub != nil { // Server 模次變動
				// 台中精機射出機 節點數值變化後會延遲3分鐘才接收到
				start := time.Now()
				valMap := make(map[string]interface{})
				var browse string
				if len(msg.NodeID.StringID()) == 0 {
					NodeIdToBrowse.Lock.Lock()
					browse = NodeIdToBrowse.DataMap[msg.NodeID.String()]
					NodeIdToBrowse.Lock.Unlock()
				} else {
					browse = msg.NodeID.StringID()
				}
				browse = browseMap[browse]
				opcuaservice.BrowseToNodeID.Lock.Lock()
				nodeID := "Device" + devName + "." + opcuaservice.BrowseToNodeID.DataMap[browse]
				opcuaservice.BrowseToNodeID.Lock.Unlock()
				var node opcuaservice.WriteAttr
				node.Value = msg.Value.Value()
				node.NodeID = nodeID
				opcuaservice.WriteVariable(node)
				gValMap.lock.Lock()
				level := gValMap.levelMap[browse]
				gValMap.lock.Unlock()
				AllNodeData.Lock.Lock()
				nodeArr := AllNodeData.DataMap[devName][level]
				AllNodeData.Lock.Unlock()
				levelData := getUnusedLevelData(devName, level)
				for i := 0; i < len(nodeArr); i++ {
					if nodeArr[i].Name == browse {
						nodeArr[i].Value = msg.Value.Value()
					}
					if browse == nodeattr.MotorStr {
						valMap[browse] = boolMap[msg.Value.Value()]
					}
					valMap[nodeArr[i].Name] = nodeArr[i].Value
				}
				for _, v := range levelData {
					valMap[v.Name] = v.Value
				}
				AllNodeData.Lock.Lock()
				AllNodeData.DataMap[devName][level] = nodeArr
				AllNodeData.Lock.Unlock()
				log.Printf("[channel ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID.String(), msg.Value.Value())
				if global.WriteDBSetting == 1 {
					if addErr := influx.AddData(influx.NodeHdaMeasure+devName, valMap); addErr != nil {
						fmt.Println("[addErr]", addErr)
					}
				}
				elapsed := time.Since(start)
				if elapsed.Milliseconds() > 0 {
					log.Printf("startChanSub took %s", elapsed)
				}
			}

		} // select
	}
}
func getUnusedLevelData(devName, level string) (tmp []NodeStruct) {
	var tmpArr []string
	for _, v := range allLevelArr {
		if v != level {
			tmpArr = append(tmpArr, v)
		}
	}
	AllNodeData.Lock.Lock()
	for _, v := range tmpArr {
		tmp = append(tmp, AllNodeData.DataMap[devName][v]...)
	}
	AllNodeData.Lock.Unlock()
	return
}
func cleanup(sub *monitor.Subscription, wg *sync.WaitGroup) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
	wg.Done()
}

// ConnectServer 建立連線
func ConnectServer(ctx context.Context, user, pass, endpointURL string) (*opcua.Client, error) {
	var c *opcua.Client
	endpoints, getErr := opcua.GetEndpoints(endpointURL)
	if getErr != nil {
		return c, getErr
	}
	opts, optErr := clientOptsFromFlags(user, pass, endpoints)
	if optErr != nil {
		return c, optErr
	}
	// Create a Client with the selected options
	c = opcua.NewClient(endpointURL, opts...)
	if err := c.Connect(ctx); err != nil {
		return c, fmt.Errorf("Connect %s error: %s", endpointURL, err)
	}
	return c, nil
}

func authFromFlags(user, pass string) (ua.UserTokenType, opcua.Option) {
	var authMode ua.UserTokenType
	var authOption opcua.Option
	if len(user) == 0 || len(pass) == 0 {
		authMode = ua.UserTokenTypeAnonymous
		authOption = opcua.AuthAnonymous()

	} else {
		authMode = ua.UserTokenTypeUserName
		authOption = opcua.AuthUsername(user, pass)
	}

	return authMode, authOption
}
func clientOptsFromFlags(user, pass string, endpoints []*ua.EndpointDescription) ([]opcua.Option, error) {
	opts := []opcua.Option{}

	// ApplicationURI is automatically read from the cert so is not required if a cert if provided
	if *nodeattr.Certfile == "" && !*nodeattr.Gencert {
		opts = append(opts, opcua.ApplicationURI(*nodeattr.Appuri))
	}

	var cert []byte
	if *nodeattr.Gencert || (*nodeattr.Certfile != "" && *nodeattr.Keyfile != "") {
		if *nodeattr.Gencert {
			generateCert(*nodeattr.Appuri, 2048, *nodeattr.Certfile, *nodeattr.Keyfile)
		}
		debug.Printf("Loading cert/key from %s/%s", *nodeattr.Certfile, *nodeattr.Keyfile)
		c, err := tls.LoadX509KeyPair(*nodeattr.Certfile, *nodeattr.Keyfile)
		if err != nil {
			return opts, fmt.Errorf("Failed to load certificate: %s", err)
		}
		pk, ok := c.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return opts, fmt.Errorf("Invalid private key")
		}
		cert = c.Certificate[0]
		opts = append(opts, opcua.PrivateKey(pk), opcua.Certificate(cert))

	}

	var secPolicy string
	secPolicy = ua.SecurityPolicyURINone

	// Select the most appropriate authentication mode from server capabilities and user input
	authMode, authOption := authFromFlags(user, pass)
	opts = append(opts, authOption)

	secMode := ua.MessageSecurityModeNone
	// Find the best endpoint based on our input and server recommendation (highest SecurityMode+SecurityLevel)
	var serverEndpoint *ua.EndpointDescription

	for _, e := range endpoints {
		if e.SecurityMode == secMode {
			serverEndpoint = e
		}
	}
	if serverEndpoint == nil { // Didn't find an endpoint with matching policy and mode.
		return opts, fmt.Errorf("unable to find suitable server endpoint with selected sec-policy and sec-mode")
	}
	secPolicy = serverEndpoint.SecurityPolicyURI
	secMode = serverEndpoint.SecurityMode
	// Check that the selected endpoint is a valid combo
	err := validateEndpointConfig(endpoints, secPolicy, secMode, authMode)
	if err != nil {
		return opts, fmt.Errorf("error validating input: %s", err)
		// 帳號密碼或其他授權資料無效
	}
	opts = append(opts, opcua.SecurityFromEndpoint(serverEndpoint, authMode))
	return opts, nil
}

func validateEndpointConfig(endpoints []*ua.EndpointDescription, secPolicy string, secMode ua.MessageSecurityMode, authMode ua.UserTokenType) error {
	for _, e := range endpoints {
		if e.SecurityMode == secMode && e.SecurityPolicyURI == secPolicy {
			for _, t := range e.UserIdentityTokens {
				if t.TokenType == authMode {
					return nil
				}
			}
		}
	}
	err := errors.Errorf("server does not support an endpoint with security : %s , %s", secPolicy, secMode)
	return err
}

func generateCert(host string, rsaBits int, certFile, keyFile string) {

	if len(host) == 0 {
		log.Fatalf("Missing required host parameter")
	}
	if rsaBits == 0 {
		rsaBits = 2048
	}
	if len(certFile) == 0 {
		certFile = "cert.pem"
	}
	if len(keyFile) == 0 {
		keyFile = "key.pem"
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gopcua Test Client"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageContentCommitment | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
		if uri, err := url.Parse(h); err == nil {
			template.URIs = append(template.URIs, uri)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}
	certOut, err := os.Create(certFile)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", certFile, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("failed to write data to %s: %s", certFile, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("error closing %s: %s", certFile, err)
	}
	log.Printf("wrote %s\n", certFile)
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("failed to open %s for writing: %s", keyFile, err)
		return
	}
	if err := pem.Encode(keyOut, pemBlockForKey(priv)); err != nil {
		log.Fatalf("failed to write data to %s: %s", keyFile, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("error closing %s: %s", keyFile, err)
	}
	log.Printf("wrote %s\n", keyFile)
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
func browse(c *opcua.Client, v *opcua.Node, devID string) {
	defer func() {
		if r := recover(); r != nil {
			HasBrowse.Lock.Lock()
			HasBrowse.DataMap[devID] = false
			HasBrowse.Lock.Unlock()
			fmt.Println(`[browse error]`, r)
		}
	}()
	node := ua.NewNumericNodeID(v.ID.Namespace(), v.ID.IntID())
	n := c.Node(node)
	attrs, err := n.Attributes(ua.AttributeIDBrowseName)
	if err != nil {
		panic(err)
	}
	browseName := attrs[0].Value.String()
	if _, ok := AllBrowseData.DataMap[devID]; !ok {
		AllBrowseData.DataMap[devID] = make(map[string]*opcua.Node)
	}

	if k, err := n.NodeClass(); err == nil {
		if k.String() == "NodeClassVariable" {
			AllBrowseData.DataMap[devID][browseName] = v
			BrowseNameMap.DataMap[devID] = append(BrowseNameMap.DataMap[devID], browseName)
		}
	}
	// Get children
	// TODO 確認是否有 HasChild 之外的參數可取得所有 browseName
	refs, err := n.ReferencedNodes(id.HasChild, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	if err != nil {
		panic(err)
	}
	for _, rn := range refs {
		browse(c, rn, devID)
	}
}

// BrowseAll 瀏覽所有節點
func BrowseAll(c *opcua.Client, devID string) {
	defer func() {
		if r := recover(); r != nil {
			HasBrowse.Lock.Lock()
			HasBrowse.DataMap[devID] = false
			HasBrowse.Lock.Unlock()
			fmt.Println(`[browse all error]`, r)
		}
	}()
	node := ua.NewNumericNodeID(0, 85) // root node id
	n := c.Node(node)
	refs, err := n.ReferencedNodes(id.Organizes, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	if err != nil {
		panic(err)
	}
	BrowseNameMap.Lock.Lock()
	AllBrowseData.Lock.Lock()
	for _, v := range refs {
		browse(c, v, devID)
	}
	if len(BrowseNameMap.DataMap) > 0 {
		HasBrowse.DataMap[devID] = true
	}
	BrowseNameMap.Lock.Unlock()
	AllBrowseData.Lock.Unlock()
}

// GetActVal 取得實際值
func GetActVal(devID, nodeID, nsStr, browse string, dev nodeattr.DevInfo) (string, error) {
	var conv nodeattr.Converter
	var con nodeattr.ConInfo
	if conArr, err := routerservice.GetAllConInfo(); err != nil {
		return "", err
	} else {
		for _, v := range conArr {
			if v.ID == dev.ConID {
				con = v
				break
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var endpoint string
		if con.Port == "80" {
			endpoint = nodeattr.OpcTitle + con.IP
		} else {
			endpoint = nodeattr.OpcTitle + con.IP + ":" + con.Port
		}
		if c, conErr := ConnectServer(ctx, con.Account, con.Password, endpoint); conErr != nil {
			return "", conErr
		} else { // 可連到連線
			defer c.Close()
			if nodeID != "" && nsStr != "" {
				ns, err := strconv.ParseInt(nsStr, 10, 64)
				if err != nil {
					return "", err
				}
				conv.SrcNodeid = nodeID
				conv.SrcNamespace = int(ns)
				val, _ := ReadNodeByID(c, conv)
				return val, nil
			} else if browse != "" {
				conv.SrcBrowse = browse
				val, _ := ReadNodeByBrowseName(c, devID, conv)
				return val, nil
			} else {
				return "", nil
			}
		}
	}
}
