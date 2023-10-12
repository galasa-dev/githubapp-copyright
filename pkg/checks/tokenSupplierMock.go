/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package checks

type TokenSupplierMock struct {
	tokenToReturn string
}

func NewTokenSupplierMock() (TokenSupplier, error) {

	var err error = nil
	this := new(TokenSupplierMock)

	// Note: This is a valid JWT but doesn't contain any secret information.
	// So we can use it in unit testing.
	// It was generated using the OpenSSL tools.
	this.tokenToReturn = `-----BEGIN PRIVATE KEY-----
	MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBALcwprHtHB9Q4I/p
	bwpg1u+ye8XtRQxzTSS2ATeHOsebY/7831/kOOrX7hLqkuT6xShHOUYbhd7GkcKd
	pXidLnKk4Xn9Mvmj0YMxzt6lT+dMwUrK49tMOsyL6m9uZLuk2EsgX9oOiopojR2J
	r9mxeoAoHuMw1EYbZBRFqrBlqpEHAgMBAAECgYB4a7HYmn53E4pa799/mgMQlGqK
	xJs0QQNAE6ifIPUBy+Mi9oW8GmFT91fX9X1Uqog6Hv/GV0dcF3ovzcO9ks25j4dL
	0Tn+6ipJJSyyasPIHFmc1etCVSRI5wtIK3W0eop1FVz0D+P4ZlVbuwVPwUyMCp19
	KG5cJBJtJCDLU2VMqQJBAOrEGMbvTENuq10kXgQTTi+QnT/vKWyVpucM7HRGptOC
	irMb9VAig+n0BjN7TbsklQXrjBuD5M1D2YKRGTBnMBsCQQDHwlX1yR9d2PlcQJbk
	kthmD+k2UC/UfmKcKKICXDbdmCQmxwmWQQrNBp6A/VID6vwgqg6Cr04LK7JoK93Q
	MumFAkBYl1VuRME3tRyPkniz+wEHLABbLwonwrVv/U2Bd2Pe7yUd/8/rxIqZD5AD
	f2VO2Lgvurpta7E80HzVK6IgxN+/AkEAj9fIvmxNQe7z4RJBleaIHTZn4MxtJL69
	k2VPBBQTDg54OdQpeyDq/ig+CvRfEqMrWvoZ6NEDide1aH3uA/YlYQJAA8YgwtOJ
	4hWEXiQNUZtWzFmgsXBip14+opLMOe8VUTOIv0M5QsVhrOSDVvDL6Wu0fCoyEYmr
	vBOJm1wjPtjrTw==
	-----END PRIVATE KEY-----`

	return this, err
}

func (this *TokenSupplierMock) GetToken(installation int) (string, error) {
	return this.tokenToReturn, nil
}
