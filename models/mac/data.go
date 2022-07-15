package mac

import "encoding/xml"

// Info EG 搜尋到的裝置資訊
type Info struct {
	IP         string `json:"ip"`
	MacAddress string `json:"macAddress"`
	Group      string `json:"group"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}
// GoogleInfo google 的chromecast 資訊
type GoogleInfo struct {
	XMLName     xml.Name `xml:"root"`
	Text        string   `xml:",chardata"`
	Xmlns       string   `xml:"xmlns,attr"`
	SpecVersion struct {
		Text  string `xml:",chardata"`
		Major string `xml:"major"`
		Minor string `xml:"minor"`
	} `xml:"specVersion"`
	URLBase string `xml:"URLBase"`
	Device  struct {
		Text         string `xml:",chardata"`
		DeviceType   string `xml:"deviceType"`
		FriendlyName string `xml:"friendlyName"`
		Manufacturer string `xml:"manufacturer"`
		ModelName    string `xml:"modelName"`
		UDN          string `xml:"UDN"`
		IconList     struct {
			Text string `xml:",chardata"`
			Icon struct {
				Text     string `xml:",chardata"`
				Mimetype string `xml:"mimetype"`
				Width    string `xml:"width"`
				Height   string `xml:"height"`
				Depth    string `xml:"depth"`
				URL      string `xml:"url"`
			} `xml:"icon"`
		} `xml:"iconList"`
		ServiceList struct {
			Text    string `xml:",chardata"`
			Service struct {
				Text        string `xml:",chardata"`
				ServiceType string `xml:"serviceType"`
				ServiceID   string `xml:"serviceId"`
				ControlURL  string `xml:"controlURL"`
				EventSubURL string `xml:"eventSubURL"`
				SCPDURL     string `xml:"SCPDURL"`
			} `xml:"service"`
		} `xml:"serviceList"`
	} `xml:"device"`
}
