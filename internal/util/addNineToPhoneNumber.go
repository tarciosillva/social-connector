package util

func AddNineToPhoneNumber(phoneNumber string) string {
	return phoneNumber[:4] + "9" + phoneNumber[4:]
}
