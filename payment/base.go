package payment

import (
	"encoding/xml"
	"errors"
)

const (
	baseURI = "https://api.mch.weixin.qq.com"
)

type BaseParam struct {
	XMLName  xml.Name `xml:"xml" json:"-"`
	AppID    string   `xml:"appid"`               // 小程序ID
	MchID    string   `xml:"mch_id"`              // 商户号
	NonceStr string   `xml:"nonce_str"`           // 随机字符串
	Sign     string   `xml:"sign"`                // 签名
	SignType string   `xml:"sign_type,omitempty"` // 签名类型: 目前支持HMAC-SHA256和MD5，默认为MD5
	TotalFee int      `xml:"total_fee"`           // 标价金额(单位分)
}

// BaseResponse 基础返回数据
type BaseResponse struct {
	XMLName    xml.Name `xml:"xml" json:"-"`
	ReturnCode string   `xml:"return_code"`            // 返回状态码: SUCCESS/FAIL
	ReturnMsg  string   `xml:"return_msg,omitempty"`   // 返回信息: 返回信息，如非空，为错误原因
	ResultCode string   `xml:"result_code,omitempty"`  // 业务结果: SUCCESS/FAIL SUCCESS退款申请接收成功，结果通过退款查询接口查询 FAIL 提交业务失败
	ErrCode    string   `xml:"err_code,omitempty"`     // 错误码
	ErrCodeDes string   `xml:"err_code_des,omitempty"` // 错误代码描述

	AppID    string `xml:"appid,omitempty"`     // 小程序ID
	MchID    string `xml:"mch_id,omitempty"`    // 商户号
	NonceStr string `xml:"nonce_str,omitempty"` // 随机字符串
	Sign     string `xml:"sign,omitempty"`      // 签名
	SignType string `xml:"sign_type,omitempty"` // 签名类型
}

// Check 检测返回信息是否包含错误
func (res BaseResponse) ReturnCheck() error {

	switch res.ReturnCode {
	case "SUCCESS":
		return nil
	case "FAIL":
		return errors.New(res.ErrCodeDes)
	default:
		return errors.New("未知微信返回状态码: " + res.ReturnCode)
	}
}

// Check 检测返回信息是否包含错误
func (res BaseResponse) ResultCheck() error {

	switch res.ResultCode {
	case "SUCCESS":
		return nil
	case "FAIL":
		return errors.New(res.ErrCodeDes)
	default:
		return errors.New("未知微信返回业务结果代码: " + res.ResultCode)
	}
}
