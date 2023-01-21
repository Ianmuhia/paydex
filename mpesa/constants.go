package mpesa

const (
	SandBox                = "sandbox"
	BusinessPayBill        = "BusinessPayBill"
	TransactionReversal    = "TransactionReversal"
	SalaryPayment          = "SalaryPayment"
	BusinessPayment        = "BusinessPayment"
	PromotionPayment       = "PromotionPayment"
	AccountBalance         = "AccountBalance"
	CustomerPayBillOnline  = "CustomerPayBillOnline"
	TransactionStatusQuery = "TransactionStatusQuery"
	BusinessBuyGoods       = "BusinessBuyGoods"

	// Identifier Types
	// see https://developer.safaricom.co.ke/docs#identifier-types
	PayBillIdentifier    = "4"
	TillNumberIdentifier = "2"
	MSISDNIdentifier     = "1"
	AccessToken          = "access_token"
	B2CAccessToken       = "B2c_access_token"

	// mpesa types.
	C2B = "c2b"
	B2B = "b2b"
)
