package wiseservice

import (
	"FSRV_Edge/dao/settingdao"
	"FSRV_Edge/datacollect"
	"FSRV_Edge/global"
	"FSRV_Edge/go-cast/controllers"
	"FSRV_Edge/googlehome"
	"FSRV_Edge/influx"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/lib"
	"FSRV_Edge/models/collect"
	"FSRV_Edge/models/dcinfo"
	"FSRV_Edge/models/mac"
	"FSRV_Edge/models/machineio"
	"FSRV_Edge/models/status"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/opcuaservice"
	"FSRV_Edge/service/routerservice"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mostlygeek/arp"
)

const (
	root                = "/ioms5data"
	dcLogMessageAPI     = "/log_message"
	dcOutputAPI         = "/log_output"
	dcMacStr            = "00D0C9"
	wiseBrandStr        = "WISE-4051/Advantech"
	defaultAccount      = "root"
	defaultPwd          = "mitroot"
	interval            = 1000
	idleTime            = 300
	idleTimeStr         = "300"
	lastTimeStr         = "10800"
	noRowErr            = "no row found"
	disConMsg           = 0
	collectInfoAPI      = root + "/datacollect/machine"
	fetchTimeAPI        = root + "/datacollect/fetchInterval"
	statusReturnAPI     = root + "/datacollect/statusReturn"
	lastDcDataAPI       = root + "/datacollect/lastDcData"
	dcDataAPI           = root + "/datacollect/dcData"
	rdDataAPI           = root + "/datacollect/rd"
	dsDataAPI           = root + "/datacollect/ds"
	deviceInfoAPI       = root + "/datacollect/macinfo"
	checkMachineIdleAPI = root + "/datacollect/checkidle"
	testAPI             = root + "/test"
)

var tr *http.Transport
var dcTr *http.Transport
var addrs []net.Addr
var doingMap sync.Map
var macAddressIPMap sync.Map
var ipArr []string
var deviceMap macMapLock

func (c *macMapLock) put(info mac.Info) {
	c.mutex.Lock()
	c.dataMap[info.MacAddress] = info
	c.mutex.Unlock()
}
func (c *macMapLock) get(macAddress string) (info mac.Info) {
	c.mutex.Lock()
	info = c.dataMap[macAddress]
	c.mutex.Unlock()
	return
}
func (c *macMapLock) delete(macAddress string) {
	c.mutex.Lock()
	delete(c.dataMap, macAddress)
	c.mutex.Unlock()
	return
}

func (c *macMapLock) getAll() (arr []mac.Info) {
	c.mutex.Lock()
	for _, val := range c.dataMap {
		arr = append(arr, val)
	}
	c.mutex.Unlock()
	return
}

type macLock struct {
	arr   []mac.Info
	mutex sync.Mutex
}

func (c *macLock) append(info mac.Info) {
	c.mutex.Lock()
	c.arr = append(c.arr, info)
	c.mutex.Unlock()
}

type macMapLock struct {
	dataMap map[string]mac.Info
	mutex   sync.Mutex
}

// now 現在的時間(微秒)
func now() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// now 現在的時間(微秒)
func nowSecond() int64 {
	return time.Now().Unix()
}

var (
	globalLogger = initlog.GetLogger()
	AllDcData    dcData
	lastDcMap    sync.Map
	errtimeData  errTime
	dcMoldCount  errTime
)

type dcData struct {
	DataMap map[string]nodeattr.HistoryData
	Lock    sync.Mutex
}
type errTime struct {
	DataMap map[string]int
	Lock    sync.Mutex
}

func init() {
	errtimeData.Lock.Lock()
	errtimeData.DataMap = make(map[string]int)
	errtimeData.Lock.Unlock()
	AllDcData.Lock.Lock()
	AllDcData.DataMap = make(map[string]nodeattr.HistoryData)
	AllDcData.Lock.Unlock()
	dcMoldCount.Lock.Lock()
	dcMoldCount.DataMap = make(map[string]int)
	dcMoldCount.Lock.Unlock()
}
func writeDcData(devId string, cycleTime float64, status int) {
	var node opcuaservice.WriteAttr
	node.Value = cycleTime
	node.NodeID = "Device" + devId + "." + nodeattr.CycleTimeStr
	opcuaservice.WriteVariable(node)
	node.Value = status
	node.NodeID = "Device" + devId + "." + nodeattr.StatusStr
	opcuaservice.WriteVariable(node)
}

// UpdateDcStatus 更新DC狀態
func UpdateDcStatus() {
	start := time.Now()
	var dev nodeattr.DevInfo
	go tryToAddNewCollection()
	go startPassMacInfo(passMacInfo)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("[Update dc status error]", err)
		}
	}()
	mac, ip := getDcIP()
	elapsed := time.Since(start)
	if elapsed.Milliseconds() > 0 {
		log.Printf("UpdateDcStatus took %s", elapsed)
	}
	// 需要檢查現有設備列表DC
	if len(mac) == 0 || len(ip) == 0 {
		return
	}
	for _, v := range global.Devs {
		if mac == v.Mac && v.Protocol == global.HttpStr {
			return
		}
	}
	dev.Protocol = global.HttpStr
	dev.ConName = ip
	dev.Mac = mac
	dev.ID = 0
	dev.Brand = wiseBrandStr
	dev.TempID = routerservice.DefaultTemp.ID
	dev.TempName = routerservice.DefaultTemp.Name
	go firstClean2(dev)
	// checkMachineIdle(mac, idleTimeStr)
	AllDcData.Lock.Lock()
	dev.Status = AllDcData.DataMap[mac].MachineStatus
	AllDcData.Lock.Unlock()
	id, err := routerservice.InsertDevInfo(dev)
	if err == nil {
		devName := "Device" + strconv.Itoa(int(id))
		opcuaservice.BuildAllDevNode(devName)
	}

}
func CheckNowDcStatus() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(`[CheckNowDcStatus]`, err)
		}
	}()
	for _, v := range global.Devs {
		if v.Protocol == global.HttpStr {
			checkMachineIdle2(v.Mac, idleTimeStr)
			AllDcData.Lock.Lock()
			status := AllDcData.DataMap[v.Mac].MachineStatus
			AllDcData.Lock.Unlock()
			if v.Status == datacollect.Stop && status == 0 {
				v.Status = datacollect.Stop
			} else {
				v.Status = status
			}
			routerservice.UpdateDevStatusByMac(v)
		}
	}
}

func firstClean2(dev nodeattr.DevInfo) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(`[first clean]`, err)
		}
	}()
	var arr []datacollect.MachineIO
	// 取得目前Dc設定
	dcSettingBody, getDcBodyErr := getDcBody2(dev.ConName)
	if getDcBodyErr != nil {
		errtimeData.Lock.Lock()
		errtimeData.DataMap[dev.Mac]++
		errTimes := errtimeData.DataMap[dev.Mac]
		errtimeData.Lock.Unlock()
		if errTimes%6 == 0 {
			//跟後台說停機瞜
			//停機後清掉暫存
			errtimeData.Lock.Lock()
			errtimeData.DataMap[dev.Mac]--
			errtimeData.Lock.Unlock()
			lastDcMap.Delete(dev.Mac)
			dev.Status = datacollect.Stop
			routerservice.UpdateDevStatusByMac(dev)
		}
		return
		// panic(getDcBodyErr.Error() + " error times:" + strconv.Itoa(errTimes))
	}
	errtimeData.Lock.Lock()
	errtimeData.DataMap[dev.Mac] = 0
	errtimeData.Lock.Unlock()
	// 去di table取得最後一筆狀態的時間
	// lastTime, getLastTimeErr = datacollectdao.GetDataLastTime(machineNumber, o)
	var min, max int64
	lastTime, getLastTimeErr := influx.GetDataLastTime(dev.Mac)
	if getLastTimeErr != nil {
		panic(getLastTimeErr)
	}
	lastTime /= 1000
	if lastTime == 0 {
		min = time.Now().UnixNano()/int64(time.Second) - interval
	} else {
		min = lastTime + 1
	}
	max = min + interval
	//===================================================================
	if max > dcSettingBody.TFst-2 {
		max = dcSettingBody.TFst - 2
	}
	if min < dcSettingBody.TLst {
		min = dcSettingBody.TLst
		max = min + 10
	}
	fetchTime := ((max - min) + 1)
	if fetchTime <= 0 {
		panic("no need to fetch")
	}
	var body datacollect.SettingBody //Body 是sbody類型並給值
	body.UID = 1
	body.MAC = 0
	body.TmF = 0
	body.Fltr = 1
	body.TSt = min
	body.TEnd = max
	body.Amt = 0
	// =========================================================
	if err := putDc2(dev.ConName, body); err != nil {
		panic(err)
	} else {
		if err := checkDc2(dev.ConName, body); err != nil {
			panic(err)
		}
		if msgs, err := getDc2(dev.ConName); err != nil { // 無法連線至 WISE IP
			panic(err)
		} else {
			for _, msg := range msgs.LogMsg {
				arr = append(arr, toDc2(msg, dev.Mac))
			}
			if createDcDataErr := CreateDcData(dev.Mac, lastTimeStr, idleTimeStr, arr); err != nil {
				panic(createDcDataErr)
			}
		}
	}
}

func putDc2(ip string, info datacollect.SettingBody) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(`[Put DC]`, err)
		}
	}()
	b, marshalErr := json.Marshal(info) //轉為json
	if marshalErr != nil {
		fmt.Println("227", marshalErr)
		return marshalErr
	}
	payload := bytes.NewBuffer(b)
	url := "http://" + ip + dcOutputAPI
	req, newRequestErr := http.NewRequest("PUT", url, payload)
	if newRequestErr != nil {
		return newRequestErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(defaultAccount, defaultPwd)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		Transport: global.DcTr,
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		return doErr
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code error")
	}
	return nil
}

func getDc2(ip string) (datacollect.WiseMsg, error) {
	var message datacollect.WiseMsg
	url := "http://" + ip + dcLogMessageAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return message, requestErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(defaultAccount, defaultPwd)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		Transport: global.DcTr,
		//Timeout:   time.Duration(40 * time.Second),
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		return message, doErr
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return message, fmt.Errorf("error status code:" + strconv.Itoa(resp.StatusCode) + " error message:" + resp.Status)
	}

	if resp.Body == nil {
		return message, fmt.Errorf("response body is nil")
	}

	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return message, err
	}
	return message, nil
}

func getDcBody(info collect.Info, ip string) (dcInfo dcinfo.SettingBody, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	url := "http://" + ip + dcOutputAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		panic(requestErr)
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", info.DcAuthorization)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: dcTr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		panic(resp.Status)
	}

	body, readAllErr := ioutil.ReadAll(resp.Body)

	if readAllErr != nil {
		panic(readAllErr)
	}

	if err := json.Unmarshal(body, &dcInfo); err != nil {
		panic(err)
	}
	return
}
func getDcBody2(ip string) (dcInfo datacollect.SettingBody, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	url := "http://" + ip + dcOutputAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		panic(requestErr)
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(defaultAccount, defaultPwd)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: global.DcTr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	if err := json.Unmarshal(body, &dcInfo); err != nil {
		panic(err)
	}
	return
}
func toDc2(msg datacollect.LogMsg, mac string) datacollect.MachineIO {
	var dcData datacollect.MachineIO
	dcData.Di0 = msg.Record[0][3]
	dcData.Di1 = msg.Record[1][3]
	dcData.Di2 = msg.Record[2][3]
	dcData.Di3 = msg.Record[3][3]
	dcData.Di4 = msg.Record[4][3]
	dcData.Di5 = msg.Record[5][3]
	dcData.Di6 = msg.Record[6][3]
	dcData.Di7 = msg.Record[7][3]
	dcData.InfluxMeasurement = influx.DiTable + mac
	i, _ := strconv.ParseInt(msg.TIM, 10, 64)
	//轉成微秒
	dcData.Timestamp = i * 1000
	return dcData
}
func toDc(msg dcinfo.LogMsg) machineio.MachineIO {
	var dcData machineio.MachineIO
	dcData.Di0 = msg.Record[0][3]
	dcData.Di1 = msg.Record[1][3]
	dcData.Di2 = msg.Record[2][3]
	dcData.Di3 = msg.Record[3][3]
	dcData.Di4 = msg.Record[4][3]
	dcData.Di5 = msg.Record[5][3]
	dcData.Di6 = msg.Record[6][3]
	dcData.Di7 = msg.Record[7][3]
	i, _ := strconv.ParseInt(msg.TIM, 10, 64)
	//轉成微秒
	dcData.Timestamp = i * 1000
	return dcData
}

func getDcIP() (macVal string, ipVal string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("arp error ", err)
		}
	}()
	if err := exec.Command("arp", "-da").Run(); err != nil {
		fmt.Println("arp:" + err.Error())
	}
	arp.CacheUpdate()
	var addrs []net.Addr
	var err error
	if addrs, err = net.InterfaceAddrs(); err != nil {
		panic(err)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				i := strings.LastIndex(ip, ".")
				if i != -1 {
					ip = ip[:i+1]
				}
				for idx := 1; idx < 255; idx++ {
					go pingTCP(ip + strconv.Itoa(idx))
				}
			}
		}
	}
	ipMap := make(map[string]string)
	for ip := range arp.Table() {
		mac := strings.Replace(arp.Search(ip), ":", "", -1)
		mac = strings.ToUpper(mac)
		startsWith := strings.HasPrefix(mac, dcMacStr)
		if startsWith {
			macVal = mac
			ipVal = ip
			ipMap[mac] = ip
			// fmt.Println(ip, mac)
			macAddressIPMap.Store(mac, ip)
		}
	}
	macAddressIPMap.Range(func(k, v interface{}) bool {
		mac := k.(string)
		if ipMap[mac] == "" {
			macAddressIPMap.Delete(mac)
		}
		return true
	})
	return
}

func checkDc(info collect.Info, ip string, min, max int64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	url := "http://" + ip + dcOutputAPI
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", info.DcAuthorization)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		Transport: dcTr,
		Timeout:   time.Duration(20 * time.Second),
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			panic(readErr)
		}
		var settingBody = new(dcinfo.SettingBody)
		if err := json.Unmarshal(body, &settingBody); err != nil {
			panic(err)
		}
		if settingBody.Fltr == 1 &&
			settingBody.MAC == 0 &&
			settingBody.Total != 0 &&
			settingBody.TSt == min &&
			settingBody.TEnd == max {
			return nil
		}
		msg := "check dc setting failed 。 dc time interval:" + strconv.FormatInt(settingBody.TSt, 10) + " to " + strconv.FormatInt(settingBody.TEnd, 10)
		panic(msg)
	}
	panic(resp.Status)
}
func checkDc2(ip string, info datacollect.SettingBody) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	url := "http://" + ip + dcOutputAPI
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(defaultAccount, defaultPwd)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		Transport: global.DcTr,
		Timeout:   time.Duration(20 * time.Second),
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			panic(readErr)
		}
		var settingBody = new(datacollect.SettingBody)
		if err := json.Unmarshal(body, &settingBody); err != nil {
			panic(err)
		}
		if settingBody.Fltr == 1 &&
			settingBody.MAC == 0 &&
			settingBody.Total != 0 &&
			settingBody.TSt == info.TSt &&
			settingBody.TEnd == info.TEnd {
			return nil
		}
		msg := "check dc setting failed 。 dc time interval:" + strconv.FormatInt(settingBody.TSt, 10) + " to " + strconv.FormatInt(settingBody.TEnd, 10)
		panic(msg)
	}
	panic(resp.Status)
}

func pingTCP(ip string) {
	// TODO 造成goroutine io wait
	conn, err := net.DialTimeout("tcp", ip+":80", 20*time.Second)
	if err == nil {
		defer conn.Close()
	}
}

// CreateDcData 新增一次清洗資料 並且 更新最後的時間
func CreateDcData(mac, lastTimeStr, idleTimeStr string, arr []datacollect.MachineIO) (err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(`[CreateDcData]`, err)
		}
	}()
	lastTime, parseIntErr := strconv.ParseInt(lastTimeStr, 10, 64)
	idleTime, parseIdleErr := strconv.ParseInt(idleTimeStr, 10, 64)
	if parseIntErr != nil || parseIdleErr != nil {
		err = fmt.Errorf("Convert integer error")
		return
	}
	settingdao.MacData.Lock.Lock()
	tmp := settingdao.MacData.DataMap[mac]
	settingdao.MacData.Lock.Unlock()

	if mac == "" {
		return fmt.Errorf("no machine number")
	} else if lastTime == 0 {
		return fmt.Errorf("no last time")
	}
	if err := influx.WriteDi(arr); err != nil {
		return err
	}
	lastStatus, getLastStatusError := influx.GetLastStatus(tmp)
	lastStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
	if getLastStatusError != nil {
		return getLastStatusError
	}
	analyzeArr, analyzeErr := analyzeDevStatus(mac, idleTime, arr, lastStatus)
	if analyzeErr != nil {
		return analyzeErr
	}
	if err := influx.WriteStatus(analyzeArr); err != nil {
		return err
	}
	dcMoldCount.Lock.Lock()
	dcMoldCount.DataMap[tmp]++
	dcMoldCount.Lock.Unlock()
	if len(analyzeArr) == 0 {
		AllDcData.Lock.Lock()
		AllDcData.DataMap[mac] = lastStatus
		AllDcData.Lock.Unlock()
	} else {
		AllDcData.Lock.Lock()
		AllDcData.DataMap[mac] = analyzeArr[len(analyzeArr)-1]
		AllDcData.Lock.Unlock()
	}
	// insertStatusErr := wisedao.InsertMultiStatusData(mac, analyzeArr, o)
	//若是現在距離最後一筆狀態太遠 而且採集的資料是一小時內的 那麼寫入一筆狀態為無法識別的資料
	if now()-lastStatus.Timestamp >= int64(time.Hour/time.Millisecond) && (nowSecond()-lastTime) <= int64(time.Hour/time.Second) {
		var status nodeattr.HistoryData
		status.MachineStatus = datacollect.DcAbnormal
		status.InfluxMeasurement = influx.NodeHdaMeasure + tmp
		status.Timestamp = now() - int64(time.Hour/time.Millisecond)
		global.LastStatusMap.Store(mac, status)
	}
	return err
}

func analyzeDevStatus(mac string, idleTime int64, arr []datacollect.MachineIO, lastStatus nodeattr.HistoryData) (analyzeArr []nodeattr.HistoryData, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if len(arr) == 0 {
		return analyzeArr, fmt.Errorf("Input DIO is empty")
	}
	noStatusDcArr, getNoStatusDcArrErr := influx.GetDcDataAfterTimestamp(mac, lastStatus.Timestamp)
	if getNoStatusDcArrErr != nil {
		return analyzeArr, getNoStatusDcArrErr
	}
	arr = append(arr, noStatusDcArr...)
	analyzeArr = getStatusByCleanDc(mac, arr, idleTime)
	return
}

func cleanDc(info collect.Info, wiseMessage dcinfo.WiseMessage, lastData machineio.MachineIO) (arr []machineio.MachineIO, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var tempStr = getCompareStr(lastData)
	var tempTime int64
	var index int64

	for _, value := range wiseMessage.LogMsg {
		if len(value.UID) > 12 {
			tmp := value.UID[len(value.UID)-12 : len(value.UID)]
			if tmp != info.MacAddress {
				return arr, errors.New("mac address not the same , " + tmp + " ,  " + info.MacAddress)
			}
		}
		tempDc := toDc(value)
		if tempDc.Di1 == 1 && tempDc.Di2 == 1 {
			continue
		}
		temp := getCompareStr(tempDc)
		if value.SysTk == 0 {
			if tempTime != tempDc.Timestamp {
				index = 0
				tempTime = tempDc.Timestamp

			} else {
				index += 100
				if index > 900 {
					index = 950
				}
			}
		} else {

			digit := value.SysTk / 100
			digit = digit % 10
			index = digit * 100
		}
		tempDc.Timestamp = tempDc.Timestamp + index
		if temp != tempStr {
			arr = append(arr, tempDc)
			tempStr = temp
		}

	}

	return arr, nil
}

func checkMachineIdle2(mac, idleTimeStr string) (err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(`[checkMachineIdle]`, err)
		}
	}()
	if mac == "" {
		return errors.New("mac is empty")
	}
	settingdao.MacData.Lock.Lock()
	tmp := settingdao.MacData.DataMap[mac]
	settingdao.MacData.Lock.Unlock()
	idleTime, err := strconv.ParseInt(idleTimeStr, 10, 64)
	if err != nil {
		return err
	}
	var lastStatus nodeattr.HistoryData
	var lastData datacollect.MachineIO
	if v, ok := global.LastStatusMap.Load(mac); ok {
		lastStatus = v.(nodeattr.HistoryData)
	}
	if v, ok := global.LastDataMap.Load(mac); ok {
		lastData = v.(datacollect.MachineIO)
	}
	if lastData.Timestamp == 0 {
		if data, err := influx.GetLastDcData(mac); err != nil {
			panic(err)
		} else {
			if data.Timestamp == 0 {
				data.Di0 = 1
				data.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
			}
			lastData = data
		}

	}
	var nowStamp = time.Now().UnixNano() / int64(time.Millisecond)
	if lastData.Di0 == 0 && lastData.Di1 == 0 && lastData.Di2 == 0 && lastData.Di3 == 0 && lastData.Di4 == 0 && lastData.Di5 == 0 && lastData.Di6 == 0 && lastData.Di7 == 0 && lastData.Timestamp != 0 {
		return nil
	}

	if nowStamp-lastData.Timestamp > idleTime*1000 {
		if lastStatus.MachineStatus == datacollect.Idle && nowStamp-lastStatus.Timestamp <= idleTime*1000 {
			return nil
		}
		var statusData nodeattr.HistoryData
		var statusDataArr []nodeattr.HistoryData
		statusData.InfluxMeasurement = influx.NodeHdaMeasure + tmp
		statusData.MachineStatus = datacollect.Idle
		statusData.CycleTime = 0
		statusData.Timestamp = nowStamp
		statusDataArr = append(statusDataArr, statusData)
		if err := influx.WriteStatus(statusDataArr); err != nil {
			return err
		}
		writeDcData(tmp, statusData.CycleTime, statusData.MachineStatus)
		dcMoldCount.Lock.Lock()
		dcMoldCount.DataMap[tmp]++
		dcMoldCount.Lock.Unlock()
		return nil
	}
	lastStatusData, getLastStatusDataErr := influx.GetLastStatus(tmp)
	if getLastStatusDataErr == nil && lastData.Timestamp != 0 && lastData.Timestamp-lastStatusData.Timestamp > 60*60*1000 {
		if v, ok := global.LastStatusMap.Load(mac); ok {
			tmpStatus := v.(nodeattr.HistoryData)
			tmpStatus.MachineStatus = datacollect.DcAbnormal
		}
	}
	return nil
}
func getStatusByCleanDc(mac string, arr []datacollect.MachineIO, idleTime int64) (analyzeArr []nodeattr.HistoryData) {
	// 當arr為空時，到DB取最後一筆狀態放入即時資料
	var tempArr []datacollect.MachineIO
	var nextTimestamp int64
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Timestamp < arr[j].Timestamp
	})
	settingdao.MacData.Lock.Lock()
	tmp := settingdao.MacData.DataMap[mac]
	settingdao.MacData.Lock.Unlock()
	var firstTime int64
	for i, value := range arr {
		if value.InfluxMeasurement == "" {
			continue
		}
		if (value.Di0 + value.Di1 + value.Di2 + value.Di3 +
			value.Di4 + value.Di5 + value.Di6 + value.Di7) == 0 {
			var tmpStatus nodeattr.HistoryData
			tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
			tmpStatus.CycleTime = 0
			tmpStatus.MachineStatus = datacollect.Stop
			tmpStatus.Timestamp = value.Timestamp
			analyzeArr = append(analyzeArr, tmpStatus)
			writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
			continue
		} else if value.Di3 == 1 {
			var tmpStatus nodeattr.HistoryData
			tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
			tmpStatus.Timestamp = value.Timestamp
			tmpStatus.CycleTime = 0
			tmpStatus.MachineStatus = datacollect.Abnormal
			analyzeArr = append(analyzeArr, tmpStatus)
			writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
		} else if value.Di4 == 1 {
			var tmpStatus nodeattr.HistoryData
			tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
			tmpStatus.Timestamp = value.Timestamp
			tmpStatus.CycleTime = 0
			tmpStatus.MachineStatus = datacollect.NG
			analyzeArr = append(analyzeArr, tmpStatus)
			writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
		}
		if i < len(arr)-1 {
			nextTimestamp = arr[i+1].Timestamp
			if nextTimestamp-value.Timestamp > idleTime*1000 {
				var tmpStatus nodeattr.HistoryData
				tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
				tmpStatus.Timestamp = value.Timestamp + idleTime*1000
				tmpStatus.CycleTime = 0
				tmpStatus.MachineStatus = datacollect.Idle
				analyzeArr = append(analyzeArr, tmpStatus)
				writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
				continue
			}
		}
		if (value.Di0+value.Di1+value.Di2+value.Di3+
			value.Di4+value.Di5+value.Di6+value.Di7) != 0 && value.Di3 != 1 && value.Di4 != 1 {
			if len(tempArr) == 0 && value.Di2 == 1 {
				tempArr = append(tempArr, value)
			} else if len(tempArr) != 0 {
				lastValue := tempArr[len(tempArr)-1]
				if ((lastValue.Di2 == 1) && (value.Di1 == 1)) || ((lastValue.Di1 == 1) && (value.Di2 == 1)) {
					tempArr = append(tempArr, value)
				}
			}
			if len(tempArr) == 3 {
				if i >= 3 && arr[i-3].Di2 == 1 {
					firstTime = arr[i-3].Timestamp
				} else {
					firstTime = tempArr[0].Timestamp
				}
				var tmpStatus nodeattr.HistoryData
				tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
				tmpStatus.Timestamp = tempArr[2].Timestamp
				var diff = (tempArr[2].Timestamp - firstTime)
				tmpStatus.CycleTime = float64(diff) / float64(1000)
				tmpStatus.MachineStatus = datacollect.Running
				tempArr = tempArr[2:3]
				if tmpStatus.CycleTime > float64(idleTime) {
					continue
				}
				// 存SCT
				var sctArr []nodeattr.HistoryData
				if v, ok := global.SCTMap.Load(mac); ok {
					sctArr = v.([]nodeattr.HistoryData)
					tmpStatus.CycleTime = lib.SCTCaculator(sctArr, 0.5)
				}
				sctArr = append(sctArr, tmpStatus)
				if len(sctArr) > 15 {
					sctArr = sctArr[len(sctArr)-15:]
				}
				analyzeArr = append(analyzeArr, tmpStatus)
				writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
			}
		}
	}
	if len(arr) != 0 {
		lastData := arr[len(arr)-1]
		if lastData.Di0 == 1 && lastData.Timestamp < (time.Now().UnixNano()/int64(time.Millisecond)-idleTime*1000) {
			var tmpStatus nodeattr.HistoryData
			tmpStatus.InfluxMeasurement = influx.NodeHdaMeasure + tmp
			tmpStatus.Timestamp = lastData.Timestamp + idleTime*1000
			tmpStatus.CycleTime = 0
			tmpStatus.MachineStatus = datacollect.Idle
			analyzeArr = append(analyzeArr, tmpStatus)
			writeDcData(tmp, tmpStatus.CycleTime, tmpStatus.MachineStatus)
		}
	}
	return
}

func check(machineNumber string) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			globalLogger.Criticalf("check err %s", err)
		}
	}()

	for {
		if value, ok := doingMap.Load(machineNumber); ok {
			info := value.(collect.Info)
			checkMachineIdle(info)
		} else {
			break
		}
		scanInterval := 5000 + rand.Intn(3000)
		time.Sleep(time.Duration(scanInterval) * time.Millisecond)
	}

}

func checkMachineIdle(info collect.Info) error {
	var url = "http://" + global.MainIP + checkMachineIdleAPI
	req, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		return newRequestErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("machineNumber", info.MachineNumber)
	req.Header.Set("idleTime", strconv.Itoa(info.IdleTime))

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		return doErr
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return nil
}
func startPassMacInfo(f func()) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			globalLogger.Criticalf("startPassMacInfo err %s", err)
		}
	}()
	for {
		f()
		time.Sleep(time.Minute)
	}
}

func passMacInfo() {
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			globalLogger.Criticalf("passMacInfo err %s", err)
		}
	}()

	tmpMap := make(map[string]string)
	macAddressIPMap.Range(func(k, v interface{}) bool {
		mac := k.(string)
		ip := v.(string)
		if mac != "" && ip != "" {
			tmpMap[mac] = ip
		}

		return true
	})
	var wg sync.WaitGroup
	for k, v := range tmpMap {
		if deviceMap.get(k).MacAddress == "" {
			wg.Add(1)
			go checkType(&wg, k, v)
		}

	}

	wg.Wait()
	passMac(global.MainIP)
	go passMac(global.BackUpIP)
}

func checkType(wg *sync.WaitGroup, macAddress, ip string) {

	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			globalLogger.Criticalf("checkType err %s", err)
		}
	}()
	defer wg.Done()

	var m mac.Info
	startsWith := strings.HasPrefix(macAddress, "00D0C9")
	if startsWith {
		m.Type = "DC"
		m.MacAddress = macAddress
		m.IP = ip
		m.Group = global.WorkShopNumber
		deviceMap.put(m)
	} else {
		isDs, name := checkDs(ip)
		if isDs {
			m.Type = "DS"
			m.MacAddress = macAddress
			m.IP = ip
			m.Group = global.WorkShopNumber
			m.Name = name
			deviceMap.put(m)
		}
	}
}

func checkDs(ip string) (bool, string) {
	var url = "http://" + ip + ":8008/ssdp/device-desc.xml"
	req, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		return false, ""
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		return false, ""
	}

	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		return false, ""
	}
	info := new(mac.GoogleInfo)
	if err := xml.Unmarshal(body, &info); err != nil {
		return false, ""
	}

	if info.Device.Manufacturer == "Google Inc." {
		return true, info.Device.FriendlyName
	}
	return false, ""
}
func passMac(ip string) {
	info := deviceMap.getAll()

	if ip == "" || len(info) == 0 {
		return
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			globalLogger.Criticalf("main err %s", err)
		}
	}()
	var url = "http://" + ip + deviceInfoAPI
	b, marshalErr := json.Marshal(info)
	if marshalErr != nil {
		panic(marshalErr)
	}
	req, newRequestErr := http.NewRequest("POST", url, bytes.NewBuffer(b))

	if newRequestErr != nil {
		panic(newRequestErr)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: tr,
	}
	if resp, err := client.Do(req); err != nil {
		panic(err)
	} else {
		defer resp.Body.Close()
	}
}

func showCollectInfo(arr []collect.Info) {
	globalLogger.Debugf("init finish 。 dc num : %d ", len(arr))
}

// getRDInfo 取得所有的RD 資料
func getRDInfo() (arr []collect.RD, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + rdDataAPI
	req, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		panic(newRequestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("workShopNumber", global.WorkShopNumber)

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	if err := json.Unmarshal(body, &arr); err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	return
}

// getDSInfo 取得所有的DS 資料
func getDSInfo() (arr []collect.DS, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + dsDataAPI
	req, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		panic(newRequestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("workShopNumber", global.WorkShopNumber)

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	if err := json.Unmarshal(body, &arr); err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	return
}

// getCollectionInfo 取得所有的DC 資料
func getCollectionInfo() (arr []collect.Info, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + collectInfoAPI

	req, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		panic(newRequestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("workShopNumber", global.WorkShopNumber)

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	var res collect.DatacollectRes
	if err := json.Unmarshal(body, &res); err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(res.Response)
	}
	return res.Info, err
}

func castStatus(cli *googlehome.Client) (status *controllers.ReceiverStatus, getStatusErr error) {
	defer func() {
		if err := recover(); err != nil {
			globalLogger.Criticalf("cast status error %s", err)
		}
	}()
	defer cli.Client.Close()

	status, getStatusErr = cli.GetStatus()

	return
}

func checkCast(ip string, value collect.DS) {
	cli, err := googlehome.NewClientWithConfig(googlehome.Config{
		Hostname: ip,
		Lang:     "en",
		Accent:   "GB",
	})
	if err != nil {
		go updateDSStatus(value.MacAddress, "OFFLINE")
		return
	}
	status, _ := castStatus(cli)
	var statusText string
	var url string
	if status == nil {
		go updateDSStatus(value.MacAddress, "AVAILABLE")
	} else {

		if len(status.Applications) > 0 {
			statusText = *status.Applications[0].StatusText
		}
		statusText = strings.ToUpper(statusText)

		if strings.Contains(statusText, "NOW PLAYING: ") {
			go updateDSStatus(value.MacAddress, "PLAYING")
		} else {
			go updateDSStatus(value.MacAddress, "AVAILABLE")
		}
		statusText = strings.Replace(statusText, "NOW PLAYING: ", "", 1)
	}
	url = strings.ToUpper(value.URL)
	if url == "" && statusText != "" {
		go castQuit(cli)
	} else if url != "" && url != statusText {
		go castURL(cli, value.URL, "iframe")
	}
}

func updateDSStatus(macAddress, status string) error {
	var url = "http://" + global.MainIP + dsDataAPI
	req, newRequestErr := http.NewRequest("PUT", url, nil)
	if newRequestErr != nil {
		return newRequestErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("macAddress", macAddress)
	req.Header.Set("status", status)

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		return doErr
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

func castURL(cli *googlehome.Client, url, castType string) {
	defer func() {
		if err := recover(); err != nil {
			globalLogger.Criticalf("cast url error %s", err)
		}
	}()
	defer cli.Client.Close()
	cli.URL(url, castType)
}

func castQuit(cli *googlehome.Client) {
	defer func() {
		if err := recover(); err != nil {
			globalLogger.Criticalf("cast quit error %s", err)
		} else {
			cli.Client.Close()
		}
	}()

	go cli.QuitApp()
}

func saveDc(info collect.Info, dcArr []machineio.MachineIO, lastTime int64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + dcDataAPI
	b, marshalErr := json.Marshal(dcArr)
	if marshalErr != nil {
		panic(marshalErr)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("machineNumber", info.MachineNumber)
	req.Header.Set("lastTime", strconv.FormatInt(lastTime, 10))
	req.Header.Set("idleTime", strconv.Itoa(info.IdleTime))

	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		panic("response body is empty")
	}
	var res = new(collect.Response)

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(res.Response)
	}

	return err

}

func getCompareStr(dcData machineio.MachineIO) string {
	return strconv.Itoa(dcData.Di0) + "_" +
		strconv.Itoa(dcData.Di1) + "_" +
		strconv.Itoa(dcData.Di2) + "_" +
		strconv.Itoa(dcData.Di3) + "_" +
		strconv.Itoa(dcData.Di4) + "_" +
		strconv.Itoa(dcData.Di5) + "_" +
		strconv.Itoa(dcData.Di6) + "_" +
		strconv.Itoa(dcData.Di7)
}

func statusReturn(info collect.Info, status int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + statusReturnAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		panic(requestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("machineNumber", info.MachineNumber)
	req.Header.Set("status", strconv.Itoa(status))

	client := &http.Client{Transport: tr}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	var res collect.FetchTimeInfo
	if err := json.Unmarshal(body, &res); err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(res.Response)
	}
	return

}

func getFetchTimeInterval(info collect.Info) (min int64, max int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + fetchTimeAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		panic(requestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("machineNumber", info.MachineNumber)

	client := &http.Client{Transport: tr}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}
	var res collect.FetchTimeInfo
	if err := json.Unmarshal(body, &res); err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(res.Response)
	}
	return res.Min, res.Max, err

}
func firstCleanLoop(machineNumber string) {

	errorTimes := new(uint)
	for {
		if value, ok := doingMap.Load(machineNumber); ok {
			info := value.(collect.Info)
			firstClean(info, errorTimes)
		} else {
			break
		}
		scanInterval := 5000 + rand.Intn(3000)
		time.Sleep(time.Duration(scanInterval) * time.Millisecond)
	}
}

func putDc(info collect.Info, ip string, min, max int64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var body dcinfo.SettingBody //Body 是sbody類型並給值
	body.UID = 1
	body.MAC = 0
	body.TmF = 0
	body.Fltr = 1
	body.TSt = min
	body.TEnd = max
	body.Amt = 0
	b, marshalErr := json.Marshal(body) //轉為json
	if marshalErr != nil {
		panic(marshalErr)
	}
	payload := bytes.NewBuffer(b)
	url := "http://" + ip + dcOutputAPI
	req, newRequestErr := http.NewRequest("PUT", url, payload)
	if newRequestErr != nil {
		return newRequestErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", info.DcAuthorization)
	req.Header.Set("Cache-Control", "no-store")

	client := &http.Client{
		Transport: dcTr,
	}

	resp, doErr := client.Do(req)

	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("status code error")
	}
	return nil
}

func getDc(info collect.Info, ip string, fetchTime int64) (message dcinfo.WiseMessage, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	url := "http://" + ip + dcLogMessageAPI
	req, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		panic(requestErr)
	}
	//req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", info.DcAuthorization)
	req.Header.Set("Cache-Control", "no-store")
	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: dcTr,
	}
	resp, doErr := client.Do(req)

	if doErr != nil {
		panic(doErr)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("error status code:" + strconv.Itoa(resp.StatusCode) + " error message:" + resp.Status)
	}

	if resp.Body == nil {
		panic("response body is nil")
	}

	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		panic(err)
	}

	if message.Err != 0 {
		panic(message.Msg)
	}
	if int64(len(message.LogMsg)) > fetchTime*15 {
		msg := "Data more than expected ,i fetch " + strconv.FormatInt(fetchTime, 10) + " seconds 。 there are " + strconv.Itoa(len(message.LogMsg)) + " data"
		panic(msg)
	}
	return message, err
}

func getLastDcData(info collect.Info) (lastData machineio.MachineIO, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	var url = "http://" + global.MainIP + lastDcDataAPI
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("machineNumber", info.MachineNumber)
	client := &http.Client{
		//Timeout:   time.Duration(40 * time.Second),
		Transport: tr,
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		panic(doErr)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		panic("response is empty")
	}

	var response = new(collect.LastDataInfo)

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		msg := "error status code:" + strconv.Itoa(resp.StatusCode) + " error message:" + response.Response
		panic(msg)
	}
	return response.Data, nil

}

func firstClean(info collect.Info, errorTimes *uint) {

	defer func() {
		if err := recover(); err != nil {
			globalLogger.Criticalf("firstClean : machineNumber : %s, dc address : %s, err : %s", info.MachineNumber, info.MacAddress, err)

		}
	}()
	ip := ""
	if v, ok := macAddressIPMap.Load(info.MacAddress); ok {
		ip = v.(string)
		fmt.Println("aaaaaaaaaaa", ip)
	} else {
		time.Sleep(30 * time.Second)

	}

	min, max, fetchTimeErr := getFetchTimeInterval(info)
	if fetchTimeErr != nil {
		panic(fetchTimeErr)
	}
	dcSettingBody, getDcBodyErr := getDcBody(info, ip)

	if getDcBodyErr != nil {
		*errorTimes++

		if *errorTimes%6 == 0 {
			//跟後台說停機瞜
			if err := statusReturn(info, status.STOP); err != nil {
				//寫入失敗嘗試下次再寫一次
				*errorTimes--
				panic(err)
			} else {
				//停機後清掉暫存
				lastDcMap.Delete(info.MachineNumber)

			}
		}
		if ip == "" {
			panic("ip is empty error times:" + strconv.Itoa(int(*errorTimes)))
		}
		panic(getDcBodyErr.Error() + " error times:" + strconv.Itoa(int(*errorTimes)))
	}

	*errorTimes = 0
	if max > dcSettingBody.TFst-2 {
		max = dcSettingBody.TFst - 2
	}
	if min < dcSettingBody.TLst {
		min = dcSettingBody.TLst
		max = min + 10
	}
	fetchTime := ((max - min) + 1)
	fmt.Println("max:", max, "min:", min)

	if fetchTime <= 0 {
		panic("no need to fetch")
	}

	if fetchTime >= int64(info.PutTimeInterval) {
		statusReturn(info, status.SYNC)
	}

	// server 記得刷成同步狀態
	if err := putDc(info, ip, min, max); err != nil {
		panic(err)
	}
	globalLogger.Emergencyf("%s %s %s fetch time interval from %d to %d total %d second (remaining %d s)", "Machine: "+info.MachineNumber, ip, info.MacAddress, min, max, fetchTime, (dcSettingBody.TFst - max))
	if err := putDc(info, ip, min, max); err != nil {
		panic(err)
	}
	if err := checkDc(info, ip, min, max); err != nil {
		panic(err)
	}
	wiseMessage, getDcErr := getDc(info, ip, fetchTime)
	if getDcErr != nil {
		panic(getDcErr)
	}

	var lastDcData machineio.MachineIO
	if v, ok := lastDcMap.Load(info.MachineNumber); ok {
		lastDcData = v.(machineio.MachineIO)
	}
	if lastDcData.Timestamp == 0 {
		lastDc, getLastDataErr := getLastDcData(info)
		if getLastDataErr != nil {
			panic(getLastDataErr)
		}
		if lastDc.Timestamp == 0 {
			lastDc.Timestamp = 1
			lastDcMap.Store(info.MachineNumber, lastDc)
			lastDcData = lastDc
		}

	}
	dcArr, cleanErr := cleanDc(info, wiseMessage, lastDcData)
	if cleanErr != nil {
		panic(cleanErr)
	}
	globalLogger.Noticef("Machine: %s saved data:%d,Wise data:%d", info.MachineNumber, len(dcArr), len(wiseMessage.LogMsg))
	if len(dcArr) != 0 {
		lastData := dcArr[len(dcArr)-1]
		lastDcMap.Store(info.MachineNumber, lastData)
	}

	saveErr := saveDc(info, dcArr, max)
	// go saveDcBackUp(dcSetting, dcArr, max)
	if saveErr != nil {
		panic(saveErr)
	}

}

func tryToAddNewCollection() {

	// if rdArr, err := getRDInfo(); err == nil {
	// 	for _, rd := range rdArr {
	// 		if v, ok := macAddressIPMap.Load(rd.MacAddress); ok {
	// 			ip := v.(string)
	// 			toRD(ip, ipArr)
	// 		}
	// 	}
	// }

	if dsArr, err := getDSInfo(); err == nil {
		for _, value := range dsArr {
			if v, ok := macAddressIPMap.Load(value.MacAddress); ok {
				ip := v.(string)
				checkCast(ip, value)
			} else {
				go updateDSStatus(value.MacAddress, "OFFLINE")
			}
		}
	}
	if collectInfoArr, err := getCollectionInfo(); err != nil {
		globalLogger.Warning(err.Error())
		time.Sleep(30 * time.Second)
	} else {
		for _, value := range collectInfoArr {
			if _, ok := doingMap.Load(value.MachineNumber); !ok {
				doingMap.Store(value.MachineNumber, value)
				go firstCleanLoop(value.MachineNumber)
				go check(value.MachineNumber)
				globalLogger.Infof("Add new dcSetting ,machineNumber: %s , dc mac address : %s", value.MachineNumber, value.MacAddress)
			} else {
				doingMap.Store(value.MachineNumber, value)
			}

		}
		numberMap := make(map[string]bool)
		for _, value := range collectInfoArr {
			numberMap[value.MachineNumber] = true
		}
		doingMap.Range(func(k, v interface{}) bool {
			info := v.(collect.Info)
			if !numberMap[info.MachineNumber] {
				doingMap.Delete(info.MachineNumber)
			}
			return true
		})
	}

}
