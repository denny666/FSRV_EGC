package lib

import (
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/routerservice"
	"math"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robbert229/jwt"
)

const (
	Token    = "_jwt"
	loginN   = 1
	logoutN  = 2
	abnormal = 0
	maxAge   = 3600
	OpcStr   = 0
	HttpStr  = 1
)

// AuthInterceptor 每個router都必須透過interceptor 檢查權限
func AuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if "/edge/login" == c.Request.RequestURI || "/edge/logout" == c.Request.RequestURI {
			c.Next()
		} else if token, err := c.Cookie(Token); err != nil {
			var errRes nodeattr.ErrMsg
			errRes.Msg = "token invalid"
			c.JSON(http.StatusForbidden, errRes) // gin.H{"error": "invalid request, restricted endpoint????????"}
			c.Abort()
		} else {
			if newToken, err := GenerateNewToken(token); err != nil {
				var errRes nodeattr.ErrMsg
				errRes.Msg = "token invalid"
				c.JSON(http.StatusForbidden, errRes) // gin.H{"error": "invalid request, restricted endpoint????????"}
				c.Abort()
			} else {
				c.SetCookie(Token, newToken, maxAge, "/", "", false, true)
				c.Next()
			}
		}
	}
}

// RecordInterceptor 每個router都必須透過 record 來做紀錄
func RecordInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var record nodeattr.ConnectRecord
		ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		account := c.Request.Header.Get("account")
		record.IP = ip
		record.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		record.URL = c.Request.RequestURI
		record.Account = account
		if record.URL == "/edge/logout" {
			record.Type = logoutN
			routerservice.InsertConRecordInfo(record)
		}
		c.Next()
	}
}

// GeneratorJWT 產生token
func GeneratorJWT(userName string) (token string, err error) {
	algorithm := jwt.HmacSha256("test")
	claims := jwt.NewClaim()
	claims.Set("UserName", userName)
	claims.SetTime("exp", time.Now().Add(3600*time.Second))
	token, err = algorithm.Encode(claims)
	return
}

// GenerateNewToken 根據token取得 account、新token
func GenerateNewToken(token string) (newToken string, err error) {
	algorithm := jwt.HmacSha256("test")
	if err := algorithm.Validate(token); err != nil {
		return newToken, err
	}

	loadedClaims, err := algorithm.Decode(token)
	if err != nil {
		return newToken, err
	}

	if userName, err := loadedClaims.Get("UserName"); err != nil {
		return newToken, err
	} else {
		algorithm := jwt.HmacSha256("test")
		claims := jwt.NewClaim()
		claims.Set("UserName", userName)
		claims.SetTime("exp", time.Now().Add(3600*time.Second))
		if token, err = algorithm.Encode(claims); err != nil {
			return newToken, err
		} else {
			return token, nil
		}

	}
}

// SCTCaculator 計算SCT
func SCTCaculator(dataorg []nodeattr.HistoryData, firstrange float64) (tempsct float64) {
	var lenSize = len(dataorg)
	if lenSize == 0 {
		return 0
	} else if lenSize <= 3 {
		var tempsum float64
		for _, value := range dataorg {
			tempsum += value.CycleTime
		}

		tempsct = tempsum / float64(lenSize)
	} else if lenSize > 3 {
		var data []float64
		var middlenum float64
		var everagenum float64
		var sqrtnum float64
		var standard float64
		var dataclean1, dataclean2 []float64
		dataclean1 = nil
		dataclean2 = nil
		for _, value := range dataorg {
			data = append(data, value.CycleTime)
		}

		sort.Float64s(data) //排序
		// logger.Emergency("排序data:	", data)
		if len(data)%2 == 0 {
			// println("偶數")
			middlenum = (data[(len(data)/2)] + data[(len(data)/2)-1]) / 2
		} else {
			// println("奇數")
			middlenum = data[(len(data)/2)+1]
		}

		for i := 0; i < len(data); i++ {
			if data[i] > (middlenum-middlenum*firstrange) && data[i] < (middlenum+middlenum*firstrange) { //取第一區間
				dataclean1 = append(dataclean1, data[i])
			}
		}
		if len(dataclean1) == 0 {
			for i := 0; i < len(data); i++ {
				if data[i] > (middlenum-middlenum*0.9) && data[i] < (middlenum+middlenum*0.9) { //雙峰取第一區間
					dataclean1 = append(dataclean1, data[i])
				}
			}
		}

		for k := 0; k < len(dataclean1); k++ {
			everagenum += dataclean1[k]
			sqrtnum += math.Pow(dataclean1[k], 2)
		}
		standard = math.Sqrt(sqrtnum/float64(len(dataclean1)) - math.Pow(everagenum/float64(len(dataclean1)), 2))
		for i := 0; i < len(dataclean1); i++ {
			if dataclean1[i] > middlenum-standard && dataclean1[i] < middlenum+standard { //取第二區間
				dataclean2 = append(dataclean2, dataclean1[i])
			}
		}
		for j := 0; j < len(dataclean2); j++ {
			tempsct += dataclean2[j]
		}
		tempsct = tempsct / float64(len(dataclean2))
		if math.IsNaN(tempsct) == true {

			tempsct = 0
		}
	}

	return tempsct

}
