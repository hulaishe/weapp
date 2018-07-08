package notify

import (
	"encoding/xml"
	"weapp/payment"
	"errors"
	"weapp/util"
)

// PaidNotify 支付结果返回数据
type PaidNotify struct {
	payment.BaseResponse

	DeviceInfo         string `xml:"device_info,omitempty"`          // 设备号:微信支付分配的终端设备号
	OpenID             string `xml:"openid,omitempty"`               // 用户标识:用户在商户appid下的唯一标识
	IsSubscribe        string `xml:"is_subscribe,omitempty"`         // 是否关注公众账号:用户是否关注公众账号，Y-关注，N-未关注，仅在公众账号类型支付有效
	TradeType          string `xml:"trade_type,omitempty"`           // 交易类型:JSAPI、NATIVE、APP
	BankType           string `xml:"bank_type,omitempty"`            // 付款银行:银行类型，采用字符串类型的银行标识
	TotalFee           int    `xml:"total_fee,omitempty"`            // 订单金额:订单总金额，单位为分
	SettlementTotalFee int    `xml:"settlement_total_fee,omitempty"` // 应结订单金额:应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	FeeType            string `xml:"fee_type,omitempty"`             // 货币种类: 符合ISO4217标准的三位字母代码，默认人民币：CNY
	CashFee            string `xml:"cash_fee,omitempty"`             // 现金支付金额: 现金支付金额订单现金支付金额
	CashFeeType        string `xml:"cash_fee_type,omitempty"`        // 现金支付货币类型: 符合ISO4217标准的三位字母代码，默认人民币：CNY
	TransactionID      string `xml:"transaction_id,omitempty"`       // 微信支付订单号
	OutTradeNo         string `xml:"out_trade_no,omitempty"`         // 商户订单号:商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|*@ ，且在同一个商户号下唯一
	Attach             string `xml:"attach,omitempty"`               // 商家数据包:商家数据包，原样返回
	TimeEnd            string `xml:"time_end,omitempty"`             // 支付完成时间:支付完成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
}

// 支付回调：根据请求表单解析支付回调参数
func ParsePaidNotify(requestBody []byte) (*PaidNotify, error) {

	if len(requestBody) == 0 {
		return nil, errors.New("支付回调参数为空")
	}

	ntf := new(PaidNotify)
	err := xml.Unmarshal(requestBody, ntf)
	if err != nil {
		return nil, err
	}

	if err = ntf.ReturnCheck(); err != nil {
		return nil, err
	}
	if err = ntf.ResultCheck(); err != nil {
		return nil, err
	}

	return ntf, nil
}

// RefundedNotify 退款结果通知
type RefundedNotify struct {
	payment.BaseResponse

	ReqInfoStr      string `xml:"req_info,omitempty"` // 加密信息
	RefundedReqInfo *RefundedReqInfo                  // 解密后的退款消息
}

type RefundedReqInfo struct {
	XMLName             xml.Name `xml:"xml" json:"-"`
	TransactionID       string   `xml:"transaction_id,omitempty"`        // 微信订单号
	OutTradeNo          string   `xml:"out_trade_no,omitempty"`          // 商户订单号:商户系统内部的订单号
	RefundID            string   `xml:"refund_id,omitempty"`             // 微信退款单号
	OutRefundNo         string   `xml:"out_refund_no,omitempty"`         // 商户退款单号
	TotalFee            int      `xml:"total_fee,omitempty"`             // 订单金额:订单总金额，单位为分，只能为整数
	SettlementRefundFee int      `xml:"settlement_refund_fee,omitempty"` // 应结订单金额:当该订单有使用非充值券时，返回此字段。应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额
	RefundFee           int      `xml:"refund_fee,omitempty"`            // 申请退款金额:退款总金额,单位为分
	SettlementTotalFee  int      `xml:"settlement_total_fee,omitempty"`  // 退款金额:退款金额=申请退款金额-非充值代金券退款金额，退款金额<=申请退款金额
	RefundStatus        string   `xml:"refund_status,omitempty"`         // 退款状态:SUCCESS-退款成功 CHANGE-退款异常 REFUNDCLOSE—退款关闭
	SuccessTime         string   `xml:"success_time,omitempty"`          // 退款成功时间:资金退款至用户帐号的时间，格式2017-12-15 09:46:01
	// 取当前退款单的退款入账方
	// 1）退回银行卡：
	// {银行名称}{卡类型}{卡尾号}
	// 2）退回支付用户零钱:
	// 支付用户零钱
	// 3）退还商户:
	// 商户基本账户
	// 商户结算银行账户
	// 4）退回支付用户零钱通:
	// 支付用户零钱通
	RefundRecvAccout    string `xml:"refund_recv_accout,omitempty"`    // 退款入账账户
	RefundAccount       string `xml:"refund_account,omitempty"`        // 退款资金来源:REFUND_SOURCE_RECHARGE_FUNDS 可用余额退款/基本账户 REFUND_SOURCE_UNSETTLED_FUNDS 未结算资金退款
	RefundRequestSource string `xml:"refund_request_source,omitempty"` // 退款发起来源:API接口 VENDOR_PLATFORM商户平台
}

// 退款回调：根据请求表单解析退款回调参数
func ParseRefundedNotify(requestBody []byte, key string) (*RefundedNotify, error) {

	if len(requestBody) == 0 {
		return nil, errors.New("支付回调参数为空")
	}

	rtf := new(RefundedNotify)
	err := xml.Unmarshal(requestBody, rtf)
	if err != nil {
		return nil, err
	}

	if err = rtf.ReturnCheck(); err != nil {
		return nil, err
	}

	reqInfoStr, err := util.AesECBDecrypt(rtf.ReqInfoStr, key)
	if err != nil {
		return nil, err
	}

	rtfReqInfo := new(RefundedReqInfo)
	err = xml.Unmarshal([]byte(reqInfoStr), rtfReqInfo)
	if err != nil {
		return nil, err
	}

	rtf.RefundedReqInfo = rtfReqInfo
	return rtf, nil
}
