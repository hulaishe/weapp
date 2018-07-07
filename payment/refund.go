package payment

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"

	"weapp"
	"weapp/util"
	"strconv"
)

const (
	refundAPI = "/secapi/pay/refund"
)

// Refunder 退款表单数据
type Refunder struct {
	BaseParam

	TransactionID string `xml:"transaction_id,emitempty"`  // 微信订单号: 微信生成的订单号，在支付通知中有返回。和商户订单号二选一
	OutTradeNo    string `xml:"out_trade_no,emitempty"`    // 商户订单号: 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一。 和微信订单号二选一
	OutRefundNo   string `xml:"out_refund_no"`             // 商户退款单号: 商户系统内部的退款单号，商户系统内部唯一，只能是数字、大小写字母_-|*@ ，同一退款单号多次请求只退一笔。
	RefundFee     int    `xml:"refund_fee"`                // 退款金额: 退款总金额，订单总金额，单位为分，只能为整数
	RefundFeeType string `xml:"refund_fee_type,emitempty"` // 货币种类: 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	RefundDesc    string `xml:"refund_desc,emitempty"`     // 退款原因: 若商户传入，会在下发给用户的退款消息中体现退款原因

	// 退款资金来源: 仅针对老资金流商户使用
	// REFUND_SOURCE_UNSETTLED_FUNDS---未结算资金退款（默认使用未结算资金退款）
	// REFUND_SOURCE_RECHARGE_FUNDS---可用余额退款
	RefundAccount string `xml:"refund_account,emitempty"`

	// 退款结果通知url: 异步接收微信支付退款结果通知的回调地址
	// 通知 URL 必须为外网可访问且不允许带参数
	// 如果参数中传了notify_url，则商户平台上配置的回调地址将不会生效。
	NotifyURL string `xml:"notify_url,omitempty"`
}

// 退款 请求返回结果
type RefundResponse struct {
	BaseResponse

	TransactionID       string `xml:"transaction_id,omitempty"`        // 微信订单号
	OutTradeNo          string `xml:"out_trade_no,omitempty"`          // 商户订单号: 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一。 和微信订单号二选一
	OutRefundNo         string `xml:"out_refund_no,omitempty"`         // 商户退款单号: 商户系统内部的退款单号，商户系统内部唯一，只能是数字、大小写字母_-|*@ ，同一退款单号多次请求只退一笔。
	RefundID            string `xml:"refund_id,omitempty"`             // 微信退款单号
	RefundFee           int    `xml:"refund_fee,omitempty"`            // 退款金额: 退款总金额，订单总金额，单位为分，只能为整数
	SettlementRefundFee int    `xml:"settlement_refund_fee,omitempty"` // 应结退款金额: 去掉非充值代金券退款金额后的退款金额，退款金额=申请退款金额-非充值代金券退款金额，退款金额<=申请退款金额
	TotalFee            int    `xml:"total_fee,omitempty"`             // 标价金额: 订单总金额，单位为分，只能为整数
	SettlementTotalFee  int    `xml:"settlement_total_fee,omitempty"`  // 应结订单金额: 去掉非充值代金券金额后的订单总金额，应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	FeeType             string `xml:"fee_type,omitempty"`              // 标价币种: 订单金额货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashFee             int    `xml:"cash_fee,omitempty"`              // 现金支付金额: 现金支付金额，单位为分，只能为整数
	CashFeeType         string `xml:"cash_fee_type,omitempty"`         // 现金支付币种: 货币类型，符合ISO 4217标准的三位字母代码，默认人民币：CNY
	CashRefundFee       int    `xml:"cash_refund_fee,omitempty"`       // 现金退款金额: 现金退款金额，单位为分，只能为整数
}

// Refund 发起退款请求
func (r *Refunder) Refund() (*RefundResponse, error) {

	// 拼接参数
	data := make(map[string]string)

	if r.AppID == "" {
		return nil, errors.New("app_id 不能为空")
	}
	data["appid"] = r.AppID

	if r.MchID == "" {
		return nil, errors.New("mch_id 不能为空")
	}
	data["mch_id"] = r.MchID

	if r.NonceStr == "" {
		r.NonceStr = util.NonceStr(6)
	}
	data["nonce_str"] = r.NonceStr

	r.SignType = "MD5"
	data["sign_type"] = "MD5"

	if r.TransactionID == "" && r.OutTradeNo == "" {
		return nil, errors.New("out_trade_no 和 out_trade_no 必须填写一个")
	}
	if r.TransactionID != "" {
		data["transaction_id"] = r.TransactionID
	}
	if r.OutTradeNo == "" {
		data["out_trade_no"] = r.OutTradeNo
	}

	if r.OutRefundNo == "" {
		return nil, errors.New("out_refund_no 不能为空")
	}
	data["out_refund_no"] = r.OutRefundNo

	if r.TotalFee == 0 {
		return nil, errors.New("total_fee 不能为空")
	}
	data["total_fee"] = strconv.Itoa(r.TotalFee)

	if r.RefundFee == 0 {
		return nil, errors.New("refund_fee 不能为空")
	}
	data["refund_fee"] = strconv.Itoa(r.RefundFee)

	if r.RefundFeeType != "" {
		data["refund_fee_type"] = r.RefundFeeType
	}

	if r.RefundDesc != "" {
		data["refund_desc"] = r.RefundDesc
	}

	if r.RefundAccount != "" {
		data["refund_account"] = r.RefundAccount
	}

	if r.NotifyURL != "" {
		data["notify_url"] = r.NotifyURL
	}

	sign, err := util.SignByMD5(data)
	if err != nil {
		return nil, err
	}

	r.Sign = sign

	body, err := xml.Marshal(r)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(baseURI+refundAPI, "application/xml", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(weapp.WeChatServerError)
	}

	var resp RefundResponse
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
