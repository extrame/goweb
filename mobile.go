package goweb

import (
	"strings"
)

//detect whether the client is from mobile device
func (c *Context) IsMobile() bool {
	//header detect
	accept := c.Request.Header.Get("HTTP_ACCEPT")
	if strings.Contains(accept, "application/x-obml2d") ||
		strings.Contains(accept, "application/vnd.rim.html") ||
		strings.Contains(accept, "text/vnd.wap.wml") ||
		strings.Contains(accept, "application/vnd.wap.xhtml+xml") {
		return true
	}
	if c.Request.Header.Get("HTTP_X_WAP_PROFILE") != "" ||
		c.Request.Header.Get("HTTP_X_WAP_CLIENTID") != "" ||
		c.Request.Header.Get("HTTP_WAP_CONNECTION") != "" ||
		c.Request.Header.Get("HTTP_PROFILE") != "" ||
		c.Request.Header.Get("HTTP_X_OPERAMINI_PHONE_UA") != "" || // Reported by Nokia devices (eg. C3)
		c.Request.Header.Get("HTTP_X_NOKIA_IPADDRESS") != "" ||
		c.Request.Header.Get("HTTP_X_NOKIA_GATEWAY_ID") != "" ||
		c.Request.Header.Get("HTTP_X_ORANGE_ID") != "" ||
		c.Request.Header.Get("HTTP_X_VODAFONE_3GPDPCONTEXT") != "" ||
		c.Request.Header.Get("HTTP_X_HUAWEI_USERID") != "" ||
		c.Request.Header.Get("HTTP_UA_OS") != "" || // Reported by Windows Smartphones.
		c.Request.Header.Get("HTTP_X_MOBILE_GATEWAY") != "" || // Reported by Verizon, Vodafone proxy system.
		c.Request.Header.Get("HTTP_X_ATT_DEVICEID") != "" ||
		c.Request.Header.Get("HTTP_UA_CPU") == "ARM" {
		return true
	}
	return false
}
