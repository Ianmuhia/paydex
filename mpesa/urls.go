package mpesa

type MURL string

const (
	LiveURL           MURL = "https://api.safaricom.co.ke/"
	SandboxURL        MURL = "https://sandbox.safaricom.co.ke/"
	tokenURL          MURL = "oauth/v1/generate?grant_type=client_credentials"
	b2cURL            MURL = "mpesa/b2c/v1/paymentrequest"
	b2bURL            MURL = "mpesa/b2b/v1/paymentrequest"
	balance           MURL = "mpesa/accountbalance/v1/query"
	transactionStatus MURL = "mpesa/transactionstatus/v1/query"
	registerURL       MURL = "mpesa/c2b/v1/registerurl"
	simulateC2BURL    MURL = "mpesa/c2b/v1/simulate"
	stkPush           MURL = "mpesa/stkpush/v1/processrequest"
	stkPushQuery      MURL = "mpesa/stkpushquery/v1/query"
)

func (m *Mpesa) getMpesaURL(s string) string {
	switch MURL(s) {
	case b2bURL:
		return m.getBaseURL() + string(b2bURL)
	case b2cURL:
		return m.getBaseURL() + string(b2cURL)
	case registerURL:
		return m.getBaseURL() + string(registerURL)
	case balance:
		return m.getBaseURL() + string(balance)
	case transactionStatus:
		return m.getBaseURL() + string(transactionStatus)
	case simulateC2BURL:
		return m.getBaseURL() + string(simulateC2BURL)
	case stkPush:
		return m.getBaseURL() + string(stkPush)
	case stkPushQuery:
		return m.getBaseURL() + string(stkPushQuery)
	case tokenURL:
		return m.getBaseURL() + string(tokenURL)
	default:
		return ""
	}
}

func (m *Mpesa) getBaseURL() string {
	if !m.Live {
		return string(SandboxURL)
	}
	return string(LiveURL)
}
