// Package payment 微信支付
package payment

import (
	"encoding/xml"
	"errors"
	"net"
	"strconv"

	"weapp/util"
	"net/http"
	"strings"
	"encoding/json"
	"weapp"
)

const (
	unifyAPI = "/pay/unifiedorder"
)

// Params 前端调用支付必须的参数
// 注意返回后得大小写格式不能变动
type FrontPayParams struct {
	Timestamp int64  `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
	Package   string `json:"package"`
}

// Order 商户统一订单
type UnifiedOrder struct {
	BaseParam

	NotifyURL      string `xml:"notify_url"`            // 异步接收微信支付结果通知的回调地址，通知url必须为外网可访问的url，不能携带参数。
	OpenID         string `xml:"openid"`                // 下单用户ID
	DeviceInfo     string `xml:"device_info,omitempty"` // 设备号， 自定义参数，可以为终端设备号(门店号或收银设备ID)，PC网页或公众号内支付可以传"WEB"
	Body           string `xml:"body"`                  // 商品描述
	Detail         string `xml:"detail,omitempty"`      // 商品详情
	Attach         string `xml:"attach,omitempty"`      // 附加数据
	OutTradeNo     string `xml:"out_trade_no"`          // 商户订单号
	FeeType        string `xml:"fee_type,omitempty"`    // 标价币种
	SPBillCreateIP net.IP `xml:"spbill_create_ip"`      // 终端IP
	TimeStart      string `xml:"time_start,omitempty"`  // 交易起始时间 格式为yyyyMMddHHmmss
	TimeExpire     string `xml:"time_expire,omitempty"` // 交易结束时间 订单失效时间 格式为yyyyMMddHHmmss
	GoodsTag       string `xml:"goods_tag,omitempty"`   // 订单优惠标记，使用代金券或立减优惠功能时需要的参数，
	TradeType      string `xml:"trade_type"`            // 小程序取值如下：JSAPI
	LimitPay       string `xml:"limit_pay,omitempty"`   // 上传此参数 no_credit 可限制用户不能使用信用卡支付
}

// 统一下单 请求返回结果
type UnifiedOrderResponse struct {
	BaseResponse

	DeviceInfo string `xml:"device_info,omitempty"` // 设备号
	TradeType  string `xml:"trade_type,omitempty"`  // 小程序取值如下：JSAPI
	PrepayID   string `xml:"prepay_id,omitempty"`   // 预支付交易会话标识
	CodeUrl    string `xml:"code_url,omitempty"`    // 二维码链接
}

// Unify 统一下单
func (o *UnifiedOrder) Unify() (*UnifiedOrderResponse, error) {

	// 拼接参数
	data := make(map[string]string)

	if o.AppID == "" {
		return nil, errors.New("app_id 不能为空")
	}
	data["appid"] = o.AppID

	if o.MchID == "" {
		return nil, errors.New("mch_id 不能为空")
	}
	data["mch_id"] = o.MchID

	if o.DeviceInfo != "" {
		data["device_info"] = o.DeviceInfo
	}

	if o.NonceStr == "" {
		o.NonceStr = util.NonceStr(6)
	}
	data["nonce_str"] = o.NonceStr

	o.SignType = "MD5"
	data["sign_type"] = "MD5"

	if o.Body == "" {
		return nil, errors.New("body 不能为空")
	}
	data["body"] = o.Body

	if o.Detail != "" {
		data["detail"] = o.Detail
	}

	if o.Attach != "" {
		data["attach"] = o.Attach
	}

	if o.OutTradeNo == "" {
		return nil, errors.New("out_trade_no 不能为空")
	}
	data["out_trade_no"] = o.OutTradeNo

	if o.FeeType != "" {
		data["fee_type"] = o.FeeType
	}

	if o.TotalFee == 0 {
		return nil, errors.New("total_fee 不能为空")
	}
	data["total_fee"] = strconv.Itoa(o.TotalFee)

	if o.SPBillCreateIP == nil {
		return nil, errors.New("spbill_create_ip 不能为空")
	}
	data["spbill_create_ip"] = o.SPBillCreateIP.String()

	if o.TimeStart != "" {
		data["time_start"] = o.TimeStart
	}

	if o.TimeExpire != "" {
		data["time_expire"] = o.TimeExpire
	}

	if o.GoodsTag != "" {
		data["goods_tag"] = o.GoodsTag
	}

	if o.NotifyURL == "" {
		return nil, errors.New("notify_url 不能为空")
	}
	data["notify_url"] = o.NotifyURL

	o.TradeType = "JSAPI"
	data["trade_type"] = o.TradeType

	if o.LimitPay != "" {
		data["limit_pay"] = o.LimitPay
	}

	if o.OpenID == "" {
		return nil, errors.New("openid 不能为空")
	}
	data["openid"] = o.OpenID

	// 计算签名
	sign, err := util.SignByMD5(data)
	if err != nil {
		return nil, err
	}

	o.Sign = sign

	// 发送统一下单请求
	body, err := xml.Marshal(o)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(baseURI+unifyAPI, "application/xml", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(weapp.WeChatServerError)
	}

	var resp UnifiedOrderResponse
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if err = resp.ReturnCheck(); err != nil {
		return nil, err
	}
	if err = resp.ResultCheck(); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetParams 获取支付参数
//
// @key
// @timestamp
func (res *UnifiedOrderResponse) GetParams(key string, timestamp int64) (*FrontPayParams, error) {

	sign, err := util.SignByMD5(map[string]string{
		"key":       key,
		"appId":     res.AppID,
		"signType":  "MD5",
		"nonceStr":  res.NonceStr,
		"package":   "prepay_id" + "=" + res.PrepayID,
		"timeStamp": strconv.FormatInt(timestamp, 10),
	})
	if err != nil {
		return nil, err
	}

	var p = FrontPayParams{
		Timestamp: timestamp,
		NonceStr:  res.NonceStr,
		SignType:  "MD5",
		PaySign:   sign,
	}

	return &p, nil
}
