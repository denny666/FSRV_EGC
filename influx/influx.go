package influx

import (
	"FSRV_Edge/datacollect"
	"FSRV_Edge/global"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cbrake/influxdbhelper/v2"
	client "github.com/influxdata/influxdb1-client/v2"
)

const (
	NodeHdaMeasure = "data_node_"
	username       = "admin"
	password       = "82589155"
	DatabaseName   = "data_node_history"
	typeIDStr      = "TypeIDString"
	typeIDUint32   = "TypeIDUint32"
	typeIDFloat    = "TypeIDFloat"
	typeIDDouble   = "TypeIDDouble"
	typeIDUint16   = "TypeIDUint16"
	timestampStr   = "timestamp"
	timeStr        = "time"
	DiTable        = "di_"
	addr           = "http://localhost:8086"
)

var (
	influxClient influxData
	clientHelp   influxdbhelper.Client
	globalLogger = initlog.GetLogger()
)

type influxData struct {
	Client client.Client
	Lock   sync.Mutex
}

func init() {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Init influx error] %v", r)
		}
	}()
	if c, err := influxdbhelper.NewClient(addr, "", "", ""); err == nil {
		clientHelp = c.UseDB(DatabaseName)
	}
	if c, err := getInfluxClient(); err != nil {
		panic(err)
	} else {
		influxClient.Client = c
	}
	if err := createDatabase(); err != nil {
		panic(err)
	}
}

// WriteDi 寫入DI資料至influx (DI有變動才寫入)
func WriteDi(data []datacollect.MachineIO) error {
	for _, v := range data {
		if err := clientHelp.WritePoint(v); err != nil {
			return err
		}
	}
	return nil
}

// WriteStatus 寫入Status資料至influx
func WriteStatus(data []nodeattr.HistoryData) error {
	for _, v := range data {
		if err := clientHelp.WritePoint(v); err != nil {
			return err
		}
	}
	return nil
}
func createDatabase() error {
	q := client.NewQuery("CREATE DATABASE "+DatabaseName+" WITH DURATION 7d", "", "")
	if _, err := influxClient.Client.Query(q); err != nil {
		return err
	}
	return nil
}

// AddData 將變化資料寫入 influx
func AddData(tableName string, valMap map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Influx read data error] %v", r)
		}
	}()
	tags := make(map[string]string)
	now := time.Now()
	points, newBpErr := client.NewBatchPoints(client.BatchPointsConfig{Database: DatabaseName})
	if newBpErr != nil {
		return fmt.Errorf("New batch points error: %s", newBpErr.Error())
	}
	pt, err := client.NewPoint(tableName, tags, valMap, now)
	if err != nil {
		return fmt.Errorf("New points error: %s", err.Error())
	}
	points.AddPoint(pt)
	err = influxClient.Client.Write(points)
	if err != nil {
		return fmt.Errorf("Write points error: %s", err.Error())
	}
	return nil
}

// GetLastStatus 取得最後狀態資料
func GetLastStatus(mac string) (nodeattr.HistoryData, error) {
	var lastStatusArr []nodeattr.HistoryData
	var lastStatus nodeattr.HistoryData
	q := `SELECT * FROM ` + NodeHdaMeasure + mac + ` order by time desc limit 1 `
	err := clientHelp.UseDB(DatabaseName).DecodeQuery(q, &lastStatusArr)
	for _, v := range lastStatusArr {
		lastStatus = v
	}
	return lastStatus, err
}

// GetDataLastTime 取得最後一筆DI時間
func GetDataLastTime(mac string) (int64, error) {
	var data []map[string]interface{}
	var lastTime int64
	q := `SELECT MAX("timestamp") FROM ` + DiTable + mac + ` WHERE "analyzed" != 1`
	if err := clientHelp.UseDB(DatabaseName).DecodeQuery(q, &data); err != nil {
		return lastTime, err
	}
	for _, m := range data {
		if max, err := m["max"].(json.Number).Int64(); err == nil {
			lastTime = max
		}
	}
	return lastTime, nil
}

// GetDcDataAfterTimestamp 取得某時間點之後的資料
func GetDcDataAfterTimestamp(mac string, timestamp int64) ([]datacollect.MachineIO, error) {
	tmStr := strconv.FormatInt(timestamp, 10)
	var arr []datacollect.MachineIO
	q := `SELECT * FROM ` + DiTable + mac + ` where "timestamp" >=` + tmStr + ` and "analyzed" != 1`
	if err := clientHelp.UseDB(DatabaseName).DecodeQuery(q, &arr); err != nil {
		return arr, err
	}
	return arr, nil
}

// GetLastDcData 取得最後一筆資料
func GetLastDcData(mac string) (datacollect.MachineIO, error) {
	var machineIoArr []datacollect.MachineIO
	var machineIO datacollect.MachineIO
	sql := `SELECT * FROM ` + DiTable + mac + ` WHERE "analyzed" != 1 order by time desc limit 1`
	if err := clientHelp.UseDB(DatabaseName).DecodeQuery(sql, &machineIoArr); err != nil {
		return machineIO, err
	}
	for _, v := range machineIoArr {
		machineIO = v
	}
	return machineIO, nil
}

// GetAllDcData 取得所有資料
func GetAllDcData(name, selectT string, showArr []string) ([]map[string]interface{}, error) {
	var hdaData []map[string]interface{}
	// var lastStatusArr []nodeattr.HistoryData
	i, err := strconv.ParseInt(selectT, 10, 64)
	if err != nil {
		return nil, err
	}
	tm := time.Unix(i*int64(time.Millisecond)/int64(time.Second), 0)
	start := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.UTC().Location())
	end := time.Date(tm.Year(), tm.Month(), tm.Day(), 23, 59, 59, 0, tm.UTC().Location())
	querySql := "SELECT "
	if len(showArr) == 0 {
		querySql += " * FROM "
	} else {
		for i, v := range showArr {
			if i != len(showArr)-1 {
				querySql += `"` + v + `"` + ","
			} else {
				querySql += `"` + v + `"`
			}
		}
		querySql += " FROM "
	}
	q := querySql + NodeHdaMeasure + name + " where time >= '" + start.UTC().Format(time.RFC3339) + "' and time <= '" + end.UTC().Format(time.RFC3339) + "' order by time asc"
	err = clientHelp.UseDB(DatabaseName).DecodeQuery(q, &hdaData)
	// var tmpArr []nodeattr.HistoryData
	var tmpArr []map[string]interface{}
	count := 1
	for _, v := range hdaData {
		v["time"] = covertRFCtoTimestamp(v["time"].(string))
		v[nodeattr.MoldCountStr] = count
		tmpArr = append(tmpArr, v)
		count++
	}
	return hdaData, err
}
func covertRFCtoTimestamp(v string) (res int64) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return
	}
	res = t.UnixNano() / int64(time.Millisecond)
	return
}

// GetMoldCount 取得DC模次
func GetMoldCount(name, selectT string) (int64, error) {
	var data []map[string]interface{}
	var mCount int64
	var err error
	i, err := strconv.ParseInt(selectT, 10, 64)
	if err != nil {
		return mCount, err
	}
	i += 1000
	tm := time.Unix(i*int64(time.Millisecond)/int64(time.Second), 0)
	start := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.UTC().Location())
	q := "SELECT count(Mold_Count) FROM " + NodeHdaMeasure + name + " WHERE time >= '" + start.UTC().Format(time.RFC3339) + "' and time <= '" + tm.UTC().Format(time.RFC3339) + "'"
	if err := clientHelp.UseDB(DatabaseName).DecodeQuery(q, &data); err != nil {
		return mCount, err
	}
	for _, m := range data {
		if mCount, err = m["count"].(json.Number).Int64(); err == nil {
			return mCount, err
		}
	}
	return mCount, err
}

// ReadData 讀取 influx
func ReadData(database, devName, selectT string, showArr []string, protocol int, all bool) ([]map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Influx read data error] %v", r)
		}
	}()
	var hdaData []map[string]interface{}
	i, err := strconv.ParseInt(selectT, 10, 64)
	i += 1000
	if err != nil {
		return nil, err
	}
	tm := time.Unix(i*int64(time.Millisecond)/int64(time.Second), 0)
	start := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.UTC().Location())
	end := time.Date(tm.Year(), tm.Month(), tm.Day(), 23, 59, 59, 0, tm.UTC().Location())
	var sql string
	querySql := "SELECT "
	if len(showArr) == 0 {
		querySql += " * FROM "
	} else {
		for i, v := range showArr {
			if i != len(showArr)-1 {
				querySql += `"` + v + `"` + ","
			} else {
				querySql += `"` + v + `"`
			}
		}
		querySql += " FROM "
	}
	if all {
		if protocol == global.OpcStr {
			sql = querySql + NodeHdaMeasure + devName + " WHERE time >= '" + start.UTC().Format(time.RFC3339) + "' and time <= '" + end.UTC().Format(time.RFC3339) + "' order by time asc"
		} else if protocol == global.HttpStr {
			sql = querySql + NodeHdaMeasure + devName + " WHERE time >= '" + start.UTC().Format(time.RFC3339) + "' and time <= '" + end.UTC().Format(time.RFC3339) + "' order by time asc"
		}
	} else {
		if protocol == global.OpcStr {
			sql = querySql + NodeHdaMeasure + devName + " WHERE time <= '" + tm.UTC().Format(time.RFC3339) + `' ORDER BY DESC LIMIT 1`
		} else if protocol == global.HttpStr {
			sql = querySql + NodeHdaMeasure + devName + " WHERE time <= '" + tm.UTC().Format(time.RFC3339) + `' ORDER BY DESC LIMIT 1`
		}
	}
	query := client.NewQuery(sql, database, "")
	resp, err := influxClient.Client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Read data error: %s", err.Error())
	}
	for _, result := range resp.Results {
		for _, row := range result.Series {
			for _, vals := range row.Values {
				data := make(map[string]interface{})
				for i, val := range vals {
					if row.Columns[i] == timeStr {
						t, err := time.Parse(time.RFC3339, val.(string))
						if err == nil {
							val = t.UnixNano() / int64(time.Millisecond)
						}
					} else if row.Columns[i] == nodeattr.StatusStr {
						switch val.(type) {
						case json.Number:
							if tmp, err := strconv.Atoi(val.(json.Number).String()); err == nil {
								val = tmp
							}
						case string:
							if tmp, err := strconv.Atoi(val.(string)); err == nil {
								val = tmp
							}
						default:
							val = 0
						}
					} else if row.Columns[i] == nodeattr.MoldCountStr {
						if protocol == global.HttpStr {
							if v, err := GetMoldCount(devName, selectT); err == nil {
								val = v
							} else {
								val = 0
								fmt.Println(err)
							}
						} else if protocol == global.OpcStr {
							switch val.(type) {
							case json.Number:
								if tmp, err := strconv.Atoi(val.(json.Number).String()); err == nil {
									val = tmp
								}
							case string:
								if tmp, err := strconv.Atoi(val.(string)); err == nil {
									val = tmp
								}
							default:
								val = 0
							}
						}
					}
					data[row.Columns[i]] = val
				}
				hdaData = append(hdaData, data)
			}
		}
	}
	return hdaData, err
}

func getInfluxClient() (client.Client, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: addr,
	})
	return c, err
}
